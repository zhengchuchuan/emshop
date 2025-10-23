package consumer

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "emshop/internal/app/coupon/srv/data/v1/interfaces"
    redismgr "emshop/internal/app/coupon/srv/data/v1/redis"
    "emshop/internal/app/coupon/srv/domain/do"
    "emshop/pkg/log"

    rocketmq "github.com/apache/rocketmq-client-go/v2"
    "github.com/apache/rocketmq-client-go/v2/primitive"
    "github.com/apache/rocketmq-client-go/v2/producer"
    redisClient "github.com/go-redis/redis/v8"
)

// FlashSaleTxnProducer 事务消息生产者接口
type FlashSaleTxnProducer interface {
    SendFlashSaleSuccessTxn(ctx context.Context, event *FlashSaleSuccessEvent) error
    Shutdown() error
}

// flashSaleTxnProducer 实现
type flashSaleTxnProducer struct {
    tp           rocketmq.TransactionProducer
    topic        string
}

// txnListener 事务监听器：在生产者本地执行事务与回查
type txnListener struct {
    data         interfaces.DataFactory
    redis        *redisClient.Client
    stockManager *redismgr.StockManager
    topic        string
}

func NewFlashSaleTxnProducer(data interfaces.DataFactory, redis *redisClient.Client, nameServers []string, groupName, topic string) (FlashSaleTxnProducer, error) {
    if data == nil || redis == nil {
        return nil, fmt.Errorf("missing dependencies for txn producer")
    }
    listener := &txnListener{
        data:  data,
        redis: redis,
        stockManager: redismgr.NewStockManagerWithOptions(redis, true), // 抑制日志，降低生产者侧IO
        topic: topic,
    }
    // NewTransactionProducer(listener, options...)
    tp, err := rocketmq.NewTransactionProducer(
        listener,
        producer.WithNameServer(nameServers),
        producer.WithGroupName(groupName),
        producer.WithRetry(2),
    )
    if err != nil {
        return nil, fmt.Errorf("创建事务生产者失败: %w", err)
    }
    if err := tp.Start(); err != nil {
        return nil, fmt.Errorf("启动事务生产者失败: %w", err)
    }
    log.Infof("事务生产者启动成功, group=%s, topic=%s", groupName, topic)
    return &flashSaleTxnProducer{tp: tp, topic: topic}, nil
}

// SendFlashSaleSuccessTxn 发送事务消息（半消息），本地事务会在监听器中执行
func (p *flashSaleTxnProducer) SendFlashSaleSuccessTxn(ctx context.Context, event *FlashSaleSuccessEvent) error {
    if p == nil || p.tp == nil || event == nil {
        return fmt.Errorf("transaction producer not ready")
    }
    // 序列化事件
    body, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("序列化事务事件失败: %v", err)
    }
    msg := &primitive.Message{Topic: p.topic, Body: body}
    msg.WithTag("FLASH_SALE_TXN_COMMIT")
    msg.WithKeys([]string{fmt.Sprintf("user_%d_activity_%d", event.UserID, event.ActivityID)})
    // 统一事件类型为 flash_sale_success，交由消费者处理持久化扣减
    msg.WithProperty("event_type", "flash_sale_success")
    msg.WithProperty("user_id", fmt.Sprintf("%d", event.UserID))
    msg.WithProperty("activity_id", fmt.Sprintf("%d", event.ActivityID))
    msg.WithProperty("timestamp", fmt.Sprintf("%d", event.Timestamp))
    if event.RequestID != "" { msg.WithProperty("request_id", event.RequestID) }

    // 发送半消息并触发本地事务
    _, err = p.tp.SendMessageInTransaction(ctx, msg)
    if err != nil {
        log.Errorf("事务消息发送失败: %v", err)
        return err
    }
    // log.Infof("事务消息发送成功: msgID=%s queue=%d", res.MsgID, res.MessageQueue.QueueId)
    return nil
}

func (p *flashSaleTxnProducer) Shutdown() error {
    if p != nil && p.tp != nil {
        log.Info("正在关闭事务生产者...")
        if err := p.tp.Shutdown(); err != nil {
            log.Errorf("关闭事务生产者失败: %v", err)
            return err
        }
    }
    return nil
}

