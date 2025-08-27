package calculator

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"emshop/internal/app/coupon/srv/domain/do"
	"emshop/internal/app/coupon/srv/domain/dto"
	"emshop/pkg/log"
)

// CalculationEngine 优惠券计算引擎
type CalculationEngine struct {
	strategies   map[do.CouponType]CalculationStrategy
	validators   []CouponValidator
	ruleEngine   *BusinessRuleEngine
	optimizer    *CombinationOptimizer
	cacheManager CacheManager
}

// CacheManager 缓存管理器接口
type CacheManager interface {
	GetCouponTemplate(ctx context.Context, couponID int64) (*CouponTemplate, error)
	GetUserCoupon(ctx context.Context, userCouponID int64) (*UserCoupon, error)
}

// CouponTemplate 优惠券模板缓存结构
type CouponTemplate struct {
	ID                int64     `json:"id"`
	Name              string    `json:"name"`
	Type              int32     `json:"type"`
	DiscountType      int32     `json:"discount_type"`
	DiscountValue     float64   `json:"discount_value"`
	MinAmount         float64   `json:"min_amount"`
	MaxDiscountAmount float64   `json:"max_discount_amount"`
	ValidStart        time.Time `json:"valid_start"`
	ValidEnd          time.Time `json:"valid_end"`
	Status            int32     `json:"status"`
}

// UserCoupon 用户优惠券缓存结构
type UserCoupon struct {
	ID             int64     `json:"id"`
	CouponID       int64     `json:"coupon_id"`
	UserID         int64     `json:"user_id"`
	CouponSn       string    `json:"coupon_sn"`
	Status         int32     `json:"status"`
	ObtainTime     time.Time `json:"obtain_time"`
	ValidStartTime time.Time `json:"valid_start_time"`
	ValidEndTime   time.Time `json:"valid_end_time"`
}

// CalculationContext 计算上下文
type CalculationContext struct {
	UserID         int64
	OrderAmount    float64
	OrderItems     []*dto.OrderItemDTO
	CouponIDs      []int64
	UserCoupons    []*EnhancedUserCoupon
	AppliedCoupons []*AppliedCoupon
	CurrentTime    time.Time
}

// EnhancedUserCoupon 增强的用户优惠券
type EnhancedUserCoupon struct {
	UserCoupon *UserCoupon
	Template   *CouponTemplate
	Priority   int
	Score      float64
}

// AppliedCoupon 已应用的优惠券
type AppliedCoupon struct {
	CouponID       int64
	DiscountAmount float64
	AppliedAmount  float64
	Strategy       string
}

// ValidationResult 验证结果
type ValidationResult struct {
	IsValid bool
	Reason  string
	Code    string
}

// CalculationStrategy 计算策略接口
type CalculationStrategy interface {
	Calculate(ctx *CalculationContext, coupon *EnhancedUserCoupon) (*CalculationResult, error)
	GetPriority() int
	CanApply(ctx *CalculationContext, coupon *EnhancedUserCoupon) bool
}

// CalculationResult 单个优惠券计算结果
type CalculationResult struct {
	CouponID        int64
	DiscountAmount  float64
	AppliedAmount   float64
	CalculationInfo string
	IsApplicable    bool
	Reason          string
}

// CouponValidator 优惠券验证器接口
type CouponValidator interface {
	Validate(ctx *CalculationContext, coupon *EnhancedUserCoupon) *ValidationResult
}

// NewCalculationEngine 创建计算引擎
func NewCalculationEngine(cacheManager CacheManager) *CalculationEngine {
	engine := &CalculationEngine{
		strategies:   make(map[do.CouponType]CalculationStrategy),
		validators:   make([]CouponValidator, 0),
		ruleEngine:   NewBusinessRuleEngine(),
		optimizer:    NewCombinationOptimizer(),
		cacheManager: cacheManager,
	}

	// 注册计算策略
	engine.registerStrategies()
	
	// 注册验证器
	engine.registerValidators()

	return engine
}

// registerStrategies 注册计算策略
func (e *CalculationEngine) registerStrategies() {
	e.strategies[do.CouponTypeThreshold] = &ThresholdCouponStrategy{}
	e.strategies[do.CouponTypeDiscount] = &DiscountCouponStrategy{}
	e.strategies[do.CouponTypeInstant] = &InstantCouponStrategy{}
	e.strategies[do.CouponTypeFreeShip] = &FreeShipCouponStrategy{}
}

