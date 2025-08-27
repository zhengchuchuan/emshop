package calculator

import (
	"fmt"

	"emshop/internal/app/coupon/srv/domain/do"
	"emshop/internal/app/coupon/srv/domain/dto"
)

// BasicCouponValidator 基础优惠券验证器
type BasicCouponValidator struct{}

func (v *BasicCouponValidator) Validate(ctx *CalculationContext, coupon *EnhancedUserCoupon) *ValidationResult {
	userCoupon := coupon.UserCoupon
	template := coupon.Template

	// 检查用户ID是否匹配
	if userCoupon.UserID != ctx.UserID {
		return &ValidationResult{
			IsValid: false,
			Reason:  "优惠券不属于当前用户",
			Code:    "USER_MISMATCH",
		}
	}

	// 检查优惠券状态
	if userCoupon.Status != int32(do.UserCouponStatusUnused) {
		return &ValidationResult{
			IsValid: false,
			Reason:  "优惠券不可用",
			Code:    "COUPON_UNAVAILABLE",
		}
	}

	// 检查优惠券模板状态
	if template.Status != int32(do.CouponStatusActive) {
		return &ValidationResult{
			IsValid: false,
			Reason:  "优惠券模板已停用",
			Code:    "TEMPLATE_INACTIVE",
		}
	}

	return &ValidationResult{IsValid: true}
}

// TimingValidator 时间验证器
type TimingValidator struct{}

func (v *TimingValidator) Validate(ctx *CalculationContext, coupon *EnhancedUserCoupon) *ValidationResult {
	userCoupon := coupon.UserCoupon
	template := coupon.Template
	currentTime := ctx.CurrentTime

	// 检查用户优惠券是否过期
	if userCoupon.ValidEndTime.Before(currentTime) {
		return &ValidationResult{
			IsValid: false,
			Reason:  "用户优惠券已过期",
			Code:    "USER_COUPON_EXPIRED",
		}
	}

	// 检查用户优惠券是否生效
	if userCoupon.ValidStartTime.After(currentTime) {
		return &ValidationResult{
			IsValid: false,
			Reason:  "用户优惠券尚未生效",
			Code:    "USER_COUPON_NOT_ACTIVE",
		}
	}

	// 检查优惠券模板有效期
	if template.ValidEnd.Before(currentTime) {
		return &ValidationResult{
			IsValid: false,
			Reason:  "优惠券模板已过期",
			Code:    "TEMPLATE_EXPIRED",
		}
	}

	if template.ValidStart.After(currentTime) {
		return &ValidationResult{
			IsValid: false,
			Reason:  "优惠券模板尚未生效",
			Code:    "TEMPLATE_NOT_ACTIVE",
		}
	}

	return &ValidationResult{IsValid: true}
}

// AmountValidator 金额验证器
type AmountValidator struct{}

func (v *AmountValidator) Validate(ctx *CalculationContext, coupon *EnhancedUserCoupon) *ValidationResult {
	template := coupon.Template

	// 检查最小订单金额要求
	if ctx.OrderAmount < template.MinAmount {
		return &ValidationResult{
			IsValid: false,
			Reason:  fmt.Sprintf("订单金额不满足最低%.2f元要求", template.MinAmount),
			Code:    "MIN_AMOUNT_NOT_SATISFIED",
		}
	}

	// 检查订单金额是否为正数
	if ctx.OrderAmount <= 0 {
		return &ValidationResult{
			IsValid: false,
			Reason:  "订单金额必须大于0",
			Code:    "INVALID_ORDER_AMOUNT",
		}
	}

	return &ValidationResult{IsValid: true}
}

// UserLimitValidator 用户限制验证器
type UserLimitValidator struct{}

func (v *UserLimitValidator) Validate(ctx *CalculationContext, coupon *EnhancedUserCoupon) *ValidationResult {
	// 这里可以添加用户级别的限制检查
	// 例如：用户类型限制、会员等级限制等
	
	// 检查是否重复使用同一张优惠券
	for _, applied := range ctx.AppliedCoupons {
		if applied.CouponID == coupon.UserCoupon.ID {
			return &ValidationResult{
				IsValid: false,
				Reason:  "不能重复使用同一张优惠券",
				Code:    "DUPLICATE_COUPON",
			}
		}
	}

	return &ValidationResult{IsValid: true}
}

// CombinationValidator 组合验证器
type CombinationValidator struct{}

func (v *CombinationValidator) Validate(ctx *CalculationContext, coupon *EnhancedUserCoupon) *ValidationResult {
	template := coupon.Template
	couponType := do.CouponType(template.Type)

	// 检查优惠券组合规则
	for _, applied := range ctx.AppliedCoupons {
		// 某些类型的优惠券不能同时使用
		if v.isConflictingCombination(couponType, applied.Strategy) {
			return &ValidationResult{
				IsValid: false,
				Reason:  "该优惠券不能与已选择的优惠券同时使用",
				Code:    "CONFLICTING_COMBINATION",
			}
		}
	}

	return &ValidationResult{IsValid: true}
}

// isConflictingCombination 检查优惠券组合冲突
func (v *CombinationValidator) isConflictingCombination(currentType do.CouponType, appliedStrategy string) bool {
	// 定义冲突规则
	// 例如：同一类型的优惠券不能同时使用
	switch currentType {
	case do.CouponTypeThreshold:
		return appliedStrategy == "ThresholdCouponStrategy"
	case do.CouponTypeDiscount:
		return appliedStrategy == "DiscountCouponStrategy"
	default:
		return false
	}
}

