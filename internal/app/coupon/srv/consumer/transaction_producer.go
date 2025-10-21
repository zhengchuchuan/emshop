package consumer

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

	"emshop/internal/app/coupon/srv/data/v1/interfaces"
	"emshop/pkg/log"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	rocketmq "github.com/apache/rocketmq-client-go/v2"
	redisClient "github.com/go-redis/redis/v8"
)

// TransactionProducer 事务消息生产者
type TransactionProducer struct {
	producer    rocketmq.TransactionProducer
	topic       string
	data        interfaces.DataFactory
	redisClient *redisClient.Client
}

// TransactionConfig 事务配置
type TransactionConfig struct {
	NameServers []string `json:"name_servers"`
	GroupName   string   `json:"group_name"`
	Topic       string   `json:"topic"`
}

// TransactionContext 事务上下文
type TransactionContext struct {
    ActivityID int64  `json:"activity_id"`
    UserID     int64  `json:"user_id"`
    CouponID   int64  `json:"coupon_id"`
    CouponSn   string `json:"coupon_sn"`
    RequestID  string `json:"request_id,omitempty"`
    Action     string `json:"action"` // "create_user_coupon", "update_stats", etc.
}

// NewTransactionProducer 创建事务消息生产者
func NewTransactionProducer(config *TransactionConfig, data interfaces.DataFactory, redisClient *redisClient.Client) (*TransactionProducer, error) {
	tp := &TransactionProducer{
		topic:       config.Topic,
		data:        data,
		redisClient: redisClient,
	}

	// 创建事务生产者
	p, err := rocketmq.NewTransactionProducer(
		tp,
		producer.WithNameServer(config.NameServers),
		producer.WithRetry(3),
		producer.WithGroupName(config.GroupName),
	)
	if err != nil {
		return nil, fmt.Errorf("创建事务生产者失败: %v", err)
	}

	tp.producer = p

	// 启动生产者
	if err := p.Start(); err != nil {
		return nil, fmt.Errorf("启动事务生产者失败: %v", err)
	}

	log.Infof("事务消息生产者启动成功, nameServers: %v, group: %s, topic: %s",
		config.NameServers, config.GroupName, config.Topic)

	return tp, nil
}

// SendTransactionMessage 发送事务消息
func (tp *TransactionProducer) SendTransactionMessage(ctx context.Context, event *FlashSaleSuccessEvent) error {
    // 构建事务上下文
    txnContext := &TransactionContext{
        ActivityID: event.ActivityID,
        UserID:     event.UserID,
        CouponID:   event.CouponID,
        CouponSn:   event.CouponSn,
        RequestID:  event.RequestID,
        Action:     "create_user_coupon",
    }

	// 序列化事务上下文
	contextData, err := json.Marshal(txnContext)
	if err != nil {
		return fmt.Errorf("序列化事务上下文失败: %v", err)
	}

	// 序列化事件数据
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("序列化事件数据失败: %v", err)
	}

	// 构建事务消息
	msg := &primitive.Message{
		Topic: tp.topic,
		Body:  eventData,
	}
	msg.WithTag("FLASH_SALE_SUCCESS_TXN")
	msg.WithKeys([]string{fmt.Sprintf("txn_user_%d_activity_%d", event.UserID, event.ActivityID)})
	msg.WithProperty("transaction_id", fmt.Sprintf("txn_%d_%d_%d", event.ActivityID, event.UserID, event.Timestamp))
	msg.WithProperty("context", string(contextData))
    msg.WithProperty("event_type", "flash_sale_success")
    msg.WithProperty("user_id", fmt.Sprintf("%d", event.UserID))
    msg.WithProperty("activity_id", fmt.Sprintf("%d", event.ActivityID))
    if event.RequestID != "" {
        msg.WithProperty("request_id", event.RequestID)
    }

	// 发送事务消息
	result, err := tp.producer.SendMessageInTransaction(ctx, msg)
	if err != nil {
		log.Errorf("发送事务消息失败: %v, userID=%d, activityID=%d", 
			err, event.UserID, event.ActivityID)
		return fmt.Errorf("发送事务消息失败: %v", err)
	}

	log.Infof("发送事务消息成功: userID=%d, activityID=%d, msgID=%s, txnID=%s",
		event.UserID, event.ActivityID, result.MsgID, result.TransactionID)

	return nil
}

// ExecuteLocalTransaction 执行本地事务
func (tp *TransactionProducer) ExecuteLocalTransaction(msg *primitive.Message) primitive.LocalTransactionState {
	// 注意: primitive.Message 没有GetMsgId()方法，使用topic和timestamp作为标识
	msgKey := fmt.Sprintf("%s_%d", msg.Topic, time.Now().UnixNano())
	log.Infof("开始执行本地事务: msgKey=%s", msgKey)

	// 获取事务上下文
	contextStr := msg.GetProperty("context")
	if contextStr == "" {
		log.Errorf("事务上下文为空: msgKey=%s", msgKey)
		return primitive.RollbackMessageState
	}

	var txnContext TransactionContext
	if err := json.Unmarshal([]byte(contextStr), &txnContext); err != nil {
		log.Errorf("解析事务上下文失败: %v, msgKey=%s", err, msgKey)
		return primitive.RollbackMessageState
	}

    // 创建事务ID用于幂等检查
    txnID := msg.GetProperty("transaction_id")
    if txnID == "" {
        txnID = fmt.Sprintf("txn_%d_%d_%d", txnContext.ActivityID, txnContext.UserID, time.Now().Unix())
    }
    // 仅设置幂等标记并提交（扣减在入口已完成，落库由消费者完成）
    ctx := context.Background()
    idempotentKey := fmt.Sprintf("txn:executed:%s", txnID)
    if err := tp.redisClient.SetEX(ctx, idempotentKey, "committed", time.Hour).Err(); err != nil {
        log.Errorf("设置事务幂等标记失败: %v, txnID=%s", err, txnID)
        return primitive.UnknowState
    }
    return primitive.CommitMessageState
}

