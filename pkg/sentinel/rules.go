package sentinel

import (
	"encoding/json"

	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/hotspot"
	"github.com/alibaba/sentinel-golang/core/system"
)

// BusinessRules 业务场景规则定义
type BusinessRules struct {
	// 秒杀活动规则
	FlashSale *FlashSaleRules `json:"flashSale"`
	// 支付流程规则
	Payment *PaymentRules `json:"payment"`
	// 库存管理规则
	Inventory *InventoryRules `json:"inventory"`
	// 优惠券规则
	Coupon *CouponRules `json:"coupon"`
	// 用户服务规则
	User *UserRules `json:"user"`
	// 商品服务规则
	Goods *GoodsRules `json:"goods"`
	// 订单服务规则
	Order *OrderRules `json:"order"`
}

// FlashSaleRules 秒杀活动规则
type FlashSaleRules struct {
	// 秒杀接口QPS限制
	FlashSaleQPS     float64 `json:"flashSaleQPS"`
	// 秒杀商品详情QPS限制  
	ProductDetailQPS float64 `json:"productDetailQPS"`
	// 秒杀下单QPS限制
	OrderQPS         float64 `json:"orderQPS"`
	// 熔断阈值
	CircuitBreaker   CircuitBreakerConfig `json:"circuitBreaker"`
}

// PaymentRules 支付流程规则
type PaymentRules struct {
	// 支付接口QPS限制
	PaymentQPS       float64 `json:"paymentQPS"`
	// 支付查询QPS限制
	PaymentQueryQPS  float64 `json:"paymentQueryQPS"`
	// 退款QPS限制
	RefundQPS        float64 `json:"refundQPS"`
	// 熔断配置
	CircuitBreaker   CircuitBreakerConfig `json:"circuitBreaker"`
}

// InventoryRules 库存管理规则
type InventoryRules struct {
	// 库存扣减QPS限制
	DeductQPS        float64 `json:"deductQPS"`
	// 库存查询QPS限制
	QueryQPS         float64 `json:"queryQPS"`
	// 库存恢复QPS限制
	RestoreQPS       float64 `json:"restoreQPS"`
	// 熔断配置
	CircuitBreaker   CircuitBreakerConfig `json:"circuitBreaker"`
}

// CouponRules 优惠券规则
type CouponRules struct {
	// 优惠券发放QPS限制
	IssueQPS         float64 `json:"issueQPS"`
	// 优惠券使用QPS限制
	UseQPS           float64 `json:"useQPS"`
	// 优惠券查询QPS限制
	QueryQPS         float64 `json:"queryQPS"`
	// 热点参数限制(用户ID)
	UserHotspot      HotspotConfig `json:"userHotspot"`
	// 熔断配置
	CircuitBreaker   CircuitBreakerConfig `json:"circuitBreaker"`
}

// UserRules 用户服务规则
type UserRules struct {
	// 用户登录QPS限制
	LoginQPS         float64 `json:"loginQPS"`
	// 用户注册QPS限制
	RegisterQPS      float64 `json:"registerQPS"`
	// 用户查询QPS限制
	QueryQPS         float64 `json:"queryQPS"`
	// 热点参数限制
	MobileHotspot    HotspotConfig `json:"mobileHotspot"`
	// 熔断配置
	CircuitBreaker   CircuitBreakerConfig `json:"circuitBreaker"`
}

// GoodsRules 商品服务规则
type GoodsRules struct {
	// 商品列表QPS限制
	ListQPS          float64 `json:"listQPS"`
	// 商品详情QPS限制
	DetailQPS        float64 `json:"detailQPS"`
	// 商品搜索QPS限制
	SearchQPS        float64 `json:"searchQPS"`
	// 热点商品限制
	ProductHotspot   HotspotConfig `json:"productHotspot"`
	// 熔断配置
	CircuitBreaker   CircuitBreakerConfig `json:"circuitBreaker"`
}

