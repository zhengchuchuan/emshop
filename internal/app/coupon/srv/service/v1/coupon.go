package v1

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"emshop/internal/app/coupon/srv/data/v1/interfaces"
	"emshop/internal/app/coupon/srv/domain/do"
	"emshop/internal/app/coupon/srv/domain/dto"
	"emshop/internal/app/coupon/srv/pkg/cache"
	"emshop/internal/app/coupon/srv/pkg/calculator"
	"emshop/internal/app/coupon/srv/pkg/scripts"
	"emshop/internal/app/pkg/code"
	"emshop/internal/app/pkg/options"
	"emshop/pkg/errors"
	"emshop/pkg/log"
	v1 "emshop/pkg/common/meta/v1"
	"github.com/go-redis/redis/v8"
)

// CouponSrv 优惠券服务接口
type CouponSrv interface {
	// 优惠券模板管理
	CreateCouponTemplate(ctx context.Context, req *dto.CreateCouponTemplateDTO) (*dto.CouponTemplateDTO, error)
	GetCouponTemplate(ctx context.Context, id int64) (*dto.CouponTemplateDTO, error)
	UpdateCouponTemplate(ctx context.Context, req *dto.UpdateCouponTemplateDTO) (*dto.CouponTemplateDTO, error)
	ListCouponTemplates(ctx context.Context, req *dto.ListCouponTemplatesDTO) (*dto.CouponTemplateListDTO, error)
	
	// 用户优惠券操作
	ReceiveCoupon(ctx context.Context, req *dto.ReceiveCouponDTO) (*dto.UserCouponDTO, error)
	GetUserCoupons(ctx context.Context, req *dto.GetUserCouponsDTO) (*dto.UserCouponListDTO, error)
	GetAvailableCoupons(ctx context.Context, req *dto.GetAvailableCouponsDTO) (*dto.UserCouponListDTO, error)
	
	// 优惠券计算和使用
	CalculateCouponDiscount(ctx context.Context, req *dto.CalculateCouponDiscountDTO) (*dto.CouponDiscountResultDTO, error)
	UseCoupons(ctx context.Context, req *dto.UseCouponsDTO) (*dto.UseCouponsResultDTO, error)
	ReleaseCoupons(ctx context.Context, req *dto.ReleaseCouponsDTO) error
}

type couponService struct {
	data             interfaces.DataFactory
	redisClient      *redis.Client
	dtmOpts          *options.DtmOptions
	keyFormatter     *scripts.RedisKeyFormatter
	calculationEngine *calculator.CalculationEngine
	cacheManager     interface {
		GetCouponTemplate(ctx context.Context, couponID int64) (*cache.CouponTemplate, error)
		GetUserCoupon(ctx context.Context, userCouponID int64) (*cache.UserCoupon, error)
		InvalidateCache(keys ...string)
	}
}

// NewCouponService 创建优惠券服务实例
func NewCouponService(data interfaces.DataFactory, redisClient *redis.Client, dtmOpts *options.DtmOptions, cacheManager interface{
	GetCouponTemplate(ctx context.Context, couponID int64) (*cache.CouponTemplate, error)
	GetUserCoupon(ctx context.Context, userCouponID int64) (*cache.UserCoupon, error)
	InvalidateCache(keys ...string)
}) CouponSrv {
	// 创建增强计算引擎
	var calculationEngine *calculator.CalculationEngine
	if cacheManager != nil {
		// 创建缓存管理器适配器
		cacheAdapter := &cacheManagerAdapter{cacheManager: cacheManager}
		calculationEngine = calculator.NewCalculationEngine(cacheAdapter)
	}

	return &couponService{
		data:             data,
		redisClient:      redisClient,
		dtmOpts:          dtmOpts,
		keyFormatter:     scripts.NewRedisKeyFormatter(),
		calculationEngine: calculationEngine,
		cacheManager:     cacheManager,
	}
}

// cacheManagerAdapter 缓存管理器适配器
type cacheManagerAdapter struct {
	cacheManager interface {
		GetCouponTemplate(ctx context.Context, couponID int64) (*cache.CouponTemplate, error)
		GetUserCoupon(ctx context.Context, userCouponID int64) (*cache.UserCoupon, error)
	}
}

