package srv

import (
	gapp "emshop-admin/gin-micro/app"
	"emshop-admin/gin-micro/registry"
	"emshop-admin/gin-micro/registry/consul"
	"emshop-admin/internal/app/pkg/options"
	"emshop-admin/internal/app/user/srv/config"
	"emshop-admin/pkg/app"
	"emshop-admin/pkg/log"

	"github.com/hashicorp/consul/api"
)

func NewApp(basename string) *app.App {
	cfg := config.New()
	appl := app.NewApp("user",
		"emshop-admin",
		app.WithOptions(cfg),
		app.WithRunFunc(run(cfg)),
		//app.WithNoConfig(), //设置不读取配置文件
	)
	return appl
}


func NewUserApp(cfg *config.Config) (*gapp.App, error) {
	// 初始化log
	log.Init(cfg.Log)
	defer log.Flush()

	// 服务注册
	register := NewRegistrar(cfg.Registry)
	// 生成rpc服务
	rpcServer, err := NewUserRPCServer(cfg)
	if err != nil {
		log.Errorf("failed to create user rpc server: %v", err)
		return nil, err
	}

	return gapp.New(
		gapp.WithRPCServer(rpcServer),
		gapp.WithRegistrar(register),
		gapp.WithName(cfg.Server.Name),
		), nil
}

func NewRegistrar(registry *options.RegistryOptions) registry.Registrar {
	// 创建Consul客户端
	c := api.DefaultConfig()
	c.Address = registry.Address
	c.Scheme = registry.Scheme
	cli, err := api.NewClient(c)
	if err != nil {
		panic(err)
	}
	// 
	r := consul.New(cli, consul.WithHealthCheck(true))
	return r
}

func run(cfg *config.Config) app.RunFunc {
	return func(baseName string) error {
		userApp, err := NewUserApp(cfg)
		if err != nil {
			log.Errorf("failed to create user app: %v", err)
			return err
		}
		if err := userApp.Run(); err != nil {
			log.Errorf("failed to run user app: %v", err)
			return err
		}
		return nil
	}
}

