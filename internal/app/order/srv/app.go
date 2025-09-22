package srv

import (
    gapp "emshop/gin-micro/app"
    "emshop/gin-micro/registry"
    "emshop/gin-micro/registry/consul"
    rpcserver "emshop/gin-micro/server/rpc-server"
    "emshop/gin-micro/server/rpc-server/selector"
    "emshop/gin-micro/server/rpc-server/selector/p2c"
    "emshop/internal/app/order/srv/config"
    "emshop/internal/app/pkg/options"
    "emshop/pkg/app"
    "emshop/pkg/log"

	"github.com/hashicorp/consul/api"
)

func NewApp(basename string) *app.App {
	cfg := config.New()
	appl := app.NewApp("order",
		"emshop",
		app.WithOptions(cfg),
		app.WithRunFunc(run(cfg)),
		//app.WithNoConfig(), //设置不读取配置文件
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

func NeworderApp(cfg *config.Config) (*gapp.App, error) {
    //初始化log
    log.Init(cfg.Log)
    defer log.Flush()

    // 初始化全局 gRPC 负载均衡策略为 p2c，并注册自定义 balancer
    selector.SetGlobalSelector(p2c.NewBuilder())
    rpcserver.InitBuilder()

    //服务注册
    register := NewRegistrar(cfg.Registry)

	//生成rpc服务
	rpcServer, err := NewOrderRPCServer(cfg)
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
		orderApp, err := NeworderApp(cfg)
		if err != nil {
			return err
		}

		//启动
		if err := orderApp.Run(); err != nil {
			log.Errorf("run user app error: %s", err)
			return err
		}
		return nil
	}
}
