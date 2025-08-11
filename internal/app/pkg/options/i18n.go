package options

import (
	"github.com/spf13/pflag"
)

// I18nOptions 国际化配置选项
type I18nOptions struct {
	// 默认语言环境: zh-CN, en
	Locale string `json:"locale" mapstructure:"locale"`
	// 翻译文件目录路径，为空时使用内置翻译
	LocalesDir string `json:"locales-dir" mapstructure:"locales-dir"`
}

// NewI18nOptions 创建默认的国际化配置
func NewI18nOptions() *I18nOptions {
	return &I18nOptions{
		Locale:     "zh-CN",
		LocalesDir: "./locales", // 默认翻译文件目录
	}
}

// Validate 验证配置参数
func (o *I18nOptions) Validate() []error {
	var errors []error
	
	// 验证语言环境
	if o.Locale != "zh-CN" && o.Locale != "en" {
		// 这里可以添加更多支持的语言环境验证
		// 目前只是警告，不返回错误，使用默认值
	}
	
	return errors
}

// AddFlags 添加命令行参数
func (o *I18nOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Locale, "i18n.locale", o.Locale, 
		"Default locale for internationalization (zh-CN, en)")
	fs.StringVar(&o.LocalesDir, "i18n.locales-dir", o.LocalesDir, 
		"Directory path for translation files, empty for built-in translations")
}