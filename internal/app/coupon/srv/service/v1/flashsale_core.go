package v1

import (
    "context"
    "fmt"
    "strconv"
    "strings"
    "time"
    "sync"

    "emshop/internal/app/coupon/srv/consumer"
    "emshop/internal/app/coupon/srv/data/v1/interfaces"
    "emshop/internal/app/coupon/srv/data/v1/redis"
    "emshop/internal/app/coupon/srv/domain/do"
    "emshop/internal/app/coupon/srv/domain/dto"
    "emshop/internal/app/coupon/srv/pkg/cache"
    "emshop/internal/app/coupon/srv/pkg/scripts"
    "emshop/pkg/log"
    redisClient "github.com/go-redis/redis/v8"
    metav1 "emshop/pkg/common/meta/v1"
)

// FlashSaleSrvCore 秒杀服务核心接口
type FlashSaleSrvCore interface {
    // 秒杀核心功能
    StartFlashSaleActivity(ctx context.Context, req *dto.StartFlashSaleDTO) error
    StopFlashSaleActivity(ctx context.Context, req *dto.StopFlashSaleDTO) error
    FlashSaleCoupon(ctx context.Context, req *dto.FlashSaleRequestDTO) (*dto.FlashSaleResultDTO, error)
    GetFlashSaleStatus(ctx context.Context, req *dto.FlashSaleStatusDTO) (*dto.FlashSaleStatusResultDTO, error)
    
    // 管理功能
    CreateFlashSaleActivity(ctx context.Context, req *dto.CreateFlashSaleActivityDTO) (*dto.FlashSaleActivityDTO, error)
    UpdateFlashSaleActivity(ctx context.Context, req *dto.UpdateFlashSaleActivityDTO) error
    GetFlashSaleActivity(ctx context.Context, activityID int64) (*dto.FlashSaleActivityDTO, error)
    ListFlashSaleActivities(ctx context.Context, req *dto.ListFlashSaleActivitiesDTO) (*dto.FlashSaleActivityListDTO, error)

    // 查询功能
    GetFlashSaleStock(ctx context.Context, flashSaleID int64) (*dto.FlashSaleStockDTO, error)
    GetUserFlashSaleRecords(ctx context.Context, req *dto.GetUserFlashSaleRecordsDTO) (*dto.FlashSaleRecordListDTO, error)
}

// flashSaleSrvCore 秒杀服务核心实现
type flashSaleSrvCore struct {
    data          interfaces.DataFactory
    redisClient   *redisClient.Client
    cacheManager  cache.CacheManager
    stockManager  *redis.StockManager
    eventProducer consumer.FlashSaleEventProducer
    txnProducer   consumer.FlashSaleTxnProducer
    skipUserLimit bool
    // db-config cache
    lastSkipCheck time.Time
    cachedSkip    bool

    // per_user_limit 本地3秒缓存，减少DB读取
    perLimitCache map[int64]int32
    perLimitLast  map[int64]time.Time
    perLimitMu    sync.RWMutex
}

// NewFlashSaleSrvCore 创建秒杀服务核心
func NewFlashSaleSrvCore(data interfaces.DataFactory, redisClient *redisClient.Client, cacheManager cache.CacheManager, eventProducer consumer.FlashSaleEventProducer, txnProducer consumer.FlashSaleTxnProducer, skipUserLimit bool) FlashSaleSrvCore {
    return &flashSaleSrvCore{
        data:          data,
        redisClient:   redisClient,
        cacheManager:  cacheManager,
        stockManager:  redis.NewStockManager(redisClient),
        eventProducer: eventProducer,
        txnProducer:   txnProducer,
        skipUserLimit: skipUserLimit,
    }
}