// GetCouponTemplate 实现 calculator.CacheManager 接口
func (c *cacheManagerAdapter) GetCouponTemplate(ctx context.Context, couponID int64) (*calculator.CouponTemplate, error) {
	cacheTemplate, err := c.cacheManager.GetCouponTemplate(ctx, couponID)
	if err != nil {
		return nil, err
	}
	
	// 转换缓存模板为计算引擎模板
	return &calculator.CouponTemplate{
		ID:                cacheTemplate.ID,
		Name:              cacheTemplate.Name,
		Type:              cacheTemplate.Type,
		DiscountType:      cacheTemplate.DiscountType,
		DiscountValue:     cacheTemplate.DiscountValue,
		MinAmount:         cacheTemplate.MinAmount,
		MaxDiscountAmount: 0, // 缓存模板中没有此字段，设为0表示无限制
		ValidStart:        cacheTemplate.ValidStart,
		ValidEnd:          cacheTemplate.ValidEnd,
		Status:            cacheTemplate.Status,
	}, nil
}

// GetUserCoupon 实现 calculator.CacheManager 接口
func (c *cacheManagerAdapter) GetUserCoupon(ctx context.Context, userCouponID int64) (*calculator.UserCoupon, error) {
	cacheUserCoupon, err := c.cacheManager.GetUserCoupon(ctx, userCouponID)
	if err != nil {
		return nil, err
	}
	
	// 转换缓存用户优惠券为计算引擎用户优惠券
	return &calculator.UserCoupon{
		ID:             cacheUserCoupon.ID,
		CouponID:       cacheUserCoupon.CouponID,
		UserID:         cacheUserCoupon.UserID,
		CouponSn:       cacheUserCoupon.CouponSn,
		Status:         cacheUserCoupon.Status,
		ObtainTime:     cacheUserCoupon.ObtainTime,
		ValidStartTime: cacheUserCoupon.ValidStartTime,
		ValidEndTime:   cacheUserCoupon.ValidEndTime,
	}, nil
}

// generateCouponCode 生成优惠券码
func (cs *couponService) generateCouponCode() string {
	timestamp := time.Now().Format("20060102150405")
	n, _ := rand.Int(rand.Reader, big.NewInt(9999))
	return fmt.Sprintf("CPN%s%04d", timestamp, n.Int64())
}

// CreateCouponTemplate 创建优惠券模板
func (cs *couponService) CreateCouponTemplate(ctx context.Context, req *dto.CreateCouponTemplateDTO) (*dto.CouponTemplateDTO, error) {
	log.Infof("创建优惠券模板: %s", req.Name)
	
	// 验证时间范围
	if req.ValidEndTime.Before(req.ValidStartTime) {
		return nil, errors.WithCode(code.ErrInvalidRequest, "结束时间不能早于开始时间")
	}
	
	// 构建DO对象
	templateDO := &do.CouponTemplateDO{
		Name:               req.Name,
		Type:               do.CouponType(req.Type),
		DiscountType:       do.DiscountType(req.DiscountType),
		DiscountValue:      req.DiscountValue,
		MinOrderAmount:     req.MinOrderAmount,
		MaxDiscountAmount:  req.MaxDiscountAmount,
		TotalCount:         req.TotalCount,
		UsedCount:          0,
		PerUserLimit:       req.PerUserLimit,
		ValidStartTime:     req.ValidStartTime,
		ValidEndTime:       req.ValidEndTime,
		ValidDays:          req.ValidDays,
		Status:             do.CouponStatusActive,
		Description:        req.Description,
	}
	
	// 保存到数据库
	if err := cs.data.CouponTemplates().Create(ctx, cs.data.DB(), templateDO); err != nil {
		log.Errorf("创建优惠券模板失败: %v", err)
		return nil, errors.WithCode(code.ErrDatabase, "创建优惠券模板失败")
	}
	
	// 转换为DTO
	return cs.convertTemplateToDTO(templateDO), nil
}

