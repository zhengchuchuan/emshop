package config

import (
	"emshop/internal/app/pkg/options"
	cliflag "emshop/pkg/common/cli/flag"
	"emshop/pkg/log"
)

// Config 用户服务的主配置结构体
// 这是CLI配置系统的顶层聚合器，展示了以下设计模式：
// 1. 组合模式：将各个子配置模块组合成完整配置
// 2. 接口统一：所有子配置都实现AddFlags/Validate方法
// 3. 分组管理：通过NamedFlagSets实现标志分组显示
type Config struct {
	// 日志配置选项
	Log *log.Options `json:"log" mapstructure:"log"`

	// 服务器配置选项（gRPC、HTTP端口等）
	Server *options.ServerOptions `json:"server" mapstructure:"server"`
	// 服务注册发现配置（Consul等）
	Registry *options.RegistryOptions `json:"registry" mapstructure:"registry"`

	// 链路追踪配置（Jaeger等）
	Telemetry *options.TelemetryOptions `json:"telemetry" mapstructure:"telemetry"`
	// MySQL数据库配置
	MySQLOptions *options.MySQLOptions `json:"mysql" mapstructure:"mysql"`
	// Nacos配置中心选项
	Nacos *options.NacosOptions `json:"nacos" mapstructure:"nacos"`
}

// Validate 验证所有配置选项的有效性
// 聚合所有子配置的验证结果，确保整体配置的正确性
func (c *Config) Validate() []error {
	var errors []error
	// 调用各个子配置的Validate方法，收集所有验证错误
	errors = append(errors, c.Log.Validate()...)
	errors = append(errors, c.Server.Validate()...)
	errors = append(errors, c.Registry.Validate()...)
	errors = append(errors, c.Telemetry.Validate()...)
	errors = append(errors, c.MySQLOptions.Validate()...)
	errors = append(errors, c.Nacos.Validate()...)
	return errors
}

// Flags 生成分组的命令行标志集合
// 这是pflag与CLI系统集成的关键方法，展示了以下特性：
// 1. 分组管理：将相关标志归类到不同组（logs、server、registry等）
// 2. 统一接口：所有子配置都通过AddFlags方法添加标志
// 3. 帮助显示：分组后的标志在--help中按组显示，提升用户体验
func (c *Config) Flags() (fss cliflag.NamedFlagSets) {
	// 为每个功能模块创建独立的标志组
	// fss.FlagSet(name) 创建或获取指定名称的pflag.FlagSet
	c.Log.AddFlags(fss.FlagSet("logs"))           // 日志相关标志组
	c.Server.AddFlags(fss.FlagSet("server"))      // 服务器相关标志组  
	c.Registry.AddFlags(fss.FlagSet("registry"))  // 注册中心相关标志组
	c.Telemetry.AddFlags(fss.FlagSet("telemetry")) // 链路追踪相关标志组
	c.MySQLOptions.AddFlags(fss.FlagSet("mysql"))  // MySQL相关标志组
	c.Nacos.AddFlags(fss.FlagSet("nacos"))        // Nacos相关标志组
	return fss
}

// New 创建具有默认值的Config实例
// 采用构造函数模式，确保所有子配置都被正确初始化
func New() *Config {
	// 使用各个子配置的NewXXXOptions构造函数，获得合理的默认值
	return &Config{
		Log:          log.NewOptions(),
		Server:       options.NewServerOptions(),
		Registry:     options.NewRegistryOptions(),
		Telemetry:    options.NewTelemetryOptions(),
		MySQLOptions: options.NewMySQLOptions(),
		Nacos:        options.NewNacosOptions(),
	}
}