// StartFlashSaleActivity 启动秒杀活动
func (fss *flashSaleSrvCore) StartFlashSaleActivity(ctx context.Context, req *dto.StartFlashSaleDTO) error {
	log.Infof("启动秒杀活动: activityID=%d", req.ActivityID)

	// 获取活动信息
	activityDO, err := fss.data.FlashSales().Get(ctx, fss.data.DB(), req.ActivityID)
	if err != nil {
		return fmt.Errorf("获取活动信息失败: %v", err)
	}
	if activityDO == nil {
		return fmt.Errorf("活动不存在")
	}

    // 检查活动状态：允许 Pending 或已 Active 幂等启动；Finished 禁止
    if activityDO.Status == do.FlashSaleStatusFinished {
        return fmt.Errorf("活动状态不允许启动")
    }

	// 获取优惠券模板信息
	templateDO, err := fss.data.CouponTemplates().Get(ctx, fss.data.DB(), activityDO.CouponTemplateID)
	if err != nil {
		return fmt.Errorf("获取优惠券模板失败: %v", err)
	}
	if templateDO == nil {
		return fmt.Errorf("优惠券模板不存在")
	}

	// 创建Redis活动信息
	activityInfo := &redis.ActivityInfo{
		ID:           activityDO.ID,
		CouponID:     activityDO.CouponTemplateID,
		Status:       int32(do.FlashSaleStatusActive),
		StartTime:    activityDO.StartTime,
		EndTime:      activityDO.EndTime,
		TotalCount:   activityDO.FlashSaleCount,
		SuccessCount: 0,
		PerUserLimit: activityDO.PerUserLimit,
	}

	// 启动Redis秒杀活动（预热库存）
	if err := fss.stockManager.StartActivity(ctx, activityInfo); err != nil {
		return fmt.Errorf("启动Redis活动失败: %v", err)
	}

	// 更新数据库活动状态
	updateData := &do.FlashSaleActivityDO{
		ID:     activityDO.ID,
		Status: do.FlashSaleStatusActive,
	}
	if err := fss.data.FlashSales().Update(ctx, fss.data.DB(), updateData); err != nil {
		// 如果数据库更新失败，回滚Redis状态
		fss.stockManager.StopActivity(ctx, activityDO.ID)
		return fmt.Errorf("更新活动状态失败: %v", err)
	}

	log.Infof("秒杀活动启动成功: activityID=%d, couponID=%d", req.ActivityID, activityDO.CouponTemplateID)
	return nil
}

// StopFlashSaleActivity 停止秒杀活动
func (fss *flashSaleSrvCore) StopFlashSaleActivity(ctx context.Context, req *dto.StopFlashSaleDTO) error {
	log.Infof("停止秒杀活动: activityID=%d", req.ActivityID)

	// 获取活动信息
	activityDO, err := fss.data.FlashSales().Get(ctx, fss.data.DB(), req.ActivityID)
	if err != nil {
		return fmt.Errorf("获取活动信息失败: %v", err)
	}
	if activityDO == nil {
		return fmt.Errorf("活动不存在")
	}

	// 停止Redis活动
	if err := fss.stockManager.StopActivity(ctx, req.ActivityID); err != nil {
		log.Errorf("停止Redis活动失败: %v", err)
	}

	// 更新数据库活动状态
	updateData := &do.FlashSaleActivityDO{
		ID:     activityDO.ID,
		Status: do.FlashSaleStatusFinished,
	}
	if err := fss.data.FlashSales().Update(ctx, fss.data.DB(), updateData); err != nil {
		return fmt.Errorf("更新活动状态失败: %v", err)
	}

	// 如果需要，可以选择清理活动数据
	if req.CleanupData {
		go func() {
			// 异步清理，避免影响响应时间
			time.Sleep(5 * time.Second)
			fss.stockManager.ClearActivityData(context.Background(), req.ActivityID, activityDO.CouponTemplateID)
		}()
	}

	log.Infof("秒杀活动停止成功: activityID=%d", req.ActivityID)
	return nil
}