// GetCouponTemplate 获取优惠券模板 (支持三层缓存)
func (cs *couponService) GetCouponTemplate(ctx context.Context, id int64) (*dto.CouponTemplateDTO, error) {
	// 优先使用缓存管理器
	if cs.cacheManager != nil {
		cacheTemplate, err := cs.cacheManager.GetCouponTemplate(ctx, id)
		if err != nil {
			log.Warnf("缓存获取优惠券模板失败，回退到数据库查询: %v", err)
		} else if cacheTemplate != nil {
			// 转换缓存数据为DTO
			return &dto.CouponTemplateDTO{
				ID:                cacheTemplate.ID,
				Name:              cacheTemplate.Name,
				Type:              cacheTemplate.Type,
				DiscountType:      cacheTemplate.DiscountType,
				DiscountValue:     cacheTemplate.DiscountValue,
				MinOrderAmount:    cacheTemplate.MinAmount,
				TotalCount:        cacheTemplate.TotalCount,
				UsedCount:         cacheTemplate.UsedCount,
				ValidStartTime:    cacheTemplate.ValidStart,
				ValidEndTime:      cacheTemplate.ValidEnd,
				Status:            cacheTemplate.Status,
			}, nil
		}
	}
	
	// 缓存失败或为空时回退到数据库查询
	templateDO, err := cs.data.CouponTemplates().Get(ctx, cs.data.DB(), id)
	if err != nil {
		log.Errorf("获取优惠券模板失败: %v", err)
		return nil, errors.WithCode(code.ErrDatabase, "获取优惠券模板失败")
	}
	
	if templateDO == nil {
		return nil, errors.WithCode(code.ErrResourceNotFound, "优惠券模板不存在")
	}
	
	return cs.convertTemplateToDTO(templateDO), nil
}

// UpdateCouponTemplate 更新优惠券模板
func (cs *couponService) UpdateCouponTemplate(ctx context.Context, req *dto.UpdateCouponTemplateDTO) (*dto.CouponTemplateDTO, error) {
	// 先获取现有模板
	templateDO, err := cs.data.CouponTemplates().Get(ctx, cs.data.DB(), req.ID)
	if err != nil {
		return nil, errors.WithCode(code.ErrDatabase, "获取优惠券模板失败")
	}
	
	if templateDO == nil {
		return nil, errors.WithCode(code.ErrResourceNotFound, "优惠券模板不存在")
	}
	
	// 更新字段
	if req.Name != nil {
		templateDO.Name = *req.Name
	}
	if req.Status != nil {
		templateDO.Status = do.CouponStatus(*req.Status)
	}
	if req.Description != nil {
		templateDO.Description = *req.Description
	}
	
	// 保存更新
	if err := cs.data.CouponTemplates().Update(ctx, cs.data.DB(), templateDO); err != nil {
		log.Errorf("更新优惠券模板失败: %v", err)
		return nil, errors.WithCode(code.ErrDatabase, "更新优惠券模板失败")
	}
	
	// 缓存失效
	if cs.cacheManager != nil {
		cacheKey := fmt.Sprintf("coupon:template:%d", req.ID)
		cs.cacheManager.InvalidateCache(cacheKey)
		log.Debugf("失效优惠券模板缓存: %s", cacheKey)
	}
	
	return cs.convertTemplateToDTO(templateDO), nil
}

// ListCouponTemplates 获取优惠券模板列表
func (cs *couponService) ListCouponTemplates(ctx context.Context, req *dto.ListCouponTemplatesDTO) (*dto.CouponTemplateListDTO, error) {
	var status do.CouponStatus
	if req.Status != nil {
		status = do.CouponStatus(*req.Status)
	}
	
	meta := v1.ListMeta{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}
	
	templateListDO, err := cs.data.CouponTemplates().List(ctx, cs.data.DB(), status, meta, []string{"created_at DESC"})
	if err != nil {
		log.Errorf("获取优惠券模板列表失败: %v", err)
		return nil, errors.WithCode(code.ErrDatabase, "获取优惠券模板列表失败")
	}
	
	// 转换为DTO
	items := make([]*dto.CouponTemplateDTO, 0, len(templateListDO.Items))
	for _, templateDO := range templateListDO.Items {
		items = append(items, cs.convertTemplateToDTO(templateDO))
	}
	
	return &dto.CouponTemplateListDTO{
		TotalCount: templateListDO.TotalCount,
		Items:      items,
	}, nil
}

