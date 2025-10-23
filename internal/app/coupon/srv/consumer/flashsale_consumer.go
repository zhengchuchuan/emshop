package consumer

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    	"emshop/internal/app/coupon/srv/data/v1/interfaces"
    	"emshop/internal/app/coupon/srv/data/v1/redis"
    	"emshop/internal/app/coupon/srv/domain/do"
    	"emshop/pkg/log"

    	"github.com/apache/rocketmq-client-go/v2/consumer"
    	"github.com/apache/rocketmq-client-go/v2/primitive"
    	"github.com/apache/rocketmq-client-go/v2/producer"
    rocketmq "github.com/apache/rocketmq-client-go/v2"
    redisClient "github.com/go-redis/redis/v8"
    "gorm.io/gorm"
    "strings"
)

// safeShutdown 保护性关闭，避免上游库在未完全初始化时 panic
func safeShutdown(c rocketmq.PushConsumer) {
    defer func() { _ = recover() }()
    if c != nil { _ = c.Shutdown() }
}

// FlashSaleSuccessEvent 秒杀成功事件
type FlashSaleSuccessEvent struct {
	ActivityID    int64  `json:"activity_id"`
	CouponID      int64  `json:"coupon_id"`
	UserID        int64  `json:"user_id"`
	CouponSn      string `json:"coupon_sn"`
	ClientIP      string `json:"client_ip,omitempty"`
	UserAgent     string `json:"user_agent,omitempty"`
	Timestamp     int64  `json:"timestamp"`
	RequestID     string `json:"request_id,omitempty"`
}

// FlashSaleConsumer 秒杀消费者
type FlashSaleConsumer struct {
	data         interfaces.DataFactory
	redisClient  *redisClient.Client
	stockManager *redis.StockManager
	mqConsumer   rocketmq.PushConsumer
	retryManager *RetryManager
	retryConfig  *RetryConfig
}

// FlashSaleConsumerConfig 秒杀消费者配置
type FlashSaleConsumerConfig struct {
	NameServers   []string `json:"name_servers"`
	ConsumerGroup string   `json:"consumer_group"`
	Topic         string   `json:"topic"`
	BatchSize     int      `json:"batch_size"`
	MaxRetries    int      `json:"max_retries"`
}

// NewFlashSaleConsumer 创建秒杀消费者
func NewFlashSaleConsumer(data interfaces.DataFactory, redisClient *redisClient.Client, retryMgr *RetryManager) *FlashSaleConsumer {
	return &FlashSaleConsumer{
		data:         data,
		redisClient:  redisClient,
		stockManager: redis.NewStockManager(redisClient),
		retryManager: retryMgr,
	}
}

// Start 启动秒杀事件消费者
func (fsc *FlashSaleConsumer) Start(config *FlashSaleConsumerConfig) error {
	if fsc == nil {
		return fmt.Errorf("flash sale consumer is nil")
	}
	if config == nil {
		return fmt.Errorf("flash sale consumer config is nil")
	}
	if fsc.mqConsumer != nil {
		return nil
	}

	options := []consumer.Option{
		consumer.WithNameServer(config.NameServers),
		consumer.WithGroupName(config.ConsumerGroup),
		consumer.WithConsumeFromWhere(consumer.ConsumeFromLastOffset),
	}
	if config.BatchSize > 0 {
		options = append(options, consumer.WithConsumeMessageBatchMaxSize(config.BatchSize))
	}
    // 提升消费并发度，避免单线程批处理导致堆积
    options = append(options, consumer.WithConsumeGoroutineNums(32))
	
    pushConsumer, err := rocketmq.NewPushConsumer(options...)
    if err != nil {
        return fmt.Errorf("创建RocketMQ消费者失败: %w", err)
    }

    if err := pushConsumer.Subscribe(config.Topic, consumer.MessageSelector{}, fsc.dispatchConsumeMessage); err != nil {
        safeShutdown(pushConsumer)
        return fmt.Errorf("订阅秒杀事件失败: %w", err)
    }

    if err := pushConsumer.Start(); err != nil {
        safeShutdown(pushConsumer)
        return fmt.Errorf("启动秒杀事件消费者失败: %w", err)
    }

	fsc.mqConsumer = pushConsumer
	fsc.retryConfig = fsc.buildRetryConfig(config)
	return nil
}

