package options

import "github.com/spf13/pflag"

// ServerOptions 服务器配置选项结构体
// 这个结构体展示了pflag的核心使用模式：
// 1. 通过结构体字段定义配置项
// 2. 使用mapstructure标签支持配置文件映射
// 3. 通过AddFlags方法定义命令行参数
type ServerOptions struct {
	// 是否开启pprof性能分析
	// mapstructure标签用于Viper配置文件到结构体的映射
	EnableProfiling bool `json:"profiling" mapstructure:"profiling"`
	// 是否开启限流
	EnableLimit bool `json:"limit" mapstructure:"limit"`
	// 是否开启metrics监控
	EnableMetrics bool `json:"enable-metrics" mapstructure:"enable-metrics"`
	// 是否开启健康检查
	EnableHealthCheck bool `json:"enable-health-check" mapstructure:"enable-health-check"`
	// 服务器主机地址
	Host string `json:"host,omitempty" mapstructure:"host"`
	// gRPC服务端口
	Port int `json:"port,omitempty" mapstructure:"port"`
	// HTTP服务端口
	HttpPort int `json:"http-port,omitempty" mapstructure:"http-port"`
	// 服务名称
	Name string `json:"name,omitempty" mapstructure:"name"`
	// 中间件列表
	Middlewares []string `json:"middlewares,omitempty" mapstructure:"middlewares"`
}

// NewServerOptions 创建带默认值的ServerOptions实例
// 这是Options模式的标准实现：提供合理的默认值
func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		EnableHealthCheck: true,
		EnableProfiling:   true,
		EnableMetrics:     true,
		Host:              "127.0.0.1",
		Port:              8078,
		HttpPort:          8079,
		Name:              "emshop-user-srv",
	}
}

// Validate 验证配置选项的有效性
// 在配置绑定后调用，确保配置值符合要求
func (so *ServerOptions) Validate() []error {
	errs := []error{}
	// TODO: 添加具体的验证逻辑
	return errs
}

// AddFlags 将服务器相关的标志添加到pflag.FlagSet
// 这是pflag使用的核心模式：将结构体字段绑定到命令行标志
func (so *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	// BoolVar: 绑定布尔类型字段到命令行标志
	// 参数：(字段指针, 标志名称, 默认值, 帮助信息)
	fs.BoolVar(&so.EnableProfiling, "server.enable-profiling", so.EnableProfiling,
		"enable-profiling, if true, will add <host>:<port>/debug/pprof/, default is true")
	fs.BoolVar(&so.EnableMetrics, "server.enable-metrics", so.EnableMetrics,
		"enable-metrics, if true, will add /metrics, default is true")
	fs.BoolVar(&so.EnableHealthCheck, "server.enable-health-check", so.EnableHealthCheck,
		"enable-health-check, if true, will add health check route, default is true")

	// StringVar: 绑定字符串类型字段到命令行标志
	fs.StringVar(&so.Host, "server.host", so.Host, "server host default is 127.0.0.1")
	fs.StringVar(&so.Name, "server.name", so.Name, "server name default is emshop-user-srv")

	// IntVar: 绑定整数类型字段到命令行标志
	fs.IntVar(&so.Port, "server.port", so.Port, "server port default is 8078")
	fs.IntVar(&so.HttpPort, "server.http-port", so.HttpPort, "server http port default is 8079")
}
