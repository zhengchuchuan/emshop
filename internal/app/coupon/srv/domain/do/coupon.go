package do

import (
	"emshop/pkg/db"
	"time"
)

// CouponType 优惠券类型定义
type CouponType int32

const (
	CouponTypeThreshold CouponType = 1 // 满减券
	CouponTypeDiscount  CouponType = 2 // 折扣券
	CouponTypeInstant   CouponType = 3 // 立减券
	CouponTypeFreeShip  CouponType = 4 // 包邮券
)

// DiscountType 折扣类型定义
type DiscountType int32

const (
	DiscountTypeFixed   DiscountType = 1 // 固定金额
	DiscountTypePercent DiscountType = 2 // 百分比
)

// CouponStatus 优惠券状态定义
type CouponStatus int32

const (
	CouponStatusActive    CouponStatus = 1 // 活跃
	CouponStatusPaused    CouponStatus = 2 // 暂停
	CouponStatusFinished  CouponStatus = 3 // 结束
)

// UserCouponStatus 用户优惠券状态定义
type UserCouponStatus int32

const (
	UserCouponStatusUnused  UserCouponStatus = 1 // 未使用
	UserCouponStatusUsed    UserCouponStatus = 2 // 已使用
	UserCouponStatusExpired UserCouponStatus = 3 // 已过期
	UserCouponStatusFrozen  UserCouponStatus = 4 // 已冻结
)

// CouponTemplateDO 优惠券模板数据对象
type CouponTemplateDO struct {
	db.BaseModel
	ID        int64     `gorm:"primarykey" json:"id"`
	Name               string        `json:"name" gorm:"column:name;type:varchar(100);not null;comment:优惠券名称"`
	Type               CouponType    `json:"type" gorm:"column:type;type:tinyint;not null;index:idx_type;comment:优惠券类型"`
	DiscountType       DiscountType  `json:"discount_type" gorm:"column:discount_type;type:tinyint;not null;comment:折扣类型"`
	DiscountValue      float64       `json:"discount_value" gorm:"column:discount_value;type:decimal(10,2);not null;comment:折扣值"`
	MinOrderAmount     float64       `json:"min_order_amount" gorm:"column:min_order_amount;type:decimal(10,2);default:0.00;comment:最小订单金额"`
	MaxDiscountAmount  float64       `json:"max_discount_amount" gorm:"column:max_discount_amount;type:decimal(10,2);default:0.00;comment:最大折扣金额"`
	TotalCount         int32         `json:"total_count" gorm:"column:total_count;type:int;not null;default:0;comment:总发放数量"`
	UsedCount          int32         `json:"used_count" gorm:"column:used_count;type:int;not null;default:0;comment:已使用数量"`
	PerUserLimit       int32         `json:"per_user_limit" gorm:"column:per_user_limit;type:int;not null;default:1;comment:每用户限领数量"`
	ValidStartTime     time.Time     `json:"valid_start_time" gorm:"column:valid_start_time;type:timestamp;not null;index:idx_valid_time;comment:有效期开始时间"`
	ValidEndTime       time.Time     `json:"valid_end_time" gorm:"column:valid_end_time;type:timestamp;not null;index:idx_valid_time;comment:有效期结束时间"`
	ValidDays          int32         `json:"valid_days" gorm:"column:valid_days;type:int;default:0;comment:有效天数"`
	Status             CouponStatus  `json:"status" gorm:"column:status;type:tinyint;not null;default:1;index:idx_status;comment:状态"`
	Description        string        `json:"description" gorm:"column:description;type:text;comment:使用说明"`
}

// TableName 指定表名
func (CouponTemplateDO) TableName() string {
	return "coupon_templates"
}