// OrderRules 订单服务规则
type OrderRules struct {
	// 创建订单QPS限制
	CreateQPS        float64 `json:"createQPS"`
	// 订单查询QPS限制
	QueryQPS         float64 `json:"queryQPS"`
	// 订单取消QPS限制
	CancelQPS        float64 `json:"cancelQPS"`
	// 熔断配置
	CircuitBreaker   CircuitBreakerConfig `json:"circuitBreaker"`
}

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	// 是否启用
	Enabled          bool    `json:"enabled"`
	// 失败率阈值
	ErrorRatio       float64 `json:"errorRatio"`
	// 慢调用比例阈值
	SlowRatio        float64 `json:"slowRatio"`
	// 慢调用时间阈值(毫秒)
	SlowTimeMs       int64   `json:"slowTimeMs"`
	// 最小请求数
	MinRequestAmount uint64  `json:"minRequestAmount"`
	// 熔断持续时间(秒)
	RecoveryTimeoutSec uint32 `json:"recoveryTimeoutSec"`
}

// HotspotConfig 热点参数配置
type HotspotConfig struct {
	// 是否启用
	Enabled          bool               `json:"enabled"`
	// 参数索引
	ParamIndex       int                `json:"paramIndex"`
	// 默认阈值
	Count            int64              `json:"count"`
	// 时间窗口(秒)
	DurationInSec    int64              `json:"durationInSec"`
	// 特定值的阈值
	SpecificItems    []HotspotItem      `json:"specificItems"`
}

// HotspotItem 热点参数特定项
type HotspotItem struct {
	Value     string `json:"value"`
	Threshold int64  `json:"threshold"`
}

// DefaultBusinessRules 获取默认业务规则
func DefaultBusinessRules() *BusinessRules {
	return &BusinessRules{
		FlashSale: &FlashSaleRules{
			FlashSaleQPS:     1000,
			ProductDetailQPS: 2000,
			OrderQPS:         500,
			CircuitBreaker: CircuitBreakerConfig{
				Enabled:              true,
				ErrorRatio:           0.5,
				SlowRatio:           0.6,
				SlowTimeMs:          1000,
				MinRequestAmount:    20,
				RecoveryTimeoutSec:  10,
			},
		},
		Payment: &PaymentRules{
			PaymentQPS:      200,
			PaymentQueryQPS: 1000,
			RefundQPS:       100,
			CircuitBreaker: CircuitBreakerConfig{
				Enabled:              true,
				ErrorRatio:           0.3,
				SlowRatio:           0.5,
				SlowTimeMs:          2000,
				MinRequestAmount:    10,
				RecoveryTimeoutSec:  15,
			},
		},
		Inventory: &InventoryRules{
			DeductQPS:  300,
			QueryQPS:   2000,
			RestoreQPS: 200,
			CircuitBreaker: CircuitBreakerConfig{
				Enabled:              true,
				ErrorRatio:           0.4,
				SlowRatio:           0.6,
				SlowTimeMs:          1000,
				MinRequestAmount:    15,
				RecoveryTimeoutSec:  10,
			},
		},
		Coupon: &CouponRules{
			IssueQPS: 500,
			UseQPS:   800,
			QueryQPS: 1500,
			UserHotspot: HotspotConfig{
				Enabled:       true,
				ParamIndex:    0,
				Count:         10,
				DurationInSec: 1,
				SpecificItems: []HotspotItem{
					{Value: "vip_user", Threshold: 50},
				},
			},
			CircuitBreaker: CircuitBreakerConfig{
				Enabled:              true,
				ErrorRatio:           0.3,
				SlowRatio:           0.5,
				SlowTimeMs:          800,
				MinRequestAmount:    10,
				RecoveryTimeoutSec:  8,
			},
		},
		User: &UserRules{
			LoginQPS:    300,
			RegisterQPS: 100,
			QueryQPS:    1000,
			MobileHotspot: HotspotConfig{
				Enabled:       true,
				ParamIndex:    0,
				Count:         5,
				DurationInSec: 60,
			},
			CircuitBreaker: CircuitBreakerConfig{
				Enabled:              true,
				ErrorRatio:           0.4,
				SlowRatio:           0.6,
				SlowTimeMs:          1000,
				MinRequestAmount:    10,
				RecoveryTimeoutSec:  10,
			},
		},
		Goods: &GoodsRules{
			ListQPS:   2000,
			DetailQPS: 3000,
			SearchQPS: 1500,
			ProductHotspot: HotspotConfig{
				Enabled:       true,
				ParamIndex:    0,
				Count:         100,
				DurationInSec: 1,
			},
			CircuitBreaker: CircuitBreakerConfig{
				Enabled:              true,
				ErrorRatio:           0.5,
				SlowRatio:           0.7,
				SlowTimeMs:          1000,
				MinRequestAmount:    20,
				RecoveryTimeoutSec:  10,
			},
		},
		Order: &OrderRules{
			CreateQPS: 500,
			QueryQPS:  1500,
			CancelQPS: 200,
			CircuitBreaker: CircuitBreakerConfig{
				Enabled:              true,
				ErrorRatio:           0.4,
				SlowRatio:           0.6,
				SlowTimeMs:          1500,
				MinRequestAmount:    15,
				RecoveryTimeoutSec:  12,
			},
		},
	}
}

