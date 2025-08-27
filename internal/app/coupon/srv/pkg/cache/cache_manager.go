package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/go-redis/redis/v8"
	"emshop/pkg/log"
)

// CouponRepository 优惠券数据仓库接口 (简化版，用于缓存层)
type CouponRepository interface {
	GetCouponTemplate(ctx context.Context, couponID int64) (*CouponTemplate, error)
	GetHotCouponTemplates(ctx context.Context, limit int) ([]*CouponTemplate, error)
	GetUserCoupon(ctx context.Context, userCouponID int64) (*UserCoupon, error)
	GetFlashSaleActivity(ctx context.Context, activityID int64) (*FlashSaleActivity, error)
}

// CouponTemplate 优惠券模板
type CouponTemplate struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Type          int32     `json:"type"`
	DiscountType  int32     `json:"discount_type"`
	DiscountValue float64   `json:"discount_value"`
	MinAmount     float64   `json:"min_amount"`
	TotalCount    int32     `json:"total_count"`
	UsedCount     int32     `json:"used_count"`
	ValidStart    time.Time `json:"valid_start"`
	ValidEnd      time.Time `json:"valid_end"`
	Status        int32     `json:"status"`
}

// UserCoupon 用户优惠券
type UserCoupon struct {
	ID             int64     `json:"id"`
	CouponID       int64     `json:"coupon_id"`
	UserID         int64     `json:"user_id"`
	CouponSn       string    `json:"coupon_sn"`
	Status         int32     `json:"status"`
	ObtainTime     time.Time `json:"obtain_time"`
	ValidStartTime time.Time `json:"valid_start_time"`
	ValidEndTime   time.Time `json:"valid_end_time"`
}

// FlashSaleActivity 秒杀活动
type FlashSaleActivity struct {
	ID           int64     `json:"id"`
	CouponID     int64     `json:"coupon_id"`
	Name         string    `json:"name"`
	TotalCount   int32     `json:"total_count"`
	SuccessCount int32     `json:"success_count"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Status       int32     `json:"status"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	RistrettoConfig RistrettoConfig `yaml:"ristretto"`
	L1TTL           time.Duration   `yaml:"l1_ttl"`
	L2TTL           time.Duration   `yaml:"l2_ttl"`
	WarmupCount     int             `yaml:"warmup_count"`
	EnableWarmup    bool            `yaml:"enable_warmup"`
}

// RistrettoConfig Ristretto缓存配置
type RistrettoConfig struct {
	NumCounters int64 `yaml:"num_counters"`
	MaxCost     int64 `yaml:"max_cost"`
	BufferItems int64 `yaml:"buffer_items"`
	Metrics     bool  `yaml:"metrics"`
}

// CouponCacheManager 三层缓存管理器
type CouponCacheManager struct {
	// L1: Ristretto本地缓存 (1ms响应，95%命中率)
	localCache *ristretto.Cache
	// L2: Redis集群缓存 (5ms响应，90%命中率)  
	redis *redis.Client
	// L3: MySQL数据库 (20ms响应，100%命中率)
	repository CouponRepository
	// 缓存配置
	config *CacheConfig
}

// NewCouponCacheManager 创建三层缓存管理器
func NewCouponCacheManager(redis *redis.Client, repo CouponRepository, config *CacheConfig) (*CouponCacheManager, error) {
	// 初始化Ristretto缓存
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: config.RistrettoConfig.NumCounters,   // 1M个key的统计信息
		MaxCost:     config.RistrettoConfig.MaxCost,       // 100MB最大内存
		BufferItems: config.RistrettoConfig.BufferItems,   // 缓冲区大小
		Metrics:     config.RistrettoConfig.Metrics,       // 开启监控指标
	})
	if err != nil {
		return nil, fmt.Errorf("初始化Ristretto缓存失败: %v", err)
	}

	ccm := &CouponCacheManager{
		localCache: cache,
		redis:      redis,
		repository: repo,
		config:     config,
	}

	// 启动缓存预热
	if config.EnableWarmup {
		go func() {
			if err := ccm.WarmupCache(context.Background()); err != nil {
				log.Errorf("缓存预热失败: %v", err)
			}
		}()
	}

	log.Info("三层缓存管理器初始化成功")
	return ccm, nil
}