// registerValidators 注册验证器
func (e *CalculationEngine) registerValidators() {
	e.validators = append(e.validators,
		&BasicCouponValidator{},
		&TimingValidator{},
		&AmountValidator{},
		&UserLimitValidator{},
		&CombinationValidator{},
	)
}

// Calculate 主计算方法
func (e *CalculationEngine) Calculate(ctx context.Context, req *dto.CalculateCouponDiscountDTO) (*dto.CouponDiscountResultDTO, error) {
	log.Infof("开始计算优惠券折扣: userID=%d, coupons=%v, amount=%.2f", 
		req.UserID, req.CouponIDs, req.OrderAmount)

	// 创建计算上下文
	calcCtx := &CalculationContext{
		UserID:         req.UserID,
		OrderAmount:    req.OrderAmount,
		OrderItems:     req.OrderItems,
		CouponIDs:      req.CouponIDs,
		AppliedCoupons: make([]*AppliedCoupon, 0),
		CurrentTime:    time.Now(),
	}

	// 加载用户优惠券信息
	if err := e.loadUserCoupons(ctx, calcCtx); err != nil {
		return nil, fmt.Errorf("加载用户优惠券失败: %v", err)
	}

	// 执行预处理
	if err := e.preProcess(calcCtx); err != nil {
		return nil, fmt.Errorf("预处理失败: %v", err)
	}

	// 执行验证
	validationResults := e.validateCoupons(calcCtx)

	// 执行计算
	calculationResults := e.calculateDiscounts(calcCtx, validationResults)

	// 优化组合
	if e.optimizer != nil {
		optimizedResults, err := e.optimizer.OptimizeCombination(calcCtx, calculationResults)
		if err != nil {
			log.Warnf("优化组合失败，使用原结果: %v", err)
		} else {
			calculationResults = optimizedResults
		}
	}

	// 构造最终结果
	return e.buildFinalResult(req.OrderAmount, calculationResults, validationResults), nil
}

// loadUserCoupons 加载用户优惠券
func (e *CalculationEngine) loadUserCoupons(ctx context.Context, calcCtx *CalculationContext) error {
	calcCtx.UserCoupons = make([]*EnhancedUserCoupon, 0, len(calcCtx.CouponIDs))

	for _, couponID := range calcCtx.CouponIDs {
		// 从缓存获取用户优惠券
		var userCoupon *UserCoupon
		var template *CouponTemplate
		var err error

		if e.cacheManager != nil {
			userCoupon, err = e.cacheManager.GetUserCoupon(ctx, couponID)
			if err != nil {
				log.Warnf("从缓存获取用户优惠券失败: %v", err)
				continue
			}

			template, err = e.cacheManager.GetCouponTemplate(ctx, userCoupon.CouponID)
			if err != nil {
				log.Warnf("从缓存获取优惠券模板失败: %v", err)
				continue
			}
		}

		if userCoupon == nil || template == nil {
			continue
		}

		enhanced := &EnhancedUserCoupon{
			UserCoupon: userCoupon,
			Template:   template,
			Priority:   e.calculatePriority(template),
			Score:      e.calculateScore(template, calcCtx.OrderAmount),
		}

		calcCtx.UserCoupons = append(calcCtx.UserCoupons, enhanced)
	}

	// 按优先级排序
	sort.Slice(calcCtx.UserCoupons, func(i, j int) bool {
		if calcCtx.UserCoupons[i].Priority == calcCtx.UserCoupons[j].Priority {
			return calcCtx.UserCoupons[i].Score > calcCtx.UserCoupons[j].Score
		}
		return calcCtx.UserCoupons[i].Priority > calcCtx.UserCoupons[j].Priority
	})

	log.Infof("加载了%d个有效优惠券", len(calcCtx.UserCoupons))
	return nil
}

// preProcess 预处理
func (e *CalculationEngine) preProcess(calcCtx *CalculationContext) error {
	// 执行业务规则检查
	if e.ruleEngine != nil {
		if err := e.ruleEngine.ExecuteRules(calcCtx); err != nil {
			return fmt.Errorf("业务规则检查失败: %v", err)
		}
	}

	return nil
}

// validateCoupons 验证优惠券
func (e *CalculationEngine) validateCoupons(calcCtx *CalculationContext) map[int64]*ValidationResult {
	results := make(map[int64]*ValidationResult)

	for _, coupon := range calcCtx.UserCoupons {
		// 执行所有验证器
		for _, validator := range e.validators {
			result := validator.Validate(calcCtx, coupon)
			if !result.IsValid {
				results[coupon.UserCoupon.ID] = result
				break
			}
		}

		// 如果没有验证失败，标记为有效
		if _, exists := results[coupon.UserCoupon.ID]; !exists {
			results[coupon.UserCoupon.ID] = &ValidationResult{
				IsValid: true,
			}
		}
	}

	return results
}