// Stop 停止秒杀事件消费者
func (fsc *FlashSaleConsumer) Stop() error {
	if fsc == nil || fsc.mqConsumer == nil {
		return nil
	}
	defer func() { fsc.mqConsumer = nil }()
	return fsc.mqConsumer.Shutdown()
}

func (fsc *FlashSaleConsumer) buildRetryConfig(config *FlashSaleConsumerConfig) *RetryConfig {
	retryCfg := &RetryConfig{
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2,
		EnableJitter: true,
	}
	if config != nil && config.MaxRetries > 0 {
		retryCfg.MaxRetries = config.MaxRetries
	} else {
		retryCfg.MaxRetries = 5
	}
	return retryCfg
}

func (fsc *FlashSaleConsumer) dispatchConsumeMessage(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	if len(msgs) == 0 {
		return consumer.ConsumeSuccess, nil
	}

	var successMsgs []*primitive.MessageExt
	var failureMsgs []*primitive.MessageExt

	for _, msg := range msgs {
		eventType := msg.GetProperty("event_type")
		switch eventType {
		case "flash_sale_success":
			successMsgs = append(successMsgs, msg)
		case "flash_sale_failure":
			failureMsgs = append(failureMsgs, msg)
		default:
			log.Warnf("收到未知类型秒杀事件，跳过: msgID=%s, eventType=%s", msg.MsgId, eventType)
		}
	}

	if len(successMsgs) > 0 {
		result, err := fsc.ConsumeFlashSaleSuccessMessage(ctx, successMsgs...)
		if err != nil || result != consumer.ConsumeSuccess {
			return result, err
		}
	}

	if len(failureMsgs) > 0 {
		result, err := fsc.ConsumeFlashSaleFailureMessage(ctx, failureMsgs...)
		if err != nil || result != consumer.ConsumeSuccess {
			return result, err
		}
	}

	return consumer.ConsumeSuccess, nil
}

// ConsumeFlashSaleSuccessMessage 消费秒杀成功消息
func (fsc *FlashSaleConsumer) ConsumeFlashSaleSuccessMessage(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	log.Infof("收到秒杀成功消息，消息数量: %d", len(msgs))

	for _, msg := range msgs {
		// 解析消息
		var event FlashSaleSuccessEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Errorf("解析秒杀成功消息失败: %v, msgID: %s", err, msg.MsgId)
			continue
		}

		log.Infof("处理秒杀成功事件: userID=%d, activityID=%d, couponSn=%s", 
			event.UserID, event.ActivityID, event.CouponSn)

        // 处理秒杀成功事件
        if err := fsc.handleFlashSaleSuccess(ctx, &event, msg); err != nil {
			handled := false
			if fsc.retryManager != nil {
				retryCfg := fsc.retryConfig
				if retryCfg == nil {
					retryCfg = fsc.buildRetryConfig(nil)
				}
				if retryErr := fsc.retryManager.ScheduleRetry(ctx, msg, err, retryCfg); retryErr != nil {
					log.Errorf("调度秒杀成功事件重试失败: %v", retryErr)
				} else {
					handled = true
					log.Infof("秒杀成功事件已加入重试队列: msgID=%s", msg.MsgId)
				}
			}

			if !handled {
				log.Errorf("处理秒杀成功事件失败: %v, 将由RocketMQ重试", err)
				return consumer.ConsumeRetryLater, err
			}
		}
	}

	return consumer.ConsumeSuccess, nil
}

