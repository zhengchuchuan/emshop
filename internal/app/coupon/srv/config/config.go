package config

import (
	"time"

	"emshop/internal/app/coupon/srv/pkg/cache"
	"emshop/internal/app/pkg/options"
)

// Config 优惠券服务配置
type Config struct {
	Server    *options.ServerOptions    `yaml:"server"`
	Log       *options.LogOptions       `yaml:"log"`
	Registry  *options.RegistryOptions  `yaml:"registry"`
	Telemetry *options.TelemetryOptions `yaml:"telemetry"`
	MySQL     *options.MySQLOptions     `yaml:"mysql"`
	Redis     *options.RedisOptions     `yaml:"redis"`
	RocketMQ  *options.RocketMQOptions  `yaml:"rocketmq"`
	DTM       *options.DtmOptions       `yaml:"dtm"`
	Ristretto *RistrettoOptions         `yaml:"ristretto"`
	Canal     *CanalOptions             `yaml:"canal"`
	Business  *BusinessOptions          `yaml:"business"`
}

// RistrettoOptions Ristretto缓存配置
type RistrettoOptions struct {
	NumCounters int64 `yaml:"num_counters"`
	MaxCost     int64 `yaml:"max_cost"`
	BufferItems int64 `yaml:"buffer_items"`
	Metrics     bool  `yaml:"metrics"`
}

// CanalOptions Canal配置
type CanalOptions struct {
	ConsumerGroup string   `yaml:"consumer_group"`
	Topic         string   `yaml:"topic"`
	WatchTables   []string `yaml:"watch_tables"`
	BatchSize     int32    `yaml:"batch_size"`
}

// BusinessOptions 业务配置
type BusinessOptions struct {
	FlashSale *FlashSaleOptions `yaml:"flashsale"`
	Coupon    *CouponOptions    `yaml:"coupon"`
	Cache     *CacheOptions     `yaml:"cache"`
}

// FlashSaleOptions 秒杀配置
type FlashSaleOptions struct {
	MaxQpsPerUser  int           `yaml:"max_qps_per_user"`
	StockCacheTTL  time.Duration `yaml:"stock_cache_ttl"`
	UserLimitTTL   time.Duration `yaml:"user_limit_ttl"`
	BatchSize      int           `yaml:"batch_size"`
}

// CouponOptions 优惠券配置
type CouponOptions struct {
	MaxStackCount int           `yaml:"max_stack_count"`
	LockTTL       time.Duration `yaml:"lock_ttl"`
	CalcTimeout   time.Duration `yaml:"calc_timeout"`
}

// CacheOptions 缓存配置
type CacheOptions struct {
	L1TTL        time.Duration `yaml:"l1_ttl"`
	L2TTL        time.Duration `yaml:"l2_ttl"`
	WarmupCount  int           `yaml:"warmup_count"`
	EnableWarmup bool          `yaml:"enable_warmup"`
}

// ToCacheConfig 转换为缓存配置
func (c *Config) ToCacheConfig() *cache.CacheConfig {
	return &cache.CacheConfig{
		RistrettoConfig: cache.RistrettoConfig{
			NumCounters: c.Ristretto.NumCounters,
			MaxCost:     c.Ristretto.MaxCost,
			BufferItems: c.Ristretto.BufferItems,
			Metrics:     c.Ristretto.Metrics,
		},
		L1TTL:        c.Business.Cache.L1TTL,
		L2TTL:        c.Business.Cache.L2TTL,
		WarmupCount:  c.Business.Cache.WarmupCount,
		EnableWarmup: c.Business.Cache.EnableWarmup,
	}
}