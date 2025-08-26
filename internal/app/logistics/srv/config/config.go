package config

import (
	"emshop/internal/app/pkg/options"
	cliflag "emshop/pkg/common/cli/flag"
	"emshop/pkg/log"
)

// Config 物流服务配置结构
type Config struct {
	// 日志配置选项
	Log *log.Options `json:"log" mapstructure:"log"`
	// 服务器配置选项
	Server *options.ServerOptions `json:"server" mapstructure:"server"`
	// MySQL数据库配置
	MySQLOptions *options.MySQLOptions `json:"mysql" mapstructure:"mysql"`
	// Redis缓存配置
	Redis *options.RedisOptions `json:"redis" mapstructure:"redis"`
	// 服务注册发现配置
	Registry *options.RegistryOptions `json:"registry" mapstructure:"registry"`
	// 链路追踪配置
	Telemetry *options.TelemetryOptions `json:"telemetry" mapstructure:"telemetry"`
}

// Validate 验证所有配置选项的有效性
func (c *Config) Validate() []error {
	var errors []error
	errors = append(errors, c.Log.Validate()...)
	errors = append(errors, c.Server.Validate()...)
	errors = append(errors, c.MySQLOptions.Validate()...)
	errors = append(errors, c.Redis.Validate()...)
	errors = append(errors, c.Registry.Validate()...)
	errors = append(errors, c.Telemetry.Validate()...)
	return errors
}

// Flags 生成分组的命令行标志集合
func (c *Config) Flags() (fss cliflag.NamedFlagSets) {
	c.Log.AddFlags(fss.FlagSet("logs"))
	c.Server.AddFlags(fss.FlagSet("server"))
	c.MySQLOptions.AddFlags(fss.FlagSet("mysql"))
	c.Redis.AddFlags(fss.FlagSet("redis"))
	c.Registry.AddFlags(fss.FlagSet("registry"))
	c.Telemetry.AddFlags(fss.FlagSet("telemetry"))
	return fss
}

// New 创建默认配置
func New() *Config {
	return &Config{
		Log:          log.NewOptions(),
		Server:       options.NewServerOptions(),
		MySQLOptions: options.NewMySQLOptions(),
		Redis:        options.NewRedisOptions(),
		Registry:     options.NewRegistryOptions(),
		Telemetry:    options.NewTelemetryOptions(),
	}
}