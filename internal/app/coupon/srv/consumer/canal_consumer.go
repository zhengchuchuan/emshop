package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/prometheus/client_golang/prometheus"

	"emshop/internal/app/coupon/srv/pkg/cache"
	"emshop/pkg/log"
)

// CanalMessage Canal消息结构体
type CanalMessage struct {
	Database    string                   `json:"database"`
	Table       string                   `json:"table"`
	Type        string                   `json:"type"` // INSERT, UPDATE, DELETE
	Data        []map[string]interface{} `json:"data"`
	Old         []map[string]interface{} `json:"old"`
	Es          int64                    `json:"es"`
	Ts          int64                    `json:"ts"`
	IsDdl       bool                     `json:"isDdl"`
	ExecuteTime int64                    `json:"executeTime"`
}

// CanalConsumerConfig Canal消费者配置
type CanalConsumerConfig struct {
	NameServers   []string `yaml:"nameservers"`
	ConsumerGroup string   `yaml:"consumer_group"`
	Topic         string   `yaml:"topic"`
	WatchTables   []string `yaml:"watch_tables"`
	BatchSize     int32    `yaml:"batch_size"`
}

// CouponCanalConsumer 优惠券Canal消费者
type CouponCanalConsumer struct {
	config       *CanalConsumerConfig
	consumer     rocketmq.PushConsumer
	cacheManager cache.CacheManager
	watchTables  map[string]bool

	// 监控指标
	messageTotal prometheus.Counter
	syncLatency  prometheus.Histogram
	errorTotal   prometheus.Counter
}

// NewCouponCanalConsumer 创建优惠券Canal消费者
func NewCouponCanalConsumer(config *CanalConsumerConfig, cacheManager cache.CacheManager) *CouponCanalConsumer {
	// 构建监听表映射
	watchTables := make(map[string]bool)
	for _, table := range config.WatchTables {
		watchTables[table] = true
	}

	// 初始化监控指标
	messageTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "coupon_canal_messages_total",
		Help: "Canal消息处理总数",
	})
	
	syncLatency := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "coupon_canal_sync_duration_seconds",
		Help:    "Canal缓存同步延迟",
		Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
	})
	
	errorTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "coupon_canal_errors_total", 
		Help: "Canal消息处理错误总数",
	})

	// 注册监控指标
	prometheus.MustRegister(messageTotal, syncLatency, errorTotal)

	return &CouponCanalConsumer{
		config:       config,
		cacheManager: cacheManager,
		watchTables:  watchTables,
		messageTotal: messageTotal,
		syncLatency:  syncLatency,
		errorTotal:   errorTotal,
	}
}

// Start 启动Canal消费者
func (ccc *CouponCanalConsumer) Start() error {
	// 创建RocketMQ消费者
	c, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer(ccc.config.NameServers),
		consumer.WithGroupName(ccc.config.ConsumerGroup),
		consumer.WithConsumeFromWhere(consumer.ConsumeFromLastOffset),
		consumer.WithConsumerModel(consumer.Clustering),
		consumer.WithConsumeMessageBatchMaxSize(int(ccc.config.BatchSize)),
	)
	if err != nil {
		return fmt.Errorf("创建Canal消费者失败: %v", err)
	}

	ccc.consumer = c

	// 订阅Canal主题
	err = c.Subscribe(ccc.config.Topic, consumer.MessageSelector{}, ccc.ConsumeCanalMessage)
	if err != nil {
		return fmt.Errorf("订阅Canal主题失败: %v", err)
	}

	// 启动消费者
	err = c.Start()
	if err != nil {
		return fmt.Errorf("启动Canal消费者失败: %v", err)
	}

	log.Infof("Canal消费者启动成功, topic: %s, group: %s", ccc.config.Topic, ccc.config.ConsumerGroup)
	return nil
}

// Stop 停止Canal消费者
func (ccc *CouponCanalConsumer) Stop() error {
	if ccc.consumer != nil {
		err := ccc.consumer.Shutdown()
		if err != nil {
			return fmt.Errorf("停止Canal消费者失败: %v", err)
		}
	}
	log.Info("Canal消费者已停止")
	return nil
}

// ConsumeCanalMessage 消费Canal消息，实现缓存一致性
func (ccc *CouponCanalConsumer) ConsumeCanalMessage(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	timer := prometheus.NewTimer(ccc.syncLatency)
	defer timer.ObserveDuration()

	for _, msg := range msgs {
		ccc.messageTotal.Inc()

		var canalMsg CanalMessage
		if err := json.Unmarshal(msg.Body, &canalMsg); err != nil {
			log.Errorf("Canal消息解析失败: %v", err)
			ccc.errorTotal.Inc()
			continue
		}

		// 只处理优惠券相关表
		if !ccc.watchTables[canalMsg.Table] {
			continue
		}

		log.Infof("收到Canal消息: database=%s, table=%s, type=%s, dataCount=%d",
			canalMsg.Database, canalMsg.Table, canalMsg.Type, len(canalMsg.Data))

		// 根据表名和操作类型处理缓存更新
		if err := ccc.handleTableChange(&canalMsg); err != nil {
			log.Errorf("处理Canal消息失败: %v", err)
			ccc.errorTotal.Inc()
			// 继续处理下一条消息，不返回错误
		}
	}

	return consumer.ConsumeSuccess, nil
}

