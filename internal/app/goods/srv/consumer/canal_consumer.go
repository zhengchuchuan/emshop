package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/prometheus/client_golang/prometheus"

	"emshop/internal/app/goods/srv/data/v1/sync"
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
	NameServers   []string `json:"nameservers"`
	ConsumerGroup string   `json:"consumer_group"`
	Topic         string   `json:"topic"`
	MaxReconsume  int32    `json:"max_reconsume"`
}

// CanalConsumer Canal消息消费者
type CanalConsumer struct {
	config      *CanalConsumerConfig
	consumer    rocketmq.PushConsumer
	syncManager sync.DataSyncManagerInterface
	
	// 监控指标
	messageTotal prometheus.Counter
	syncLatency  prometheus.Histogram
	errorTotal   prometheus.Counter
}

// NewCanalConsumer 创建Canal消费者
func NewCanalConsumer(config *CanalConsumerConfig, syncManager sync.DataSyncManagerInterface) *CanalConsumer {
	return &CanalConsumer{
		config:      config,
		syncManager: syncManager,
		messageTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "canal_message_total",
			Help: "Total number of canal messages processed",
		}),
		syncLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "canal_sync_latency_seconds",
			Help: "Latency of canal message sync operation",
		}),
		errorTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "canal_error_total",
			Help: "Total number of canal processing errors",
		}),
	}
}

// Start 启动消费者
func (c *CanalConsumer) Start() error {
	log.Info("starting canal consumer")
	
	// 创建消费者
	pushConsumer, err := rocketmq.NewPushConsumer(
		consumer.WithGroupName(c.config.ConsumerGroup),
		consumer.WithNameServer(c.config.NameServers),
		consumer.WithConsumeFromWhere(consumer.ConsumeFromLastOffset),
		consumer.WithMaxReconsumeTimes(c.config.MaxReconsume),
	)
	if err != nil {
		return fmt.Errorf("failed to create push consumer: %w", err)
	}
	
	c.consumer = pushConsumer
	
	// 订阅主题
	err = c.consumer.Subscribe(c.config.Topic, consumer.MessageSelector{}, c.handleMessage)
	if err != nil {
		return fmt.Errorf("failed to subscribe topic %s: %w", c.config.Topic, err)
	}
	
	// 启动消费者
	err = c.consumer.Start()
	if err != nil {
		return fmt.Errorf("failed to start consumer: %w", err)
	}
	
	log.Infof("canal consumer started successfully, group=%s, topic=%s", 
		c.config.ConsumerGroup, c.config.Topic)
	
	return nil
}

// Stop 停止消费者
func (c *CanalConsumer) Stop() error {
	if c.consumer != nil {
		log.Info("stopping canal consumer")
		err := c.consumer.Shutdown()
		if err != nil {
			log.Errorf("failed to shutdown consumer: %v", err)
			return err
		}
		log.Info("canal consumer stopped successfully")
	}
	return nil
}

// handleMessage 处理RocketMQ消息
func (c *CanalConsumer) handleMessage(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	startTime := time.Now()
	defer func() {
		c.syncLatency.Observe(time.Since(startTime).Seconds())
	}()
	
	for _, msg := range msgs {
		c.messageTotal.Inc()
		
		// 解析Canal消息
		var canalMsg CanalMessage
		if err := json.Unmarshal(msg.Body, &canalMsg); err != nil {
			c.errorTotal.Inc()
			log.Errorf("failed to unmarshal canal message: %v, message: %s", err, string(msg.Body))
			continue // 跳过无法解析的消息，避免重复消费
		}
		
		// 处理消息
		if err := c.processMessage(ctx, &canalMsg); err != nil {
			c.errorTotal.Inc()
			log.Errorf("failed to process canal message: %v, message: %+v", err, canalMsg)
			return consumer.ConsumeRetryLater, err
		}
		
		log.Debugf("successfully processed canal message: table=%s, type=%s, data_count=%d", 
			canalMsg.Table, canalMsg.Type, len(canalMsg.Data))
	}
	
	return consumer.ConsumeSuccess, nil
}

// processMessage 处理Canal消息
func (c *CanalConsumer) processMessage(ctx context.Context, msg *CanalMessage) error {
	// 只处理emshop_goods_srv数据库的消息
	if msg.Database != "emshop_goods_srv" {
		log.Debugf("ignoring message from database: %s", msg.Database)
		return nil
	}
	
	// 忽略DDL操作
	if msg.IsDdl {
		log.Debugf("ignoring DDL operation for table: %s", msg.Table)
		return nil
	}
	
	// 根据表类型处理消息
	switch msg.Table {
	case "goods":
		return c.handleGoodsChange(ctx, msg)
	case "brands":
		return c.handleBrandsChange(ctx, msg)
	case "category":
		return c.handleCategoryChange(ctx, msg)
	case "category_brand":
		return c.handleCategoryBrandChange(ctx, msg)
	case "banner":
		return c.handleBannerChange(ctx, msg)
	default:
		log.Debugf("ignoring message for unsupported table: %s", msg.Table)
		return nil
	}
}