// FlashSaleCoupon 执行秒杀
func (fss *flashSaleSrvCore) FlashSaleCoupon(ctx context.Context, req *dto.FlashSaleRequestDTO) (*dto.FlashSaleResultDTO, error) {
    log.Infof("执行秒杀请求(事务模式): activityID=%d, userID=%d", req.ActivityID, req.UserID)

    if fss.txnProducer == nil {
        return nil, fmt.Errorf("事务消息生产者未配置")
    }

    // 统一生成并注入同一个 request_id，贯穿 Redis/消息/落库
    rid := getRequestID(ctx)
    ctx = context.WithValue(ctx, "request_id", rid)

	// 获取活动信息（优先从Redis获取）
    activityInfo, err := fss.stockManager.GetActivityStatus(ctx, req.ActivityID)
    if err != nil {
        // 高并发路径严格依赖 Redis，不回退 DB，避免打爆数据库
        log.Warnf("Redis未找到活动或读取失败: activityID=%d, err=%v", req.ActivityID, err)
        return &dto.FlashSaleResultDTO{Success: false, Message: "活动数据暂不可用", Code: -3}, nil
    }
	// log.Infof("获取秒杀活动信息: activityID=%d, status=%d, total=%d, success=%d", activityInfo.ID, activityInfo.Status, activityInfo.TotalCount, activityInfo.SuccessCount)

    // 构建秒杀请求：人均限购直接取 Redis 活动信息，避免热点 DB 访问
    perLimit := activityInfo.PerUserLimit

    flashSaleReq := &redis.FlashSaleRequest{
        ActivityID:   req.ActivityID,
        CouponID:     activityInfo.CouponID,
        UserID:       req.UserID,
        RequestCount: 1, // 每次只能秒杀1张
        ClientIP:     req.ClientIP,
        UserAgent:    req.UserAgent,
        RequestID:    rid,
        PerUserLimitOverride: perLimit,
    }

    // 执行Redis预留（不落日志、不增成功数）
    result, err := fss.stockManager.Reserve(ctx, flashSaleReq)
    if err != nil {
        log.Errorf("秒杀执行失败: %v", err)
        return &dto.FlashSaleResultDTO{
            Success: false,
            Message: "系统繁忙，请稍后重试",
            Code:    -4,
        }, nil
    }

	// 构建返回结果
	flashSaleResult := &dto.FlashSaleResultDTO{
		Success:      result.Success,
		Message:      result.Message,
		Code:         result.Code,
		CouponSn:     result.CouponSn,
		RemainStock:  int32(result.RemainStock),
		Timestamp:    result.Timestamp,
	}

    // 如果预留成功，发送事务消息，本地事务内创建用户优惠券，失败则回滚预留
    if result.Success {
        successEvent := &consumer.FlashSaleSuccessEvent{
            ActivityID: req.ActivityID,
            CouponID:   activityInfo.CouponID,
            UserID:     req.UserID,
            CouponSn:   result.CouponSn,
            ClientIP:   req.ClientIP,
            UserAgent:  req.UserAgent,
            Timestamp:  time.Now().Unix(),
            RequestID:  rid,
        }
        // 发送事务消息（会在本地事务内建券并决定提交/回滚）；若不可用则内联落库回退
        if fss.txnProducer == nil {
            // 无事务消息能力，回滚预留并失败返回
            _ = fss.stockManager.RollbackStock(ctx, req.ActivityID, req.UserID, activityInfo.CouponID, 1)
            log.Errorf("事务消息生产者未配置，已回滚预留")
            return &dto.FlashSaleResultDTO{Success: false, Message: "系统繁忙，请稍后重试", Code: -4}, nil
        }
        if err := fss.txnProducer.SendFlashSaleSuccessTxn(ctx, successEvent); err != nil {
            // 发送失败回滚预留
            _ = fss.stockManager.RollbackStock(ctx, req.ActivityID, req.UserID, activityInfo.CouponID, 1)
            log.Errorf("事务消息发送失败，已回滚预留: %v", err)
            return &dto.FlashSaleResultDTO{Success: false, Message: "系统繁忙，请稍后重试", Code: -4}, nil
        }
        // log.Infof("预留成功并触发事务提交: userID=%d, activityID=%d, couponSn=%s", req.UserID, req.ActivityID, result.CouponSn)
    } else {
        // 失败分支：不做额外处理
    }

	return flashSaleResult, nil
}

