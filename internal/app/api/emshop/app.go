package admin

import (
    "context"
    gapp "emshop/gin-micro/app"
    "emshop/gin-micro/core/trace"
    rpcserver "emshop/gin-micro/server/rpc-server"
    "emshop/gin-micro/server/rpc-server/selector"
    "emshop/gin-micro/server/rpc-server/selector/p2c"
    "emshop/internal/app/api/emshop/config"
    "emshop/internal/app/pkg/options"
    "emshop/pkg/app"
    "emshop/pkg/log"
    "emshop/pkg/storage"

	"github.com/hashicorp/consul/api"

	"emshop/gin-micro/registry"
	"emshop/gin-micro/registry/consul"
)

// 创建基础应用程序
func NewApp(basename string) *app.App {
	cfg := config.New()
	appl := app.NewApp("api",
		"emshop",
		app.WithOptions(cfg),
		app.WithRunFunc(run(cfg)),
	)
	return appl
}

func run(cfg *config.Config) app.RunFunc {
	return func(baseName string) error {
		apiApp, err := NewAPIApp(cfg)
		if err != nil {
			return err
		}

		//启动
		if err := apiApp.Run(); err != nil {
			log.Errorf("run api app error: %s", err)
			return err
		}
		return nil
	}
}

// NewRegistrar 自定义consul注册
//
//	@param registry
//	@return registry.Registrar
func NewRegistrar(registry *options.RegistryOptions, dev bool) registry.Registrar {
    c := api.DefaultConfig()
    c.Address = registry.Address
    c.Scheme = registry.Scheme
    cli, err := api.NewClient(c)
    if err != nil {
        panic(err)
    }
    // 创建自定义consul注册器实例
    opts := []consul.Option{consul.WithHealthCheck(true)}
    if registry.HealthCheckInterval > 0 {
        opts = append(opts, consul.WithHealthCheckInterval(registry.HealthCheckInterval))
    }
    if registry.CheckTimeout > 0 {
        opts = append(opts, consul.WithCheckTimeout(registry.CheckTimeout))
    }
    if dev {
        opts = append(opts, consul.WithDeregisterCriticalServiceAfter(60))
    } else if registry.DeregisterCriticalAfter > 0 {
        opts = append(opts, consul.WithDeregisterCriticalServiceAfter(registry.DeregisterCriticalAfter))
    }
    r := consul.New(cli, opts...)
    return r
}

func NewAPIApp(cfg *config.Config) (*gapp.App, error) {
    //初始化log
    log.Init(cfg.Log)
    defer log.Flush()

    // 初始化全局 gRPC 负载均衡策略为 p2c，并注册自定义 balancer
    selector.SetGlobalSelector(p2c.NewBuilder())
    rpcserver.InitBuilder()

    //服务注册
    register := NewRegistrar(cfg.Registry, cfg.Log.Development)

	// 初始化链路追踪
	trace.InitAgent(trace.Options{
		Name:     cfg.Telemetry.Name,
		Endpoint: cfg.Telemetry.Endpoint,
		Sampler:  cfg.Telemetry.Sampler,
		Batcher:  cfg.Telemetry.Batcher,
	})

	//连接redis
	redisConfig := &storage.Config{
		Host:                  cfg.Redis.Host,
		Port:                  cfg.Redis.Port,
		Addrs:                 cfg.Redis.Addrs,
		MasterName:            cfg.Redis.MasterName,
		Username:              cfg.Redis.Username,
		Password:              cfg.Redis.Password,
		Database:              cfg.Redis.Database,
		MaxIdle:               cfg.Redis.MaxIdle,
		MaxActive:             cfg.Redis.MaxActive,
		Timeout:               cfg.Redis.Timeout,
		EnableCluster:         cfg.Redis.EnableCluster,
		UseSSL:                cfg.Redis.UseSSL,
		SSLInsecureSkipVerify: cfg.Redis.SSLInsecureSkipVerify,
		EnableTracing:         cfg.Redis.EnableTracing,
	}
	go storage.ConnectToRedis(context.Background(), redisConfig)

	//生成http服务
	rpcServer, err := NewAPIHTTPServer(cfg)
	if err != nil {
		return nil, err
	}

	return gapp.New(
		gapp.WithName(cfg.Server.Name),
		gapp.WithRestServer(rpcServer),
		gapp.WithRegistrar(register),
	), nil
}
