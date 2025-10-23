package v1

import (
    "context"
    "emshop/internal/app/coupon/srv/config"
    "emshop/internal/app/coupon/srv/consumer"
    "emshop/internal/app/coupon/srv/data/v1/interfaces"
    "emshop/internal/app/coupon/srv/pkg/cache"
    "emshop/internal/app/pkg/options"
    "emshop/pkg/log"
    v1 "emshop/pkg/common/meta/v1"
    "github.com/go-redis/redis/v8"
    "fmt"
)

// Service 优惠券服务工厂
type Service struct {
    CouponSrv           CouponSrv
    FlashSaleCore       FlashSaleSrvCore  // 新的秒杀核心服务
    CacheManager        cache.CacheManager
    EventProducer       consumer.FlashSaleEventProducer // RocketMQ事件生产者
    TxnProducer         consumer.FlashSaleTxnProducer   // RocketMQ事务生产者（方案A）
    RetryManager        *consumer.RetryManager // 重试管理器
    asyncFlashSale      bool
    data                interfaces.DataFactory
}

// NewService 创建优惠券服务工厂
func NewService(
    data interfaces.DataFactory,
    redisClient *redis.Client,
    rocketmqOpts *options.RocketMQOptions,
    cacheConfig *cache.CacheConfig,
    bizOpts *config.BusinessOptions,
) *Service {
	// 创建缓存适配器，将数据层接口适配为缓存仓库接口
	cacheRepository := newCacheRepositoryAdapter(data)
	
	// 创建缓存管理器
	cacheManager, err := cache.NewCouponCacheManager(redisClient, cacheRepository, cacheConfig)
	if err != nil {
		log.Errorf("初始化缓存管理器失败: %v", err)
		// 在缓存初始化失败时，仍然可以运行，只是性能会下降
		cacheManager = nil
	}
	
	// 创建RocketMQ事件生产者
    var eventProducer consumer.FlashSaleEventProducer
    if rocketmqOpts != nil {
        producer, err := consumer.NewFlashSaleEventProducer(
            rocketmqOpts.NameServers,
            "coupon-producer-group", // 生产者组名
            rocketmqOpts.Topic,
            rocketmqOpts.AsyncSend,
        )
        if err != nil {
            log.Errorf("初始化RocketMQ事件生产者失败: %v", err)
            // 可以考虑使用fallback或mock实现
            eventProducer = nil
        } else {
            // 包装为内存队列生产者，后台SendOneWay，满则丢弃（压测优先吞吐）
            eventProducer = consumer.NewQueuedFlashSaleEventProducer(producer, 8, 100000)
            log.Info("RocketMQ事件生产者初始化成功")
        }
    }
	
	// 创建重试管理器
    var retryManager *consumer.RetryManager
    if rocketmqOpts != nil {
        retryMgr, err := consumer.NewRetryManager(
            rocketmqOpts.NameServers,
            "coupon-retry-group",
            rocketmqOpts.Topic,
            redisClient,
            5, // 最大重试次数
        )
        if err != nil {
            log.Errorf("初始化重试管理器失败: %v", err)
        } else {
            retryManager = retryMgr
            log.Info("重试管理器初始化成功")
        }
    }
    
    // 同时尝试创建事务生产者（方案A）
    var txnProducer consumer.FlashSaleTxnProducer
    if rocketmqOpts != nil {
        if p, err := consumer.NewFlashSaleTxnProducer(
            data, redisClient, rocketmqOpts.NameServers, "coupon-txn-group", rocketmqOpts.Topic,
        ); err != nil {
            log.Errorf("初始化事务生产者失败: %v", err)
        } else {
            txnProducer = p
            log.Info("事务生产者初始化成功")
        }
    }

    // 使用普通RocketMQ生产者（非事务）
    finalEventProducer := eventProducer
    
    service := &Service{
        CouponSrv:           NewCouponService(data, redisClient, cacheManager),
        FlashSaleCore:       NewFlashSaleSrvCore(
            data, redisClient, cacheManager, finalEventProducer, txnProducer,
            bizOpts != nil && bizOpts.FlashSale != nil && bizOpts.FlashSale.SkipUserLimitForTest,
        ),
        CacheManager:        cacheManager,
        EventProducer:       eventProducer,
        TxnProducer:         txnProducer,
        RetryManager:        retryManager,
        data:                data,
    }

    if bizOpts != nil && bizOpts.FlashSale != nil && bizOpts.FlashSale.EnableAsync {
        // 方案A优先（事务生产者）；否则退化为普通生产者
        service.asyncFlashSale = (txnProducer != nil) || (finalEventProducer != nil)
    }

    return service
}

// AsyncFlashSaleEnabled 返回是否启用异步秒杀链路
func (s *Service) AsyncFlashSaleEnabled() bool {
    return s != nil && s.asyncFlashSale && s.FlashSaleCore != nil
}

// UsingTxnFlashSale 是否使用事务消息方案
func (s *Service) UsingTxnFlashSale() bool {
    return s != nil && s.TxnProducer != nil
}

// ShouldUseAsync 根据活动配置+全局可用性判断是否使用异步链路
// 规则：
// 1) 活动 async_enabled=0 → 同步
// 2) async_enabled=1 且全局 AsyncFlashSaleEnabled()=true → 异步
// 3) 其它情况（如MQ未就绪）→ 同步（降级）
func (s *Service) ShouldUseAsync(ctx context.Context, activityID int64) bool {
    // 高并发路径避免 DB 访问，依据全局可用性判定
    return s.AsyncFlashSaleEnabled()
}

