package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	couponpb "emshop/api/coupon/v1"
	"emshop/gin-micro/core/trace"
	"emshop/gin-micro/registry"
	"emshop/gin-micro/registry/consul"
	rpcserver "emshop/gin-micro/server/rpc-server"
	"emshop/internal/app/coupon/srv/config"
	"emshop/internal/app/coupon/srv/consumer"
	controllerv1 "emshop/internal/app/coupon/srv/controller/v1"
	datav1 "emshop/internal/app/coupon/srv/data/v1"
	"emshop/internal/app/coupon/srv/data/v1/interfaces"
	"emshop/internal/app/coupon/srv/pkg/cache"
	servicev1 "emshop/internal/app/coupon/srv/service/v1"
	"emshop/internal/app/pkg/options"
	appframework "emshop/pkg/app"
	"emshop/pkg/log"

	redis "github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
)

// NewApp returns a CLI application wired for the coupon service.
func NewApp(basename string) *appframework.App {
	cfg := config.New()
	return appframework.NewApp(
		"coupon",
		basename,
		appframework.WithOptions(cfg),
		appframework.WithRunFunc(run(cfg)),
	)
}

func run(cfg *config.Config) appframework.RunFunc {
	return func(basename string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		couponApp, err := NewCouponApp(cfg)
		if err != nil {
			return err
		}

		if err := couponApp.Run(ctx); err != nil {
			_ = couponApp.Stop()
			return err
		}

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(quit)
		<-quit

		log.Info("正在优雅关闭优惠券服务...")

		if err := couponApp.Stop(); err != nil {
			log.Errorf("停止优惠券服务失败: %v", err)
			return err
		}

		log.Info("优惠券服务已停止")
		return nil
	}
}

// CouponApp 优惠券应用
type CouponApp struct {
	config          *config.Config
	cacheManager    cache.CacheManager
	canalConsumer   *consumer.CouponCanalConsumer
	flashSaleConfig *consumer.FlashSaleConsumerConfig
	flashSaleConsumer *consumer.FlashSaleConsumer
	redisClient     *redis.Client
	dataFactory     interfaces.DataFactory
	factoryManager  *datav1.FactoryManager
	rpcServer       *rpcserver.Server
	service         *servicev1.Service
	registrar       registry.Registrar
	serviceInstance *registry.ServiceInstance
}

