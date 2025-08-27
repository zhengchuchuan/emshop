package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"emshop/pkg/log"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

// CanalMessage Canal消息结构
type CanalMessage struct {
	ID            int64                    `json:"id"`
	Database      string                   `json:"database"`
	Table         string                   `json:"table"`
	PKNames       []string                 `json:"pkNames"`
	IsDDL         bool                     `json:"isDdl"`
	Type          string                   `json:"type"`
	Es            int64                    `json:"es"`
	Ts            int64                    `json:"ts"`
	SQL           string                   `json:"sql"`
	SqlType       map[string]int           `json:"sqlType"`
	MySQLType     map[string]string        `json:"mysqlType"`
	Data          []map[string]interface{} `json:"data"`
	Old           []map[string]interface{} `json:"old"`
}

// CanalSyncManager Canal缓存同步管理器
type CanalSyncManager struct {
	cacheManager CacheManager
	consumer     rocketmq.PushConsumer
	watchTables  []string
	batchSize    int
	database     string
}

// CanalSyncConfig Canal同步配置
type CanalSyncConfig struct {
	NameServers   []string
	ConsumerGroup string
	Topic         string
	WatchTables   []string
	BatchSize     int
	Database      string
}

// NewCanalSyncManager 创建Canal同步管理器
func NewCanalSyncManager(cacheManager CacheManager, config *CanalSyncConfig) (*CanalSyncManager, error) {
	// 创建RocketMQ消费者
	pushConsumer, err := rocketmq.NewPushConsumer(
		consumer.WithGroupName(config.ConsumerGroup),
		consumer.WithNameServer(config.NameServers),
		consumer.WithConsumeFromWhere(consumer.ConsumeFromFirstOffset),
		consumer.WithConsumerModel(consumer.Clustering),
	)
	if err != nil {
		return nil, fmt.Errorf("创建RocketMQ消费者失败: %v", err)
	}

	manager := &CanalSyncManager{
		cacheManager: cacheManager,
		consumer:     pushConsumer,
		watchTables:  config.WatchTables,
		batchSize:    config.BatchSize,
		database:     config.Database,
	}

	// 订阅Canal消息
	err = pushConsumer.Subscribe(config.Topic, consumer.MessageSelector{}, manager.handleCanalMessage)
	if err != nil {
		return nil, fmt.Errorf("订阅Canal消息失败: %v", err)
	}

	log.Infof("Canal缓存同步管理器初始化成功, 监听表: %v", config.WatchTables)
	return manager, nil
}

// Start 启动Canal同步
func (csm *CanalSyncManager) Start() error {
	err := csm.consumer.Start()
	if err != nil {
		return fmt.Errorf("启动Canal消费者失败: %v", err)
	}
	log.Info("Canal缓存同步服务启动成功")
	return nil
}

// Stop 停止Canal同步
func (csm *CanalSyncManager) Stop() error {
	err := csm.consumer.Shutdown()
	if err != nil {
		log.Errorf("停止Canal消费者失败: %v", err)
		return err
	}
	log.Info("Canal缓存同步服务停止")
	return nil
}

// handleCanalMessage 处理Canal消息
func (csm *CanalSyncManager) handleCanalMessage(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for _, msg := range msgs {
		if err := csm.processCanalMessage(msg); err != nil {
			log.Errorf("处理Canal消息失败: %v, msgId: %s", err, msg.MsgId)
			// 对于缓存同步失败，通常选择继续处理其他消息，而不是重试
			// 因为缓存可以通过其他方式恢复（如查询数据库）
			continue
		}
	}
	return consumer.ConsumeSuccess, nil
}

// processCanalMessage 处理单个Canal消息
func (csm *CanalSyncManager) processCanalMessage(msg *primitive.MessageExt) error {
	var canalMsg CanalMessage
	if err := json.Unmarshal(msg.Body, &canalMsg); err != nil {
		return fmt.Errorf("解析Canal消息失败: %v", err)
	}

	// 检查数据库是否匹配
	if canalMsg.Database != csm.database {
		log.Debugf("跳过非目标数据库的消息: %s", canalMsg.Database)
		return nil
	}

	// 检查表是否需要监听
	if !csm.isWatchedTable(canalMsg.Table) {
		log.Debugf("跳过非监听表的消息: %s", canalMsg.Table)
		return nil
	}

	// 根据操作类型处理缓存
	switch canalMsg.Type {
	case "INSERT", "UPDATE", "DELETE":
		return csm.handleDataChange(canalMsg)
	case "CREATE", "ALTER", "DROP":
		if canalMsg.IsDDL {
			log.Infof("收到DDL操作: %s, table: %s", canalMsg.Type, canalMsg.Table)
			return csm.handleDDLChange(canalMsg)
		}
	}

	return nil
}

