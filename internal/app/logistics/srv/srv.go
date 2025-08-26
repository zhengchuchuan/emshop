package srv

import (
	"github.com/hashicorp/consul/api"
	"emshop/internal/app/logistics/srv/config"
	"emshop/internal/app/pkg/options"
	gapp "emshop/gin-micro/app"
	"emshop/pkg/app"
	"emshop/pkg/log"

	"emshop/gin-micro/registry"
	"emshop/gin-micro/registry/consul"
)

func NewApp(basename string) *app.App {
	cfg := config.New()
	appl := app.NewApp("logistics",
		"emshop",
		app.WithOptions(cfg),
		app.WithRunFunc(run(cfg)),
		//app.WithNoConfig(), //设置不读取配置文件，保持注释以启用配置文件
	)
	return appl
}

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

func NewLogisticsApp(cfg *config.Config) (*gapp.App, error) {
	//初始化log
	log.Init(cfg.Log)
	defer log.Flush()

	//服务注册
	register := NewRegistrar(cfg.Registry)

	//生成rpc服务
	rpcServer, err := NewLogisticsRPCServer(cfg)
	if err != nil {
		return nil, err
	}

	return gapp.New(
		gapp.WithName(cfg.Server.Name),
		gapp.WithRPCServer(rpcServer),
		gapp.WithRegistrar(register),
	), nil
}

func run(cfg *config.Config) app.RunFunc {
	return func(baseName string) error {
		logisticsApp, err := NewLogisticsApp(cfg)
		if err != nil {
			return err
		}

		//启动
		if err := logisticsApp.Run(); err != nil {
			log.Errorf("run logistics app error: %s", err)
			return err
		}
		return nil
	}
}