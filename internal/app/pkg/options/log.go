package options

import "github.com/spf13/pflag"

// LogOptions 日志配置选项结构体
type LogOptions struct {
	// 日志级别 (debug, info, warn, error)
	Level string `json:"level" mapstructure:"level"`
	// 日志格式 (json, console)
	Format string `json:"format" mapstructure:"format"`
	// 是否输出颜色
	Color bool `json:"color" mapstructure:"color"`
	// 日志文件路径
	LogFile string `json:"log-file,omitempty" mapstructure:"log-file"`
	// 日志文件最大大小 (MB)
	MaxSize int `json:"max-size" mapstructure:"max-size"`
	// 日志文件最大备份数
	MaxBackups int `json:"max-backups" mapstructure:"max-backups"`
	// 日志文件最大保留天数
	MaxAge int `json:"max-age" mapstructure:"max-age"`
	// 是否压缩备份文件
	Compress bool `json:"compress" mapstructure:"compress"`
	// 是否输出到stdout
	ToStdout bool `json:"to-stdout" mapstructure:"to-stdout"`
	// 是否输出到文件
	ToFile bool `json:"to-file" mapstructure:"to-file"`
}

// NewLogOptions 创建带默认值的LogOptions实例
func NewLogOptions() *LogOptions {
	return &LogOptions{
		Level:      "info",
		Format:     "console",
		Color:      true,
		MaxSize:    100,
		MaxBackups: 10,
		MaxAge:     30,
		Compress:   true,
		ToStdout:   true,
		ToFile:     false,
	}
}

// Validate 验证日志配置选项的有效性
func (lo *LogOptions) Validate() []error {
	errs := []error{}
	// TODO: 添加具体的验证逻辑
	return errs
}

// AddFlags 将日志相关的标志添加到pflag.FlagSet
func (lo *LogOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&lo.Level, "log.level", lo.Level, "log level (debug, info, warn, error)")
	fs.StringVar(&lo.Format, "log.format", lo.Format, "log format (json, console)")
	fs.BoolVar(&lo.Color, "log.color", lo.Color, "enable colored logs")
	fs.StringVar(&lo.LogFile, "log.log-file", lo.LogFile, "log file path")
	fs.IntVar(&lo.MaxSize, "log.max-size", lo.MaxSize, "log file max size in MB")
	fs.IntVar(&lo.MaxBackups, "log.max-backups", lo.MaxBackups, "log file max backups")
	fs.IntVar(&lo.MaxAge, "log.max-age", lo.MaxAge, "log file max age in days")
	fs.BoolVar(&lo.Compress, "log.compress", lo.Compress, "compress log backups")
	fs.BoolVar(&lo.ToStdout, "log.to-stdout", lo.ToStdout, "output logs to stdout")
	fs.BoolVar(&lo.ToFile, "log.to-file", lo.ToFile, "output logs to file")
}