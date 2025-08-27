package calculator

import (
	"fmt"
	"time"

	"emshop/internal/app/coupon/srv/domain/do"
	"emshop/pkg/log"
)

// BusinessRule 业务规则接口
type BusinessRule interface {
	Execute(ctx *CalculationContext) error
	GetPriority() int
	GetRuleName() string
}

// BusinessRuleEngine 业务规则引擎
type BusinessRuleEngine struct {
	rules []BusinessRule
}

// NewBusinessRuleEngine 创建业务规则引擎
func NewBusinessRuleEngine() *BusinessRuleEngine {
	engine := &BusinessRuleEngine{
		rules: make([]BusinessRule, 0),
	}

	// 注册默认规则
	engine.registerDefaultRules()
	
	return engine
}

// registerDefaultRules 注册默认规则
func (bre *BusinessRuleEngine) registerDefaultRules() {
	bre.AddRule(&OrderAmountRule{})
	bre.AddRule(&CouponCountRule{maxCount: 5})
	bre.AddRule(&TimePeriodRule{})
	bre.AddRule(&UserTypeRule{})
	bre.AddRule(&CombinationRule{})
}

// AddRule 添加规则
func (bre *BusinessRuleEngine) AddRule(rule BusinessRule) {
	bre.rules = append(bre.rules, rule)
	
	// 按优先级排序
	for i := len(bre.rules) - 1; i > 0; i-- {
		if bre.rules[i].GetPriority() > bre.rules[i-1].GetPriority() {
			bre.rules[i], bre.rules[i-1] = bre.rules[i-1], bre.rules[i]
		} else {
			break
		}
	}
}

// ExecuteRules 执行所有规则
func (bre *BusinessRuleEngine) ExecuteRules(ctx *CalculationContext) error {
	log.Infof("开始执行业务规则检查，规则数量: %d", len(bre.rules))
	
	for _, rule := range bre.rules {
		if err := rule.Execute(ctx); err != nil {
			log.Warnf("业务规则 %s 执行失败: %v", rule.GetRuleName(), err)
			return fmt.Errorf("业务规则 %s 执行失败: %v", rule.GetRuleName(), err)
		}
	}
	
	log.Info("业务规则检查完成")
	return nil
}

// OrderAmountRule 订单金额规则
type OrderAmountRule struct{}

func (r *OrderAmountRule) Execute(ctx *CalculationContext) error {
	if ctx.OrderAmount <= 0 {
		return fmt.Errorf("订单金额必须大于0")
	}
	
	if ctx.OrderAmount > 100000 { // 10万元限制
		return fmt.Errorf("订单金额超过系统限制")
	}
	
	return nil
}

func (r *OrderAmountRule) GetPriority() int {
	return 100
}

func (r *OrderAmountRule) GetRuleName() string {
	return "OrderAmountRule"
}

// CouponCountRule 优惠券数量规则
type CouponCountRule struct {
	maxCount int
}

func (r *CouponCountRule) Execute(ctx *CalculationContext) error {
	if len(ctx.CouponIDs) > r.maxCount {
		return fmt.Errorf("单次最多使用%d张优惠券", r.maxCount)
	}
	
	return nil
}

func (r *CouponCountRule) GetPriority() int {
	return 90
}

func (r *CouponCountRule) GetRuleName() string {
	return "CouponCountRule"
}

// TimePeriodRule 时间周期规则
type TimePeriodRule struct{}

func (r *TimePeriodRule) Execute(ctx *CalculationContext) error {
	currentTime := ctx.CurrentTime
	
	// 检查是否在禁用时间段（例如：系统维护时间）
	if r.isMaintenanceTime(currentTime) {
		return fmt.Errorf("系统维护中，暂时无法使用优惠券")
	}
	
	// 检查是否在活动时间段
	if r.isRestrictedTime(currentTime) {
		log.Warnf("当前时间段优惠券使用受限: %v", currentTime)
	}
	
	return nil
}

func (r *TimePeriodRule) isMaintenanceTime(currentTime time.Time) bool {
	// 例如：每天凌晨2-4点为维护时间
	hour := currentTime.Hour()
	return hour >= 2 && hour < 4
}

func (r *TimePeriodRule) isRestrictedTime(currentTime time.Time) bool {
	// 例如：周末的某些时间段限制使用
	return false
}

func (r *TimePeriodRule) GetPriority() int {
	return 80
}

func (r *TimePeriodRule) GetRuleName() string {
	return "TimePeriodRule"
}

// UserTypeRule 用户类型规则
type UserTypeRule struct{}

func (r *UserTypeRule) Execute(ctx *CalculationContext) error {
	// 这里可以根据用户ID查询用户类型
	// 然后应用相应的规则
	userType := r.getUserType(ctx.UserID)
	
	switch userType {
	case "blacklist":
		return fmt.Errorf("用户被限制使用优惠券")
	case "restricted":
		if len(ctx.CouponIDs) > 1 {
			return fmt.Errorf("受限用户每次只能使用1张优惠券")
		}
	}
	
	return nil
}

func (r *UserTypeRule) getUserType(userID int64) string {
	// 简化实现：实际应该查询用户服务
	return "normal"
}

func (r *UserTypeRule) GetPriority() int {
	return 70
}

func (r *UserTypeRule) GetRuleName() string {
	return "UserTypeRule"
}