// UserCouponDO 用户优惠券数据对象
type UserCouponDO struct {
	db.BaseModel
	ID        int64     `gorm:"primarykey" json:"id"`
	CouponTemplateID int64            `json:"coupon_template_id" gorm:"column:coupon_template_id;type:bigint;not null;index:idx_coupon_template_id;comment:优惠券模板ID"`
	UserID           int64            `json:"user_id" gorm:"column:user_id;type:bigint;not null;index:idx_user_id,idx_user_status;comment:用户ID"`
	CouponCode       string           `json:"coupon_code" gorm:"column:coupon_code;type:varchar(32);not null;uniqueIndex:idx_coupon_code;comment:优惠券码"`
	Status           UserCouponStatus `json:"status" gorm:"column:status;type:tinyint;not null;default:1;index:idx_status,idx_user_status;comment:状态"`
	OrderSn          *string          `json:"order_sn" gorm:"column:order_sn;type:varchar(64);index:idx_order_sn;comment:使用的订单号"`
	ReceivedAt       time.Time        `json:"received_at" gorm:"column:received_at;type:timestamp;default:CURRENT_TIMESTAMP;comment:领取时间"`
	UsedAt           *time.Time       `json:"used_at" gorm:"column:used_at;type:timestamp;comment:使用时间"`
	ExpiredAt        time.Time        `json:"expired_at" gorm:"column:expired_at;type:timestamp;not null;index:idx_expired_at;comment:过期时间"`
}

// TableName 指定表名
func (UserCouponDO) TableName() string {
	return "user_coupons"
}

// CouponUsageLogDO 优惠券使用记录数据对象
type CouponUsageLogDO struct {
	db.BaseModel
	ID        int64     `gorm:"primarykey" json:"id"`
	UserCouponID   int64   `json:"user_coupon_id" gorm:"column:user_coupon_id;type:bigint;not null;index:idx_user_coupon_id;comment:用户优惠券ID"`
	UserID         int64   `json:"user_id" gorm:"column:user_id;type:bigint;not null;index:idx_user_id;comment:用户ID"`
	OrderSn        string  `json:"order_sn" gorm:"column:order_sn;type:varchar(64);not null;index:idx_order_sn;comment:订单号"`
	OriginalAmount float64 `json:"original_amount" gorm:"column:original_amount;type:decimal(10,2);not null;comment:原始订单金额"`
	DiscountAmount float64 `json:"discount_amount" gorm:"column:discount_amount;type:decimal(10,2);not null;comment:优惠金额"`
	FinalAmount    float64 `json:"final_amount" gorm:"column:final_amount;type:decimal(10,2);not null;comment:最终订单金额"`
	Action         string  `json:"action" gorm:"column:action;type:varchar(32);not null;index:idx_action;comment:操作类型"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;index:idx_created_at"`
}

// TableName 指定表名
func (CouponUsageLogDO) TableName() string {
	return "coupon_usage_logs"
}

// CouponConfigDO 优惠券配置数据对象
type CouponConfigDO struct {
	db.BaseModel
	ConfigKey   string `json:"config_key" gorm:"column:config_key;type:varchar(64);not null;uniqueIndex:idx_config_key;comment:配置键"`
	ConfigValue string `json:"config_value" gorm:"column:config_value;type:text;not null;comment:配置值"`
	Description string `json:"description" gorm:"column:description;type:varchar(255);comment:配置说明"`
}

// TableName 指定表名
func (CouponConfigDO) TableName() string {
	return "coupon_configs"
}

// CouponTemplateDOList 优惠券模板列表
type CouponTemplateDOList struct {
	TotalCount int64               `json:"totalCount"`
	Items      []*CouponTemplateDO `json:"items"`
}

// UserCouponDOList 用户优惠券列表
type UserCouponDOList struct {
	TotalCount int64           `json:"totalCount"`
	Items      []*UserCouponDO `json:"items"`
}

// CouponUsageLogDOList 优惠券使用记录列表
type CouponUsageLogDOList struct {
	TotalCount int64               `json:"totalCount"`
	Items      []*CouponUsageLogDO `json:"items"`
}

// CouponCalculationResult 优惠券计算结果
type CouponCalculationResult struct {
	CouponID         int64   `json:"coupon_id"`
	OriginalAmount   float64 `json:"original_amount"`
	DiscountAmount   float64 `json:"discount_amount"`
	FinalAmount      float64 `json:"final_amount"`
	AppliedCoupons   []int64 `json:"applied_coupons"`
	RejectedCoupons  []CouponRejectionReason `json:"rejected_coupons"`
}

// CouponRejectionReason 优惠券拒绝原因
type CouponRejectionReason struct {
	CouponID int64  `json:"coupon_id"`
	Reason   string `json:"reason"`
}

// CouponAvailabilityCheck 优惠券可用性检查结果
type CouponAvailabilityCheck struct {
	CouponID    int64  `json:"coupon_id"`
	IsAvailable bool   `json:"is_available"`
	Reason      string `json:"reason,omitempty"`
}