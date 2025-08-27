package calculator

import (
	"fmt"
	"math"

	"emshop/internal/app/coupon/srv/domain/do"
	"emshop/pkg/log"
)

// ThresholdCouponStrategy 满减券计算策略
type ThresholdCouponStrategy struct{}

func (s *ThresholdCouponStrategy) Calculate(ctx *CalculationContext, coupon *EnhancedUserCoupon) (*CalculationResult, error) {
	template := coupon.Template
	
	// 检查是否满足最小订单金额
	if ctx.OrderAmount < template.MinAmount {
		return &CalculationResult{
			CouponID:     coupon.UserCoupon.ID,
			IsApplicable: false,
			Reason:       fmt.Sprintf("订单金额不满足最低%.2f元要求", template.MinAmount),
		}, nil
	}

	var discount float64
	if do.DiscountType(template.DiscountType) == do.DiscountTypeFixed {
		discount = template.DiscountValue
	} else {
		// 按比例计算
		discount = ctx.OrderAmount * template.DiscountValue / 100
		if template.MaxDiscountAmount > 0 && discount > template.MaxDiscountAmount {
			discount = template.MaxDiscountAmount
		}
	}

	return &CalculationResult{
		CouponID:        coupon.UserCoupon.ID,
		DiscountAmount:  discount,
		AppliedAmount:   ctx.OrderAmount,
		IsApplicable:    true,
		CalculationInfo: fmt.Sprintf("满减券: 满%.2f减%.2f", template.MinAmount, discount),
	}, nil
}

func (s *ThresholdCouponStrategy) GetPriority() int {
	return 3
}

func (s *ThresholdCouponStrategy) CanApply(ctx *CalculationContext, coupon *EnhancedUserCoupon) bool {
	return ctx.OrderAmount >= coupon.Template.MinAmount
}

// DiscountCouponStrategy 折扣券计算策略
type DiscountCouponStrategy struct{}

func (s *DiscountCouponStrategy) Calculate(ctx *CalculationContext, coupon *EnhancedUserCoupon) (*CalculationResult, error) {
	template := coupon.Template
	
	// 检查是否满足最小订单金额
	if ctx.OrderAmount < template.MinAmount {
		return &CalculationResult{
			CouponID:     coupon.UserCoupon.ID,
			IsApplicable: false,
			Reason:       fmt.Sprintf("订单金额不满足最低%.2f元要求", template.MinAmount),
		}, nil
	}

	var discount float64
	var discountRate float64

	if do.DiscountType(template.DiscountType) == do.DiscountTypePercent {
		// 百分比折扣：DiscountValue 表示折扣比例（如：10表示9折，20表示8折）
		discountRate = template.DiscountValue / 100
		discount = ctx.OrderAmount * discountRate
		
		// 检查最大折扣限制
		if template.MaxDiscountAmount > 0 && discount > template.MaxDiscountAmount {
			discount = template.MaxDiscountAmount
		}
	} else {
		// 固定金额折扣（少见情况）
		discount = template.DiscountValue
	}

	// 确保折扣不超过订单金额
	discount = math.Min(discount, ctx.OrderAmount)

	calculationInfo := ""
	if do.DiscountType(template.DiscountType) == do.DiscountTypePercent {
		actualRate := (100 - template.DiscountValue) / 100
		calculationInfo = fmt.Sprintf("折扣券: %.1f折，折扣金额%.2f", actualRate*10, discount)
	} else {
		calculationInfo = fmt.Sprintf("折扣券: 固定减%.2f", discount)
	}

	return &CalculationResult{
		CouponID:        coupon.UserCoupon.ID,
		DiscountAmount:  discount,
		AppliedAmount:   ctx.OrderAmount,
		IsApplicable:    true,
		CalculationInfo: calculationInfo,
	}, nil
}

func (s *DiscountCouponStrategy) GetPriority() int {
	return 4
}

func (s *DiscountCouponStrategy) CanApply(ctx *CalculationContext, coupon *EnhancedUserCoupon) bool {
	return ctx.OrderAmount >= coupon.Template.MinAmount
}