// ReceiveCoupon 领取优惠券
func (cs *couponService) ReceiveCoupon(ctx context.Context, req *dto.ReceiveCouponDTO) (*dto.UserCouponDTO, error) {
	log.Infof("用户领取优惠券: userID=%d, templateID=%d", req.UserID, req.CouponTemplateID)
	
	// 获取优惠券模板
	templateDO, err := cs.data.CouponTemplates().Get(ctx, cs.data.DB(), req.CouponTemplateID)
	if err != nil || templateDO == nil {
		return nil, errors.WithCode(code.ErrResourceNotFound, "优惠券模板不存在")
	}
	
	currentTime := time.Now()
	
	// 检查模板可用性
	available, err := cs.data.CouponTemplates().CheckTemplateAvailability(ctx, cs.data.DB(), req.CouponTemplateID, currentTime)
	if err != nil {
		return nil, errors.WithCode(code.ErrDatabase, "检查优惠券模板可用性失败")
	}
	if !available {
		return nil, errors.WithCode(code.ErrResourceNotAvailable, "优惠券模板不可用或已过期")
	}
	
	// 检查用户限领数量
	userCouponCount, err := cs.data.UserCoupons().CountUserCouponsByTemplate(ctx, cs.data.DB(), req.UserID, req.CouponTemplateID)
	if err != nil {
		return nil, errors.WithCode(code.ErrDatabase, "检查用户优惠券数量失败")
	}
	if userCouponCount >= int64(templateDO.PerUserLimit) {
		return nil, errors.WithCode(code.ErrResourceLimitExceeded, "超出个人限领数量")
	}
	
	// 开始事务
	tx := cs.data.Begin()
	
	// 创建用户优惠券
	expiredAt := templateDO.ValidEndTime
	if templateDO.ValidDays > 0 {
		expiredAt = currentTime.AddDate(0, 0, int(templateDO.ValidDays))
	}
	
	userCouponDO := &do.UserCouponDO{
		CouponTemplateID: req.CouponTemplateID,
		UserID:           req.UserID,
		CouponCode:       cs.generateCouponCode(),
		Status:           do.UserCouponStatusUnused,
		ReceivedAt:       currentTime,
		ExpiredAt:        expiredAt,
	}
	
	if err := cs.data.UserCoupons().Create(ctx, tx, userCouponDO); err != nil {
		tx.Rollback()
		log.Errorf("创建用户优惠券失败: %v", err)
		return nil, errors.WithCode(code.ErrDatabase, "领取优惠券失败")
	}
	
	// 更新模板使用数量
	if err := cs.data.CouponTemplates().UpdateUsedCount(ctx, tx, req.CouponTemplateID, 1); err != nil {
		tx.Rollback()
		log.Errorf("更新优惠券模板使用数量失败: %v", err)
		return nil, errors.WithCode(code.ErrDatabase, "领取优惠券失败")
	}
	
	if err := tx.Commit().Error; err != nil {
		log.Errorf("提交事务失败: %v", err)
		return nil, errors.WithCode(code.ErrDatabase, "领取优惠券失败")
	}
	
	// 转换为DTO
	return cs.convertUserCouponToDTO(userCouponDO, templateDO), nil
}

// GetUserCoupons 获取用户优惠券列表
func (cs *couponService) GetUserCoupons(ctx context.Context, req *dto.GetUserCouponsDTO) (*dto.UserCouponListDTO, error) {
	var status do.UserCouponStatus
	if req.Status != nil {
		status = do.UserCouponStatus(*req.Status)
	}
	
	meta := v1.ListMeta{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}
	
	userCouponListDO, err := cs.data.UserCoupons().GetUserCoupons(ctx, cs.data.DB(), req.UserID, status, meta)
	if err != nil {
		log.Errorf("获取用户优惠券列表失败: %v", err)
		return nil, errors.WithCode(code.ErrDatabase, "获取用户优惠券列表失败")
	}
	
	// 转换为DTO
	items := make([]*dto.UserCouponDTO, 0, len(userCouponListDO.Items))
	for _, userCouponDO := range userCouponListDO.Items {
		items = append(items, cs.convertUserCouponToDTO(userCouponDO, nil))
	}
	
	return &dto.UserCouponListDTO{
		TotalCount: userCouponListDO.TotalCount,
		Items:      items,
	}, nil
}

