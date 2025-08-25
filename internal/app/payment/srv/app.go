package srv

import (
	gapp "emshop/gin-micro/app"
	"emshop/gin-micro/registry"
	"emshop/gin-micro/registry/consul"
	"emshop/internal/app/payment/srv/config"
	"emshop/internal/app/pkg/options"
	"emshop/pkg/app"
	"emshop/pkg/log"

	"github.com/hashicorp/consul/api"
)

func NewApp(basename string) *app.App {
	cfg := config.New()
	appl := app.NewApp("payment",
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

func NewPaymentApp(cfg *config.Config) (*gapp.App, error) {
	//初始化log
	log.Init(cfg.Log)
	defer log.Flush()

	//服务注册
	register := NewRegistrar(cfg.Registry)

	//生成rpc服务
	rpcServer, err := NewPaymentRPCServer(cfg)
	if err != nil {
		return nil, err
	}

	opts := []gapp.Option{
		gapp.WithName(cfg.Server.Name),
		gapp.WithRegistrar(register),
		gapp.WithRPCServer(rpcServer),
	}
	return gapp.New(opts...)
}

func run(cfg *config.Config) app.RunFunc {
	return func(basename string) error {
		paymentApp, err := NewPaymentApp(cfg)
		if err != nil {
			return err
		}

		if err := paymentApp.Start(); err != nil {
			log.Fatalf("start payment server failed: %s", err.Error())
		}

		log.Info("payment server stop")
		return paymentApp.Stop()
	}
}