// InstantCouponStrategy 立减券计算策略
type InstantCouponStrategy struct{}

func (s *InstantCouponStrategy) Calculate(ctx *CalculationContext, coupon *EnhancedUserCoupon) (*CalculationResult, error) {
	template := coupon.Template
	
	var discount float64
	if do.DiscountType(template.DiscountType) == do.DiscountTypeFixed {
		discount = template.DiscountValue
	} else {
		// 按比例立减
		discount = ctx.OrderAmount * template.DiscountValue / 100
		if template.MaxDiscountAmount > 0 && discount > template.MaxDiscountAmount {
			discount = template.MaxDiscountAmount
		}
	}

	// 确保折扣不超过订单金额
	discount = math.Min(discount, ctx.OrderAmount)

	return &CalculationResult{
		CouponID:        coupon.UserCoupon.ID,
		DiscountAmount:  discount,
		AppliedAmount:   ctx.OrderAmount,
		IsApplicable:    true,
		CalculationInfo: fmt.Sprintf("立减券: 立减%.2f", discount),
	}, nil
}

func (s *InstantCouponStrategy) GetPriority() int {
	return 2
}

func (s *InstantCouponStrategy) CanApply(ctx *CalculationContext, coupon *EnhancedUserCoupon) bool {
	return true // 立减券通常无门槛
}

// FreeShipCouponStrategy 包邮券计算策略
type FreeShipCouponStrategy struct{}

func (s *FreeShipCouponStrategy) Calculate(ctx *CalculationContext, coupon *EnhancedUserCoupon) (*CalculationResult, error) {
	template := coupon.Template
	
	// 包邮券的折扣金额通常是固定的邮费金额
	var discount float64
	if do.DiscountType(template.DiscountType) == do.DiscountTypeFixed {
		discount = template.DiscountValue // 通常是邮费金额
	} else {
		// 按比例包邮（罕见情况）
		discount = ctx.OrderAmount * template.DiscountValue / 100
		if template.MaxDiscountAmount > 0 && discount > template.MaxDiscountAmount {
			discount = template.MaxDiscountAmount
		}
	}

	return &CalculationResult{
		CouponID:        coupon.UserCoupon.ID,
		DiscountAmount:  discount,
		AppliedAmount:   discount, // 包邮券应用的金额就是邮费
		IsApplicable:    true,
		CalculationInfo: fmt.Sprintf("包邮券: 免邮费%.2f", discount),
	}, nil
}

func (s *FreeShipCouponStrategy) GetPriority() int {
	return 1 // 包邮券优先级最低
}

func (s *FreeShipCouponStrategy) CanApply(ctx *CalculationContext, coupon *EnhancedUserCoupon) bool {
	// 检查是否满足包邮门槛
	if coupon.Template.MinAmount > 0 {
		return ctx.OrderAmount >= coupon.Template.MinAmount
	}
	return true
}

// StrategyFactory 策略工厂
type StrategyFactory struct{}

// GetStrategy 根据优惠券类型获取计算策略
func (f *StrategyFactory) GetStrategy(couponType do.CouponType) (CalculationStrategy, error) {
	switch couponType {
	case do.CouponTypeThreshold:
		return &ThresholdCouponStrategy{}, nil
	case do.CouponTypeDiscount:
		return &DiscountCouponStrategy{}, nil
	case do.CouponTypeInstant:
		return &InstantCouponStrategy{}, nil
	case do.CouponTypeFreeShip:
		return &FreeShipCouponStrategy{}, nil
	default:
		return nil, fmt.Errorf("不支持的优惠券类型: %v", couponType)
	}
}

// MultiCouponStrategy 多优惠券组合策略
type MultiCouponStrategy struct {
	strategies map[do.CouponType]CalculationStrategy
}

// NewMultiCouponStrategy 创建多优惠券组合策略
func NewMultiCouponStrategy() *MultiCouponStrategy {
	strategies := make(map[do.CouponType]CalculationStrategy)
	strategies[do.CouponTypeThreshold] = &ThresholdCouponStrategy{}
	strategies[do.CouponTypeDiscount] = &DiscountCouponStrategy{}
	strategies[do.CouponTypeInstant] = &InstantCouponStrategy{}
	strategies[do.CouponTypeFreeShip] = &FreeShipCouponStrategy{}

	return &MultiCouponStrategy{
		strategies: strategies,
	}
}

