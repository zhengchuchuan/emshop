package config

import (
	"emshop/internal/app/pkg/options"
	cliflag "emshop/pkg/common/cli/flag"
	"emshop/pkg/log"
)

// Config 用户操作服务的配置结构体
type Config struct {
	Log      *log.Options              `json:"log"      mapstructure:"log"`
	Server   *options.ServerOptions    `json:"server"   mapstructure:"server"`
	Registry *options.RegistryOptions  `json:"registry" mapstructure:"registry"`
	MySQL    *options.MySQLOptions     `json:"mysql"    mapstructure:"mysql"`
}

// Validate 验证配置
func (c *Config) Validate() []error {
	var errors []error
	errors = append(errors, c.Log.Validate()...)
	errors = append(errors, c.Server.Validate()...)
	errors = append(errors, c.Registry.Validate()...)
	errors = append(errors, c.MySQL.Validate()...)
	return errors
}

// Flags 生成命令行标志
func (c *Config) Flags() (fss cliflag.NamedFlagSets) {
	c.Log.AddFlags(fss.FlagSet("logs"))
	c.Server.AddFlags(fss.FlagSet("server"))
	c.Registry.AddFlags(fss.FlagSet("registry"))
	c.MySQL.AddFlags(fss.FlagSet("mysql"))
	return fss
}

// NewConfig 创建配置实例
func NewConfig() *Config {
	return &Config{
		Log:      log.NewOptions(),
		Server:   options.NewServerOptions(),
		Registry: options.NewRegistryOptions(),
		MySQL:    options.NewMySQLOptions(),
	}
}