// GetCouponTemplate 获取优惠券模板（三层缓存查询）
func (ccm *CouponCacheManager) GetCouponTemplate(ctx context.Context, couponID int64) (*CouponTemplate, error) {
	key := fmt.Sprintf("coupon:template:%d", couponID)

	// L1: Ristretto本地缓存查询
	if value, found := ccm.localCache.Get(key); found {
		template := value.(*CouponTemplate)
		log.Debugf("命中L1缓存, couponID: %d", couponID)
		return template, nil
	}

	// L2: Redis缓存查询
	if data := ccm.redis.Get(ctx, key).Val(); data != "" {
		var template CouponTemplate
		if err := json.Unmarshal([]byte(data), &template); err == nil {
			// 回写L1缓存 (成本为1，TTL从配置读取)
			ccm.localCache.SetWithTTL(key, &template, 1, ccm.config.L1TTL)
			log.Debugf("命中L2缓存, couponID: %d", couponID)
			return &template, nil
		}
	}

	// L3: 数据库查询
	template, err := ccm.repository.GetCouponTemplate(ctx, couponID)
	if err != nil {
		return nil, fmt.Errorf("查询优惠券模板失败: %v", err)
	}

	// 回写L2缓存
	data, _ := json.Marshal(template)
	ccm.redis.SetEX(ctx, key, data, ccm.config.L2TTL)

	// 回写L1缓存
	ccm.localCache.SetWithTTL(key, template, 1, ccm.config.L1TTL)

	log.Debugf("命中L3数据库, couponID: %d", couponID)
	return template, nil
}

// GetUserCoupon 获取用户优惠券（三层缓存查询）
func (ccm *CouponCacheManager) GetUserCoupon(ctx context.Context, userCouponID int64) (*UserCoupon, error) {
	key := fmt.Sprintf("coupon:user:%d", userCouponID)

	// L1: 本地缓存查询
	if value, found := ccm.localCache.Get(key); found {
		userCoupon := value.(*UserCoupon)
		log.Debugf("命中L1缓存, userCouponID: %d", userCouponID)
		return userCoupon, nil
	}

	// L2: Redis缓存查询
	if data := ccm.redis.Get(ctx, key).Val(); data != "" {
		var userCoupon UserCoupon
		if err := json.Unmarshal([]byte(data), &userCoupon); err == nil {
			ccm.localCache.SetWithTTL(key, &userCoupon, 1, ccm.config.L1TTL)
			log.Debugf("命中L2缓存, userCouponID: %d", userCouponID)
			return &userCoupon, nil
		}
	}

	// L3: 数据库查询
	userCoupon, err := ccm.repository.GetUserCoupon(ctx, userCouponID)
	if err != nil {
		return nil, fmt.Errorf("查询用户优惠券失败: %v", err)
	}

	// 回写缓存
	data, _ := json.Marshal(userCoupon)
	ccm.redis.SetEX(ctx, key, data, ccm.config.L2TTL)
	ccm.localCache.SetWithTTL(key, userCoupon, 1, ccm.config.L1TTL)

	log.Debugf("命中L3数据库, userCouponID: %d", userCouponID)
	return userCoupon, nil
}

// GetFlashSaleActivity 获取秒杀活动（三层缓存查询）
func (ccm *CouponCacheManager) GetFlashSaleActivity(ctx context.Context, activityID int64) (*FlashSaleActivity, error) {
	key := fmt.Sprintf("flashsale:activity:%d", activityID)

	// L1: 本地缓存查询
	if value, found := ccm.localCache.Get(key); found {
		activity := value.(*FlashSaleActivity)
		log.Debugf("命中L1缓存, activityID: %d", activityID)
		return activity, nil
	}

	// L2: Redis缓存查询
	if data := ccm.redis.Get(ctx, key).Val(); data != "" {
		var activity FlashSaleActivity
		if err := json.Unmarshal([]byte(data), &activity); err == nil {
			ccm.localCache.SetWithTTL(key, &activity, 1, ccm.config.L1TTL)
			log.Debugf("命中L2缓存, activityID: %d", activityID)
			return &activity, nil
		}
	}

	// L3: 数据库查询
	activity, err := ccm.repository.GetFlashSaleActivity(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("查询秒杀活动失败: %v", err)
	}

	// 回写缓存
	data, _ := json.Marshal(activity)
	ccm.redis.SetEX(ctx, key, data, ccm.config.L2TTL)
	ccm.localCache.SetWithTTL(key, activity, 1, ccm.config.L1TTL)

	log.Debugf("命中L3数据库, activityID: %d", activityID)
	return activity, nil
}

// InvalidateCache 缓存失效 (Canal调用)
func (ccm *CouponCacheManager) InvalidateCache(keys ...string) {
	for _, key := range keys {
		// 删除L1缓存
		ccm.localCache.Del(key)
		// 删除L2缓存  
		ccm.redis.Del(context.Background(), key)
		log.Infof("缓存失效: %s", key)
	}
}

// InvalidateCacheByPattern 根据模式失效缓存
func (ccm *CouponCacheManager) InvalidateCacheByPattern(ctx context.Context, pattern string) error {
	// 获取匹配的Redis键
	keys, err := ccm.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("获取匹配键失败: %v", err)
	}

	if len(keys) > 0 {
		// 批量删除Redis缓存
		ccm.redis.Del(ctx, keys...)
		log.Infof("批量失效Redis缓存: %d个键, pattern: %s", len(keys), pattern)
	}

	// 本地缓存只能逐个删除，这里简化处理
	// 在实际场景中，可以考虑增加本地缓存的标记机制
	log.Infof("模式失效缓存: %s", pattern)
	return nil
}