// handleTableChange 根据表变更处理缓存
func (ccc *CouponCanalConsumer) handleTableChange(msg *CanalMessage) error {
	switch msg.Table {
	case "coupon_templates":
		return ccc.handleCouponTemplateChange(msg)
	case "user_coupons":
		return ccc.handleUserCouponChange(msg)
	case "flash_sale_activities":
		return ccc.handleFlashSaleChange(msg)
	default:
		log.Warnf("未处理的表变更: %s", msg.Table)
		return nil
	}
}

// handleCouponTemplateChange 处理优惠券模板变更
func (ccc *CouponCanalConsumer) handleCouponTemplateChange(msg *CanalMessage) error {
	for _, data := range msg.Data {
		couponIDStr, ok := data["id"].(string)
		if !ok {
			continue
		}

		couponID, err := strconv.ParseInt(couponIDStr, 10, 64)
		if err != nil {
			log.Errorf("解析coupon ID失败: %v", err)
			continue
		}

		// 构建需要失效的缓存key
		keys := []string{
			cache.CacheKeys.CouponTemplate(couponID),
		}

		// 如果是删除操作，还需要清理相关缓存
		if msg.Type == "DELETE" {
			// 清理相关的用户优惠券列表缓存
			pattern := fmt.Sprintf("coupon:user:list:*")
			if err := ccc.cacheManager.InvalidateCacheByPattern(context.Background(), pattern); err != nil {
				log.Errorf("清理用户优惠券列表缓存失败: %v", err)
			}

			// 清理优惠券有效性缓存
			pattern = fmt.Sprintf("coupon:valid:%d:*", couponID)
			if err := ccc.cacheManager.InvalidateCacheByPattern(context.Background(), pattern); err != nil {
				log.Errorf("清理优惠券有效性缓存失败: %v", err)
			}
		}

		// 执行缓存失效
		ccc.cacheManager.InvalidateCache(keys...)

		log.Infof("优惠券模板缓存失效: couponID=%d, type=%s", couponID, msg.Type)
	}

	return nil
}

// handleUserCouponChange 处理用户优惠券变更
func (ccc *CouponCanalConsumer) handleUserCouponChange(msg *CanalMessage) error {
	for _, data := range msg.Data {
		userIDStr, ok := data["user_id"].(string)
		if !ok {
			continue
		}

		userCouponIDStr, ok := data["id"].(string)
		if !ok {
			continue
		}

		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			log.Errorf("解析user ID失败: %v", err)
			continue
		}

		userCouponID, err := strconv.ParseInt(userCouponIDStr, 10, 64)
		if err != nil {
			log.Errorf("解析user coupon ID失败: %v", err)
			continue
		}

		// 失效用户相关缓存
		keys := []string{
			cache.CacheKeys.UserCoupon(userCouponID),           // 单个用户优惠券
			cache.CacheKeys.UserCouponList(userID),             // 用户优惠券列表
			cache.CacheKeys.UserAvailableCoupons(userID),       // 用户可用优惠券
			fmt.Sprintf("coupon:user:count:%d", userID),        // 用户优惠券数量
		}

		ccc.cacheManager.InvalidateCache(keys...)
		log.Infof("用户优惠券缓存失效: userID=%d, userCouponID=%d, type=%s", userID, userCouponID, msg.Type)
	}

	return nil
}

// handleFlashSaleChange 处理秒杀活动变更
func (ccc *CouponCanalConsumer) handleFlashSaleChange(msg *CanalMessage) error {
	for _, data := range msg.Data {
		activityIDStr, ok := data["id"].(string)
		if !ok {
			continue
		}

		activityID, err := strconv.ParseInt(activityIDStr, 10, 64)
		if err != nil {
			log.Errorf("解析activity ID失败: %v", err)
			continue
		}

		// 获取关联的优惠券ID
		var couponID int64
		if couponIDStr, ok := data["coupon_id"].(string); ok {
			couponID, _ = strconv.ParseInt(couponIDStr, 10, 64)
		}

		// 失效秒杀活动相关缓存
		keys := []string{
			cache.CacheKeys.FlashSaleActivity(activityID),  // 秒杀活动信息
			cache.CacheKeys.FlashSaleStatus(activityID),    // 秒杀状态
		}

		// 如果有关联的优惠券，也要失效库存缓存
		if couponID > 0 {
			keys = append(keys, cache.CacheKeys.CouponStock(couponID))
		}

		ccc.cacheManager.InvalidateCache(keys...)
		log.Infof("秒杀活动缓存失效: activityID=%d, couponID=%d, type=%s", activityID, couponID, msg.Type)
	}

	return nil
}