package app

import (
	"context"
	"fmt"
	"net"

	couponpb "emshop/api/coupon/v1"
	"emshop/internal/app/coupon/srv/config"
	"emshop/internal/app/coupon/srv/consumer"
	controllerv1 "emshop/internal/app/coupon/srv/controller/v1"
	datav1 "emshop/internal/app/coupon/srv/data/v1"
	"emshop/internal/app/coupon/srv/data/v1/interfaces"
	"emshop/internal/app/coupon/srv/pkg/cache"
	servicev1 "emshop/internal/app/coupon/srv/service/v1"
	"emshop/internal/app/pkg/options"
	"emshop/pkg/log"
	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
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
	grpcServer     *grpc.Server
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
	if len(cfg.Redis.Addrs) > 0 {
		addr = cfg.Redis.Addrs[0] // 使用第一个地址
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.Database,
		PoolSize:     cfg.Redis.MaxActive,
		MinIdleConns: cfg.Redis.MaxIdle,
	})

	// 测试Redis连接
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("连接Redis失败: %v", err)
	}

	// 创建数据层工厂管理器
	factoryManager, err := datav1.NewFactoryManager(cfg.MySQL)
	if err != nil {
		return nil, fmt.Errorf("创建数据工厂管理器失败: %v", err)
	}

	dataFactory := factoryManager.GetDataFactory()

	// 创建缓存管理器 (暂时使用nil repository，后续实现repository适配器)
	cacheManager, err := cache.NewCouponCacheManager(redisClient, nil, cfg.ToCacheConfig())
	if err != nil {
		return nil, fmt.Errorf("创建缓存管理器失败: %v", err)
	}

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

	// 创建服务层
	service := servicev1.NewService(dataFactory, redisClient, cfg.DTM, cfg.RocketMQ, cfg.ToCacheConfig())
	
	// 创建gRPC服务器
	grpcServer := grpc.NewServer()
	
	// 注册优惠券服务
	couponServer := controllerv1.NewCouponServer(service)
	couponpb.RegisterCouponServer(grpcServer, couponServer)

	log.Info("优惠券应用初始化成功")

	return &CouponApp{
		config:         cfg,
		cacheManager:   cacheManager,
		canalConsumer:  canalConsumer,
		redisClient:    redisClient,
		dataFactory:    dataFactory,
		factoryManager: factoryManager,
		grpcServer:     grpcServer,
		service:        service,
	}, nil
}

// Run 运行应用
func (app *CouponApp) Run(ctx context.Context) error {
	log.Info("启动优惠券服务...")

	// 启动Canal消费者
	if err := app.canalConsumer.Start(); err != nil {
		return fmt.Errorf("启动Canal消费者失败: %v", err)
	}

	// 启动gRPC服务器
	go func() {
		listen, err := net.Listen("tcp", fmt.Sprintf(":%d", app.config.Server.Port))
		if err != nil {
			log.Fatalf("gRPC服务器监听失败: %v", err)
		}

		log.Infof("gRPC服务器启动成功，监听端口: %d", app.config.Server.Port)
		
		if err := app.grpcServer.Serve(listen); err != nil {
			log.Fatalf("gRPC服务器启动失败: %v", err)
		}
	}()

	log.Infof("优惠券服务启动成功 (gRPC: %d, HTTP: %d)", app.config.Server.Port, app.config.Server.HttpPort)
	return nil
}

// Stop 停止应用
func (app *CouponApp) Stop() error {
	log.Info("停止优惠券服务...")

	// 停止gRPC服务器
	if app.grpcServer != nil {
		app.grpcServer.GracefulStop()
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