// CombinationRule 组合规则
type CombinationRule struct{}

func (r *CombinationRule) Execute(ctx *CalculationContext) error {
	if len(ctx.CouponIDs) <= 1 {
		return nil // 单张优惠券无需检查组合规则
	}
	
	// 检查优惠券类型组合是否合法
	typeCount := make(map[do.CouponType]int)
	for _, coupon := range ctx.UserCoupons {
		if coupon.Template != nil {
			couponType := do.CouponType(coupon.Template.Type)
			typeCount[couponType]++
		}
	}
	
	// 应用组合规则
	if err := r.validateCombination(typeCount); err != nil {
		return err
	}
	
	return nil
}

func (r *CombinationRule) validateCombination(typeCount map[do.CouponType]int) error {
	// 规则1：最多只能有一张折扣券
	if typeCount[do.CouponTypeDiscount] > 1 {
		return fmt.Errorf("不能同时使用多张折扣券")
	}
	
	// 规则2：最多只能有一张满减券
	if typeCount[do.CouponTypeThreshold] > 1 {
		return fmt.Errorf("不能同时使用多张满减券")
	}
	
	// 规则3：折扣券和满减券不能同时使用
	if typeCount[do.CouponTypeDiscount] > 0 && typeCount[do.CouponTypeThreshold] > 0 {
		return fmt.Errorf("折扣券和满减券不能同时使用")
	}
	
	// 规则4：立减券最多使用2张
	if typeCount[do.CouponTypeInstant] > 2 {
		return fmt.Errorf("立减券最多使用2张")
	}
	
	return nil
}

func (r *CombinationRule) GetPriority() int {
	return 60
}

func (r *CombinationRule) GetRuleName() string {
	return "CombinationRule"
}

// SeasonalRule 季节性规则
type SeasonalRule struct{}

func (r *SeasonalRule) Execute(ctx *CalculationContext) error {
	currentTime := ctx.CurrentTime
	
	// 检查是否在特殊节假日期间
	if r.isSpecialHoliday(currentTime) {
		// 节假日期间可能有特殊规则
		return r.applyHolidayRules(ctx)
	}
	
	// 检查是否在促销季
	if r.isPromotionSeason(currentTime) {
		// 促销季可能有额外限制或优惠
		return r.applySeasonalRules(ctx)
	}
	
	return nil
}

func (r *SeasonalRule) isSpecialHoliday(t time.Time) bool {
	// 检查是否在双11、618等特殊日期
	month := t.Month()
	day := t.Day()
	
	// 双11
	if month == 11 && day == 11 {
		return true
	}
	
	// 618
	if month == 6 && day == 18 {
		return true
	}
	
	return false
}

func (r *SeasonalRule) isPromotionSeason(t time.Time) bool {
	month := t.Month()
	// 年末促销季
	return month == 11 || month == 12
}

func (r *SeasonalRule) applyHolidayRules(ctx *CalculationContext) error {
	// 节假日特殊规则：可能允许使用更多优惠券
	if len(ctx.CouponIDs) > 8 {
		return fmt.Errorf("节假日期间每次最多使用8张优惠券")
	}
	return nil
}

func (r *SeasonalRule) applySeasonalRules(ctx *CalculationContext) error {
	// 促销季规则：可能有额外的限制
	return nil
}

func (r *SeasonalRule) GetPriority() int {
	return 50
}

func (r *SeasonalRule) GetRuleName() string {
	return "SeasonalRule"
}

// RiskControlRule 风控规则
type RiskControlRule struct{}

func (r *RiskControlRule) Execute(ctx *CalculationContext) error {
	// 检查是否存在异常行为
	if r.isAbnormalBehavior(ctx) {
		return fmt.Errorf("检测到异常行为，请稍后重试")
	}
	
	// 检查优惠券价值是否异常
	if r.isAbnormalCouponValue(ctx) {
		log.Warnf("检测到高价值优惠券使用，用户: %d", ctx.UserID)
		// 可能需要额外验证，但不直接拒绝
	}
	
	return nil
}

func (r *RiskControlRule) isAbnormalBehavior(ctx *CalculationContext) bool {
	// 检查是否在短时间内多次计算
	// 检查优惠券组合是否异常
	// 简化实现
	return false
}

func (r *RiskControlRule) isAbnormalCouponValue(ctx *CalculationContext) bool {
	totalValue := 0.0
	for _, coupon := range ctx.UserCoupons {
		if coupon.Template != nil {
			if do.DiscountType(coupon.Template.DiscountType) == do.DiscountTypeFixed {
				totalValue += coupon.Template.DiscountValue
			} else {
				// 对于百分比折扣，估算最大可能价值
				maxValue := ctx.OrderAmount * coupon.Template.DiscountValue / 100
				if coupon.Template.MaxDiscountAmount > 0 && maxValue > coupon.Template.MaxDiscountAmount {
					maxValue = coupon.Template.MaxDiscountAmount
				}
				totalValue += maxValue
			}
		}
	}
	
	// 如果总价值超过订单金额的50%，认为异常
	return totalValue > ctx.OrderAmount*0.5
}

func (r *RiskControlRule) GetPriority() int {
	return 40
}

func (r *RiskControlRule) GetRuleName() string {
	return "RiskControlRule"
}