// CheckLocalTransaction 检查本地事务状态（事务回查）
func (tp *TransactionProducer) CheckLocalTransaction(msg *primitive.MessageExt) primitive.LocalTransactionState {
	log.Infof("开始事务回查: msgID=%s", msg.MsgId)

	// 获取事务ID
	txnID := msg.GetProperty("transaction_id")
	if txnID == "" {
		log.Errorf("事务ID为空: msgID=%s", msg.MsgId)
		return primitive.RollbackMessageState
	}

	// 检查事务状态
	idempotentKey := fmt.Sprintf("txn:executed:%s", txnID)
	ctx := context.Background()
	result := tp.redisClient.Get(ctx, idempotentKey).Val()

	switch result {
	case "committed":
		log.Infof("事务回查结果: 已提交, txnID=%s", txnID)
		return primitive.CommitMessageState
	case "rollback":
		log.Infof("事务回查结果: 已回滚, txnID=%s", txnID)
		return primitive.RollbackMessageState
	default:
		// 事务状态未知，可能正在执行或已超时
		log.Warnf("事务状态未知: txnID=%s, 返回回滚状态", txnID)
		return primitive.RollbackMessageState
	}
}

// updateCouponTemplateStats 更新优惠券模板统计
func (tp *TransactionProducer) updateCouponTemplateStats(ctx context.Context, tx interface{}, couponID int64) error {
    // 改为 Redis 累加，异步任务批量合并到DB
    if tp.redisClient != nil {
        return tp.redisClient.HIncrBy(ctx, "coupon:stats:template", fmt.Sprintf("%d", couponID), 1).Err()
    }
    return nil
}

// updateFlashSaleStats 更新秒杀活动统计
func (tp *TransactionProducer) updateFlashSaleStats(ctx context.Context, tx interface{}, activityID int64) error {
    if tp.redisClient != nil {
        return tp.redisClient.HIncrBy(ctx, "coupon:stats:flashsale", fmt.Sprintf("%d", activityID), 1).Err()
    }
    return nil
}

// Shutdown 关闭事务生产者
func (tp *TransactionProducer) Shutdown() error {
	if tp.producer != nil {
		log.Info("正在关闭事务消息生产者...")
		err := tp.producer.Shutdown()
		if err != nil {
			log.Errorf("关闭事务生产者失败: %v", err)
			return err
		}
		log.Info("事务消息生产者已关闭")
	}
	return nil
}

// TransactionFlashSaleEventProducer 事务消息版本的秒杀事件生产者
type TransactionFlashSaleEventProducer struct {
	txnProducer     *TransactionProducer
	fallbackProducer FlashSaleEventProducer // 备用的普通生产者
}

// NewTransactionFlashSaleEventProducer 创建事务版本的秒杀事件生产者
func NewTransactionFlashSaleEventProducer(config *TransactionConfig, data interfaces.DataFactory, redisClient *redisClient.Client, fallback FlashSaleEventProducer) (*TransactionFlashSaleEventProducer, error) {
	txnProducer, err := NewTransactionProducer(config, data, redisClient)
	if err != nil {
		return nil, fmt.Errorf("创建事务生产者失败: %v", err)
	}

	return &TransactionFlashSaleEventProducer{
		txnProducer:      txnProducer,
		fallbackProducer: fallback,
	}, nil
}

// SendFlashSaleSuccessEvent 发送秒杀成功事件（事务消息）
func (tfp *TransactionFlashSaleEventProducer) SendFlashSaleSuccessEvent(event *FlashSaleSuccessEvent) error {
	ctx := context.Background()

	// 优先使用事务消息
	err := tfp.txnProducer.SendTransactionMessage(ctx, event)
	if err != nil {
		log.Errorf("事务消息发送失败，使用备用生产者: %v", err)
		// 如果事务消息失败，使用备用的普通消息
		if tfp.fallbackProducer != nil {
			return tfp.fallbackProducer.SendFlashSaleSuccessEvent(event)
		}
		return err
	}

	return nil
}

// SendFlashSaleFailureEvent 发送秒杀失败事件
func (tfp *TransactionFlashSaleEventProducer) SendFlashSaleFailureEvent(event *FlashSaleFailureEvent) error {
	// 失败事件不需要事务保证，直接使用备用生产者
	if tfp.fallbackProducer != nil {
		return tfp.fallbackProducer.SendFlashSaleFailureEvent(event)
	}
	return fmt.Errorf("no fallback producer available for failure events")
}

// Shutdown 关闭生产者
func (tfp *TransactionFlashSaleEventProducer) Shutdown() error {
	var errs []error

	if tfp.txnProducer != nil {
		if err := tfp.txnProducer.Shutdown(); err != nil {
			errs = append(errs, err)
		}
	}

	if tfp.fallbackProducer != nil {
		if err := tfp.fallbackProducer.Shutdown(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("关闭生产者时发生错误: %v", errs)
	}

	return nil
}
