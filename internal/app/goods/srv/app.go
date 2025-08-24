package srv

import (
	"github.com/hashicorp/consul/api"
	"emshop/internal/app/goods/srv/config"
	"emshop/internal/app/goods/srv/consumer"
	"emshop/internal/app/pkg/options"
	gapp "emshop/gin-micro/app"
	"emshop/pkg/app"
	"emshop/pkg/log"

	"emshop/gin-micro/registry"
	"emshop/gin-micro/registry/consul"
)

func NewApp(basename string) *app.App {
	cfg := config.New()
	appl := app.NewApp("goods",
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

func NewGoodsApp(cfg *config.Config) (*gapp.App, error) {
	//初始化log
	log.Init(cfg.Log)
	defer log.Flush()

	//服务注册
	register := NewRegistrar(cfg.Registry)

	//生成rpc服务
	rpcServer, factoryManager, err := NewGoodsRPCServer(cfg)
	if err != nil {
		return nil, err
	}

	// 创建Canal消费者配置
	canalConsumerConfig := &consumer.CanalConsumerConfig{
		NameServers:   cfg.RocketMQ.NameServers,
		ConsumerGroup: cfg.RocketMQ.ConsumerGroup,
		Topic:         cfg.RocketMQ.Topic,
		MaxReconsume:  cfg.RocketMQ.MaxReconsume,
	}
	
	canalConsumer := consumer.NewCanalConsumer(canalConsumerConfig, factoryManager.GetSyncManager())

	// 在后台启动Canal消费者
	go func() {
		log.Info("starting canal consumer in background")
		if err := canalConsumer.Start(); err != nil {
			log.Errorf("failed to start canal consumer: %v", err)
		}
	}()

	return gapp.New(
		gapp.WithName(cfg.Server.Name),
		gapp.WithRPCServer(rpcServer),
		gapp.WithRegistrar(register),
	), nil
}

func run(cfg *config.Config) app.RunFunc {
	return func(baseName string) error {
		goodsApp, err := NewGoodsApp(cfg)
		if err != nil {
			return err
		}

		//启动
		if err := goodsApp.Run(); err != nil {
			log.Errorf("run user app error: %s", err)
			return err
		}
		return nil
	}
}