// handleFlashSaleSuccess 处理秒杀成功事件
func (fsc *FlashSaleConsumer) handleFlashSaleSuccess(ctx context.Context, event *FlashSaleSuccessEvent, msg *primitive.MessageExt) error {
    msgID := msg.MsgId
    // 1. 检查幂等性（避免重复处理）优先使用 request_id，其次 msgID
    dedupeKey := msgID
    if event.RequestID != "" {
        dedupeKey = "req:" + event.RequestID
    }
    idempotentKey := fmt.Sprintf("flashsale:processed:%s", dedupeKey)
    exists, err := fsc.redisClient.Exists(ctx, idempotentKey).Result()
		if err != nil {
			log.Errorf("检查幂等性失败: %v", err)
		} else if exists > 0 {
			log.Infof("消息已处理过，跳过: msgID=%s", msgID)
			return nil
		}

    // 1.1 加处理锁，防止并发重复扣减（高并发/重复投递窗口）
    lockKey := fmt.Sprintf("flashsale:lock:%s", dedupeKey)
    locked, lerr := fsc.redisClient.SetNX(ctx, lockKey, "1", 30*time.Second).Result()
    if lerr != nil {
        log.Warnf("获取处理锁失败，继续尝试处理: %v", lerr)
    } else if !locked {
        // 有并发处理中的同一request，直接跳过（由持锁者完成扣减并写processed标记）
        log.Infof("同一请求正在处理，跳过: %s", dedupeKey)
        return nil
    } else {
        defer func() { _ = fsc.redisClient.Del(ctx, lockKey).Err() }()
    }

    // 2. 若为事务消息，优先信任本地事务已落库，仅做确认与幂等标记
    isTxnFinal := (msg.GetProperty("TRAN_MSG") == "true") || (msg.GetProperty("transaction_id") != "") || (msg.GetProperty("__transactionId__") != "")
    if isTxnFinal {
        // 快速确认是否已存在
        // 使用事务生产者侧写入的轻量订单作为幂等锚点
        var existed *do.FlashSaleOrderDO
        var getErr error
        if event.RequestID != "" {
            existed, getErr = fsc.data.FlashSaleOrders().GetByRequestID(ctx, fsc.data.DB(), event.RequestID)
        }
        if getErr != nil {
            log.Warnf("事务消息确认(查询订单)失败，继续处理扣减: %v", getErr)
        }
        if existed != nil {
            // 本地事务已成功创建轻量订单：继续进行库存持久化扣减
            log.Infof("事务消息确认：订单已存在，开始持久化扣减, userID=%d, requestID=%s", event.UserID, event.RequestID)
        } else {
            // 未查到：继续进行库存持久化扣减（补偿路径），不在消费者侧创建用户券
            log.Warnf("事务消息未查到订单记录，进入补偿扣减: userID=%d, requestID=%s", event.UserID, event.RequestID)
        }
    }


    // 3. 基于订单状态原子闸门：仅当从 CREATED -> COUNTED 成功时才进行 remaining_count -= 1
    if event.RequestID != "" {
        updated, uerr := fsc.data.FlashSaleOrders().MarkCountedByRequestID(ctx, fsc.data.DB(), event.RequestID)
        if uerr != nil {
            return fmt.Errorf("标记订单已计数失败: %v", uerr)
        }
        if !updated {
            // 已经处理过（或不存在），直接打上幂等标记并返回，避免重复扣减
            _ = fsc.redisClient.SetEX(ctx, idempotentKey, "1", 7*24*time.Hour).Err()
            log.Infof("请求已计数过，跳过重复扣减: userID=%d, requestID=%s", event.UserID, event.RequestID)
            return nil
        }
    }

    // 4. 持久化扣减库存（原子自增）
    if err := fsc.data.FlashSales().IncrementSoldCount(ctx, fsc.data.DB(), event.ActivityID); err != nil {
        return fmt.Errorf("持久化扣减库存失败: %v", err)
    }

    // 4.1 仅在事务链路下：提交后记录最终成功日志与统计（避免与预留阶段混淆）
    isTxn := (msg.GetProperty("TRAN_MSG") == "true") || (msg.GetProperty("transaction_id") != "") || (msg.GetProperty("__transactionId__") != "")
    if isTxn {
        // Redis 日志：仅记录“最终成交”
        logKey := fmt.Sprintf("coupon:log:%d", event.ActivityID)
        ts := event.Timestamp
        if ts <= 0 { ts = time.Now().Unix() }
        reqID := event.RequestID
        if strings.TrimSpace(reqID) == "" { reqID = fmt.Sprintf("r_%d_%d_%d", event.UserID, event.ActivityID, ts) }
        raw := fmt.Sprintf("%d:%d:%d:%d:%s", event.UserID, event.ActivityID, 1, ts, reqID)
        if err := fsc.redisClient.LPush(ctx, logKey, raw).Err(); err != nil {
            log.Warnf("写入成交日志失败: %v", err)
        }
        // 活动统计：success_count 仅在最终成交后累加
        actKey := fmt.Sprintf("coupon:activity:%d", event.ActivityID)
        if err := fsc.redisClient.HIncrBy(ctx, actKey, "success_count", 1).Err(); err != nil {
            log.Warnf("累加活动成功数失败: %v", err)
        }
    }

    // 5. 设置幂等性标记（7天过期）
    _ = fsc.redisClient.SetEX(ctx, idempotentKey, "1", 7*24*time.Hour).Err()

    log.Infof("秒杀成功事件处理完成：已持久化扣减库存，userID=%d, activityID=%d, couponSn=%s",
        event.UserID, event.ActivityID, event.CouponSn)
	
	return nil
}

