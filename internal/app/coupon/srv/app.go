package app

import (
	"context"
	"fmt"

	couponpb "emshop/api/coupon/v1"
	"emshop/gin-micro/core/trace"
	rpcserver "emshop/gin-micro/server/rpc-server"
	"emshop/internal/app/coupon/srv/config"
	"emshop/internal/app/coupon/srv/consumer"
	controllerv1 "emshop/internal/app/coupon/srv/controller/v1"
	datav1 "emshop/internal/app/coupon/srv/data/v1"
	"emshop/internal/app/coupon/srv/data/v1/interfaces"
	"emshop/internal/app/coupon/srv/pkg/cache"
	servicev1 "emshop/internal/app/coupon/srv/service/v1"
	"emshop/internal/app/pkg/options"
	"emshop/pkg/log"
	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
	"os"
)

// CouponApp 优惠券应用
type CouponApp struct {
	config         *config.Config
	cacheManager   cache.CacheManager
	canalConsumer  *consumer.CouponCanalConsumer
	redisClient    *redis.Client
	dataFactory    interfaces.DataFactory
	factoryManager *datav1.FactoryManager
	rpcServer      *rpcserver.Server
	service        *servicev1.Service
}

// NewCouponApp 创建优惠券应用
func NewCouponApp(configFile string) (*CouponApp, error) {
	// 加载配置
	cfg, err := loadConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %v", err)
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

	// 测试Redis连接 - 暂时禁用直到解决兼容性问题
	// TODO: 调试 Redis 8.0.1 和 go-redis/v9 的兼容性问题
	/*
		if err := redisClient.Ping(context.Background()).Err(); err != nil {
			return nil, fmt.Errorf("连接Redis失败: %v", err)
		}
		log.Info("Redis连接测试成功")
	*/
	log.Warn("Redis连接测试暂时跳过，需要调试兼容性问题")

	// 创建数据层工厂管理器
	factoryManager, err := datav1.NewFactoryManager(cfg.MySQL)
	if err != nil {
		return nil, fmt.Errorf("创建数据工厂管理器失败: %v", err)
	}

	dataFactory := factoryManager.GetDataFactory()

	// 创建缓存管理器 - 暂时跳过，因为需要更新到go-redis/v9
	// TODO: 更新 cache.NewCouponCacheManager 以支持 redis/go-redis/v9
	var cacheManager cache.CacheManager
	log.Warn("缓存管理器暂时禁用，需要更新到go-redis/v9")

	// 创建Canal消费者配置
	canalConfig := &consumer.CanalConsumerConfig{
		NameServers:   cfg.RocketMQ.NameServers,
		ConsumerGroup: cfg.Canal.ConsumerGroup,
		Topic:         cfg.Canal.Topic,
		WatchTables:   cfg.Canal.WatchTables,
		BatchSize:     cfg.Canal.BatchSize,
	}

	// 创建Canal消费者
	canalConsumer := consumer.NewCouponCanalConsumer(canalConfig, cacheManager)

	// 创建服务层 - 暂时不传递redisClient，因为需要更新到go-redis/v9
	// TODO: 更新 servicev1.NewService 以支持 redis/go-redis/v9
	service := servicev1.NewService(dataFactory, nil, cfg.DTM, cfg.RocketMQ, cfg.ToCacheConfig())
	log.Warn("服务层Redis客户端暂时禁用，需要更新到go-redis/v9")

	// 初始化链路追踪
	trace.InitAgent(trace.Options{
		Name:     cfg.Telemetry.Name,
		Endpoint: cfg.Telemetry.Endpoint,
		Sampler:  cfg.Telemetry.Sampler,
		Batcher:  cfg.Telemetry.Batcher,
	})

	// 创建gRPC服务器
	rpcAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	rpcSrv := rpcserver.NewServer(
		rpcserver.WithAddress(rpcAddr),
		rpcserver.WithMetrics(cfg.Server.EnableMetrics),
	)

	// 注册优惠券服务
	couponServer := controllerv1.NewCouponServer(service)
	couponpb.RegisterCouponServer(rpcSrv.Server, couponServer)

	log.Info("优惠券应用初始化成功")

	return &CouponApp{
		config:         cfg,
		cacheManager:   cacheManager,
		canalConsumer:  canalConsumer,
		redisClient:    redisClient,
		dataFactory:    dataFactory,
		factoryManager: factoryManager,
		rpcServer:      rpcSrv,
		service:        service,
	}, nil
}

// Run 运行应用
func (app *CouponApp) Run(ctx context.Context) error {
	log.Info("启动优惠券服务...")

	// 启动Canal消费者 - 暂时禁用，需要创建RocketMQ主题
	// TODO: 创建coupon-binlog-topic主题后启用
	/*
		if err := app.canalConsumer.Start(); err != nil {
			return fmt.Errorf("启动Canal消费者失败: %v", err)
		}
	*/
	log.Warn("Canal消费者暂时禁用，需要创建RocketMQ主题")

	// 启动gRPC服务器
	go func() {
		if err := app.rpcServer.Start(context.Background()); err != nil {
			log.Fatalf("gRPC服务器启动失败: %v", err)
		}
	}()

	log.Infof("优惠券服务启动成功 (gRPC: %s, HTTP: %d)", app.rpcServer.Address(), app.config.Server.HttpPort)
	return nil
}

// Stop 停止应用
func (app *CouponApp) Stop() error {
	log.Info("停止优惠券服务...")

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

	// 停止Canal消费者
	if err := app.canalConsumer.Stop(); err != nil {
		log.Errorf("停止Canal消费者失败: %v", err)
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

	log.Info("优惠券服务停止完成")
	return nil
}

// loadConfig 加载配置文件
func loadConfig(configFile string) (*config.Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &cfg, nil
}

// initLogger 初始化日志
func initLogger(logOpts *options.LogOptions) error {
	// 这里暂时简化，实际应该根据配置初始化日志
	log.Info("日志系统初始化成功")
	return nil
}