// CalculateOptimalCombination 计算最优组合
func (ms *MultiCouponStrategy) CalculateOptimalCombination(
	ctx *CalculationContext, 
	coupons []*EnhancedUserCoupon,
) ([]*CalculationResult, error) {
	log.Infof("开始计算最优优惠券组合，候选优惠券数量: %d", len(coupons))

	// 将优惠券按类型分组
	couponGroups := make(map[do.CouponType][]*EnhancedUserCoupon)
	for _, coupon := range coupons {
		couponType := do.CouponType(coupon.Template.Type)
		couponGroups[couponType] = append(couponGroups[couponType], coupon)
	}

	results := make([]*CalculationResult, 0)
	currentAmount := ctx.OrderAmount

	// 优先使用包邮券（影响最小）
	if freeShipCoupons := couponGroups[do.CouponTypeFreeShip]; len(freeShipCoupons) > 0 {
		strategy := ms.strategies[do.CouponTypeFreeShip]
		for _, coupon := range freeShipCoupons {
			if strategy.CanApply(ctx, coupon) {
				result, err := strategy.Calculate(ctx, coupon)
				if err == nil && result.IsApplicable {
					results = append(results, result)
					break // 只使用一个包邮券
				}
			}
		}
	}

	// 然后使用折扣券或满减券（选择最优的一个）
	bestResult := ms.selectBestDiscountCoupon(ctx, currentAmount, couponGroups)
	if bestResult != nil {
		results = append(results, bestResult)
		currentAmount -= bestResult.DiscountAmount
	}

	// 最后使用立减券
	if instantCoupons := couponGroups[do.CouponTypeInstant]; len(instantCoupons) > 0 {
		strategy := ms.strategies[do.CouponTypeInstant]
		// 更新上下文中的订单金额（考虑之前的折扣）
		tempCtx := *ctx
		tempCtx.OrderAmount = currentAmount
		
		for _, coupon := range instantCoupons {
			if strategy.CanApply(&tempCtx, coupon) {
				result, err := strategy.Calculate(&tempCtx, coupon)
				if err == nil && result.IsApplicable {
					results = append(results, result)
					break // 只使用一个立减券
				}
			}
		}
	}

	log.Infof("最优组合计算完成，选中优惠券数量: %d", len(results))
	return results, nil
}

// selectBestDiscountCoupon 选择最好的折扣券
func (ms *MultiCouponStrategy) selectBestDiscountCoupon(
	ctx *CalculationContext, 
	currentAmount float64, 
	couponGroups map[do.CouponType][]*EnhancedUserCoupon,
) *CalculationResult {
	var bestResult *CalculationResult
	maxDiscount := 0.0

	// 比较满减券
	if thresholdCoupons := couponGroups[do.CouponTypeThreshold]; len(thresholdCoupons) > 0 {
		strategy := ms.strategies[do.CouponTypeThreshold]
		for _, coupon := range thresholdCoupons {
			if strategy.CanApply(ctx, coupon) {
				result, err := strategy.Calculate(ctx, coupon)
				if err == nil && result.IsApplicable && result.DiscountAmount > maxDiscount {
					maxDiscount = result.DiscountAmount
					bestResult = result
				}
			}
		}
	}

	// 比较折扣券
	if discountCoupons := couponGroups[do.CouponTypeDiscount]; len(discountCoupons) > 0 {
		strategy := ms.strategies[do.CouponTypeDiscount]
		for _, coupon := range discountCoupons {
			if strategy.CanApply(ctx, coupon) {
				result, err := strategy.Calculate(ctx, coupon)
				if err == nil && result.IsApplicable && result.DiscountAmount > maxDiscount {
					maxDiscount = result.DiscountAmount
					bestResult = result
				}
			}
		}
	}

	return bestResult
}