// GenerateFlowRules 生成流控规则
func (br *BusinessRules) GenerateFlowRules(serviceName string) ([]*flow.Rule, error) {
	rules := make([]*flow.Rule, 0)

	switch serviceName {
	case "emshop-coupon-srv":
		if br.Coupon != nil {
			rules = append(rules, 
				&flow.Rule{
					Resource:               "coupon-srv:IssueCoupon",
					TokenCalculateStrategy: flow.Direct,
					ControlBehavior:        flow.Reject,
					Threshold:              br.Coupon.IssueQPS,
					StatIntervalInMs:       1000,
				},
				&flow.Rule{
					Resource:               "coupon-srv:UseCoupon", 
					TokenCalculateStrategy: flow.Direct,
					ControlBehavior:        flow.Reject,
					Threshold:              br.Coupon.UseQPS,
					StatIntervalInMs:       1000,
				},
				&flow.Rule{
					Resource:               "coupon-srv:GetUserCoupons",
					TokenCalculateStrategy: flow.Direct,
					ControlBehavior:        flow.Reject,
					Threshold:              br.Coupon.QueryQPS,
					StatIntervalInMs:       1000,
				},
			)
		}
	case "emshop-user-srv":
		if br.User != nil {
			rules = append(rules,
				&flow.Rule{
					Resource:               "user-srv:CreateUser",
					TokenCalculateStrategy: flow.Direct,
					ControlBehavior:        flow.Reject,
					Threshold:              br.User.RegisterQPS,
					StatIntervalInMs:       1000,
				},
				&flow.Rule{
					Resource:               "user-srv:GetUserByMobile",
					TokenCalculateStrategy: flow.Direct,
					ControlBehavior:        flow.Reject,
					Threshold:              br.User.LoginQPS,
					StatIntervalInMs:       1000,
				},
				&flow.Rule{
					Resource:               "user-srv:GetUserById",
					TokenCalculateStrategy: flow.Direct,
					ControlBehavior:        flow.Reject,
					Threshold:              br.User.QueryQPS,
					StatIntervalInMs:       1000,
				},
			)
		}
	case "emshop-inventory-srv":
		if br.Inventory != nil {
			rules = append(rules,
				&flow.Rule{
					Resource:               "inventory-srv:Sell",
					TokenCalculateStrategy: flow.Direct,
					ControlBehavior:        flow.Reject,
					Threshold:              br.Inventory.DeductQPS,
					StatIntervalInMs:       1000,
				},
				&flow.Rule{
					Resource:               "inventory-srv:InvDetail",
					TokenCalculateStrategy: flow.Direct,
					ControlBehavior:        flow.Reject,
					Threshold:              br.Inventory.QueryQPS,
					StatIntervalInMs:       1000,
				},
			)
		}
	}

	return rules, nil
}