// 管理：配置读写
func (s *Service) GetManageConfig(ctx context.Context, key string) (value string, source string, err error) {
    if s == nil || s.data == nil {
        return "", "", fmt.Errorf("service not initialized")
    }
    cfg, e := s.data.CouponConfigs().Get(ctx, s.data.DB(), key)
    if e != nil { return "", "", e }
    if cfg == nil { return "", "", fmt.Errorf("config not found") }
    return cfg.ConfigValue, "db", nil
}

func (s *Service) SetManageConfig(ctx context.Context, key, value, desc string) error {
    if s == nil || s.data == nil {
        return fmt.Errorf("service not initialized")
    }
    return s.data.CouponConfigs().Set(ctx, s.data.DB(), key, value, desc)
}

// Shutdown 优雅关闭服务
func (s *Service) Shutdown() error {
    log.Info("正在关闭优惠券服务...")
    
    // 关闭RocketMQ事件生产者
    if s.EventProducer != nil {
        if err := s.EventProducer.Shutdown(); err != nil {
            log.Errorf("关闭RocketMQ事件生产者失败: %v", err)
        }
    }
	
	// 关闭重试管理器
	if s.RetryManager != nil {
		if err := s.RetryManager.Shutdown(); err != nil {
			log.Errorf("关闭重试管理器失败: %v", err)
		}
	}
	
    log.Info("优惠券服务关闭完成")
    return nil
}

// cacheRepositoryAdapter 缓存仓库适配器，将数据层接口适配为缓存需要的接口
type cacheRepositoryAdapter struct {
	data interfaces.DataFactory
}

func newCacheRepositoryAdapter(data interfaces.DataFactory) cache.CouponRepository {
	return &cacheRepositoryAdapter{data: data}
}

func (c *cacheRepositoryAdapter) GetCouponTemplate(ctx context.Context, couponID int64) (*cache.CouponTemplate, error) {
	templateDO, err := c.data.CouponTemplates().Get(ctx, c.data.DB(), couponID)
	if err != nil {
		return nil, err
	}
	if templateDO == nil {
		return nil, nil
	}
	
	return &cache.CouponTemplate{
		ID:            templateDO.ID,
		Name:          templateDO.Name,
		Type:          int32(templateDO.Type),
		DiscountType:  int32(templateDO.DiscountType),
		DiscountValue: templateDO.DiscountValue,
		MinAmount:     templateDO.MinOrderAmount,
		TotalCount:    templateDO.TotalCount,
		UsedCount:     templateDO.UsedCount,
		ValidStart:    templateDO.ValidStartTime,
		ValidEnd:      templateDO.ValidEndTime,
		Status:        int32(templateDO.Status),
	}, nil
}

func (c *cacheRepositoryAdapter) GetHotCouponTemplates(ctx context.Context, limit int) ([]*cache.CouponTemplate, error) {
	// 查询热门优惠券模板 (按使用次数排序)
	meta := v1.ListMeta{Page: 1, PageSize: limit}
	templateListDO, err := c.data.CouponTemplates().List(ctx, c.data.DB(), 0, meta, []string{"used_count DESC", "created_at DESC"})
	if err != nil {
		return nil, err
	}
	
	templates := make([]*cache.CouponTemplate, 0, len(templateListDO.Items))
	for _, templateDO := range templateListDO.Items {
		templates = append(templates, &cache.CouponTemplate{
			ID:            templateDO.ID,
			Name:          templateDO.Name,
			Type:          int32(templateDO.Type),
			DiscountType:  int32(templateDO.DiscountType),
			DiscountValue: templateDO.DiscountValue,
			MinAmount:     templateDO.MinOrderAmount,
			TotalCount:    templateDO.TotalCount,
			UsedCount:     templateDO.UsedCount,
			ValidStart:    templateDO.ValidStartTime,
			ValidEnd:      templateDO.ValidEndTime,
			Status:        int32(templateDO.Status),
		})
	}
	
	return templates, nil
}

func (c *cacheRepositoryAdapter) GetUserCoupon(ctx context.Context, userCouponID int64) (*cache.UserCoupon, error) {
	userCouponDO, err := c.data.UserCoupons().Get(ctx, c.data.DB(), userCouponID)
	if err != nil {
		return nil, err
	}
	if userCouponDO == nil {
		return nil, nil
	}
	
	return &cache.UserCoupon{
		ID:             userCouponDO.ID,
		CouponID:       userCouponDO.CouponTemplateID,
		UserID:         userCouponDO.UserID,
		CouponSn:       userCouponDO.CouponCode,
		Status:         int32(userCouponDO.Status),
		ObtainTime:     userCouponDO.ReceivedAt,
		ValidStartTime: userCouponDO.ReceivedAt, // 这里简化处理，实际可能需要计算
		ValidEndTime:   userCouponDO.ExpiredAt,
	}, nil
}

func (c *cacheRepositoryAdapter) GetFlashSaleActivity(ctx context.Context, activityID int64) (*cache.FlashSaleActivity, error) {
	activityDO, err := c.data.FlashSales().Get(ctx, c.data.DB(), activityID)
	if err != nil {
		return nil, err
	}
	if activityDO == nil {
		return nil, nil
	}
	
	return &cache.FlashSaleActivity{
		ID:           activityDO.ID,
		CouponID:     activityDO.CouponTemplateID,
		Name:         activityDO.Name,
		TotalCount:   activityDO.FlashSaleCount,
		SuccessCount: activityDO.SoldCount,
		StartTime:    activityDO.StartTime,
		EndTime:      activityDO.EndTime,
		Status:       int32(activityDO.Status),
	}, nil
}