// GetAvailableCoupons 获取用户可用优惠券
func (cs *couponService) GetAvailableCoupons(ctx context.Context, req *dto.GetAvailableCouponsDTO) (*dto.UserCouponListDTO, error) {
	currentTime := time.Now()
	
	userCouponDOs, err := cs.data.UserCoupons().GetUserAvailableCoupons(ctx, cs.data.DB(), req.UserID, req.OrderAmount, currentTime)
	if err != nil {
		log.Errorf("获取用户可用优惠券失败: %v", err)
		return nil, errors.WithCode(code.ErrDatabase, "获取用户可用优惠券失败")
	}
	
	// 转换为DTO
	items := make([]*dto.UserCouponDTO, 0, len(userCouponDOs))
	for _, userCouponDO := range userCouponDOs {
		items = append(items, cs.convertUserCouponToDTO(userCouponDO, nil))
	}
	
	return &dto.UserCouponListDTO{
		TotalCount: int64(len(items)),
		Items:      items,
	}, nil
}

// CalculateCouponDiscount 计算优惠券折扣 - 使用增强计算引擎
func (cs *couponService) CalculateCouponDiscount(ctx context.Context, req *dto.CalculateCouponDiscountDTO) (*dto.CouponDiscountResultDTO, error) {
	log.Infof("开始计算优惠券折扣: userID=%d, coupons=%v, amount=%.2f", 
		req.UserID, req.CouponIDs, req.OrderAmount)

	// 使用增强的计算引擎
	if cs.calculationEngine != nil {
		result, err := cs.calculationEngine.Calculate(ctx, req)
		if err != nil {
			log.Errorf("使用增强计算引擎失败，回退到简单模式: %v", err)
			return cs.fallbackCalculation(ctx, req)
		}
		
		log.Infof("增强计算引擎计算完成: 原始金额=%.2f, 折扣金额=%.2f, 最终金额=%.2f, 应用优惠券=%v",
			result.OriginalAmount, result.DiscountAmount, result.FinalAmount, result.AppliedCoupons)
		
		return result, nil
	}

	// 回退到简单计算模式
	log.Warn("增强计算引擎未初始化，使用简单计算模式")
	return cs.fallbackCalculation(ctx, req)
}

// fallbackCalculation 回退的简单计算模式
func (cs *couponService) fallbackCalculation(ctx context.Context, req *dto.CalculateCouponDiscountDTO) (*dto.CouponDiscountResultDTO, error) {
	result := &dto.CouponDiscountResultDTO{
		OriginalAmount:  req.OrderAmount,
		DiscountAmount:  0,
		FinalAmount:     req.OrderAmount,
		AppliedCoupons:  make([]int64, 0),
		RejectedCoupons: make([]*dto.CouponRejection, 0),
	}
	
	currentTime := time.Now()
	
	// 获取用户优惠券
	for _, couponID := range req.CouponIDs {
		userCouponDO, err := cs.data.UserCoupons().Get(ctx, cs.data.DB(), couponID)
		if err != nil || userCouponDO == nil {
			result.RejectedCoupons = append(result.RejectedCoupons, &dto.CouponRejection{
				CouponID: couponID,
				Reason:   "优惠券不存在",
			})
			continue
		}
		
		// 检查优惠券状态
		if userCouponDO.Status != do.UserCouponStatusUnused {
			result.RejectedCoupons = append(result.RejectedCoupons, &dto.CouponRejection{
				CouponID: couponID,
				Reason:   "优惠券不可用",
			})
			continue
		}
		
		// 检查过期时间
		if userCouponDO.ExpiredAt.Before(currentTime) {
			result.RejectedCoupons = append(result.RejectedCoupons, &dto.CouponRejection{
				CouponID: couponID,
				Reason:   "优惠券已过期",
			})
			continue
		}
		
		// 获取优惠券模板
		templateDO, err := cs.data.CouponTemplates().Get(ctx, cs.data.DB(), userCouponDO.CouponTemplateID)
		if err != nil || templateDO == nil {
			result.RejectedCoupons = append(result.RejectedCoupons, &dto.CouponRejection{
				CouponID: couponID,
				Reason:   "优惠券模板不存在",
			})
			continue
		}
		
		// 检查最小订单金额
		if req.OrderAmount < templateDO.MinOrderAmount {
			result.RejectedCoupons = append(result.RejectedCoupons, &dto.CouponRejection{
				CouponID: couponID,
				Reason:   fmt.Sprintf("订单金额不满足最低%.2f元要求", templateDO.MinOrderAmount),
			})
			continue
		}
		
		// 计算折扣
		var discount float64
		if templateDO.DiscountType == do.DiscountTypeFixed {
			discount = templateDO.DiscountValue
		} else {
			discount = req.OrderAmount * templateDO.DiscountValue / 100
			if templateDO.MaxDiscountAmount > 0 && discount > templateDO.MaxDiscountAmount {
				discount = templateDO.MaxDiscountAmount
			}
		}
		
		// 累加折扣
		result.DiscountAmount += discount
		result.AppliedCoupons = append(result.AppliedCoupons, couponID)
	}
	
	// 最终金额不能小于0
	result.FinalAmount = result.OriginalAmount - result.DiscountAmount
	if result.FinalAmount < 0 {
		result.FinalAmount = 0
	}
	
	return result, nil
}

