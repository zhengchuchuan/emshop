package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"emshop/pkg/common/util/homedir"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const configFlagName = "config"

// 全局配置文件路径变量，由pflag标志设置
var cfgFile string

// init函数在包初始化时自动调用，定义全局配置文件标志
// nolint: gochecknoinits
func init() {
	// 使用pflag定义配置文件标志：--config/-c
	// StringVarP: 支持长标志(--config)和短标志(-c)
	pflag.StringVarP(&cfgFile, "config", "c", cfgFile, "Read configuration from specified `FILE`, "+
		"support JSON, TOML, YAML, HCL, or Java properties formats.")
}

// addConfigFlag 将配置文件标志添加到指定的pflag.FlagSet，并设置Viper配置读取
// 这个函数集成了pflag、Viper和Cobra三个库的功能
func addConfigFlag(basename string, fs *pflag.FlagSet) {
	// 将全局定义的config标志添加到当前FlagSet
	// pflag.Lookup从全局FlagSet中查找已定义的标志
	fs.AddFlag(pflag.Lookup(configFlagName))

	// 配置Viper环境变量支持
	viper.AutomaticEnv() // 启用自动环境变量读取
	// 设置环境变量前缀，如"emshop-user-srv" -> "EMSHOP_USER_SRV_"
	viper.SetEnvPrefix(strings.Replace(strings.ToUpper(basename), "-", "_", -1))
	// 环境变量键名转换：将"."和"-"替换为"_"
	// 例如：server.port -> EMSHOP_USER_SRV_SERVER_PORT
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// 注册Cobra初始化回调，在命令执行前读取配置文件
	// 这是Cobra提供的生命周期钩子
	cobra.OnInitialize(func() {
		// 配置文件路径解析
		if cfgFile != "" {
			// 如果通过--config指定了配置文件，直接使用
			viper.SetConfigFile(cfgFile)
		} else {
			// 如果未指定配置文件，搜索默认位置
			viper.AddConfigPath(".") // 当前目录

			// 如果basename包含"-"，则在用户目录添加搜索路径
			// 如"emshop-user-srv" -> ~/.emshop/
			if names := strings.Split(basename, "-"); len(names) > 1 {
				viper.AddConfigPath(filepath.Join(homedir.HomeDir(), "."+names[0]))
			}

			// 设置配置文件名（不含扩展名）
			viper.SetConfigName(basename)
		}

		// 读取配置文件（Viper的核心功能）
		// 支持JSON、TOML、YAML、HCL、Java properties等格式
		if err := viper.ReadInConfig(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: failed to read configuration file(%s): %v\n", cfgFile, err)
			os.Exit(1)
		} else {
			printConfig() // 打印读取到的配置信息
		}

	})
}

// printConfig 打印当前Viper中加载的所有配置项
// 用于调试和确认配置是否正确加载
func printConfig() {
	keys := viper.AllKeys() // 获取Viper中所有配置键
	if len(keys) > 0 {
		fmt.Printf("%v Configuration items:\n", progressMessage)
		table := uitable.New()
		table.Separator = " "
		table.MaxColWidth = 80
		table.RightAlign(0)
		// 遍历所有配置项，以表格形式打印
		for _, k := range keys {
			// viper.Get(k) 获取配置值，支持多种数据类型
			table.AddRow(fmt.Sprintf("%s:", k), viper.Get(k))
		}
		fmt.Printf("%v", table)
	}
}