// BusinessRuleValidator 业务规则验证器
type BusinessRuleValidator struct {
	maxCouponsPerOrder int
	allowedCombinations map[string][]string
}

// NewBusinessRuleValidator 创建业务规则验证器
func NewBusinessRuleValidator() *BusinessRuleValidator {
	return &BusinessRuleValidator{
		maxCouponsPerOrder: 3, // 每个订单最多使用3张优惠券
		allowedCombinations: map[string][]string{
			"threshold": {"freeship"},
			"discount": {"freeship"},
			"instant": {"freeship", "threshold"},
		},
	}
}

func (v *BusinessRuleValidator) Validate(ctx *CalculationContext, coupon *EnhancedUserCoupon) *ValidationResult {
	// 检查优惠券使用数量限制
	if len(ctx.AppliedCoupons) >= v.maxCouponsPerOrder {
		return &ValidationResult{
			IsValid: false,
			Reason:  fmt.Sprintf("每个订单最多使用%d张优惠券", v.maxCouponsPerOrder),
			Code:    "MAX_COUPONS_EXCEEDED",
		}
	}

	// 检查组合规则
	template := coupon.Template
	currentType := v.getCouponTypeString(do.CouponType(template.Type))
	
	for _, applied := range ctx.AppliedCoupons {
		if !v.isValidCombination(currentType, applied.Strategy) {
			return &ValidationResult{
				IsValid: false,
				Reason:  "该优惠券组合不被允许",
				Code:    "INVALID_COMBINATION",
			}
		}
	}

	return &ValidationResult{IsValid: true}
}

// getCouponTypeString 获取优惠券类型字符串
func (v *BusinessRuleValidator) getCouponTypeString(couponType do.CouponType) string {
	switch couponType {
	case do.CouponTypeThreshold:
		return "threshold"
	case do.CouponTypeDiscount:
		return "discount"
	case do.CouponTypeInstant:
		return "instant"
	case do.CouponTypeFreeShip:
		return "freeship"
	default:
		return "unknown"
	}
}

// isValidCombination 检查组合是否有效
func (v *BusinessRuleValidator) isValidCombination(currentType, appliedStrategy string) bool {
	allowedTypes, exists := v.allowedCombinations[currentType]
	if !exists {
		return false
	}

	for _, allowedType := range allowedTypes {
		if appliedStrategy == allowedType {
			return true
		}
	}
	return false
}

// SpecialRuleValidator 特殊规则验证器
type SpecialRuleValidator struct{}

func (v *SpecialRuleValidator) Validate(ctx *CalculationContext, coupon *EnhancedUserCoupon) *ValidationResult {
	template := coupon.Template

	// 检查特殊商品限制（如果订单项中包含特殊商品）
	if v.hasRestrictedItems(ctx.OrderItems) {
		// 某些优惠券不能用于特殊商品
		if v.isRestrictedForSpecialItems(do.CouponType(template.Type)) {
			return &ValidationResult{
				IsValid: false,
				Reason:  "该优惠券不适用于订单中的特殊商品",
				Code:    "RESTRICTED_FOR_SPECIAL_ITEMS",
			}
		}
	}

	// 检查新用户专享优惠券
	if v.isNewUserCoupon(template) && !v.isNewUser(ctx.UserID) {
		return &ValidationResult{
			IsValid: false,
			Reason:  "该优惠券仅限新用户使用",
			Code:    "NEW_USER_ONLY",
		}
	}

	return &ValidationResult{IsValid: true}
}

// hasRestrictedItems 检查是否有受限商品
func (v *SpecialRuleValidator) hasRestrictedItems(items []*dto.OrderItemDTO) bool {
	// 这里可以根据商品ID检查是否是特殊商品
	// 例如：奢侈品、特价商品等
	return false // 简化实现
}

// isRestrictedForSpecialItems 检查优惠券是否对特殊商品受限
func (v *SpecialRuleValidator) isRestrictedForSpecialItems(couponType do.CouponType) bool {
	// 某些类型的优惠券不能用于特殊商品
	switch couponType {
	case do.CouponTypeDiscount:
		return true // 折扣券不能用于特殊商品
	default:
		return false
	}
}

// isNewUserCoupon 检查是否是新用户专享优惠券
func (v *SpecialRuleValidator) isNewUserCoupon(template *CouponTemplate) bool {
	// 这里可以根据优惠券名称或其他标识判断
	// 简化实现：假设名称包含"新用户"字样
	return false
}

// isNewUser 检查是否是新用户
func (v *SpecialRuleValidator) isNewUser(userID int64) bool {
	// 这里应该查询用户服务判断是否是新用户
	// 简化实现
	return false
}

// ValidatorChain 验证器链
type ValidatorChain struct {
	validators []CouponValidator
}

// NewValidatorChain 创建验证器链
func NewValidatorChain() *ValidatorChain {
	return &ValidatorChain{
		validators: make([]CouponValidator, 0),
	}
}

// AddValidator 添加验证器
func (vc *ValidatorChain) AddValidator(validator CouponValidator) *ValidatorChain {
	vc.validators = append(vc.validators, validator)
	return vc
}

// Validate 执行所有验证器
func (vc *ValidatorChain) Validate(ctx *CalculationContext, coupon *EnhancedUserCoupon) *ValidationResult {
	for _, validator := range vc.validators {
		result := validator.Validate(ctx, coupon)
		if !result.IsValid {
			return result
		}
	}
	return &ValidationResult{IsValid: true}
}