// UseCoupons 使用优惠券 (DTM事务用)
func (cs *couponService) UseCoupons(ctx context.Context, req *dto.UseCouponsDTO) (*dto.UseCouponsResultDTO, error) {
	log.Infof("使用优惠券: orderSn=%s, userID=%d, coupons=%v", req.OrderSn, req.UserID, req.CouponIDs)
	
	// 先计算折扣
	calcReq := &dto.CalculateCouponDiscountDTO{
		UserID:      req.UserID,
		CouponIDs:   req.CouponIDs,
		OrderAmount: req.OrderAmount,
		OrderItems:  make([]*dto.OrderItemDTO, 0), // 简化处理
	}
	
	calcResult, err := cs.CalculateCouponDiscount(ctx, calcReq)
	if err != nil {
		return nil, err
	}
	
	if len(calcResult.AppliedCoupons) == 0 {
		return nil, errors.WithCode(code.ErrInvalidRequest, "没有可用的优惠券")
	}
	
	// 使用Redis Lua脚本锁定优惠券
	currentTime := time.Now().Unix()
	lockTimeout := int64(300) // 5分钟锁定时间
	
	tx := cs.data.Begin()
	usedCoupons := make([]int64, 0)
	
	for _, couponID := range calcResult.AppliedCoupons {
		userCouponDO, err := cs.data.UserCoupons().Get(ctx, cs.data.DB(), couponID)
		if err != nil || userCouponDO == nil {
			tx.Rollback()
			return nil, errors.WithCode(code.ErrResourceNotFound, "优惠券不存在")
		}
		
		// Redis锁定检查
		lockKey := cs.keyFormatter.CouponLockKey(couponID)
		statusKey := cs.keyFormatter.CouponStatusKey(couponID)
		
		result, err := cs.redisClient.Eval(ctx, scripts.CheckCouponUsageLua, []string{lockKey, statusKey}, 
			couponID, req.UserID, currentTime, userCouponDO.ExpiredAt.Unix(), lockTimeout).Result()
		if err != nil {
			tx.Rollback()
			return nil, errors.WithCode(code.ErrRedis, "优惠券状态检查失败")
		}
		
		if result.(int64) != scripts.CouponLockSuccess {
			tx.Rollback()
			reason := scripts.GetCouponCheckResultMessage(result.(int64))
			return nil, errors.WithCode(code.ErrResourceNotAvailable, "%s", reason)
		}
		
		// 使用优惠券
		if err := cs.data.UserCoupons().UseCoupon(ctx, tx, couponID, req.OrderSn, time.Now()); err != nil {
			tx.Rollback()
			// 释放Redis锁
			cs.redisClient.Eval(ctx, scripts.ReleaseCouponLockLua, []string{lockKey, statusKey}, req.UserID, "release")
			return nil, errors.WithCode(code.ErrDatabase, "使用优惠券失败")
		}
		
		// 释放Redis锁并标记已使用
		cs.redisClient.Eval(ctx, scripts.ReleaseCouponLockLua, []string{lockKey, statusKey}, req.UserID, "use")
		usedCoupons = append(usedCoupons, couponID)
	}
	
	if err := tx.Commit().Error; err != nil {
		log.Errorf("提交使用优惠券事务失败: %v", err)
		return nil, errors.WithCode(code.ErrDatabase, "使用优惠券失败")
	}
	
	return &dto.UseCouponsResultDTO{
		DiscountAmount: calcResult.DiscountAmount,
		UsedCoupons:    usedCoupons,
	}, nil
}

