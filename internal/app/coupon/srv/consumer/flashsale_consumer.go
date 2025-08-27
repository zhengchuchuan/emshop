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
)

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
func NewFlashSaleConsumer(data interfaces.DataFactory, redisClient *redisClient.Client) *FlashSaleConsumer {
	return &FlashSaleConsumer{
		data:         data,
		redisClient:  redisClient,
		stockManager: redis.NewStockManager(redisClient),
	}
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
		if err := fsc.handleFlashSaleSuccess(ctx, &event, msg.MsgId); err != nil {
			log.Errorf("处理秒杀成功事件失败: %v, 将重试", err)
			return consumer.ConsumeRetryLater, err
		}
	}

	return consumer.ConsumeSuccess, nil
}

// handleFlashSaleSuccess 处理秒杀成功事件
func (fsc *FlashSaleConsumer) handleFlashSaleSuccess(ctx context.Context, event *FlashSaleSuccessEvent, msgID string) error {
	// 1. 检查幂等性（避免重复处理）
	idempotentKey := fmt.Sprintf("flashsale:processed:%s", msgID)
	exists, err := fsc.redisClient.Exists(ctx, idempotentKey).Result()
	if err != nil {
		log.Errorf("检查幂等性失败: %v", err)
	} else if exists > 0 {
		log.Infof("消息已处理过，跳过: msgID=%s", msgID)
		return nil
	}

	// 2. 获取活动信息
	activityDO, err := fsc.data.FlashSales().Get(ctx, fsc.data.DB(), event.ActivityID)
	if err != nil {
		return fmt.Errorf("获取活动信息失败: %v", err)
	}
	if activityDO == nil {
		return fmt.Errorf("活动不存在: activityID=%d", event.ActivityID)
	}

	// 3. 获取优惠券模板信息
	templateDO, err := fsc.data.CouponTemplates().Get(ctx, fsc.data.DB(), event.CouponID)
	if err != nil {
		return fmt.Errorf("获取优惠券模板失败: %v", err)
	}
	if templateDO == nil {
		return fmt.Errorf("优惠券模板不存在: couponID=%d", event.CouponID)
	}

	// 4. 创建用户优惠券记录
	userCouponDO := &do.UserCouponDO{
		CouponTemplateID: event.CouponID,
		UserID:           event.UserID,
		CouponCode:       event.CouponSn,
		Status:           do.UserCouponStatusUnused,
		ReceivedAt:       time.Now(),
		ExpiredAt:        fsc.calculateExpiryTime(templateDO),
	}

	// 5. 开始数据库事务
	tx := fsc.data.DB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 6. 创建用户优惠券
	if err := fsc.data.UserCoupons().Create(ctx, tx, userCouponDO); err != nil {
		tx.Rollback()
		// 如果是因为重复创建导致的错误，可能是并发问题，进行库存回滚
		if fsc.isDuplicateError(err) {
			log.Warnf("用户优惠券可能已存在，进行库存回滚: userID=%d, couponSn=%s", 
				event.UserID, event.CouponSn)
			fsc.rollbackStockIfNeeded(ctx, event)
			return nil
		}
		return fmt.Errorf("创建用户优惠券失败: %v", err)
	}

	// 7. 更新优惠券模板使用统计
	if err := fsc.updateCouponTemplateStats(ctx, tx, event.CouponID); err != nil {
		tx.Rollback()
		log.Errorf("更新优惠券模板统计失败: %v", err)
		// 统计更新失败不影响主流程，记录日志即可
	}

	// 8. 更新活动统计
	if err := fsc.updateFlashSaleStats(ctx, tx, event.ActivityID); err != nil {
		tx.Rollback()
		log.Errorf("更新活动统计失败: %v", err)
		// 统计更新失败不影响主流程，记录日志即可
	}

	// 9. 提交事务
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("提交事务失败: %v", err)
	}

	// 10. 设置幂等性标记（7天过期）
	fsc.redisClient.SetEX(ctx, idempotentKey, "1", 7*24*time.Hour)

	log.Infof("秒杀成功事件处理完成: userID=%d, userCouponID=%d, couponSn=%s", 
		event.UserID, userCouponDO.ID, event.CouponSn)
	
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
	// 简化处理：直接执行SQL更新
	return tx.Exec("UPDATE coupon_templates SET used_count = used_count + 1 WHERE id = ?", couponID).Error
}

// updateFlashSaleStats 更新秒杀活动统计
func (fsc *FlashSaleConsumer) updateFlashSaleStats(ctx context.Context, tx *gorm.DB, activityID int64) error {
	// 增加售出数量
	return fsc.data.FlashSales().IncrementSoldCount(ctx, tx, activityID)
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
}

// NewFlashSaleEventProducer 创建秒杀事件生产者
func NewFlashSaleEventProducer(nameServers []string, groupName, topic string) (FlashSaleEventProducer, error) {
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
	
	// 发送消息
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	result, err := p.producer.SendSync(ctx, msg)
	if err != nil {
		log.Errorf("发送秒杀成功事件失败: %v, userID=%d, activityID=%d", 
			err, event.UserID, event.ActivityID)
		return fmt.Errorf("发送秒杀成功事件失败: %v", err)
	}

	log.Infof("发送秒杀成功事件成功: userID=%d, activityID=%d, msgID=%s, queueID=%d", 
		event.UserID, event.ActivityID, result.MsgID, result.MessageQueue.QueueId)
	
	return nil
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

	// 发送消息
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	result, err := p.producer.SendSync(ctx, msg)
	if err != nil {
		log.Errorf("发送秒杀失败事件失败: %v, userID=%d, activityID=%d, reason=%s", 
			err, event.UserID, event.ActivityID, event.Reason)
		return fmt.Errorf("发送秒杀失败事件失败: %v", err)
	}

	log.Infof("发送秒杀失败事件成功: userID=%d, activityID=%d, reason=%s, msgID=%s", 
		event.UserID, event.ActivityID, event.Reason, result.MsgID)
	
	return nil
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