// GenerateCircuitBreakerRules 生成熔断规则
func (br *BusinessRules) GenerateCircuitBreakerRules(serviceName string) ([]*circuitbreaker.Rule, error) {
	rules := make([]*circuitbreaker.Rule, 0)

	var cbConfig *CircuitBreakerConfig

	switch serviceName {
	case "emshop-coupon-srv":
		if br.Coupon != nil {
			cbConfig = &br.Coupon.CircuitBreaker
		}
	case "emshop-user-srv":
		if br.User != nil {
			cbConfig = &br.User.CircuitBreaker
		}
	case "emshop-inventory-srv":
		if br.Inventory != nil {
			cbConfig = &br.Inventory.CircuitBreaker
		}
	case "emshop-payment-srv":
		if br.Payment != nil {
			cbConfig = &br.Payment.CircuitBreaker
		}
	}

	if cbConfig != nil && cbConfig.Enabled {
		// 定义具体的资源名列表，而不是使用通配符
		var resources []string
		switch serviceName {
		case "emshop-coupon-srv":
			resources = []string{"coupon-srv:IssueCoupon", "coupon-srv:UseCoupon", "coupon-srv:GetUserCoupons"}
		case "emshop-user-srv":
			resources = []string{"user-srv:CreateUser", "user-srv:GetUserByMobile", "user-srv:GetUserById"}
		case "emshop-inventory-srv":
			resources = []string{"inventory-srv:Sell", "inventory-srv:InvDetail", "inventory-srv:Reback"}
		case "emshop-payment-srv":
			resources = []string{"payment-srv:CreatePayment", "payment-srv:QueryPayment", "payment-srv:ProcessRefund"}
		}

		// 为每个资源创建熔断规则
		for _, resource := range resources {
			// 错误率熔断规则
			rules = append(rules, &circuitbreaker.Rule{
				Resource:                     resource,
				Strategy:                     circuitbreaker.ErrorRatio,
				RetryTimeoutMs:              uint32(cbConfig.RecoveryTimeoutSec * 1000),
				MinRequestAmount:            cbConfig.MinRequestAmount,
				StatIntervalMs:              1000,
				StatSlidingWindowBucketCount: 10,
				MaxAllowedRtMs:              uint64(cbConfig.SlowTimeMs),
				Threshold:                   cbConfig.ErrorRatio,
			})

			// 慢调用比例熔断规则
			rules = append(rules, &circuitbreaker.Rule{
				Resource:                     resource,
				Strategy:                     circuitbreaker.SlowRequestRatio,
				RetryTimeoutMs:              uint32(cbConfig.RecoveryTimeoutSec * 1000),
				MinRequestAmount:            cbConfig.MinRequestAmount,
				StatIntervalMs:              1000,
				StatSlidingWindowBucketCount: 10,
				MaxAllowedRtMs:              uint64(cbConfig.SlowTimeMs),
				Threshold:                   cbConfig.SlowRatio,
			})
		}
	}

	return rules, nil
}

// GenerateHotspotRules 生成热点参数规则
func (br *BusinessRules) GenerateHotspotRules(serviceName string) ([]*hotspot.Rule, error) {
	rules := make([]*hotspot.Rule, 0)

	switch serviceName {
	case "emshop-coupon-srv":
		if br.Coupon != nil && br.Coupon.UserHotspot.Enabled {
			rule := &hotspot.Rule{
				Resource:        "coupon-srv:GetUserCoupons",
				MetricType:      hotspot.QPS,
				ParamIndex:      br.Coupon.UserHotspot.ParamIndex,
				Threshold:       int64(br.Coupon.UserHotspot.Count),
				DurationInSec:   br.Coupon.UserHotspot.DurationInSec,
			}
			
			// 添加特定参数值的阈值
			if len(br.Coupon.UserHotspot.SpecificItems) > 0 {
				rule.SpecificItems = make(map[interface{}]int64)
				for _, item := range br.Coupon.UserHotspot.SpecificItems {
					rule.SpecificItems[item.Value] = item.Threshold
				}
			}
			
			rules = append(rules, rule)
		}
	case "emshop-user-srv":
		if br.User != nil && br.User.MobileHotspot.Enabled {
			rules = append(rules, &hotspot.Rule{
				Resource:        "user-srv:GetUserByMobile",
				MetricType:      hotspot.QPS,
				ParamIndex:      br.User.MobileHotspot.ParamIndex,
				Threshold:       int64(br.User.MobileHotspot.Count),
				DurationInSec:   br.User.MobileHotspot.DurationInSec,
			})
		}
	}

	return rules, nil
}

// GenerateSystemRules 生成系统规则
func (br *BusinessRules) GenerateSystemRules(serviceName string) ([]*system.Rule, error) {
	rules := []*system.Rule{
		// CPU使用率限制
		{
			MetricType:   system.CpuUsage,
			TriggerCount: 0.8, // CPU使用率超过80%时触发
			Strategy:     system.BBR,
		},
		// 系统负载限制
		{
			MetricType:   system.Load,
			TriggerCount: 10.0, // 系统负载超过10时触发
			Strategy:     system.BBR,
		},
	}

	return rules, nil
}

// ToJSON 将业务规则转换为JSON
func (br *BusinessRules) ToJSON() ([]byte, error) {
	return json.MarshalIndent(br, "", "  ")
}

// FromJSON 从JSON加载业务规则
func FromJSON(data []byte) (*BusinessRules, error) {
	var rules BusinessRules
	err := json.Unmarshal(data, &rules)
	return &rules, err
}