// NewCouponApp 创建优惠券应用
func NewCouponApp(cfg *config.Config) (*CouponApp, error) {
	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 初始化日志
	if err := initLogger(cfg.Log); err != nil {
		return nil, fmt.Errorf("初始化日志失败: %v", err)
	}

	// 创建Redis客户端
	addr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.Database,
	})

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := redisClient.Ping(pingCtx).Err(); err != nil {
		cancel()
		return nil, fmt.Errorf("连接Redis失败: %v", err)
	}
	cancel()
	log.Infof("Redis连接测试成功, addr: %s", addr)

	// 创建数据层工厂管理器
	factoryManager, err := datav1.NewFactoryManager(cfg.MySQL)
	if err != nil {
		return nil, fmt.Errorf("创建数据工厂管理器失败: %v", err)
	}

	dataFactory := factoryManager.GetDataFactory()

	// 创建服务层，注入Redis与RocketMQ依赖
	service := servicev1.NewService(dataFactory, redisClient, cfg.DTM, cfg.RocketMQ, cfg.ToCacheConfig(), cfg.Business)
	cacheManager := service.CacheManager
	if cacheManager == nil {
		return nil, fmt.Errorf("初始化缓存管理器失败")
	}

	if cfg.Canal == nil {
		return nil, fmt.Errorf("未配置Canal同步信息")
	}

	canalConfig := &consumer.CanalConsumerConfig{
		NameServers:   cfg.RocketMQ.NameServers,
		ConsumerGroup: cfg.Canal.ConsumerGroup,
		Topic:         cfg.Canal.Topic,
		WatchTables:   cfg.Canal.WatchTables,
		BatchSize:     cfg.Canal.BatchSize,
	}

	canalConsumer := consumer.NewCouponCanalConsumer(canalConfig, cacheManager)

	var flashSaleConsumer *consumer.FlashSaleConsumer
	var flashSaleCfg *consumer.FlashSaleConsumerConfig
	if service.AsyncFlashSaleEnabled() && cfg.Business != nil && cfg.Business.FlashSale != nil {
		flashSaleCfg = &consumer.FlashSaleConsumerConfig{
			NameServers:   cfg.RocketMQ.NameServers,
			ConsumerGroup: cfg.RocketMQ.ConsumerGroup,
			Topic:         cfg.RocketMQ.Topic,
			BatchSize:     int(cfg.Business.FlashSale.BatchSize),
			MaxRetries:    int(cfg.RocketMQ.MaxReconsume),
		}
		if flashSaleCfg.BatchSize <= 0 {
			flashSaleCfg.BatchSize = 16
		}
		if flashSaleCfg.MaxRetries <= 0 {
			flashSaleCfg.MaxRetries = 3
		}
		flashSaleConsumer = consumer.NewFlashSaleConsumer(dataFactory, redisClient, service.RetryManager)
	}

	// 初始化链路追踪
	if cfg.Telemetry != nil {
		trace.InitAgent(trace.Options{
			Name:     cfg.Telemetry.Name,
			Endpoint: cfg.Telemetry.Endpoint,
			Sampler:  cfg.Telemetry.Sampler,
			Batcher:  cfg.Telemetry.Batcher,
		})
	}

	// 创建gRPC服务器
	rpcAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	rpcSrv := rpcserver.NewServer(
		rpcserver.WithAddress(rpcAddr),
		rpcserver.WithMetrics(cfg.Server.EnableMetrics),
	)

	// 注册优惠券服务
	couponServer := controllerv1.NewCouponServer(service)
	couponpb.RegisterCouponServer(rpcSrv.Server, couponServer)

	var registrar registry.Registrar
	var serviceInstance *registry.ServiceInstance
	if cfg.Registry != nil {
		registrar, err = newConsulRegistrar(cfg.Registry)
		if err != nil {
			return nil, fmt.Errorf("创建Consul注册器失败: %v", err)
		}
		serviceInstance, err = buildServiceInstance(cfg.Server, rpcSrv)
		if err != nil {
			return nil, fmt.Errorf("构建服务实例失败: %v", err)
		}
	} else {
		log.Warn("未配置服务注册信息，将跳过Consul服务注册")
	}

	log.Info("优惠券应用初始化成功")

	return &CouponApp{
		config:          cfg,
		cacheManager:    cacheManager,
		canalConsumer:   canalConsumer,
		flashSaleConfig: flashSaleCfg,
		flashSaleConsumer: flashSaleConsumer,
		redisClient:     redisClient,
		dataFactory:     dataFactory,
		factoryManager:  factoryManager,
		rpcServer:       rpcSrv,
		service:         service,
		registrar:       registrar,
		serviceInstance: serviceInstance,
	}, nil
}

// Run 运行应用
func (app *CouponApp) Run(ctx context.Context) error {
	log.Info("启动优惠券服务...")

	if app.flashSaleConsumer != nil && app.flashSaleConfig != nil {
		if err := app.flashSaleConsumer.Start(app.flashSaleConfig); err != nil {
			return fmt.Errorf("启动秒杀事件消费者失败: %v", err)
		}
		log.Info("秒杀事件消费者启动成功，已开启异步落库")
	}

	if app.canalConsumer != nil {
		if err := app.canalConsumer.Start(); err != nil {
			return fmt.Errorf("启动Canal消费者失败: %v", err)
		}
		log.Info("Canal消费者启动成功，已开启缓存同步")
	} else {
		log.Warn("Canal消费者未初始化，跳过缓存同步")
	}

	// 启动gRPC服务器
	go func() {
		if err := app.rpcServer.Start(context.Background()); err != nil {
			log.Fatalf("gRPC服务器启动失败: %v", err)
		}
	}()

	if app.registrar != nil && app.serviceInstance != nil {
		regCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err := app.registrar.Register(regCtx, app.serviceInstance)
		cancel()
		if err != nil {
			log.Errorf("注册优惠券服务到Consul失败: %v", err)
			if app.rpcServer != nil {
				_ = app.rpcServer.Stop(context.Background())
			}
			return fmt.Errorf("注册优惠券服务失败: %w", err)
		}
		log.Infof("优惠券服务已注册到Consul (serviceID=%s)", app.serviceInstance.ID)
	} else {
		log.Warn("Consul注册器未初始化，跳过服务注册")
	}

	log.Infof("优惠券服务启动成功 (gRPC: %s, HTTP: %d)", app.rpcServer.Address(), app.config.Server.HttpPort)
	return nil
}