// ReleaseCoupons 释放优惠券 (DTM补偿用)
func (cs *couponService) ReleaseCoupons(ctx context.Context, req *dto.ReleaseCouponsDTO) error {
	log.Infof("释放优惠券: orderSn=%s", req.OrderSn)
	
	// 获取该订单的优惠券使用记录
	usageLogs, err := cs.data.CouponUsageLogs().GetByOrderSn(ctx, cs.data.DB(), req.OrderSn)
	if err != nil {
		log.Errorf("获取优惠券使用记录失败: %v", err)
		return errors.WithCode(code.ErrDatabase, "获取优惠券使用记录失败")
	}
	
	tx := cs.data.Begin()
	
	for _, usageLog := range usageLogs {
		// 恢复优惠券状态
		if err := cs.data.UserCoupons().UpdateStatus(ctx, tx, usageLog.UserCouponID, do.UserCouponStatusUnused); err != nil {
			tx.Rollback()
			return errors.WithCode(code.ErrDatabase, "恢复优惠券状态失败")
		}
		
		// 清除Redis状态
		statusKey := cs.keyFormatter.CouponStatusKey(usageLog.UserCouponID)
		cs.redisClient.Set(ctx, statusKey, 1, time.Hour) // 设置为未使用状态
	}
	
	if err := tx.Commit().Error; err != nil {
		log.Errorf("提交释放优惠券事务失败: %v", err)
		return errors.WithCode(code.ErrDatabase, "释放优惠券失败")
	}
	
	return nil
}

// convertTemplateToDTO 转换模板DO为DTO
func (cs *couponService) convertTemplateToDTO(templateDO *do.CouponTemplateDO) *dto.CouponTemplateDTO {
	return &dto.CouponTemplateDTO{
		ID:                templateDO.ID,
		Name:              templateDO.Name,
		Type:              int32(templateDO.Type),
		DiscountType:      int32(templateDO.DiscountType),
		DiscountValue:     templateDO.DiscountValue,
		MinOrderAmount:    templateDO.MinOrderAmount,
		MaxDiscountAmount: templateDO.MaxDiscountAmount,
		TotalCount:        templateDO.TotalCount,
		UsedCount:         templateDO.UsedCount,
		PerUserLimit:      templateDO.PerUserLimit,
		ValidStartTime:    templateDO.ValidStartTime,
		ValidEndTime:      templateDO.ValidEndTime,
		ValidDays:         templateDO.ValidDays,
		Status:            int32(templateDO.Status),
		Description:       templateDO.Description,
		CreatedAt:         templateDO.CreatedAt,
	}
}

// convertUserCouponToDTO 转换用户优惠券DO为DTO
func (cs *couponService) convertUserCouponToDTO(userCouponDO *do.UserCouponDO, templateDO *do.CouponTemplateDO) *dto.UserCouponDTO {
	dto := &dto.UserCouponDTO{
		ID:               userCouponDO.ID,
		CouponTemplateID: userCouponDO.CouponTemplateID,
		UserID:           userCouponDO.UserID,
		CouponCode:       userCouponDO.CouponCode,
		Status:           int32(userCouponDO.Status),
		OrderSn:          userCouponDO.OrderSn,
		ReceivedAt:       userCouponDO.ReceivedAt,
		UsedAt:           userCouponDO.UsedAt,
		ExpiredAt:        userCouponDO.ExpiredAt,
	}
	
	if templateDO != nil {
		dto.Template = cs.convertTemplateToDTO(templateDO)
	}
	
	return dto
}