// WarmupCache 缓存预热
func (ccm *CouponCacheManager) WarmupCache(ctx context.Context) error {
	log.Info("开始缓存预热...")

	// 查询热门优惠券模板
	hotCoupons, err := ccm.repository.GetHotCouponTemplates(ctx, ccm.config.WarmupCount)
	if err != nil {
		return fmt.Errorf("获取热门优惠券失败: %v", err)
	}

	// 批量预热到L1和L2缓存
	for _, coupon := range hotCoupons {
		key := fmt.Sprintf("coupon:template:%d", coupon.ID)

		// 写入L2 Redis缓存
		data, _ := json.Marshal(coupon)
		ccm.redis.SetEX(ctx, key, data, ccm.config.L2TTL)

		// 写入L1 Ristretto缓存 (高成本保证不被淘汰)
		ccm.localCache.SetWithTTL(key, coupon, 10, ccm.config.L1TTL)
	}

	log.Infof("缓存预热完成，预热%d个优惠券模板", len(hotCoupons))
	return nil
}

// GetCacheStats 获取缓存统计信息
func (ccm *CouponCacheManager) GetCacheStats() map[string]interface{} {
	metrics := ccm.localCache.Metrics

	return map[string]interface{}{
		"ristretto_hits":         metrics.Hits(),
		"ristretto_misses":       metrics.Misses(),
		"ristretto_hit_ratio":    metrics.Ratio(),
		"ristretto_keys_added":   metrics.KeysAdded(),
		"ristretto_keys_evicted": metrics.KeysEvicted(),
		"ristretto_cost_added":   metrics.CostAdded(),
		"ristretto_cost_evicted": metrics.CostEvicted(),
	}
}

// Close 关闭缓存管理器
func (ccm *CouponCacheManager) Close() {
	if ccm.localCache != nil {
		ccm.localCache.Close()
	}
	log.Info("缓存管理器已关闭")
}

// ConvertCacheConfig 转换配置选项为缓存配置 
func ConvertCacheConfig(opts *CacheOptions) *CacheConfig {
	if opts == nil {
		return &CacheConfig{
			RistrettoConfig: RistrettoConfig{
				NumCounters: 1000000,   // 1M个key的统计信息
				MaxCost:     104857600, // 100MB最大内存
				BufferItems: 64,        // 缓冲区大小
				Metrics:     true,      // 开启监控指标
			},
			L1TTL:        10 * time.Minute, // L1缓存TTL
			L2TTL:        30 * time.Minute, // L2缓存TTL
			WarmupCount:  100,              // 预热优惠券数量
			EnableWarmup: true,             // 是否开启预热
		}
	}
	
	return &CacheConfig{
		RistrettoConfig: RistrettoConfig{
			NumCounters: opts.Ristretto.NumCounters,
			MaxCost:     opts.Ristretto.MaxCost,
			BufferItems: opts.Ristretto.BufferItems,
			Metrics:     opts.Ristretto.Metrics,
		},
		L1TTL:        opts.L1TTL,
		L2TTL:        opts.L2TTL,
		WarmupCount:  opts.WarmupCount,
		EnableWarmup: opts.EnableWarmup,
	}
}

// CacheOptions 引入配置选项的类型别名，避免循环导入
type CacheOptions struct {
	Ristretto   RistrettoOptions
	L1TTL       time.Duration
	L2TTL       time.Duration
	WarmupCount int
	EnableWarmup bool
}

type RistrettoOptions struct {
	NumCounters int64
	MaxCost     int64
	BufferItems int64
	Metrics     bool
}

// NewIntegratedCacheManagerWithCanal 创建带Canal同步的集成缓存管理器
func NewIntegratedCacheManagerWithCanal(
	redisClient *redis.Client, 
	repository CouponRepository, 
	cacheConfig *CacheConfig,
	canalConfig *CanalSyncConfig,
) (CacheWithSyncManager, error) {
	// 创建基础缓存管理器
	cacheManager, err := NewCouponCacheManager(redisClient, repository, cacheConfig)
	if err != nil {
		return nil, fmt.Errorf("创建缓存管理器失败: %v", err)
	}
	
	// 如果没有Canal配置，返回不带同步的缓存管理器
	if canalConfig == nil {
		log.Info("未配置Canal同步，使用普通缓存管理器")
		return &IntegratedCacheManager{
			CacheManager: cacheManager,
			canalSync:    nil,
		}, nil
	}
	
	// 创建Canal同步管理器
	canalSync, err := NewCanalSyncManager(cacheManager, canalConfig)
	if err != nil {
		log.Warnf("创建Canal同步管理器失败，降级为普通缓存: %v", err)
		// Canal同步失败不影响缓存管理器的正常工作
		return &IntegratedCacheManager{
			CacheManager: cacheManager,
			canalSync:    nil,
		}, nil
	}
	
	log.Info("集成缓存管理器（含Canal同步）创建成功")
	return NewIntegratedCacheManager(cacheManager, canalSync), nil
}