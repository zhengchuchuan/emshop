package srv

import (
	gapp "emshop/gin-micro/app"
	"emshop/gin-micro/core/trace"
	"emshop/gin-micro/registry"
	"emshop/gin-micro/registry/consul"
	rpcserver "emshop/gin-micro/server/rpc-server"
	"emshop/internal/app/pkg/options"
	"emshop/internal/app/userop/srv/config"
	datav1 "emshop/internal/app/userop/srv/data/v1"
	"emshop/internal/app/userop/srv/domain/do"
	servicev1 "emshop/internal/app/userop/srv/service/v1"
	"emshop/pkg/app"
	"emshop/pkg/log"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/hashicorp/consul/api"
)

func NewApp(basename string) *app.App {
	cfg := config.NewConfig()
	appl := app.NewApp("userop",
		"emshop",
		app.WithOptions(cfg),
		app.WithRunFunc(run(cfg)),
	)
	return appl
}

// NewRegistrar 创建服务注册器
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

// NewDatabase 创建数据库连接
func NewDatabase(mysqlOpts *options.MySQLOptions) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		mysqlOpts.Username,
		mysqlOpts.Password,
		mysqlOpts.Host,
		mysqlOpts.Port,
		mysqlOpts.Database)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 自动迁移数据表
	if err = db.AutoMigrate(
		&do.UserFav{},
		&do.Address{},
		&do.LeavingMessages{},
	); err != nil {
		log.Errorf("auto migrate failed: %v", err)
		return nil, err
	}

	return db, nil
}

// NewUserOpRPCServer 创建RPC服务器
func NewUserOpRPCServer(telemetry *options.TelemetryOptions, serverOpts *options.ServerOptions, srv servicev1.Service) *rpcserver.Server {
	trace.InitAgent(trace.Options{
		Name:     telemetry.Name,
		Endpoint: telemetry.Endpoint,
		Sampler:  telemetry.Sampler,
		Batcher:  telemetry.Batcher,
	})
	rpcAddr := fmt.Sprintf("%s:%d", serverOpts.Host, serverOpts.Port)
	grpcServer := rpcserver.NewServer(
		rpcserver.WithAddress(rpcAddr),
		rpcserver.WithMetrics(serverOpts.EnableMetrics),
	)
	RegisterGRPCServer(grpcServer.Server, srv)
	return grpcServer
}

// NewUserOpApp 创建应用实例
func NewUserOpApp(serverOpts *options.ServerOptions, register registry.Registrar,
	rpcServer *rpcserver.Server) (*gapp.App, error) {

	return gapp.New(
		gapp.WithName(serverOpts.Name),
		gapp.WithRPCServer(rpcServer),
		gapp.WithRegistrar(register),
	), nil
}

func run(cfg *config.Config) app.RunFunc {
	return func(baseName string) error {
		// 初始化日志
		log.Init(cfg.Log)
		defer log.Flush()

		// 初始化数据库
		db, err := NewDatabase(cfg.MySQL)
		if err != nil {
			log.Errorf("init database failed: %v", err)
			return err
		}

		// 初始化数据层
		dataFactory := datav1.GetDataFactory(db)

		// 初始化服务层
		service := servicev1.NewService(dataFactory)

		// 初始化注册器
		registrar := NewRegistrar(cfg.Registry)

		// 初始化RPC服务器
		rpcServer := NewUserOpRPCServer(cfg.Telemetry, cfg.Server, service)

		// 创建应用
		userOpApp, err := NewUserOpApp(cfg.Server, registrar, rpcServer)
		if err != nil {
			return err
		}

		// 启动应用
		if err := userOpApp.Run(); err != nil {
			log.Errorf("run userop app error: %s", err)
			return err
		}
		return nil
	}
}