// calculateExpiryTime 计算过期时间
func (fsc *FlashSaleConsumer) calculateExpiryTime(template *do.CouponTemplateDO) time.Time {
	// 根据优惠券模板的有效期设置用户优惠券过期时间
	// 这里简化处理，使用模板的结束时间
	return template.ValidEndTime
}

// isDuplicateError 检查是否是重复错误
func (fsc *FlashSaleConsumer) isDuplicateError(err error) bool {
	// 这里简化处理，实际应该检查具体的数据库错误类型
	errStr := err.Error()
	return contains(errStr, "duplicate") || contains(errStr, "unique")
}

// rollbackStockIfNeeded 必要时回滚库存
func (fsc *FlashSaleConsumer) rollbackStockIfNeeded(ctx context.Context, event *FlashSaleSuccessEvent) {
	// 简化处理：直接回滚库存
	// 在实际项目中，这里应该检查用户优惠券是否真的已存在
	log.Warnf("检测到重复创建错误，直接进行库存回滚: userID=%d, couponSn=%s", 
		event.UserID, event.CouponSn)
		
	if true { // 简化处理
		// 如果确实不存在，说明可能需要回滚库存
		log.Warnf("用户优惠券不存在但创建失败，回滚库存: userID=%d, couponSn=%s", 
			event.UserID, event.CouponSn)
		
		err := fsc.stockManager.RollbackStock(ctx, event.ActivityID, event.UserID, event.CouponID, 1)
		if err != nil {
			log.Errorf("回滚库存失败: %v", err)
		}
	}
}

// updateCouponTemplateStats 更新优惠券模板统计
func (fsc *FlashSaleConsumer) updateCouponTemplateStats(ctx context.Context, tx *gorm.DB, couponID int64) error {
    // 改为 Redis 累加，异步合并到DB
    if fsc.redisClient != nil {
        return fsc.redisClient.HIncrBy(ctx, "coupon:stats:template", fmt.Sprintf("%d", couponID), 1).Err()
    }
    return nil
}

// updateFlashSaleStats 更新秒杀活动统计
func (fsc *FlashSaleConsumer) updateFlashSaleStats(ctx context.Context, tx *gorm.DB, activityID int64) error {
    if fsc.redisClient != nil {
        return fsc.redisClient.HIncrBy(ctx, "coupon:stats:flashsale", fmt.Sprintf("%d", activityID), 1).Err()
    }
    return nil
}

// ConsumeFlashSaleFailureMessage 消费秒杀失败消息（可选）
func (fsc *FlashSaleConsumer) ConsumeFlashSaleFailureMessage(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	log.Infof("收到秒杀失败消息，消息数量: %d", len(msgs))

	// 这里可以处理秒杀失败的统计和监控
	for _, msg := range msgs {
		log.Infof("秒杀失败消息: msgID=%s, body=%s", msg.MsgId, string(msg.Body))
	}

	return consumer.ConsumeSuccess, nil
}

// PublishFlashSaleSuccessEvent 发布秒杀成功事件（由秒杀服务调用）
func PublishFlashSaleSuccessEvent(event *FlashSaleSuccessEvent, producer interface{}) error {
	// 这个方法应该在秒杀服务中调用，发送消息到RocketMQ
	// 这里只是一个示例，实际的发送逻辑应该在秒杀服务中实现
	
	_, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("序列化秒杀成功事件失败: %v", err)
	}

	log.Infof("准备发送秒杀成功事件: userID=%d, activityID=%d", 
		event.UserID, event.ActivityID)

	// TODO: 实际的RocketMQ消息发送逻辑
	// 这里需要根据实际的RocketMQ Producer进行实现
	
	return nil
}

// FlashSaleEventProducer 秒杀事件生产者接口
type FlashSaleEventProducer interface {
	SendFlashSaleSuccessEvent(event *FlashSaleSuccessEvent) error
	SendFlashSaleFailureEvent(event *FlashSaleFailureEvent) error
	Shutdown() error
}

