package config

import (
	"encoding/json"
	"fmt"
	"time"

	"emshop/internal/app/coupon/srv/pkg/cache"
	"emshop/internal/app/pkg/options"
	"emshop/pkg/log"

	cliflag "emshop/pkg/common/cli/flag"
)

// Config 优惠券服务配置
type Config struct {
	Server    *options.ServerOptions    `yaml:"server"`
	Log       *log.Options              `yaml:"log"`
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

// New 创建具有合理默认值的配置对象，保证YAML加载前的字段完整性
func New() *Config {
	return &Config{
		Server:   options.NewServerOptions(),
		Log:      log.NewOptions(),
		Registry: options.NewRegistryOptions(),
		Telemetry: func() *options.TelemetryOptions {
			opt := options.NewTelemetryOptions()
			opt.Name = "coupon"
			return opt
		}(),
		MySQL: options.NewMySQLOptions(),
		Redis: options.NewRedisOptions(),
		RocketMQ: func() *options.RocketMQOptions {
			opt := options.NewRocketMQOptions()
			opt.Topic = "coupon-events"
			opt.ConsumerGroup = "coupon-consumer-group"
			return opt
		}(),
		DTM: options.NewDtmOptions(),
		Ristretto: &RistrettoOptions{
			NumCounters: 1_000_000,
			MaxCost:     100 << 20,
			BufferItems: 64,
			Metrics:     true,
		},
		Canal: &CanalOptions{
			ConsumerGroup: "coupon-cache-sync-consumer",
			Topic:         "coupon-binlog-topic",
			WatchTables:   []string{"coupon_templates", "user_coupons", "flash_sale_activities"},
			BatchSize:     32,
		},
		Business: &BusinessOptions{
			FlashSale: &FlashSaleOptions{
				MaxQpsPerUser: 5,
				StockCacheTTL: 300 * time.Second,
				UserLimitTTL:  1800 * time.Second,
				BatchSize:     100,
			},
			Coupon: &CouponOptions{
				MaxStackCount: 5,
				LockTTL:       900 * time.Second,
				CalcTimeout:   5 * time.Second,
			},
			Cache: &CacheOptions{
				L1TTL:        10 * time.Minute,
				L2TTL:        30 * time.Minute,
				WarmupCount:  100,
				EnableWarmup: false,
			},
		},
	}
}

// Flags returns grouped command-line flags for the coupon service configuration.
func (c *Config) Flags() (fss cliflag.NamedFlagSets) {
	if c.Server != nil {
		c.Server.AddFlags(fss.FlagSet("server"))
	}
	if c.Log != nil {
		c.Log.AddFlags(fss.FlagSet("logs"))
	}
	if c.Registry != nil {
		c.Registry.AddFlags(fss.FlagSet("registry"))
	}
	if c.Telemetry != nil {
		c.Telemetry.AddFlags(fss.FlagSet("telemetry"))
	}
	if c.MySQL != nil {
		c.MySQL.AddFlags(fss.FlagSet("mysql"))
	}
	if c.Redis != nil {
		c.Redis.AddFlags(fss.FlagSet("redis"))
	}
	if c.RocketMQ != nil {
		c.RocketMQ.AddFlags(fss.FlagSet("rocketmq"))
	}
	if c.DTM != nil {
		c.DTM.AddFlags(fss.FlagSet("dtm"))
	}
	return fss
}

// Validate performs basic validation across option groups.
func (c *Config) Validate() []error {
	var errs []error

	if c.Server == nil {
		errs = append(errs, fmt.Errorf("server configuration is required"))
	} else {
		errs = append(errs, c.Server.Validate()...)
	}
	if c.Log == nil {
		errs = append(errs, fmt.Errorf("log configuration is required"))
	} else {
		errs = append(errs, c.Log.Validate()...)
	}
	if c.Registry == nil {
		errs = append(errs, fmt.Errorf("registry configuration is required"))
	} else {
		errs = append(errs, c.Registry.Validate()...)
	}
	if c.Telemetry == nil {
		errs = append(errs, fmt.Errorf("telemetry configuration is required"))
	} else {
		errs = append(errs, c.Telemetry.Validate()...)
	}
	if c.MySQL == nil {
		errs = append(errs, fmt.Errorf("mysql configuration is required"))
	} else {
		errs = append(errs, c.MySQL.Validate()...)
	}
	if c.Redis == nil {
		errs = append(errs, fmt.Errorf("redis configuration is required"))
	} else {
		errs = append(errs, c.Redis.Validate()...)
	}
	if c.RocketMQ == nil {
		errs = append(errs, fmt.Errorf("rocketmq configuration is required"))
	} else {
		errs = append(errs, c.RocketMQ.Validate()...)
	}
	if c.DTM == nil {
		errs = append(errs, fmt.Errorf("dtm configuration is required"))
	} else {
		errs = append(errs, c.DTM.Validate()...)
	}
	if c.Canal == nil {
		errs = append(errs, fmt.Errorf("canal configuration is required"))
	}
	if c.Business == nil {
		errs = append(errs, fmt.Errorf("business configuration is required"))
	}

	return errs
}

// String returns the json representation of the config for logging purposes.
func (c *Config) String() string {
	data, _ := json.Marshal(c)
	return string(data)
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
	MaxQpsPerUser int           `yaml:"max_qps_per_user"`
	StockCacheTTL time.Duration `yaml:"stock_cache_ttl"`
	UserLimitTTL  time.Duration `yaml:"user_limit_ttl"`
	BatchSize     int           `yaml:"batch_size"`
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
	if c.Ristretto == nil {
		c.Ristretto = &RistrettoOptions{
			NumCounters: 1_000_000,
			MaxCost:     100 << 20,
			BufferItems: 64,
			Metrics:     true,
		}
	}

	cacheOpts := c.Business
	if cacheOpts == nil {
		cacheOpts = &BusinessOptions{Cache: &CacheOptions{L1TTL: 10 * time.Minute, L2TTL: 30 * time.Minute, WarmupCount: 100, EnableWarmup: false}}
	} else if cacheOpts.Cache == nil {
		cacheOpts.Cache = &CacheOptions{L1TTL: 10 * time.Minute, L2TTL: 30 * time.Minute, WarmupCount: 100, EnableWarmup: false}
	}

	return &cache.CacheConfig{
		RistrettoConfig: cache.RistrettoConfig{
			NumCounters: c.Ristretto.NumCounters,
			MaxCost:     c.Ristretto.MaxCost,
			BufferItems: c.Ristretto.BufferItems,
			Metrics:     c.Ristretto.Metrics,
		},
		L1TTL:        cacheOpts.Cache.L1TTL,
		L2TTL:        cacheOpts.Cache.L2TTL,
		WarmupCount:  cacheOpts.Cache.WarmupCount,
		EnableWarmup: cacheOpts.Cache.EnableWarmup,
	}
}
