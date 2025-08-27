package v1

import (
	"context"
	"emshop/internal/app/coupon/srv/consumer"
	"emshop/internal/app/coupon/srv/data/v1/interfaces"
	"emshop/internal/app/coupon/srv/pkg/cache"
	"emshop/internal/app/pkg/options"
	"emshop/pkg/log"
	v1 "emshop/pkg/common/meta/v1"
	"github.com/go-redis/redis/v8"
)

// Service 优惠券服务工厂
type Service struct {
	CouponSrv           CouponSrv
	FlashSaleSrv        FlashSaleSrv
	FlashSaleCore       FlashSaleSrvCore  // 新的秒杀核心服务
	DTMManager          *CouponDTMManager
	CacheManager        cache.CacheManager
	EventProducer       consumer.FlashSaleEventProducer // RocketMQ事件生产者
	TransactionProducer *consumer.TransactionFlashSaleEventProducer // 事务消息生产者
	RetryManager        *consumer.RetryManager // 重试管理器
}

// NewService 创建优惠券服务工厂
func NewService(data interfaces.DataFactory, redisClient *redis.Client, dtmOpts *options.DtmOptions, rocketmqOpts *options.RocketMQOptions, cacheConfig *cache.CacheConfig) *Service {
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
		)
		if err != nil {
			log.Errorf("初始化RocketMQ事件生产者失败: %v", err)
			// 可以考虑使用fallback或mock实现
			eventProducer = nil
		} else {
			eventProducer = producer
			log.Info("RocketMQ事件生产者初始化成功")
		}
	}
	
	// 创建事务消息生产者
	var transactionProducer *consumer.TransactionFlashSaleEventProducer
	if rocketmqOpts != nil && eventProducer != nil {
		txnConfig := &consumer.TransactionConfig{
			NameServers: rocketmqOpts.NameServers,
			GroupName:   "coupon-txn-producer-group",
			Topic:       rocketmqOpts.Topic,
		}
		
		txnProducer, err := consumer.NewTransactionFlashSaleEventProducer(
			txnConfig, data, redisClient, eventProducer,
		)
		if err != nil {
			log.Errorf("初始化事务消息生产者失败: %v", err)
		} else {
			transactionProducer = txnProducer
			log.Info("事务消息生产者初始化成功")
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
	
	// 优先使用事务消息生产者，fallback到普通生产者
	finalEventProducer := eventProducer
	if transactionProducer != nil {
		finalEventProducer = transactionProducer
		log.Info("使用事务消息生产者作为主要事件生产者")
	}
	
	service := &Service{
		CouponSrv:           NewCouponService(data, redisClient, dtmOpts, cacheManager),
		FlashSaleSrv:        NewFlashSaleService(data, redisClient, cacheManager),
		FlashSaleCore:       NewFlashSaleSrvCore(data, redisClient, cacheManager, finalEventProducer),
		CacheManager:        cacheManager,
		EventProducer:       eventProducer,
		TransactionProducer: transactionProducer,
		RetryManager:        retryManager,
	}
	
	// 创建DTM管理器，传入服务实例用于TCC回调
	service.DTMManager = NewCouponDTMManager(dtmOpts, service)
	
	return service
}

// Shutdown 优雅关闭服务
func (s *Service) Shutdown() error {
	log.Info("正在关闭优惠券服务...")
	
	// 关闭事务消息生产者
	if s.TransactionProducer != nil {
		if err := s.TransactionProducer.Shutdown(); err != nil {
			log.Errorf("关闭事务消息生产者失败: %v", err)
		}
	}
	
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