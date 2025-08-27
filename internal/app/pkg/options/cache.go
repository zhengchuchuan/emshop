package options

import (
	"time"
	"github.com/spf13/pflag"
)

type RedisOptions struct {
	Host                  string   `mapstructure:"host" json:"host"`
	Port                  int      `mapstructure:"port" json:"port"`
	Addrs                 []string `mapstructure:"addrs" json:"addrs"`
	Username              string   `mapstructure:"username" json:"username"`
	Password              string   `mapstructure:"password" json:"password"`
	Database              int      `mapstructure:"database" json:"database"`
	MasterName            string   `mapstructure:"master-name" json:"master-name"`
	MaxIdle               int      `mapstructure:"optimisation-max-idle" json:"optimisation-max-idle"`
	MaxActive             int      `mapstructure:"optimisation-max-active" json:"optimisation-max-active"`
	Timeout               int      `mapstructure:"timeout" json:"timeout"`
	EnableCluster         bool     `mapstructure:"enable-cluster" json:"enable-cluster"`
	UseSSL                bool     `mapstructure:"use-ssl" json:"use-ssl"`
	SSLInsecureSkipVerify bool     `mapstructure:"ssl-insecure-skip-verify" json:"ssl-insecure-skip-verify"`
	EnableTracing         bool     `mapstructure:"enabletracing" json:"enabletracing"`
}

// NewRedisOptions create a `zero` value instance.
func NewRedisOptions() *RedisOptions {
	return &RedisOptions{
		Host:                  "127.0.0.1",
		Port:                  6379,
		Addrs:                 []string{},
		Username:              "",
		Password:              "",
		Database:              0,
		MasterName:            "",
		MaxIdle:               2000,
		MaxActive:             4000,
		Timeout:               0,
		EnableCluster:         false,
		UseSSL:                false,
		SSLInsecureSkipVerify: false,
	}
}

func (o *RedisOptions) Validate() []error {
	errs := []error{}

	return errs
}

// AddFlags adds flags related to redis storage for a specific APIServer to the specified FlagSet.
func (o *RedisOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Host, "redis.host", o.Host, "Hostname of your Redis server.")
	fs.IntVar(&o.Port, "redis.port", o.Port, "The port the Redis server is listening on.")
	fs.StringSliceVar(&o.Addrs, "redis.addrs", o.Addrs, "A set of redis address(format: 127.0.0.1:6379).")
	fs.StringVar(&o.Username, "redis.username", o.Username, "Username for access to redis service.")
	fs.StringVar(&o.Password, "redis.password", o.Password, "Optional auth password for Redis db.")

	fs.IntVar(&o.Database, "redis.database", o.Database, ""+
		"By default, the database is 0. Setting the database is not supported with redis cluster. "+
		"As such, if you have --redis.enable-cluster=true, then this value should be omitted or explicitly set to 0.")

	fs.StringVar(&o.MasterName, "redis.master-name", o.MasterName, "The name of master redis instance.")

	fs.IntVar(&o.MaxIdle, "redis.optimisation-max-idle", o.MaxIdle, ""+
		"This setting will configure how many connections are maintained in the pool when idle (no traffic). "+
		"Set the --redis.optimisation-max-active to something large, we usually leave it at around 2000 for "+
		"HA deployments.")

	fs.IntVar(&o.MaxActive, "redis.optimisation-max-active", o.MaxActive, ""+
		"In order to not over commit connections to the Redis server, we may limit the total "+
		"number of active connections to Redis. We recommend for production use to set this to around 4000.")

	fs.IntVar(&o.Timeout, "redis.timeout", o.Timeout, "Timeout (in seconds) when connecting to redis service.")

	fs.BoolVar(&o.EnableCluster, "redis.enable-cluster", o.EnableCluster, ""+
		"If you are using Redis cluster, enable it here to enable the slots mode.")

	fs.BoolVar(&o.UseSSL, "redis.use-ssl", o.UseSSL, ""+
		"If set, IAM will assume the connection to Redis is encrypted. "+
		"(use with Redis providers that support in-transit encryption).")

	fs.BoolVar(&o.SSLInsecureSkipVerify, "redis.ssl-insecure-skip-verify", o.SSLInsecureSkipVerify, ""+
		"Allows usage of self-signed certificates when connecting to an encrypted Redis database.")
}

// RistrettoOptions Ristretto缓存配置选项
type RistrettoOptions struct {
	NumCounters int64 `mapstructure:"num_counters" json:"num_counters"`
	MaxCost     int64 `mapstructure:"max_cost" json:"max_cost"`
	BufferItems int64 `mapstructure:"buffer_items" json:"buffer_items"`
	Metrics     bool  `mapstructure:"metrics" json:"metrics"`
}

// CacheOptions 缓存配置选项
type CacheOptions struct {
	Ristretto   RistrettoOptions `mapstructure:"ristretto" json:"ristretto"`
	L1TTL       time.Duration    `mapstructure:"l1_ttl" json:"l1_ttl"`
	L2TTL       time.Duration    `mapstructure:"l2_ttl" json:"l2_ttl"`
	WarmupCount int              `mapstructure:"warmup_count" json:"warmup_count"`
	EnableWarmup bool            `mapstructure:"enable_warmup" json:"enable_warmup"`
}

// CanalOptions Canal配置选项
type CanalOptions struct {
	ConsumerGroup string   `mapstructure:"consumer_group" json:"consumer_group"`
	Topic         string   `mapstructure:"topic" json:"topic"`
	WatchTables   []string `mapstructure:"watch_tables" json:"watch_tables"`
	BatchSize     int      `mapstructure:"batch_size" json:"batch_size"`
}

// BusinessOptions 业务配置选项
type BusinessOptions struct {
	FlashSale FlashSaleBusinessOptions `mapstructure:"flashsale" json:"flashsale"`
	Coupon    CouponBusinessOptions    `mapstructure:"coupon" json:"coupon"`
	Cache     CacheOptions             `mapstructure:"cache" json:"cache"`
}

// FlashSaleBusinessOptions 秒杀业务配置
type FlashSaleBusinessOptions struct {
	MaxQPSPerUser  int           `mapstructure:"max_qps_per_user" json:"max_qps_per_user"`
	StockCacheTTL  time.Duration `mapstructure:"stock_cache_ttl" json:"stock_cache_ttl"`
	UserLimitTTL   time.Duration `mapstructure:"user_limit_ttl" json:"user_limit_ttl"`
	BatchSize      int           `mapstructure:"batch_size" json:"batch_size"`
}

// CouponBusinessOptions 优惠券业务配置
type CouponBusinessOptions struct {
	MaxStackCount int           `mapstructure:"max_stack_count" json:"max_stack_count"`
	LockTTL       time.Duration `mapstructure:"lock_ttl" json:"lock_ttl"`
	CalcTimeout   time.Duration `mapstructure:"calc_timeout" json:"calc_timeout"`
}

// NewCacheOptions 创建缓存配置
func NewCacheOptions() *CacheOptions {
	return &CacheOptions{
		Ristretto: RistrettoOptions{
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

// NewBusinessOptions 创建业务配置
func NewBusinessOptions() *BusinessOptions {
	return &BusinessOptions{
		FlashSale: FlashSaleBusinessOptions{
			MaxQPSPerUser:  5,
			StockCacheTTL:  300 * time.Second,
			UserLimitTTL:   1800 * time.Second,
			BatchSize:      100,
		},
		Coupon: CouponBusinessOptions{
			MaxStackCount: 5,
			LockTTL:       900 * time.Second,
			CalcTimeout:   5 * time.Second,
		},
		Cache: *NewCacheOptions(),
	}
}