// Stop 停止应用
func (app *CouponApp) Stop() error {
	log.Info("停止优惠券服务...")

	if app.registrar != nil && app.serviceInstance != nil {
		deregCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := app.registrar.Deregister(deregCtx, app.serviceInstance)
		cancel()
		if err != nil {
			log.Errorf("注销优惠券服务失败: %v", err)
		} else {
			log.Info("已从Consul注销优惠券服务")
		}
	}

	// 停止gRPC服务器
	if app.rpcServer != nil {
		_ = app.rpcServer.Stop(context.Background())
		log.Info("gRPC服务器已停止")
	}

	// 关闭服务层（包括RocketMQ生产者）
	if app.service != nil {
		if err := app.service.Shutdown(); err != nil {
			log.Errorf("关闭服务层失败: %v", err)
		}
	}

	if app.flashSaleConsumer != nil {
		if err := app.flashSaleConsumer.Stop(); err != nil {
			log.Errorf("停止秒杀事件消费者失败: %v", err)
		}
	}

	// 停止Canal消费者
	if app.canalConsumer != nil {
		if err := app.canalConsumer.Stop(); err != nil {
			log.Errorf("停止Canal消费者失败: %v", err)
		}
	}

	// 关闭缓存管理器
	if app.cacheManager != nil {
		app.cacheManager.Close()
	}

	// 关闭Redis连接
	if app.redisClient != nil {
		app.redisClient.Close()
	}

	// 关闭数据工厂管理器
	if app.factoryManager != nil {
		if err := app.factoryManager.Close(); err != nil {
			log.Errorf("关闭数据工厂管理器失败: %v", err)
		}
	}

	log.Flush()
	log.Info("优惠券服务停止完成")
	return nil
}

// initLogger 初始化日志
func initLogger(logOpts *log.Options) error {
	if logOpts == nil {
		logOpts = log.NewOptions()
	}

	log.Init(logOpts)
	log.Infof("日志系统初始化成功，level=%s", logOpts.Level)
	return nil
}

func newConsulRegistrar(registryOpts *options.RegistryOptions) (registry.Registrar, error) {
	if registryOpts == nil {
		return nil, fmt.Errorf("registry配置为空")
	}

	cfg := api.DefaultConfig()
	if registryOpts.Address != "" {
		cfg.Address = registryOpts.Address
	}
	if registryOpts.Scheme != "" {
		cfg.Scheme = registryOpts.Scheme
	}

	cli, err := api.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("创建Consul客户端失败: %w", err)
	}

	return consul.New(cli, consul.WithHealthCheck(true)), nil
}

func buildServiceInstance(serverOpts *options.ServerOptions, rpcSrv *rpcserver.Server) (*registry.ServiceInstance, error) {
	if serverOpts == nil {
		return nil, fmt.Errorf("server配置为空")
	}
	if rpcSrv == nil {
		return nil, fmt.Errorf("gRPC服务器未初始化")
	}

	endpoint := rpcSrv.Endpoint()
	if endpoint == nil {
		return nil, fmt.Errorf("无法获取gRPC服务Endpoint")
	}

	instanceID := uuid.NewString()
	if serverOpts.Name != "" {
		instanceID = fmt.Sprintf("%s-%s", serverOpts.Name, instanceID)
	}

	metadata := map[string]string{
		"host":      serverOpts.Host,
		"http_port": strconv.Itoa(serverOpts.HttpPort),
	}

	return &registry.ServiceInstance{
		ID:        instanceID,
		Name:      serverOpts.Name,
		Endpoints: []string{endpoint.String()},
		Metadata:  metadata,
	}, nil
}
