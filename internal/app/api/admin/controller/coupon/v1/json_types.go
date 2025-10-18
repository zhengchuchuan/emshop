package coupon

import (
	"emshop/internal/app/pkg/common/datetime"
)

// CreateCouponTemplateJSON 允许时间字段为秒级时间戳或日期字符串
type CreateCouponTemplateJSON struct {
	Name              string              `json:"name" binding:"required"`
	Type              int32               `json:"type" binding:"required"`
	DiscountType      int32               `json:"discount_type" binding:"required"`
	DiscountValue     float64             `json:"discount_value"`
	MinOrderAmount    float64             `json:"min_order_amount"`
	MaxDiscountAmount float64             `json:"max_discount_amount"`
	TotalCount        int32               `json:"total_count"`
	PerUserLimit      int32               `json:"per_user_limit"`
	ValidStartTime    datetime.UnixOrTime `json:"valid_start_time"`
	ValidEndTime      datetime.UnixOrTime `json:"valid_end_time"`
	ValidDays         int32               `json:"valid_days"`
	Description       string              `json:"description"`
}

// CreateFlashSaleJSON 允许时间字段为秒级时间戳或日期字符串
type CreateFlashSaleJSON struct {
	CouponTemplateID int64               `json:"coupon_template_id" binding:"required"`
	Name             string              `json:"name" binding:"required"`
	StartTime        datetime.UnixOrTime `json:"start_time" binding:"required"`
	EndTime          datetime.UnixOrTime `json:"end_time" binding:"required"`
	FlashSaleCount   int32               `json:"flash_sale_count" binding:"required"`
	PerUserLimit     int32               `json:"per_user_limit" binding:"required"`
}