// inlineCreateOrder 在无事务消息可用或发送失败时的降级方案：
// 直接本地事务创建“秒杀订单”并增加 sold_count。
func (fss *flashSaleSrvCore) inlineCreateOrder(ctx context.Context, evt *consumer.FlashSaleSuccessEvent) error {
    if fss == nil || fss.data == nil {
        return fmt.Errorf("service not initialized")
    }
    tx := fss.data.DB().Begin()
    if tx.Error != nil { return tx.Error }
    order := &do.FlashSaleOrderDO{
        OrderSn:     fmt.Sprintf("FSO-%d-%s", time.Now().Unix(), evt.RequestID),
        RequestID:   evt.RequestID,
        UserID:      evt.UserID,
        FlashSaleID: evt.ActivityID,
        CouponID:    evt.CouponID,
        Status:      "CREATED",
        Amount:      0,
    }
    if err := fss.data.FlashSaleOrders().Create(ctx, tx, order); err != nil {
        tx.Rollback()
        return err
    }
    if err := fss.data.FlashSales().IncrementSoldCount(ctx, tx, evt.ActivityID); err != nil {
        tx.Rollback()
        return err
    }
    return tx.Commit().Error
}

// GetFlashSaleStock 获取秒杀库存（优先Redis，回退DB）
func (fss *flashSaleSrvCore) GetFlashSaleStock(ctx context.Context, flashSaleID int64) (*dto.FlashSaleStockDTO, error) {
    // 先从Redis获取
    stockKey := scripts.NewRedisKeyFormatter().FlashSaleStockKey(flashSaleID)
    stockInt, err := fss.redisClient.Get(ctx, stockKey).Int()
    var remainingStock int32
    if err == redisClient.Nil {
        // Redis中没有，从数据库获取
        stockInfo, err := fss.data.FlashSales().CheckStock(ctx, fss.data.DB(), flashSaleID)
        if err != nil {
            return nil, fmt.Errorf("获取秒杀库存失败: %v", err)
        }
        remainingStock = stockInfo.RemainingStock
    } else if err != nil {
        return nil, fmt.Errorf("获取Redis库存失败: %v", err)
    } else {
        remainingStock = int32(stockInt)
    }

    // 获取活动信息
    activityDO, err := fss.data.FlashSales().Get(ctx, fss.data.DB(), flashSaleID)
    if err != nil || activityDO == nil {
        return nil, fmt.Errorf("秒杀活动不存在")
    }

    return &dto.FlashSaleStockDTO{
        FlashSaleID:    flashSaleID,
        TotalStock:     activityDO.FlashSaleCount,
        RemainingStock: remainingStock,
        SoldCount:      activityDO.SoldCount,
    }, nil
}

// GetUserFlashSaleRecords 获取用户秒杀记录
func (fss *flashSaleSrvCore) GetUserFlashSaleRecords(ctx context.Context, req *dto.GetUserFlashSaleRecordsDTO) (*dto.FlashSaleRecordListDTO, error) {
    // 使用数据层方法
    meta := metav1.ListMeta{Page: int(req.Page), PageSize: int(req.PageSize)}
    recordListDO, err := fss.data.FlashSaleRecords().GetUserRecords(ctx, fss.data.DB(), req.UserID, meta)
    if err != nil {
        return nil, fmt.Errorf("获取用户秒杀记录失败: %v", err)
    }

    // 转换为DTO
    items := make([]*dto.FlashSaleRecordDTO, 0, len(recordListDO.Items))
    for _, recordDO := range recordListDO.Items {
        items = append(items, &dto.FlashSaleRecordDTO{
            ID:             recordDO.ID,
            FlashSaleID:    recordDO.FlashSaleID,
            UserID:         recordDO.UserID,
            UserCouponID:   recordDO.UserCouponID,
            Status:         int32(recordDO.Status),
            FailReason:     nil, // 简化：保留 nil
            CreatedAt:      recordDO.CreatedAt,
            OrderCreatedAt: recordDO.OrderCreatedAt,
        })
    }

    return &dto.FlashSaleRecordListDTO{
        TotalCount: recordListDO.TotalCount,
        Items:      items,
    }, nil
}