// handleGoodsChange 处理商品表变更
func (c *CanalConsumer) handleGoodsChange(ctx context.Context, msg *CanalMessage) error {
	log.Debugf("processing goods change: type=%s, data_count=%d", msg.Type, len(msg.Data))
	
	switch msg.Type {
	case "INSERT", "UPDATE":
		// 处理新增和更新
		for _, data := range msg.Data {
			if idVal, ok := data["id"]; ok {
				// 处理不同类型的ID值
				var goodsID uint64
				switch v := idVal.(type) {
				case float64:
					goodsID = uint64(v)
				case int64:
					goodsID = uint64(v)
				case int:
					goodsID = uint64(v)
				case string:
					// 如果是字符串，尝试解析
					if parsed, err := parseUint64(v); err == nil {
						goodsID = parsed
					} else {
						log.Errorf("failed to parse goods id from string: %s", v)
						continue
					}
				default:
					log.Errorf("unsupported goods id type: %T, value: %v", v, v)
					continue
				}
				
				log.Debugf("syncing goods to search: id=%d, operation=%s", goodsID, msg.Type)
				if err := c.syncManager.SyncToSearch(ctx, "goods", goodsID); err != nil {
					log.Errorf("failed to sync goods %d to search: %v", goodsID, err)
					return err
				}
			}
		}
		
	case "DELETE":
		// 处理删除
		dataSource := msg.Data
		if len(msg.Old) > 0 {
			dataSource = msg.Old // 删除操作使用old字段
		}
		
		for _, data := range dataSource {
			if idVal, ok := data["id"]; ok {
				var goodsID uint64
				switch v := idVal.(type) {
				case float64:
					goodsID = uint64(v)
				case int64:
					goodsID = uint64(v)
				case int:
					goodsID = uint64(v)
				default:
					log.Errorf("unsupported goods id type for deletion: %T, value: %v", v, v)
					continue
				}
				
				log.Debugf("removing goods from search: id=%d", goodsID)
				if err := c.syncManager.RemoveFromSearch(ctx, "goods", goodsID); err != nil {
					log.Errorf("failed to remove goods %d from search: %v", goodsID, err)
					return err
				}
			}
		}
		
	default:
		log.Warnf("unsupported operation type: %s", msg.Type)
	}
	
	return nil
}

// handleBrandsChange 处理品牌表变更 - 可能影响商品搜索
func (c *CanalConsumer) handleBrandsChange(ctx context.Context, msg *CanalMessage) error {
	log.Debugf("processing brands change: type=%s, data_count=%d", msg.Type, len(msg.Data))
	
	// 品牌变更可能影响相关商品的搜索结果
	// 这里可以实现更复杂的逻辑，比如找到相关商品并重新同步
	// 暂时记录日志，后续可以根据业务需求扩展
	
	return nil
}

// handleCategoryChange 处理分类表变更
func (c *CanalConsumer) handleCategoryChange(ctx context.Context, msg *CanalMessage) error {
	log.Debugf("processing category change: type=%s, data_count=%d", msg.Type, len(msg.Data))
	
	// 分类变更可能影响相关商品的搜索结果
	// 暂时记录日志，后续可以根据业务需求扩展
	
	return nil
}

// handleCategoryBrandChange 处理分类品牌关联表变更
func (c *CanalConsumer) handleCategoryBrandChange(ctx context.Context, msg *CanalMessage) error {
	log.Debugf("processing category_brand change: type=%s, data_count=%d", msg.Type, len(msg.Data))
	
	// 分类品牌关联变更可能影响相关商品的搜索结果
	// 暂时记录日志，后续可以根据业务需求扩展
	
	return nil
}

// handleBannerChange 处理banner表变更
func (c *CanalConsumer) handleBannerChange(ctx context.Context, msg *CanalMessage) error {
	log.Debugf("processing banner change: type=%s, data_count=%d", msg.Type, len(msg.Data))
	
	// Banner变更通常不需要同步到商品搜索
	// 这里可以实现banner相关的同步逻辑
	
	return nil
}

// parseUint64 解析字符串为uint64
func parseUint64(s string) (uint64, error) {
	var result uint64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid character in number: %c", c)
		}
		result = result*10 + uint64(c-'0')
	}
	return result, nil
}

// IsRunning 检查消费者是否运行中
func (c *CanalConsumer) IsRunning() bool {
	return c.consumer != nil
}

// GetMetrics 获取监控指标
func (c *CanalConsumer) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"consumer_group": c.config.ConsumerGroup,
		"topic":          c.config.Topic,
		"running":        c.IsRunning(),
	}
}