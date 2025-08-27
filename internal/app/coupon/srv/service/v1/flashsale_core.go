package v1

import (
	"context"
	"fmt"
	"time"

	"emshop/internal/app/coupon/srv/consumer"
	"emshop/internal/app/coupon/srv/data/v1/interfaces"
	"emshop/internal/app/coupon/srv/data/v1/redis"
	"emshop/internal/app/coupon/srv/domain/do"
	"emshop/internal/app/coupon/srv/domain/dto"
	"emshop/internal/app/coupon/srv/pkg/cache"
	"emshop/pkg/log"
	redisClient "github.com/go-redis/redis/v8"
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
}

// flashSaleSrvCore 秒杀服务核心实现
type flashSaleSrvCore struct {
	data          interfaces.DataFactory
	redisClient   *redisClient.Client
	cacheManager  cache.CacheManager
	stockManager  *redis.StockManager
	eventProducer consumer.FlashSaleEventProducer
}

// NewFlashSaleSrvCore 创建秒杀服务核心
func NewFlashSaleSrvCore(data interfaces.DataFactory, redisClient *redisClient.Client, cacheManager cache.CacheManager, eventProducer consumer.FlashSaleEventProducer) FlashSaleSrvCore {
	return &flashSaleSrvCore{
		data:          data,
		redisClient:   redisClient,
		cacheManager:  cacheManager,
		stockManager:  redis.NewStockManager(redisClient),
		eventProducer: eventProducer,
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

	// 检查活动状态
	if activityDO.Status != do.FlashSaleStatusPending {
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
	log.Infof("执行秒杀请求: activityID=%d, userID=%d", req.ActivityID, req.UserID)

	// 获取活动信息（优先从Redis获取）
	activityInfo, err := fss.stockManager.GetActivityStatus(ctx, req.ActivityID)
	if err != nil {
		// 如果Redis中没有，从数据库获取
		activityDO, dbErr := fss.data.FlashSales().Get(ctx, fss.data.DB(), req.ActivityID)
		if dbErr != nil || activityDO == nil {
			return &dto.FlashSaleResultDTO{
				Success: false,
				Message: "活动不存在",
				Code:    -3,
			}, nil
		}
		
		// 检查活动是否可以秒杀
		if activityDO.Status != do.FlashSaleStatusActive {
			return &dto.FlashSaleResultDTO{
				Success: false,
				Message: "活动未开始或已结束",
				Code:    -3,
			}, nil
		}

		// 如果Redis中没有活动，可能需要重新启动
		log.Warnf("Redis中无活动信息，尝试恢复: activityID=%d", req.ActivityID)
		return &dto.FlashSaleResultDTO{
			Success: false,
			Message: "活动数据异常，请稍后重试",
			Code:    -3,
		}, nil
	}

	// 构建秒杀请求
	flashSaleReq := &redis.FlashSaleRequest{
		ActivityID:   req.ActivityID,
		CouponID:     activityInfo.CouponID,
		UserID:       req.UserID,
		RequestCount: 1, // 每次只能秒杀1张
		ClientIP:     req.ClientIP,
		UserAgent:    req.UserAgent,
	}

	// 执行Redis秒杀
	result, err := fss.stockManager.FlashSale(ctx, flashSaleReq)
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

	// 如果秒杀成功，发送异步消息进行后续处理
	if result.Success {
		// 发送秒杀成功事件到RocketMQ
		successEvent := &consumer.FlashSaleSuccessEvent{
			ActivityID: req.ActivityID,
			CouponID:   activityInfo.CouponID,
			UserID:     req.UserID,
			CouponSn:   result.CouponSn,
			ClientIP:   req.ClientIP,
			UserAgent:  req.UserAgent,
			Timestamp:  time.Now().Unix(),
			RequestID:  getRequestID(ctx),
		}
		
		// 异步发送，避免影响响应时间
		go func() {
			if err := fss.eventProducer.SendFlashSaleSuccessEvent(successEvent); err != nil {
				log.Errorf("发送秒杀成功事件失败: %v, userID=%d, activityID=%d", 
					err, req.UserID, req.ActivityID)
				
				// 发送失败事件
				failureEvent := &consumer.FlashSaleFailureEvent{
					ActivityID: req.ActivityID,
					UserID:     req.UserID,
					Reason:     fmt.Sprintf("消息发送失败: %v", err),
					Code:       -5,
					ClientIP:   req.ClientIP,
					UserAgent:  req.UserAgent,
					Timestamp:  time.Now().Unix(),
				}
				fss.eventProducer.SendFlashSaleFailureEvent(failureEvent)
			}
		}()
		
		log.Infof("秒杀成功，已发送异步消息: userID=%d, activityID=%d, couponSn=%s", 
			req.UserID, req.ActivityID, result.CouponSn)
	} else {
		// 秒杀失败，发送失败事件用于统计和监控
		failureEvent := &consumer.FlashSaleFailureEvent{
			ActivityID: req.ActivityID,
			UserID:     req.UserID,
			Reason:     result.Message,
			Code:       result.Code,
			ClientIP:   req.ClientIP,
			UserAgent:  req.UserAgent,
			Timestamp:  time.Now().Unix(),
		}
		
		// 异步发送失败事件
		go func() {
			if err := fss.eventProducer.SendFlashSaleFailureEvent(failureEvent); err != nil {
				log.Errorf("发送秒杀失败事件失败: %v, userID=%d, activityID=%d", 
					err, req.UserID, req.ActivityID)
			}
		}()
	}

	return flashSaleResult, nil
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

	// 如果活动正在进行中，获取实时状态
	if activityDO.Status == do.FlashSaleStatusActive {
		if currentStock, err := fss.stockManager.GetCurrentStock(ctx, activityDO.CouponTemplateID); err == nil {
			result.RemainStock = currentStock
		}
		
		if redisActivityInfo, err := fss.stockManager.GetActivityStatus(ctx, activityID); err == nil {
			result.SoldCount = redisActivityInfo.SuccessCount
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