// calculateDiscounts 计算折扣
func (e *CalculationEngine) calculateDiscounts(calcCtx *CalculationContext, validationResults map[int64]*ValidationResult) []*CalculationResult {
	results := make([]*CalculationResult, 0)

	for _, coupon := range calcCtx.UserCoupons {
		// 检查验证结果
		validationResult, exists := validationResults[coupon.UserCoupon.ID]
		if !exists || !validationResult.IsValid {
			results = append(results, &CalculationResult{
				CouponID:     coupon.UserCoupon.ID,
				IsApplicable: false,
				Reason:       validationResult.Reason,
			})
			continue
		}

		// 获取对应的计算策略
		strategy, exists := e.strategies[do.CouponType(coupon.Template.Type)]
		if !exists {
			results = append(results, &CalculationResult{
				CouponID:     coupon.UserCoupon.ID,
				IsApplicable: false,
				Reason:       "不支持的优惠券类型",
			})
			continue
		}

		// 检查策略是否适用
		if !strategy.CanApply(calcCtx, coupon) {
			results = append(results, &CalculationResult{
				CouponID:     coupon.UserCoupon.ID,
				IsApplicable: false,
				Reason:       "不满足策略条件",
			})
			continue
		}

		// 执行计算
		result, err := strategy.Calculate(calcCtx, coupon)
		if err != nil {
			log.Errorf("计算优惠券%d折扣失败: %v", coupon.UserCoupon.ID, err)
			results = append(results, &CalculationResult{
				CouponID:     coupon.UserCoupon.ID,
				IsApplicable: false,
				Reason:       "计算失败",
			})
			continue
		}

		results = append(results, result)
	}

	return results
}

// calculatePriority 计算优惠券优先级
func (e *CalculationEngine) calculatePriority(template *CouponTemplate) int {
	priority := 0

	// 根据优惠券类型设置基础优先级
	switch do.CouponType(template.Type) {
	case do.CouponTypeFreeShip:
		priority = 1
	case do.CouponTypeInstant:
		priority = 2
	case do.CouponTypeThreshold:
		priority = 3
	case do.CouponTypeDiscount:
		priority = 4
	}

	// 即将过期的优惠券提高优先级
	if template.ValidEnd.Before(time.Now().Add(24 * time.Hour)) {
		priority += 10
	}

	return priority
}

// calculateScore 计算优惠券得分
func (e *CalculationEngine) calculateScore(template *CouponTemplate, orderAmount float64) float64 {
	score := 0.0

	// 根据潜在折扣计算得分
	if do.DiscountType(template.DiscountType) == do.DiscountTypeFixed {
		score = template.DiscountValue
	} else {
		discount := orderAmount * template.DiscountValue / 100
		if template.MaxDiscountAmount > 0 && discount > template.MaxDiscountAmount {
			discount = template.MaxDiscountAmount
		}
		score = discount
	}

	// 满足门槛要求的优惠券得分更高
	if orderAmount >= template.MinAmount {
		score *= 1.2
	}

	return score
}

// buildFinalResult 构造最终结果
func (e *CalculationEngine) buildFinalResult(
	originalAmount float64,
	calculationResults []*CalculationResult,
	validationResults map[int64]*ValidationResult,
) *dto.CouponDiscountResultDTO {
	result := &dto.CouponDiscountResultDTO{
		OriginalAmount:  originalAmount,
		DiscountAmount:  0,
		FinalAmount:     originalAmount,
		AppliedCoupons:  make([]int64, 0),
		RejectedCoupons: make([]*dto.CouponRejection, 0),
	}

	for _, calcResult := range calculationResults {
		if calcResult.IsApplicable {
			result.DiscountAmount += calcResult.DiscountAmount
			result.AppliedCoupons = append(result.AppliedCoupons, calcResult.CouponID)
		} else {
			result.RejectedCoupons = append(result.RejectedCoupons, &dto.CouponRejection{
				CouponID: calcResult.CouponID,
				Reason:   calcResult.Reason,
			})
		}
	}

	// 确保最终金额不为负数
	result.FinalAmount = math.Max(0, originalAmount-result.DiscountAmount)

	log.Infof("计算完成: 原始金额=%.2f, 折扣金额=%.2f, 最终金额=%.2f, 应用优惠券=%v",
		result.OriginalAmount, result.DiscountAmount, result.FinalAmount, result.AppliedCoupons)

	return result
}