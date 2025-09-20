package app

import (
	"fmt"
	"os"

	//controller(参数校验) ->service(具体的业务逻辑) -> data(数据库的接口)
	cliflag "emshop/pkg/common/cli/flag"
	"emshop/pkg/common/cli/globalflag"
	"emshop/pkg/common/term"
	"emshop/pkg/common/version"
	"emshop/pkg/common/version/verflag"
	"emshop/pkg/errors"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"emshop/pkg/log"
)

var (
	progressMessage = color.GreenString("==>")
	//nolint: deadcode,unused,varcheck
	usageTemplate = fmt.Sprintf(`%s{{if .Runnable}}
  %s{{end}}{{if .HasAvailableSubCommands}}
  %s{{end}}{{if gt (len .Aliases) 0}}

%s
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

%s
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

%s{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  %s {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

%s
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

%s
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

%s{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "%s --help" for more information about a command.{{end}}
`,
		color.CyanString("Usage:"),
		color.GreenString("{{.UseLine}}"),
		color.GreenString("{{.CommandPath}} [command]"),
		color.CyanString("Aliases:"),
		color.CyanString("Examples:"),
		color.CyanString("Available Commands:"),
		color.GreenString("{{rpad .Name .NamePadding }}"),
		color.CyanString("Flags:"),
		color.CyanString("Global Flags:"),
		color.CyanString("Additional help topics:"),
		color.GreenString("{{.CommandPath}} [command]"),
	)
)

 // App 是 CLI 应用程序的主结构体。
 // 推荐使用 app.NewApp() 函数来创建一个应用实例。
type App struct {
	basename    string
	name        string
	description string
	options     CliOptions
	runFunc     RunFunc
	silence     bool
	noVersion   bool
	noConfig    bool
	commands    []*Command
	args        cobra.PositionalArgs
	cmd         *cobra.Command
}

 // Option 定义了初始化应用程序结构体的可选参数。
type Option func(*App)

 // WithOptions 用于开启应用程序从命令行或配置文件读取参数的功能。
func WithOptions(opt CliOptions) Option {
	return func(a *App) {
		a.options = opt
	}
}

 // RunFunc 定义了应用程序启动时的回调函数类型。
type RunFunc func(basename string) error

 // WithRunFunc 用于设置应用程序启动时的回调函数选项。
func WithRunFunc(run RunFunc) Option {
	return func(a *App) {
		a.runFunc = run
	}
}

 // WithDescription 用于设置应用程序的描述信息。
func WithDescription(desc string) Option {
	return func(a *App) {
		a.description = desc
	}
}

 // WithSilence 设置应用程序为静默模式，启动信息、配置信息和版本信息不会输出到控制台。
func WithSilence() Option {
	return func(a *App) {
		a.silence = true
	}
}

 // WithNoVersion 设置应用程序不提供版本标志。
func WithNoVersion() Option {
	return func(a *App) {
		a.noVersion = true
	}
}

 // WithNoConfig 设置应用程序不提供配置文件标志。
func WithNoConfig() Option {
	return func(a *App) {
		a.noConfig = true
	}
}

 // WithValidArgs 设置校验非 flag 参数的校验函数。
func WithValidArgs(args cobra.PositionalArgs) Option {
	return func(a *App) {
		a.args = args
	}
}

 // WithDefaultValidArgs 设置默认的非 flag 参数校验函数。
func WithDefaultValidArgs() Option {
	return func(a *App) {
		a.args = func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), args)
				}
			}

			return nil
		}
	}
}

 // NewApp 根据给定的应用名称、二进制名称和其他选项创建一个新的应用实例。
func NewApp(name string, basename string, opts ...Option) *App {
	a := &App{
		name:     name,
		basename: basename,
	}
	
	for _, o := range opts {
		o(a)
	}
	// 
	a.buildCommand()

	return a
}