// handleDataChange 处理数据变更
func (csm *CanalSyncManager) handleDataChange(canalMsg CanalMessage) error {
	log.Infof("处理数据变更: table=%s, type=%s, rows=%d", 
		canalMsg.Table, canalMsg.Type, len(canalMsg.Data))

	invalidateKeys := make([]string, 0)

	// 根据不同表生成不同的缓存键
	switch canalMsg.Table {
	case "coupon_templates":
		keys := csm.generateCouponTemplateKeys(canalMsg)
		invalidateKeys = append(invalidateKeys, keys...)
		
	case "user_coupons":
		keys := csm.generateUserCouponKeys(canalMsg)
		invalidateKeys = append(invalidateKeys, keys...)
		
	case "flash_sale_activities":
		keys := csm.generateFlashSaleKeys(canalMsg)
		invalidateKeys = append(invalidateKeys, keys...)
		
	default:
		log.Warnf("未处理的表变更: %s", canalMsg.Table)
		return nil
	}

	// 批量失效缓存
	if len(invalidateKeys) > 0 {
		csm.cacheManager.InvalidateCache(invalidateKeys...)
		log.Infof("失效缓存键: %v", invalidateKeys)
	}

	return nil
}

// handleDDLChange 处理DDL变更
func (csm *CanalSyncManager) handleDDLChange(canalMsg CanalMessage) error {
	// DDL变更通常需要清空相关表的所有缓存
	patterns := csm.getDDLCachePatterns(canalMsg.Table)
	for _, pattern := range patterns {
		if err := csm.cacheManager.InvalidateCacheByPattern(context.Background(), pattern); err != nil {
			log.Errorf("DDL缓存失效失败: pattern=%s, err=%v", pattern, err)
		}
	}
	log.Infof("DDL操作缓存清理完成: table=%s", canalMsg.Table)
	return nil
}

// generateCouponTemplateKeys 生成优惠券模板缓存键
func (csm *CanalSyncManager) generateCouponTemplateKeys(canalMsg CanalMessage) []string {
	keys := make([]string, 0)
	
	// 处理当前数据的缓存键
	for _, row := range canalMsg.Data {
		if id, ok := row["id"]; ok {
			key := fmt.Sprintf("coupon:template:%v", id)
			keys = append(keys, key)
		}
	}
	
	// 如果是UPDATE操作，还需要处理旧数据的缓存键
	if canalMsg.Type == "UPDATE" {
		for _, row := range canalMsg.Old {
			if id, ok := row["id"]; ok {
				key := fmt.Sprintf("coupon:template:%v", id)
				keys = append(keys, key)
			}
		}
	}
	
	return keys
}

// generateUserCouponKeys 生成用户优惠券缓存键
func (csm *CanalSyncManager) generateUserCouponKeys(canalMsg CanalMessage) []string {
	keys := make([]string, 0)
	
	// 处理当前数据的缓存键
	for _, row := range canalMsg.Data {
		if id, ok := row["id"]; ok {
			key := fmt.Sprintf("coupon:user:%v", id)
			keys = append(keys, key)
		}
		
		// 用户优惠券列表缓存也需要失效
		if userID, ok := row["user_id"]; ok {
			listKey := fmt.Sprintf("coupon:user:list:%v", userID)
			availableKey := fmt.Sprintf("coupon:user:available:%v", userID)
			keys = append(keys, listKey, availableKey)
		}
	}
	
	// 处理UPDATE操作的旧数据
	if canalMsg.Type == "UPDATE" {
		for _, row := range canalMsg.Old {
			if userID, ok := row["user_id"]; ok {
				listKey := fmt.Sprintf("coupon:user:list:%v", userID)
				availableKey := fmt.Sprintf("coupon:user:available:%v", userID)
				keys = append(keys, listKey, availableKey)
			}
		}
	}
	
	return keys
}

// generateFlashSaleKeys 生成秒杀活动缓存键
func (csm *CanalSyncManager) generateFlashSaleKeys(canalMsg CanalMessage) []string {
	keys := make([]string, 0)
	
	// 处理当前数据的缓存键
	for _, row := range canalMsg.Data {
		if id, ok := row["id"]; ok {
			activityKey := fmt.Sprintf("flashsale:activity:%v", id)
			statusKey := fmt.Sprintf("flashsale:status:%v", id)
			stockKey := fmt.Sprintf("flashsale:stock:%v", id)
			keys = append(keys, activityKey, statusKey, stockKey)
		}
	}
	
	return keys
}

// getDDLCachePatterns 获取DDL操作需要清理的缓存模式
func (csm *CanalSyncManager) getDDLCachePatterns(tableName string) []string {
	patterns := make([]string, 0)
	
	switch tableName {
	case "coupon_templates":
		patterns = append(patterns, "coupon:template:*")
	case "user_coupons":
		patterns = append(patterns, "coupon:user:*")
	case "flash_sale_activities":
		patterns = append(patterns, "flashsale:*")
	}
	
	return patterns
}

// isWatchedTable 检查表是否在监听列表中
func (csm *CanalSyncManager) isWatchedTable(tableName string) bool {
	for _, watchTable := range csm.watchTables {
		if strings.EqualFold(watchTable, tableName) {
			return true
		}
	}
	return false
}

// GetSyncStats 获取同步统计信息
func (csm *CanalSyncManager) GetSyncStats() map[string]interface{} {
	return map[string]interface{}{
		"watch_tables":     csm.watchTables,
		"batch_size":       csm.batchSize,
		"consumer_running": true, // 简化实现，实际应该检查消费者状态
		"last_sync_time":   time.Now(),
	}
}