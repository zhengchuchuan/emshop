package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"emshop/pkg/log"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	rocketmq "github.com/apache/rocketmq-client-go/v2"
	redisClient "github.com/go-redis/redis/v8"
)

// RetryManager 重试管理器
type RetryManager struct {
	producer    rocketmq.Producer
	redisClient *redisClient.Client
	topic       string
	maxRetries  int
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries    int           `json:"max_retries"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	Multiplier    float64       `json:"multiplier"`
	EnableJitter  bool          `json:"enable_jitter"`
}

// RetryRecord 重试记录
type RetryRecord struct {
	MessageID     string            `json:"message_id"`
	Topic         string            `json:"topic"`
	Tag           string            `json:"tag"`
	Body          string            `json:"body"`
	Properties    map[string]string `json:"properties"`
	RetryCount    int               `json:"retry_count"`
	MaxRetries    int               `json:"max_retries"`
	LastRetryTime time.Time         `json:"last_retry_time"`
	NextRetryTime time.Time         `json:"next_retry_time"`
	ErrorMessage  string            `json:"error_message"`
	Status        string            `json:"status"` // "retrying", "failed", "dead_letter"
}

// MessageFailureRecord 消息失败记录
type MessageFailureRecord struct {
	MessageID      string    `json:"message_id"`
	Topic          string    `json:"topic"`
	Tag            string    `json:"tag"`
	Body           string    `json:"body"`
	FailureReason  string    `json:"failure_reason"`
	FailureTime    time.Time `json:"failure_time"`
	RetryAttempts  int       `json:"retry_attempts"`
	LastRetryTime  time.Time `json:"last_retry_time"`
	IsDeadLetter   bool      `json:"is_dead_letter"`
}

// NewRetryManager 创建重试管理器
func NewRetryManager(nameServers []string, groupName, topic string, redisClient *redisClient.Client, maxRetries int) (*RetryManager, error) {
	// 创建RocketMQ Producer
	p, err := rocketmq.NewProducer(
		producer.WithNameServer(nameServers),
		producer.WithRetry(3),
		producer.WithGroupName(groupName+"-retry"),
	)
	if err != nil {
		return nil, fmt.Errorf("创建重试管理器失败: %v", err)
	}

	// 启动生产者
	if err := p.Start(); err != nil {
		return nil, fmt.Errorf("启动重试生产者失败: %v", err)
	}

	return &RetryManager{
		producer:    p,
		redisClient: redisClient,
		topic:       topic,
		maxRetries:  maxRetries,
	}, nil
}

// ScheduleRetry 调度重试
func (rm *RetryManager) ScheduleRetry(ctx context.Context, originalMsg *primitive.MessageExt, err error, config *RetryConfig) error {
	// 获取当前重试次数
	retryCountStr := originalMsg.GetProperty("retry_count")
	retryCount := 0
	if retryCountStr != "" {
		fmt.Sscanf(retryCountStr, "%d", &retryCount)
	}

	// 检查是否超过最大重试次数
	if retryCount >= config.MaxRetries {
		return rm.sendToDeadLetterQueue(ctx, originalMsg, err)
	}

	// 计算下次重试时间
	nextRetryTime := rm.calculateNextRetryTime(retryCount, config)
	delayLevel := rm.calculateDelayLevel(retryCount, config)

	// 创建重试记录
	retryRecord := &RetryRecord{
		MessageID:     originalMsg.MsgId,
		Topic:         originalMsg.Topic,
		Tag:           originalMsg.GetTags(),
		Body:          string(originalMsg.Body),
		Properties:    originalMsg.GetProperties(),
		RetryCount:    retryCount + 1,
		MaxRetries:    config.MaxRetries,
		LastRetryTime: time.Now(),
		NextRetryTime: nextRetryTime,
		ErrorMessage:  err.Error(),
		Status:        "retrying",
	}

	// 保存重试记录到Redis
	if err := rm.saveRetryRecord(ctx, retryRecord); err != nil {
		log.Errorf("保存重试记录失败: %v", err)
	}

	// 构建重试消息
	retryMsg := &primitive.Message{
		Topic: rm.topic,
		Body:  originalMsg.Body,
	}
	retryMsg.WithTag(originalMsg.GetTags())
	// GetKeys() 返回字符串，需要转换为字符串数组
	keys := originalMsg.GetKeys()
	if keys != "" {
		retryMsg.WithKeys([]string{keys})
	}
	retryMsg.WithDelayTimeLevel(delayLevel)

	// 复制原消息的属性并更新重试信息
	for key, value := range originalMsg.GetProperties() {
		retryMsg.WithProperty(key, value)
	}
	retryMsg.WithProperty("retry_count", fmt.Sprintf("%d", retryRecord.RetryCount))
	retryMsg.WithProperty("original_msg_id", originalMsg.MsgId)
	retryMsg.WithProperty("retry_reason", err.Error())
	retryMsg.WithProperty("retry_timestamp", fmt.Sprintf("%d", time.Now().Unix()))

	// 发送延迟消息
	result, err := rm.producer.SendSync(ctx, retryMsg)
	if err != nil {
		log.Errorf("发送重试消息失败: %v", err)
		return err
	}

	log.Infof("消息重试调度成功: originalMsgID=%s, retryCount=%d, delayLevel=%d, newMsgID=%s",
		originalMsg.MsgId, retryRecord.RetryCount, delayLevel, result.MsgID)

	return nil
}

// calculateNextRetryTime 计算下次重试时间
func (rm *RetryManager) calculateNextRetryTime(retryCount int, config *RetryConfig) time.Time {
	// 指数退避算法
	delay := config.InitialDelay
	for i := 0; i < retryCount; i++ {
		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
			break
		}
	}

	// 添加随机抖动（可选）
	if config.EnableJitter {
		jitter := time.Duration(float64(delay) * 0.1) // 10%抖动
		delay += time.Duration(jitter.Nanoseconds() * (2*time.Now().UnixNano()%2 - 1) / 10)
	}

	return time.Now().Add(delay)
}

// calculateDelayLevel 计算RocketMQ延迟级别
func (rm *RetryManager) calculateDelayLevel(retryCount int, config *RetryConfig) int {
	// RocketMQ延迟级别映射：1s 5s 10s 30s 1m 2m 3m 4m 5m 6m 7m 8m 9m 10m 20m 30m 1h 2h
	delayLevels := []time.Duration{
		1 * time.Second,    // 1
		5 * time.Second,    // 2
		10 * time.Second,   // 3
		30 * time.Second,   // 4
		1 * time.Minute,    // 5
		2 * time.Minute,    // 6
		3 * time.Minute,    // 7
		4 * time.Minute,    // 8
		5 * time.Minute,    // 9
		6 * time.Minute,    // 10
		7 * time.Minute,    // 11
		8 * time.Minute,    // 12
		9 * time.Minute,    // 13
		10 * time.Minute,   // 14
		20 * time.Minute,   // 15
		30 * time.Minute,   // 16
		1 * time.Hour,      // 17
		2 * time.Hour,      // 18
	}

	// 计算期望的延迟时间
	expectedDelay := config.InitialDelay
	for i := 0; i < retryCount; i++ {
		expectedDelay = time.Duration(float64(expectedDelay) * config.Multiplier)
		if expectedDelay > config.MaxDelay {
			expectedDelay = config.MaxDelay
			break
		}
	}

	// 找到最接近的延迟级别
	for i, levelDelay := range delayLevels {
		if expectedDelay <= levelDelay {
			return i + 1
		}
	}

	// 如果超过了最大延迟级别，返回最高级别
	return len(delayLevels)
}

// sendToDeadLetterQueue 发送到死信队列
func (rm *RetryManager) sendToDeadLetterQueue(ctx context.Context, originalMsg *primitive.MessageExt, err error) error {
	// 创建死信队列消息
	deadLetterMsg := &primitive.Message{
		Topic: rm.topic + "_DLQ", // 死信队列主题
		Body:  originalMsg.Body,
	}
	deadLetterMsg.WithTag("DEAD_LETTER")
	// GetKeys() 返回字符串，需要转换为字符串数组
	keys := originalMsg.GetKeys()
	if keys != "" {
		deadLetterMsg.WithKeys([]string{keys})
	}

	// 复制原消息的属性
	for key, value := range originalMsg.GetProperties() {
		deadLetterMsg.WithProperty(key, value)
	}
	deadLetterMsg.WithProperty("original_msg_id", originalMsg.MsgId)
	deadLetterMsg.WithProperty("dead_letter_reason", err.Error())
	deadLetterMsg.WithProperty("dead_letter_timestamp", fmt.Sprintf("%d", time.Now().Unix()))
	deadLetterMsg.WithProperty("original_topic", originalMsg.Topic)

	// 发送死信消息
	result, sendErr := rm.producer.SendSync(ctx, deadLetterMsg)
	if sendErr != nil {
		log.Errorf("发送死信消息失败: %v", sendErr)
		return sendErr
	}

	// 记录失败记录
	failureRecord := &MessageFailureRecord{
		MessageID:     originalMsg.MsgId,
		Topic:         originalMsg.Topic,
		Tag:           originalMsg.GetTags(),
		Body:          string(originalMsg.Body),
		FailureReason: err.Error(),
		FailureTime:   time.Now(),
		IsDeadLetter:  true,
	}

	if err := rm.saveFailureRecord(ctx, failureRecord); err != nil {
		log.Errorf("保存失败记录失败: %v", err)
	}

	log.Errorf("消息发送到死信队列: originalMsgID=%s, deadLetterMsgID=%s, reason=%s",
		originalMsg.MsgId, result.MsgID, err.Error())

	return nil
}

// saveRetryRecord 保存重试记录到Redis
func (rm *RetryManager) saveRetryRecord(ctx context.Context, record *RetryRecord) error {
	key := fmt.Sprintf("retry:record:%s", record.MessageID)
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	// 保存24小时
	return rm.redisClient.SetEX(ctx, key, data, 24*time.Hour).Err()
}

// saveFailureRecord 保存失败记录到Redis
func (rm *RetryManager) saveFailureRecord(ctx context.Context, record *MessageFailureRecord) error {
	key := fmt.Sprintf("failure:record:%s", record.MessageID)
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	// 保存7天
	return rm.redisClient.SetEX(ctx, key, data, 7*24*time.Hour).Err()
}

// GetRetryRecord 获取重试记录
func (rm *RetryManager) GetRetryRecord(ctx context.Context, messageID string) (*RetryRecord, error) {
	key := fmt.Sprintf("retry:record:%s", messageID)
	data := rm.redisClient.Get(ctx, key).Val()
	if data == "" {
		return nil, fmt.Errorf("重试记录不存在: %s", messageID)
	}

	var record RetryRecord
	if err := json.Unmarshal([]byte(data), &record); err != nil {
		return nil, err
	}

	return &record, nil
}

// GetFailureRecord 获取失败记录
func (rm *RetryManager) GetFailureRecord(ctx context.Context, messageID string) (*MessageFailureRecord, error) {
	key := fmt.Sprintf("failure:record:%s", messageID)
	data := rm.redisClient.Get(ctx, key).Val()
	if data == "" {
		return nil, fmt.Errorf("失败记录不存在: %s", messageID)
	}

	var record MessageFailureRecord
	if err := json.Unmarshal([]byte(data), &record); err != nil {
		return nil, err
	}

	return &record, nil
}

// ListRetryRecords 列出重试记录
func (rm *RetryManager) ListRetryRecords(ctx context.Context, limit int) ([]*RetryRecord, error) {
	keys, err := rm.redisClient.Keys(ctx, "retry:record:*").Result()
	if err != nil {
		return nil, err
	}

	var records []*RetryRecord
	for i, key := range keys {
		if limit > 0 && i >= limit {
			break
		}

		data := rm.redisClient.Get(ctx, key).Val()
		if data == "" {
			continue
		}

		var record RetryRecord
		if err := json.Unmarshal([]byte(data), &record); err != nil {
			log.Errorf("解析重试记录失败: %v", err)
			continue
		}

		records = append(records, &record)
	}

	return records, nil
}

// ListFailureRecords 列出失败记录
func (rm *RetryManager) ListFailureRecords(ctx context.Context, limit int) ([]*MessageFailureRecord, error) {
	keys, err := rm.redisClient.Keys(ctx, "failure:record:*").Result()
	if err != nil {
		return nil, err
	}

	var records []*MessageFailureRecord
	for i, key := range keys {
		if limit > 0 && i >= limit {
			break
		}

		data := rm.redisClient.Get(ctx, key).Val()
		if data == "" {
			continue
		}

		var record MessageFailureRecord
		if err := json.Unmarshal([]byte(data), &record); err != nil {
			log.Errorf("解析失败记录失败: %v", err)
			continue
		}

		records = append(records, &record)
	}

	return records, nil
}

// CleanupExpiredRecords 清理过期记录
func (rm *RetryManager) CleanupExpiredRecords(ctx context.Context) error {
	// 清理过期的重试记录
	retryKeys, err := rm.redisClient.Keys(ctx, "retry:record:*").Result()
	if err != nil {
		return err
	}

	for _, key := range retryKeys {
		ttl := rm.redisClient.TTL(ctx, key).Val()
		if ttl < 0 { // 已过期但未被删除
			rm.redisClient.Del(ctx, key)
		}
	}

	// 清理过期的失败记录
	failureKeys, err := rm.redisClient.Keys(ctx, "failure:record:*").Result()
	if err != nil {
		return err
	}

	for _, key := range failureKeys {
		ttl := rm.redisClient.TTL(ctx, key).Val()
		if ttl < 0 { // 已过期但未被删除
			rm.redisClient.Del(ctx, key)
		}
	}

	log.Infof("清理过期记录完成: retry=%d, failure=%d", len(retryKeys), len(failureKeys))
	return nil
}

// Shutdown 关闭重试管理器
func (rm *RetryManager) Shutdown() error {
	if rm.producer != nil {
		log.Info("正在关闭重试管理器...")
		err := rm.producer.Shutdown()
		if err != nil {
			log.Errorf("关闭重试生产者失败: %v", err)
			return err
		}
		log.Info("重试管理器已关闭")
	}
	return nil
}

// DefaultRetryConfig 默认重试配置
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:   5,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Minute,
		Multiplier:   2.0,
		EnableJitter: true,
	}
}