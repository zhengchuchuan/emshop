package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"emshop/internal/app/coupon/srv/data/v1/interfaces"
	"emshop/internal/app/coupon/srv/domain/do"
	"emshop/pkg/log"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	rocketmq "github.com/apache/rocketmq-client-go/v2"
	redisClient "github.com/go-redis/redis/v8"
	"gorm.io/gorm"
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

	// 执行本地事务
	ctx := context.Background()
	switch txnContext.Action {
	case "create_user_coupon":
		return tp.executeCreateUserCouponTransaction(ctx, &txnContext, txnID)
	default:
		log.Errorf("未知的事务操作: %s, msgKey=%s", txnContext.Action, msgKey)
		return primitive.RollbackMessageState
	}
}

// executeCreateUserCouponTransaction 执行创建用户优惠券事务
func (tp *TransactionProducer) executeCreateUserCouponTransaction(ctx context.Context, txnContext *TransactionContext, txnID string) primitive.LocalTransactionState {
	// 1. 检查事务幂等性
	idempotentKey := fmt.Sprintf("txn:executed:%s", txnID)
	exists, err := tp.redisClient.Exists(ctx, idempotentKey).Result()
	if err != nil {
		log.Errorf("检查事务幂等性失败: %v, txnID=%s", err, txnID)
		return primitive.UnknowState
	}
	if exists > 0 {
		// 事务已执行过，检查结果
		result := tp.redisClient.Get(ctx, idempotentKey).Val()
		if result == "committed" {
			log.Infof("事务已提交: txnID=%s", txnID)
			return primitive.CommitMessageState
		} else {
			log.Infof("事务已回滚: txnID=%s", txnID)
			return primitive.RollbackMessageState
		}
	}

	// 2. 获取活动和优惠券模板信息
	activityDO, err := tp.data.FlashSales().Get(ctx, tp.data.DB(), txnContext.ActivityID)
	if err != nil {
		log.Errorf("获取活动信息失败: %v, txnID=%s", err, txnID)
		tp.redisClient.SetEX(ctx, idempotentKey, "rollback", time.Hour)
		return primitive.RollbackMessageState
	}
	if activityDO == nil {
		log.Errorf("活动不存在: activityID=%d, txnID=%s", txnContext.ActivityID, txnID)
		tp.redisClient.SetEX(ctx, idempotentKey, "rollback", time.Hour)
		return primitive.RollbackMessageState
	}

	templateDO, err := tp.data.CouponTemplates().Get(ctx, tp.data.DB(), txnContext.CouponID)
	if err != nil {
		log.Errorf("获取优惠券模板失败: %v, txnID=%s", err, txnID)
		tp.redisClient.SetEX(ctx, idempotentKey, "rollback", time.Hour)
		return primitive.RollbackMessageState
	}
	if templateDO == nil {
		log.Errorf("优惠券模板不存在: couponID=%d, txnID=%s", txnContext.CouponID, txnID)
		tp.redisClient.SetEX(ctx, idempotentKey, "rollback", time.Hour)
		return primitive.RollbackMessageState
	}

	// 3. 开始数据库事务
	tx := tp.data.DB().Begin()
	if tx.Error != nil {
		log.Errorf("开始数据库事务失败: %v, txnID=%s", tx.Error, txnID)
		tp.redisClient.SetEX(ctx, idempotentKey, "rollback", time.Hour)
		return primitive.RollbackMessageState
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			tp.redisClient.SetEX(ctx, idempotentKey, "rollback", time.Hour)
			log.Errorf("本地事务执行异常: %v, txnID=%s", r, txnID)
		}
	}()

	// 4. 创建用户优惠券
	userCouponDO := &do.UserCouponDO{
		CouponTemplateID: txnContext.CouponID,
		UserID:           txnContext.UserID,
		CouponCode:       txnContext.CouponSn,
		Status:           do.UserCouponStatusUnused,
		ReceivedAt:       time.Now(),
		ExpiredAt:        templateDO.ValidEndTime,
	}

	if err := tp.data.UserCoupons().Create(ctx, tx, userCouponDO); err != nil {
		tx.Rollback()
		log.Errorf("创建用户优惠券失败: %v, txnID=%s", err, txnID)
		tp.redisClient.SetEX(ctx, idempotentKey, "rollback", time.Hour)
		return primitive.RollbackMessageState
	}

	// 5. 更新统计信息
	if err := tp.updateCouponTemplateStats(ctx, tx, txnContext.CouponID); err != nil {
		tx.Rollback()
		log.Errorf("更新优惠券模板统计失败: %v, txnID=%s", err, txnID)
		tp.redisClient.SetEX(ctx, idempotentKey, "rollback", time.Hour)
		return primitive.RollbackMessageState
	}

	if err := tp.updateFlashSaleStats(ctx, tx, txnContext.ActivityID); err != nil {
		tx.Rollback()
		log.Errorf("更新活动统计失败: %v, txnID=%s", err, txnID)
		tp.redisClient.SetEX(ctx, idempotentKey, "rollback", time.Hour)
		return primitive.RollbackMessageState
	}

	// 6. 提交数据库事务
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		log.Errorf("提交数据库事务失败: %v, txnID=%s", err, txnID)
		tp.redisClient.SetEX(ctx, idempotentKey, "rollback", time.Hour)
		return primitive.RollbackMessageState
	}

	// 7. 标记事务成功
	tp.redisClient.SetEX(ctx, idempotentKey, "committed", time.Hour)

	log.Infof("本地事务执行成功: userID=%d, activityID=%d, userCouponID=%d, txnID=%s",
		txnContext.UserID, txnContext.ActivityID, userCouponDO.ID, txnID)

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
	// 使用UpdateUsedCount方法
	if db, ok := tx.(*gorm.DB); ok {
		return tp.data.CouponTemplates().UpdateUsedCount(ctx, db, couponID, 1)
	}
	return fmt.Errorf("invalid transaction type")
}

// updateFlashSaleStats 更新秒杀活动统计
func (tp *TransactionProducer) updateFlashSaleStats(ctx context.Context, tx interface{}, activityID int64) error {
	// 使用IncrementSoldCount方法
	if db, ok := tx.(*gorm.DB); ok {
		return tp.data.FlashSales().IncrementSoldCount(ctx, db, activityID)
	}
	return fmt.Errorf("invalid transaction type")
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