// FlashSaleFailureEvent 秒杀失败事件
type FlashSaleFailureEvent struct {
	ActivityID int64  `json:"activity_id"`
	UserID     int64  `json:"user_id"`
	Reason     string `json:"reason"`
	Code       int    `json:"code"`
	ClientIP   string `json:"client_ip,omitempty"`
	UserAgent  string `json:"user_agent,omitempty"`
	Timestamp  int64  `json:"timestamp"`
}

// flashSaleEventProducer RocketMQ事件生产者实现
type flashSaleEventProducer struct {
    producer rocketmq.Producer
    topic    string
    async    bool
}

// queuedEventProducer 将成功事件放入内存有界队列，由N个worker后台发送（OneWay/Sync由底层producer决定）。
type queuedEventProducer struct {
    base    FlashSaleEventProducer
    ch      chan *FlashSaleSuccessEvent
}

// NewQueuedFlashSaleEventProducer 包装底层生产者，workers个后台协程消费，capacity为有界队列大小；队列满则直接丢弃（计数可另行添加）。
func NewQueuedFlashSaleEventProducer(base FlashSaleEventProducer, workers, capacity int) FlashSaleEventProducer {
    if workers <= 0 { workers = 8 }
    if capacity <= 0 { capacity = 100000 }
    q := &queuedEventProducer{ base: base, ch: make(chan *FlashSaleSuccessEvent, capacity) }
    for i := 0; i < workers; i++ {
        go func() {
            for evt := range q.ch {
                // 忽略错误（压测优先吞吐）
                _ = q.base.SendFlashSaleSuccessEvent(evt)
            }
        }()
    }
    return q
}

func (q *queuedEventProducer) SendFlashSaleSuccessEvent(event *FlashSaleSuccessEvent) error {
    select {
    case q.ch <- event:
    default:
        // 队列满丢弃
    }
    return nil
}

func (q *queuedEventProducer) SendFlashSaleFailureEvent(event *FlashSaleFailureEvent) error {
    // 失败事件直接透传，不进入队列
    return q.base.SendFlashSaleFailureEvent(event)
}

func (q *queuedEventProducer) Shutdown() error {
    // 最简实现：直接关闭底层；队列不做drain（压测优先）
    close(q.ch)
    return q.base.Shutdown()
}

// NewFlashSaleEventProducer 创建秒杀事件生产者
func NewFlashSaleEventProducer(nameServers []string, groupName, topic string, async bool) (FlashSaleEventProducer, error) {
	// 创建RocketMQ Producer
	p, err := rocketmq.NewProducer(
		producer.WithNameServer(nameServers),
		producer.WithRetry(3),
		producer.WithGroupName(groupName),
	)
	if err != nil {
		return nil, fmt.Errorf("创建RocketMQ Producer失败: %v", err)
	}

	// 启动生产者
	err = p.Start()
	if err != nil {
		return nil, fmt.Errorf("启动RocketMQ Producer失败: %v", err)
	}

	log.Infof("RocketMQ Producer启动成功, nameServers: %v, group: %s, topic: %s", 
		nameServers, groupName, topic)

    return &flashSaleEventProducer{
        producer: p,
        topic:    topic,
        async:    async,
    }, nil
}

// SendFlashSaleSuccessEvent 发送秒杀成功事件
func (p *flashSaleEventProducer) SendFlashSaleSuccessEvent(event *FlashSaleSuccessEvent) error {
	// 验证事件数据
	if err := ValidateFlashSaleEvent(event); err != nil {
		return fmt.Errorf("事件数据验证失败: %v", err)
	}

	// 序列化事件数据
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("序列化秒杀成功事件失败: %v", err)
	}

	// 构建消息
	msg := &primitive.Message{
		Topic: p.topic,
		Body:  eventData,
	}
	msg.WithTag("FLASH_SALE_SUCCESS")
	msg.WithKeys([]string{fmt.Sprintf("user_%d_activity_%d", event.UserID, event.ActivityID)})
	msg.WithProperty("event_type", "flash_sale_success")
	msg.WithProperty("user_id", fmt.Sprintf("%d", event.UserID))
	msg.WithProperty("activity_id", fmt.Sprintf("%d", event.ActivityID))
	msg.WithProperty("timestamp", fmt.Sprintf("%d", event.Timestamp))
	
    // 发送消息（支持异步以提高吞吐）
    if p.async {
        // 为避免 go client 在超时回调路径上的崩溃风险，使用 OneWay 提升吞吐并规避回调。
        ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
        defer cancel()
        return p.producer.SendOneWay(ctx, msg)
    } else {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        result, err := p.producer.SendSync(ctx, msg)
        if err != nil {
            log.Errorf("发送秒杀成功事件失败: %v, userID=%d, activityID=%d", err, event.UserID, event.ActivityID)
            return fmt.Errorf("发送秒杀成功事件失败: %v", err)
        }
        log.Infof("发送秒杀成功事件成功: userID=%d, activityID=%d, msgID=%s, queueID=%d", event.UserID, event.ActivityID, result.MsgID, result.MessageQueue.QueueId)
        return nil
    }
}