// getRequestID 安全地从context中获取request_id
func getRequestID(ctx context.Context) string {
	if requestID := ctx.Value("request_id"); requestID != nil {
		if rid, ok := requestID.(string); ok {
			return rid
		}
	}
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// getSkipUserLimit 优先读取数据库配置（coupon_configs.flashsale_skip_user_limit_for_test），
// 否则回退到YAML开关；结果缓存3s以降低DB压力。
func (fss *flashSaleSrvCore) getSkipUserLimit(ctx context.Context) bool {
    if time.Since(fss.lastSkipCheck) < 3*time.Second {
        return fss.cachedSkip
    }
    // 1) 先查Redis一级缓存
    const redisKey = "coupon:config:flashsale_skip_user_limit_for_test"
    if fss.redisClient != nil {
        if val, err := fss.redisClient.Get(ctx, redisKey).Result(); err == nil && val != "" {
            v := strings.TrimSpace(strings.ToLower(val))
            skip := (v == "1" || v == "true" || v == "yes")
            fss.cachedSkip = skip
            fss.lastSkipCheck = time.Now()
            return skip
        }
    }
    // 2) 查DB，回写Redis（60s）
    skip := fss.skipUserLimit // YAML默认
    if fss.data != nil {
        if cfg, err := fss.data.CouponConfigs().Get(ctx, fss.data.DB(), "flashsale_skip_user_limit_for_test"); err == nil && cfg != nil {
            v := strings.TrimSpace(strings.ToLower(cfg.ConfigValue))
            if v == "1" || v == "true" || v == "yes" {
                skip = true
            }
            if v == "0" || v == "false" || v == "no" {
                skip = false
            }
            if fss.redisClient != nil {
                _ = fss.redisClient.SetEX(ctx, redisKey, strconv.FormatBool(skip), 60*time.Second).Err()
            }
        }
    }
    fss.cachedSkip = skip
    fss.lastSkipCheck = time.Now()
    return skip
}

// GetFlashSaleStatus 获取秒杀状态
func (fss *flashSaleSrvCore) GetFlashSaleStatus(ctx context.Context, req *dto.FlashSaleStatusDTO) (*dto.FlashSaleStatusResultDTO, error) {
	// 优先从Redis获取实时状态
	activityInfo, err := fss.stockManager.GetActivityStatus(ctx, req.ActivityID)
	if err != nil {
		// Redis获取失败，从数据库获取
		activityDO, dbErr := fss.data.FlashSales().Get(ctx, fss.data.DB(), req.ActivityID)
		if dbErr != nil || activityDO == nil {
			return nil, fmt.Errorf("活动不存在")
		}

		return &dto.FlashSaleStatusResultDTO{
			ActivityID:   activityDO.ID,
			CouponID:     activityDO.CouponTemplateID,
			Status:       int32(activityDO.Status),
			TotalCount:   activityDO.FlashSaleCount,
			SuccessCount: activityDO.SoldCount,
			RemainStock:  activityDO.FlashSaleCount - activityDO.SoldCount,
			StartTime:    activityDO.StartTime,
			EndTime:      activityDO.EndTime,
		}, nil
	}

	// 获取实时库存
	currentStock, err := fss.stockManager.GetCurrentStock(ctx, activityInfo.CouponID)
	if err != nil {
		log.Errorf("获取实时库存失败: %v", err)
		currentStock = 0
	}

	// 获取用户参与状态（如果提供了用户ID）
	var userParticipated bool
	var userParticipationCount int32
	if req.UserID > 0 {
		count, err := fss.stockManager.GetUserParticipationCount(ctx, req.ActivityID, req.UserID)
		if err == nil {
			userParticipationCount = count
			userParticipated = count > 0
		}
	}

	return &dto.FlashSaleStatusResultDTO{
		ActivityID:             activityInfo.ID,
		CouponID:               activityInfo.CouponID,
		Status:                 activityInfo.Status,
		TotalCount:             activityInfo.TotalCount,
		SuccessCount:           activityInfo.SuccessCount,
		RemainStock:            currentStock,
		StartTime:              activityInfo.StartTime,
		EndTime:                activityInfo.EndTime,
		UserParticipated:       userParticipated,
		UserParticipationCount: userParticipationCount,
	}, nil
}

// CreateFlashSaleActivity 创建秒杀活动
func (fss *flashSaleSrvCore) CreateFlashSaleActivity(ctx context.Context, req *dto.CreateFlashSaleActivityDTO) (*dto.FlashSaleActivityDTO, error) {
	log.Infof("创建秒杀活动: couponID=%d, name=%s", req.CouponTemplateID, req.Name)

	// 验证优惠券模板是否存在
	templateDO, err := fss.data.CouponTemplates().Get(ctx, fss.data.DB(), req.CouponTemplateID)
	if err != nil {
		return nil, fmt.Errorf("获取优惠券模板失败: %v", err)
	}
	if templateDO == nil {
		return nil, fmt.Errorf("优惠券模板不存在")
	}

	// 创建活动DO
	activityDO := &do.FlashSaleActivityDO{
		CouponTemplateID: req.CouponTemplateID,
		Name:             req.Name,
		FlashSaleCount:   req.FlashSaleCount,
		PerUserLimit:     req.PerUserLimit,
		StartTime:        req.StartTime,
		EndTime:          req.EndTime,
		Status:           do.FlashSaleStatusPending, // 待开始
		SoldCount:        0,
	}

	// 保存到数据库
	if err := fss.data.FlashSales().Create(ctx, fss.data.DB(), activityDO); err != nil {
		return nil, fmt.Errorf("创建活动失败: %v", err)
	}

	// 构建返回结果
	result := &dto.FlashSaleActivityDTO{
		ID:               activityDO.ID,
		CouponTemplateID: activityDO.CouponTemplateID,
		CouponID:         activityDO.CouponTemplateID, // 兼容字段
		Name:             activityDO.Name,
		FlashSaleCount:   activityDO.FlashSaleCount,
		PerUserLimit:     activityDO.PerUserLimit,
		StartTime:        activityDO.StartTime,
		EndTime:          activityDO.EndTime,
		Status:           int32(activityDO.Status),
		SoldCount:        activityDO.SoldCount,
		CreatedAt:        activityDO.CreatedAt,
		UpdatedAt:        activityDO.UpdatedAt,
	}

	// 填充优惠券模板信息
	result.CouponName = templateDO.Name
	result.CouponType = int32(templateDO.Type)
	result.DiscountValue = templateDO.DiscountValue

	log.Infof("秒杀活动创建成功: activityID=%d", activityDO.ID)
	return result, nil
}

// UpdateFlashSaleActivity 更新秒杀活动
func (fss *flashSaleSrvCore) UpdateFlashSaleActivity(ctx context.Context, req *dto.UpdateFlashSaleActivityDTO) error {
	// 获取现有活动
	activityDO, err := fss.data.FlashSales().Get(ctx, fss.data.DB(), req.ID)
	if err != nil {
		return fmt.Errorf("获取活动信息失败: %v", err)
	}
	if activityDO == nil {
		return fmt.Errorf("活动不存在")
	}

	// 检查是否可以更新（只有待开始状态才能更新）
	if activityDO.Status != do.FlashSaleStatusPending {
		return fmt.Errorf("活动已开始或结束，无法更新")
	}

	// 构建更新数据
	updateData := &do.FlashSaleActivityDO{
		ID: req.ID,
	}

	// 更新字段
	if req.Name != "" {
		updateData.Name = req.Name
	}
	if req.FlashSaleCount > 0 {
		updateData.FlashSaleCount = req.FlashSaleCount
	}
	if req.PerUserLimit > 0 {
		updateData.PerUserLimit = req.PerUserLimit
	}
	if !req.StartTime.IsZero() {
		updateData.StartTime = req.StartTime
	}
	if !req.EndTime.IsZero() {
		updateData.EndTime = req.EndTime
	}

	// 更新数据库
	if err := fss.data.FlashSales().Update(ctx, fss.data.DB(), updateData); err != nil {
		return fmt.Errorf("更新活动失败: %v", err)
	}

	log.Infof("秒杀活动更新成功: activityID=%d", req.ID)
	return nil
}

// GetFlashSaleActivity 获取秒杀活动详情
func (fss *flashSaleSrvCore) GetFlashSaleActivity(ctx context.Context, activityID int64) (*dto.FlashSaleActivityDTO, error) {
	// 从数据库获取活动信息
	activityDO, err := fss.data.FlashSales().Get(ctx, fss.data.DB(), activityID)
	if err != nil {
		return nil, fmt.Errorf("获取活动信息失败: %v", err)
	}
	if activityDO == nil {
		return nil, fmt.Errorf("活动不存在")
	}

	// 获取优惠券模板信息
	templateDO, err := fss.data.CouponTemplates().Get(ctx, fss.data.DB(), activityDO.CouponTemplateID)
	if err != nil {
		log.Errorf("获取优惠券模板失败: %v", err)
	}

	// 构建返回结果
	result := &dto.FlashSaleActivityDTO{
		ID:               activityDO.ID,
		CouponTemplateID: activityDO.CouponTemplateID,
		CouponID:         activityDO.CouponTemplateID, // 兼容字段
		Name:             activityDO.Name,
		FlashSaleCount:   activityDO.FlashSaleCount,
		PerUserLimit:     activityDO.PerUserLimit,
		StartTime:        activityDO.StartTime,
		EndTime:          activityDO.EndTime,
		Status:           int32(activityDO.Status),
		SoldCount:        activityDO.SoldCount,
		CreatedAt:        activityDO.CreatedAt,
		UpdatedAt:        activityDO.UpdatedAt,
	}

	// 填充优惠券模板信息
	if templateDO != nil {
		result.CouponName = templateDO.Name
		result.CouponType = int32(templateDO.Type)
		result.DiscountValue = templateDO.DiscountValue
	}

    // 如果活动正在进行中，获取实时库存（售出数以DB为准，避免事务模式下Redis success_count未维护导致不一致）
    if activityDO.Status == do.FlashSaleStatusActive {
        if currentStock, err := fss.stockManager.GetCurrentStock(ctx, activityDO.CouponTemplateID); err == nil {
            result.RemainStock = currentStock
        }
    }

	return result, nil
}

// ListFlashSaleActivities 获取秒杀活动列表
func (fss *flashSaleSrvCore) ListFlashSaleActivities(ctx context.Context, req *dto.ListFlashSaleActivitiesDTO) (*dto.FlashSaleActivityListDTO, error) {
	// 构建查询条件
	meta := req.ListMeta
	if meta.PageSize == 0 {
		meta.PageSize = 20
	}
	if meta.PageSize > 100 {
		meta.PageSize = 100
	}

	// 转换状态类型
	var status do.FlashSaleStatus
	if req.Status != nil {
		status = do.FlashSaleStatus(*req.Status)
	}

	// 从数据库查询
	listDO, err := fss.data.FlashSales().List(ctx, fss.data.DB(), status, meta, nil)
	if err != nil {
		return nil, fmt.Errorf("查询活动列表失败: %v", err)
	}

	// 转换为DTO
	activities := make([]*dto.FlashSaleActivityDTO, 0, len(listDO.Items))
	for _, activityDO := range listDO.Items {
		activityDTO := &dto.FlashSaleActivityDTO{
			ID:               activityDO.ID,
			CouponTemplateID: activityDO.CouponTemplateID,
			CouponID:         activityDO.CouponTemplateID, // 兼容字段
			Name:             activityDO.Name,
			FlashSaleCount:   activityDO.FlashSaleCount,
			PerUserLimit:     activityDO.PerUserLimit,
			StartTime:        activityDO.StartTime,
			EndTime:          activityDO.EndTime,
			Status:           int32(activityDO.Status),
			SoldCount:        activityDO.SoldCount,
			CreatedAt:        activityDO.CreatedAt,
			UpdatedAt:        activityDO.UpdatedAt,
		}

		// 如果是进行中的活动，获取实时数据
		if activityDO.Status == do.FlashSaleStatusActive {
			if currentStock, err := fss.stockManager.GetCurrentStock(ctx, activityDO.CouponTemplateID); err == nil {
				activityDTO.RemainStock = currentStock
			}
		}

		activities = append(activities, activityDTO)
	}

	return &dto.FlashSaleActivityListDTO{
		Items:      activities,
		TotalCount: listDO.TotalCount,
		ListMeta:   meta,
	}, nil
}
