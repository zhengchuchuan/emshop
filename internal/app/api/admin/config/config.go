package config

import (
	"emshop/internal/app/pkg/options"
	cliflag "emshop/pkg/common/cli/flag"
	"emshop/pkg/log"
)

type Config struct {
	Log *log.Options `json:"log" mapstructure:"log"`

	Server    *options.ServerOptions    `json:"server" mapstructure:"server"`
	Registry  *options.RegistryOptions  `json:"registry" mapstructure:"registry"`
	MySQL     *options.MySQLOptions     `json:"mysql" mapstructure:"mysql"`
	Jwt       *options.JwtOptions       `json:"jwt" mapstructure:"jwt"`
	Telemetry *options.TelemetryOptions `json:"telemetry" mapstructure:"telemetry"`
}

func (c *Config) Validate() []error {
	var errors []error
	errors = append(errors, c.Log.Validate()...)
	errors = append(errors, c.Server.Validate()...)
	errors = append(errors, c.Registry.Validate()...)
	errors = append(errors, c.MySQL.Validate()...)
	errors = append(errors, c.Jwt.Validate()...)
	errors = append(errors, c.Telemetry.Validate()...)
	return errors
}

func (c *Config) Flags() (fss cliflag.NamedFlagSets) {
	c.Log.AddFlags(fss.FlagSet("logs"))
	c.Server.AddFlags(fss.FlagSet("server"))
	c.Registry.AddFlags(fss.FlagSet("registry"))
	c.MySQL.AddFlags(fss.FlagSet("mysql"))
	c.Jwt.AddFlags(fss.FlagSet("jwt"))
	c.Telemetry.AddFlags(fss.FlagSet("telemetry"))
	return fss
}

func New() *Config {
	//配置默认初始化
	return &Config{
		Log:       log.NewOptions(),
		Server:    options.NewServerOptions(),
		Registry:  options.NewRegistryOptions(),
		MySQL:     options.NewMySQLOptions(),
		Jwt:       options.NewJwtOptions(),
		Telemetry: options.NewTelemetryOptions(),
	}
}