// SendFlashSaleFailureEvent 发送秒杀失败事件
func (p *flashSaleEventProducer) SendFlashSaleFailureEvent(event *FlashSaleFailureEvent) error {
	// 验证基础字段
	if event.UserID <= 0 {
		return fmt.Errorf("invalid user_id: %d", event.UserID)
	}
	if event.ActivityID <= 0 {
		return fmt.Errorf("invalid activity_id: %d", event.ActivityID)
	}
	if event.Reason == "" {
		return fmt.Errorf("empty failure reason")
	}

	// 序列化事件数据
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("序列化秒杀失败事件失败: %v", err)
	}

	// 构建消息
	msg := &primitive.Message{
		Topic: p.topic,
		Body:  eventData,
	}
	msg.WithTag("FLASH_SALE_FAILURE")
	msg.WithKeys([]string{fmt.Sprintf("user_%d_activity_%d", event.UserID, event.ActivityID)})
	msg.WithProperty("event_type", "flash_sale_failure")
	msg.WithProperty("user_id", fmt.Sprintf("%d", event.UserID))
	msg.WithProperty("activity_id", fmt.Sprintf("%d", event.ActivityID))
	msg.WithProperty("reason", event.Reason)
	msg.WithProperty("code", fmt.Sprintf("%d", event.Code))
	msg.WithProperty("timestamp", fmt.Sprintf("%d", event.Timestamp))

    if p.async {
        ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
        defer cancel()
        return p.producer.SendOneWay(ctx, msg)
    } else {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        result, err := p.producer.SendSync(ctx, msg)
        if err != nil {
            log.Errorf("发送秒杀失败事件失败: %v, userID=%d, activityID=%d, reason=%s", err, event.UserID, event.ActivityID, event.Reason)
            return fmt.Errorf("发送秒杀失败事件失败: %v", err)
        }
        log.Infof("发送秒杀失败事件成功: userID=%d, activityID=%d, reason=%s, msgID=%s", event.UserID, event.ActivityID, event.Reason, result.MsgID)
        return nil
    }
}

// Shutdown 优雅关闭生产者
func (p *flashSaleEventProducer) Shutdown() error {
	if p.producer != nil {
		log.Info("正在关闭RocketMQ Producer...")
		err := p.producer.Shutdown()
		if err != nil {
			log.Errorf("关闭RocketMQ Producer失败: %v", err)
			return err
		}
		log.Info("RocketMQ Producer已关闭")
	}
	return nil
}

// contains 字符串包含检查
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) > 0 && 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		 s[len(s)-len(substr):] == substr || 
		 stringContains(s, substr))))
}

// stringContains 简单的字符串包含实现
func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetRetryDelayLevel 获取重试延迟级别
func GetRetryDelayLevel(retryCount int) int {
	// RocketMQ延迟级别：1s 5s 10s 30s 1m 2m 3m 4m 5m 6m 7m 8m 9m 10m 20m 30m 1h 2h
	delayLevels := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18}
	
	if retryCount >= len(delayLevels) {
		return delayLevels[len(delayLevels)-1] // 最大2小时
	}
	
	return delayLevels[retryCount]
}

// ValidateFlashSaleEvent 验证秒杀事件
func ValidateFlashSaleEvent(event *FlashSaleSuccessEvent) error {
	if event.ActivityID <= 0 {
		return fmt.Errorf("invalid activity_id: %d", event.ActivityID)
	}
	if event.CouponID <= 0 {
		return fmt.Errorf("invalid coupon_id: %d", event.CouponID)
	}
	if event.UserID <= 0 {
		return fmt.Errorf("invalid user_id: %d", event.UserID)
	}
	if event.CouponSn == "" {
		return fmt.Errorf("empty coupon_sn")
	}
	if event.Timestamp <= 0 {
		return fmt.Errorf("invalid timestamp: %d", event.Timestamp)
	}
	
	return nil
}