// ExecuteLocalTransaction 生产者本地事务：快速本地事务（仅插入轻量订单，不更新已售）
func (l *txnListener) ExecuteLocalTransaction(msg *primitive.Message) primitive.LocalTransactionState {
    defer func() { if r := recover(); r != nil { log.Errorf("执行本地事务panic: %v", r) } }()

    var evt FlashSaleSuccessEvent
    if err := json.Unmarshal(msg.Body, &evt); err != nil {
        log.Errorf("解析事务事件失败: %v", err)
        return primitive.RollbackMessageState
    }
    ctx := context.Background()
    // 本地事务：创建用户优惠券（幂等依赖DB唯一约束+request_id）
    tx := l.data.DB().Begin()
    if tx.Error != nil { log.Errorf("开启事务失败: %v", tx.Error); return primitive.RollbackMessageState }
    commit := func() primitive.LocalTransactionState {
        if err := tx.Commit().Error; err != nil { log.Errorf("提交事务失败: %v", err); return primitive.RollbackMessageState }
        return primitive.CommitMessageState
    }
    rollback := func() primitive.LocalTransactionState {
        _ = tx.Rollback().Error
        // 回滚Redis预留
        if err := l.stockManager.RollbackStock(ctx, evt.ActivityID, evt.UserID, evt.CouponID, 1); err != nil {
            log.Warnf("回滚库存失败: %v", err)
        }
        return primitive.RollbackMessageState
    }

    // 创建本地“秒杀订单”（轻量订单），不在此处更新 DB 库存，库存扣减交由消费者完成
    order := &do.FlashSaleOrderDO{
        OrderSn:     fmt.Sprintf("FSO-%d-%s", time.Now().Unix(), evt.RequestID),
        RequestID:   evt.RequestID,
        UserID:      evt.UserID,
        FlashSaleID: evt.ActivityID,
        CouponID:    evt.CouponID,
        Status:      "CREATED",
        Amount:      0,
    }
    if err := l.data.FlashSaleOrders().Create(ctx, tx, order); err != nil {
        // 幂等：若重复，视为成功
        if contains(err.Error(), "duplicate") || contains(err.Error(), "unique") {
            log.Warnf("秒杀订单已存在，幂等通过: user=%d req=%s", evt.UserID, evt.RequestID)
        } else {
            log.Errorf("创建秒杀订单失败: %v", err)
            return rollback()
        }
    }
    return commit()
}

// CheckLocalTransaction 事务回查：以 request_id 或 coupon_sn 判断是否已落库
func (l *txnListener) CheckLocalTransaction(msg *primitive.MessageExt) primitive.LocalTransactionState {
    defer func() { if r := recover(); r != nil { log.Errorf("回查panic: %v", r) } }()
    var evt FlashSaleSuccessEvent
    if err := json.Unmarshal(msg.Body, &evt); err != nil {
        log.Errorf("回查事件解析失败: %v", err)
        return primitive.RollbackMessageState
    }
    ctx := context.Background()
    var existed *do.FlashSaleOrderDO
    var err error
    if evt.RequestID != "" {
        existed, err = l.data.FlashSaleOrders().GetByRequestID(ctx, l.data.DB(), evt.RequestID)
    }
    if err != nil {
        // 在高并发/DB抖动情况下，避免长时间占用预留库存，直接回滚库存并让消息回滚
        if rerr := l.stockManager.RollbackStock(ctx, evt.ActivityID, evt.UserID, evt.CouponID, 1); rerr != nil {
            log.Warnf("回查时回滚库存失败: %v", rerr)
        }
        log.Warnf("回查查询失败，已回滚库存并回退事务: %v", err)
        return primitive.RollbackMessageState
    }
    if existed != nil {
        return primitive.CommitMessageState
    }
    // 未查到：回滚预留库存
    if rerr := l.stockManager.RollbackStock(ctx, evt.ActivityID, evt.UserID, evt.CouponID, 1); rerr != nil {
        log.Warnf("回查回滚库存失败: %v", rerr)
    }
    return primitive.RollbackMessageState
}
