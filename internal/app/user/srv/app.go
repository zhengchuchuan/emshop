
package srv

import (
	gapp "emshop/gin-micro/app"
	"emshop/gin-micro/registry"
	"emshop/gin-micro/registry/consul"
	rpcserver "emshop/gin-micro/server/rpc-server"
	"emshop/internal/app/pkg/options"
	"emshop/internal/app/user/srv/config"
	"emshop/pkg/app"
	"emshop/pkg/log"

	"github.com/google/wire"
	"github.com/hashicorp/consul/api"
)

// wire provider 获取wire的注入依赖
var ProviderSet = wire.NewSet(NewUserApp, NewRegistrar, NewUserRPCServer, NewNacosDataSource)

func NewApp(basename string) *app.App {
	cfg := config.New()
	appl := app.NewApp("user",
		"emshop",
		app.WithOptions(cfg),
		app.WithRunFunc(run(cfg)), // 此处的run函数会在app.Run()时被调用
		//app.WithNoConfig(), 		//设置不读取配置文件,从命令行中读取
	)
	return appl
}

// NewRegistrar 将http服务注册到Consul,并执行健康检查
//	@param registry 
//	@return registry.Registrar 
func NewRegistrar(registry *options.RegistryOptions) registry.Registrar {
	c := api.DefaultConfig()
	c.Address = registry.Address
	c.Scheme = registry.Scheme
	cli, err := api.NewClient(c)
	if err != nil {
		panic(err)
	}
	r := consul.New(cli, consul.WithHealthCheck(true))
	return r
}

func NewUserApp(logOpts *log.Options, register registry.Registrar,
	serverOpts *options.ServerOptions, rpcServer *rpcserver.Server) (*gapp.App, error) {
	//初始化log
	log.Init(logOpts)
	defer log.Flush()

	return gapp.New(
		gapp.WithName(serverOpts.Name),
		gapp.WithRPCServer(rpcServer),
		gapp.WithRegistrar(register),
	), nil
}

func run(cfg *config.Config) app.RunFunc {
	return func(baseName string) error {
		// 通过wire生成的依赖注入
		userApp, err := initApp(cfg.Nacos, cfg.Log, cfg.Server, cfg.Registry, cfg.Telemetry, cfg.MySQLOptions)
		if err != nil {
			return err
		}

		//启动
		if err := userApp.Run(); err != nil {
			log.Errorf("run user app error: %s", err)
			return err
		}
		return nil
	}
}