// buildCommand 构建并配置应用程序的 Cobra 命令实例
// 这是应用程序初始化的核心方法，负责设置命令行界面的所有方面
func (a *App) buildCommand() {
	// 1. 创建基础的 Cobra 命令结构
	cmd := cobra.Command{
		Use:   FormatBaseName(a.basename),  // 命令名称（格式化后的可执行文件名）
		Short: a.name,                      // 简短描述
		Long:  a.description,               // 详细描述
		// 禁用错误时自动打印使用说明，由应用程序自己控制错误处理
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          a.args,  // 位置参数验证函数
	}
	
	// 2. 配置命令的输入输出流
	// cmd.SetUsageTemplate(usageTemplate)  // 可选：自定义使用说明模板
	cmd.SetOut(os.Stdout)          // 标准输出流
	cmd.SetErr(os.Stderr)          // 标准错误流
	cmd.Flags().SortFlags = true   // 按字母顺序排序标志
	
	cliflag.InitFlags(cmd.Flags()) // 初始化命令行标志系统

	// 3. 添加子命令支持
	if len(a.commands) > 0 {
		// 将应用程序定义的子命令添加到根命令
		for _, command := range a.commands {
			cmd.AddCommand(command.cobraCommand())
		}
		// 设置帮助命令
		cmd.SetHelpCommand(helpCommand(a.name))
	}
	
	// 4. 设置命令执行函数
	if a.runFunc != nil {
		cmd.RunE = a.runCommand  // 绑定运行函数到 Cobra 命令
	}

	// 5. 处理应用程序选项和标志（pflag与Cobra的集成核心）
	var namedFlagSets cliflag.NamedFlagSets
	if a.options != nil {
		// 从应用程序选项中获取命令行标志定义
		// 这里调用Config.Flags()方法，它会聚合所有子选项的pflag定义
		namedFlagSets = a.options.Flags()
		fs := cmd.Flags() // 获取Cobra命令的pflag.FlagSet
		// 将所有标志集合添加到Cobra命令的标志系统中
		// 这是pflag与Cobra集成的关键步骤：Cobra接收pflag定义的标志
		for _, f := range namedFlagSets.FlagSets {
			fs.AddFlagSet(f) // Cobra内部使用pflag.FlagSet.AddFlagSet()
		}

		// 6. 自定义帮助和使用说明格式
		usageFmt := "Usage:\n  %s\n"
		cols, _, _ := term.TerminalSize(cmd.OutOrStdout()) // 获取终端宽度
		
		// 自定义帮助函数，支持分组显示标志
		cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
			cliflag.PrintSections(cmd.OutOrStdout(), namedFlagSets, cols)
		})
		
		// 自定义使用说明函数
		cmd.SetUsageFunc(func(cmd *cobra.Command) error {
			fmt.Fprintf(cmd.OutOrStderr(), usageFmt, cmd.UseLine())
			cliflag.PrintSections(cmd.OutOrStderr(), namedFlagSets, cols)
			return nil
		})
	}

	// 7. 添加全局标志
	// 版本标志（如 --version, -v）
	if !a.noVersion {
		verflag.AddFlags(namedFlagSets.FlagSet("global"))
	}

	// 配置文件标志（如 --config, -c）
	if !a.noConfig {
		addConfigFlag(a.basename, namedFlagSets.FlagSet("global"))
	}

	// 其他全局标志（如日志级别等）
	globalflag.AddGlobalFlags(namedFlagSets.FlagSet("global"), cmd.Name())

	// 8. 保存构建好的命令实例
	a.cmd = &cmd
}

 // Run 用于启动应用程序。
func (a *App) Run() {
	if err := a.cmd.Execute(); err != nil {
		fmt.Printf("%v %v\n", color.RedString("Error:"), err)
		os.Exit(1)
	}
}

 // Command 返回应用程序内部的 cobra.Command 实例。
func (a *App) Command() *cobra.Command {
	return a.cmd
}

// runCommand 是Cobra命令的实际执行函数
// 当用户运行CLI命令时，Cobra会调用这个函数
func (a *App) runCommand(cmd *cobra.Command, args []string) error {
	printWorkingDir()
	cliflag.PrintFlags(cmd.Flags()) // 打印所有pflag标志的调试信息
	if !a.noVersion {
		// 检查是否请求版本信息并退出
		verflag.PrintAndExitIfRequested()
	}

	// 配置文件与命令行参数的绑定和合并（Viper的核心功能）
	if !a.noConfig {
		// 步骤1: 将pflag标志绑定到Viper
		// 这使得Viper可以从pflag获取命令行参数值
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}

		// 步骤2: 将Viper中的配置（文件+环境变量+命令行参数）反序列化到Go结构体
		// 这是配置系统的最终步骤：将所有配置源合并后绑定到应用程序的配置结构体
		if err := viper.Unmarshal(a.options); err != nil {
			return err
		}
	}

	if !a.silence {
		log.Infof("%v Starting %s ...", progressMessage, a.name)
		if !a.noVersion {
			log.Infof("%v Version: `%s`", progressMessage, version.Get().ToJSON())
		}
		if !a.noConfig {
			log.Infof("%v Config file used: `%s`", progressMessage, viper.ConfigFileUsed())
		}
	}
	if a.options != nil {
		if err := a.applyOptionRules(); err != nil {
			return err
		}
	}
	// run application
	if a.runFunc != nil {
		return a.runFunc(a.basename)
	}

	return nil
}

func (a *App) applyOptionRules() error {
	if completeableOptions, ok := a.options.(CompleteableOptions); ok {
		if err := completeableOptions.Complete(); err != nil {
			return err
		}
	}

	if errs := a.options.Validate(); len(errs) != 0 {
		return errors.NewAggregate(errs)
	}

	if printableOptions, ok := a.options.(PrintableOptions); ok && !a.silence {
		log.Infof("%v Config: `%s`", progressMessage, printableOptions.String())
	}

	return nil
}

func printWorkingDir() {
	wd, _ := os.Getwd()
	log.Infof("%v WorkingDir: %s", progressMessage, wd)
}
