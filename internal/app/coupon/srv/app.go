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
    dto "emshop/internal/app/coupon/srv/domain/dto"
	"emshop/internal/app/coupon/srv/tasks"
	"emshop/internal/app/pkg/options"
	appframework "emshop/pkg/app"
	"emshop/pkg/log"

    redis "github.com/go-redis/redis/v8"
    "github.com/google/uuid"
    "github.com/hashicorp/consul/api"
    restserver "emshop/gin-micro/server/rest-server"
    // 控制 RocketMQ Go 客户端日志级别，避免刷屏 INFO 日志
    "github.com/apache/rocketmq-client-go/v2/rlog"
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
    restServer      *restserver.Server
	service         *servicev1.Service
	registrar       registry.Registrar
	serviceInstance *registry.ServiceInstance

	// background tasks
	stockLogSyncer *tasks.StockLogSyncer
	reconciler     *tasks.Reconciler
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
    service := servicev1.NewService(dataFactory, redisClient, cfg.RocketMQ, cfg.ToCacheConfig(), cfg.Business)
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
    // 无论是否使用事务消息，都启动秒杀事件消费者，用于统一异步落库（包含 remaining_count 持久化等）
    if service.AsyncFlashSaleEnabled() && cfg.Business != nil && cfg.Business.FlashSale != nil {
        // 注意：RocketMQ PushConsumer 批量范围 [1,1024]
        mqBatch := int(cfg.Business.FlashSale.BatchSize)
        if mqBatch <= 0 { mqBatch = 16 }
        if mqBatch > 1024 {
            log.Warnf("RocketMQ batch_size=%d 超出范围，自动调整为 1024", mqBatch)
            mqBatch = 1024
        }
        flashSaleCfg = &consumer.FlashSaleConsumerConfig{
            NameServers:   cfg.RocketMQ.NameServers,
            ConsumerGroup: cfg.RocketMQ.ConsumerGroup,
            Topic:         cfg.RocketMQ.Topic,
            BatchSize:     mqBatch,
            MaxRetries:    int(cfg.RocketMQ.MaxReconsume),
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
        registrar, err = newConsulRegistrar(cfg.Registry, cfg.Log.Development)
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

	app := &CouponApp{
		config:          cfg,
		cacheManager:    cacheManager,
		canalConsumer:   canalConsumer,
		flashSaleConfig: flashSaleCfg,
		flashSaleConsumer: flashSaleConsumer,
		redisClient:     redisClient,
		dataFactory:     dataFactory,
		factoryManager:  factoryManager,
        rpcServer:       rpcSrv,
        restServer:      nil,
		service:         service,
		registrar:       registrar,
		serviceInstance: serviceInstance,
	}

	// 后台任务：Redis 扣减日志落库 & 对账
    if service.AsyncFlashSaleEnabled() && !service.UsingTxnFlashSale() {
        batch := 100
        if cfg.Business != nil && cfg.Business.FlashSale != nil && cfg.Business.FlashSale.BatchSize > 0 {
            batch = int(cfg.Business.FlashSale.BatchSize)
        }
        app.stockLogSyncer = tasks.NewStockLogSyncer(dataFactory, redisClient, batch, 100*time.Millisecond, 8, 2*time.Second)
        // 使用最终事件生产者（事务或普通）做补偿
        var interval = 30 * time.Second
        var threshold = 0
        var maxPerRun = 100
        var cooldown = 60 * time.Second
        if cfg.Business != nil && cfg.Business.FlashSale != nil {
            if cfg.Business.FlashSale.ReconcileInterval > 0 { interval = cfg.Business.FlashSale.ReconcileInterval }
            if cfg.Business.FlashSale.ReconcileThreshold >= 0 { threshold = cfg.Business.FlashSale.ReconcileThreshold }
            if cfg.Business.FlashSale.CompensationMaxPerRun > 0 { maxPerRun = cfg.Business.FlashSale.CompensationMaxPerRun }
            if cfg.Business.FlashSale.CompensationCooldown > 0 { cooldown = cfg.Business.FlashSale.CompensationCooldown }
        }
        // 当 interval <= 0 时，关闭对账任务
        if interval > 0 {
        // 方案B下（非事务消息），使用普通事件生产者做补偿；避免传入 typed-nil 事务生产者引发 panic。
        var compProducer consumer.FlashSaleEventProducer
        if service.EventProducer != nil {
            compProducer = service.EventProducer
        }
        app.reconciler = tasks.NewReconciler(dataFactory, redisClient, interval, compProducer, threshold, maxPerRun, cooldown)
        } else {
            app.reconciler = nil
            log.Warn("库存对账任务已禁用（reconcile_interval<=0）")
        }

        // 启动统计合并任务（将 Redis 增量合并到 DB）
        statsMerge := tasks.NewStatsMergeSyncer(dataFactory, redisClient, 1*time.Second)
        statsMerge.Start()
    }

	return app, nil
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

    // 自动启动配置中的秒杀活动（预热Redis，避免压测时打爆数据库）
    go app.autoStartFlashSales()

    // 启动后台任务
    if app.stockLogSyncer != nil {
        app.stockLogSyncer.Start()
        log.Info("库存扣减日志落库任务已启动")
    }
	if app.reconciler != nil {
		app.reconciler.Start()
		log.Info("库存对账任务已启动")
	}

	// 启动gRPC服务器
	go func() {
		if err := app.rpcServer.Start(context.Background()); err != nil {
			log.Fatalf("gRPC服务器启动失败: %v", err)
		}
	}()

    // 已移除内部管理REST服务器（改用 gRPC 管理接口）

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

    // 无REST管理服务可停止

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

	// 停止后台任务
	if app.stockLogSyncer != nil { app.stockLogSyncer.Stop() }
	if app.reconciler != nil { app.reconciler.Stop() }

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

    // 降低 RocketMQ 客户端内部日志级别，屏蔽诸如
    // "update offset to broker success" 等 INFO 级别日志
    // 注意：必须在创建任何 RocketMQ Producer/Consumer 之前调用
    // 可按需调整为 "warn"/"error"，这里选用更静默的 error
    // 参考: github.com/apache/rocketmq-client-go/v2/rlog
    rlog.SetLogLevel("error")
    return nil
}

func newConsulRegistrar(registryOpts *options.RegistryOptions, dev bool) (registry.Registrar, error) {
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

    opts := []consul.Option{consul.WithHealthCheck(true), consul.WithHeartbeat(false)}
    if registryOpts.HealthCheckInterval > 0 {
        opts = append(opts, consul.WithHealthCheckInterval(registryOpts.HealthCheckInterval))
    }
    if registryOpts.CheckTimeout > 0 {
        opts = append(opts, consul.WithCheckTimeout(registryOpts.CheckTimeout))
    }
    if dev {
        opts = append(opts, consul.WithDeregisterCriticalServiceAfter(60))
    } else if registryOpts.DeregisterCriticalAfter > 0 {
        opts = append(opts, consul.WithDeregisterCriticalServiceAfter(registryOpts.DeregisterCriticalAfter))
    }
    return consul.New(cli, opts...), nil
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

// autoStartFlashSales 自动启动配置或环境变量指定的活动ID
func (app *CouponApp) autoStartFlashSales() {
    if app == nil || app.service == nil || app.service.FlashSaleCore == nil {
        return
    }
    var ids []int64
    // 优先读取环境变量：COUPON_AUTO_START_IDS=1,2,3
    if v := os.Getenv("COUPON_AUTO_START_IDS"); v != "" {
        for _, part := range splitAndTrim(v, ',') {
            if n, err := strconv.ParseInt(part, 10, 64); err == nil && n > 0 {
                ids = append(ids, n)
            }
        }
    } else if app.config != nil && app.config.Business != nil && app.config.Business.FlashSale != nil {
        if len(app.config.Business.FlashSale.AutoStartIDs) > 0 {
            ids = append(ids, app.config.Business.FlashSale.AutoStartIDs...)
        }
    }
    if len(ids) == 0 {
        return
    }
    for _, id := range ids {
        ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
        err := app.service.FlashSaleCore.StartFlashSaleActivity(ctx, &dto.StartFlashSaleDTO{ActivityID: id})
        cancel()
        if err != nil {
            log.Warnf("自动启动秒杀活动失败: id=%d err=%v", id, err)
        } else {
            log.Infof("自动启动秒杀活动成功: id=%d", id)
        }
    }
}

func splitAndTrim(s string, sep rune) []string {
    var out []string
    cur := make([]rune, 0, len(s))
    for _, r := range s {
        if r == sep {
            str := string(cur)
            if t := trimSpaces(str); t != "" { out = append(out, t) }
            cur = cur[:0]
        } else {
            cur = append(cur, r)
        }
    }
    if len(cur) > 0 { if t := trimSpaces(string(cur)); t != "" { out = append(out, t) } }
    return out
}

func trimSpaces(s string) string {
    i, j := 0, len(s)
    for i < j && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') { i++ }
    for j > i && (s[j-1] == ' ' || s[j-1] == '\t' || s[j-1] == '\n' || s[j-1] == '\r') { j-- }
    return s[i:j]
}
