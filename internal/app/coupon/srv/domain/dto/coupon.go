package dto

import "time"

// CreateCouponTemplateDTO 创建优惠券模板DTO
type CreateCouponTemplateDTO struct {
	Name              string    `json:"name" validate:"required,max=100"`
	Type              int32     `json:"type" validate:"required,min=1,max=4"`
	DiscountType      int32     `json:"discount_type" validate:"required,min=1,max=2"`
	DiscountValue     float64   `json:"discount_value" validate:"required,min=0"`
	MinOrderAmount    float64   `json:"min_order_amount" validate:"min=0"`
	MaxDiscountAmount float64   `json:"max_discount_amount" validate:"min=0"`
	TotalCount        int32     `json:"total_count" validate:"required,min=1"`
	PerUserLimit      int32     `json:"per_user_limit" validate:"required,min=1"`
	ValidStartTime    time.Time `json:"valid_start_time" validate:"required"`
	ValidEndTime      time.Time `json:"valid_end_time" validate:"required"`
	ValidDays         int32     `json:"valid_days" validate:"min=0"`
	Description       string    `json:"description" validate:"max=500"`
}

// UpdateCouponTemplateDTO 更新优惠券模板DTO
type UpdateCouponTemplateDTO struct {
	ID          int64   `json:"id" validate:"required"`
	Name        *string `json:"name,omitempty" validate:"omitempty,max=100"`
	Status      *int32  `json:"status,omitempty" validate:"omitempty,min=1,max=3"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
}

// CouponTemplateDTO 优惠券模板DTO
type CouponTemplateDTO struct {
	ID                int64     `json:"id"`
	Name              string    `json:"name"`
	Type              int32     `json:"type"`
	DiscountType      int32     `json:"discount_type"`
	DiscountValue     float64   `json:"discount_value"`
	MinOrderAmount    float64   `json:"min_order_amount"`
	MaxDiscountAmount float64   `json:"max_discount_amount"`
	TotalCount        int32     `json:"total_count"`
	UsedCount         int32     `json:"used_count"`
	PerUserLimit      int32     `json:"per_user_limit"`
	ValidStartTime    time.Time `json:"valid_start_time"`
	ValidEndTime      time.Time `json:"valid_end_time"`
	ValidDays         int32     `json:"valid_days"`
	Status            int32     `json:"status"`
	Description       string    `json:"description"`
	CreatedAt         time.Time `json:"created_at"`
}

// ListCouponTemplatesDTO 优惠券模板列表DTO
type ListCouponTemplatesDTO struct {
	Status   *int32 `json:"status,omitempty" validate:"omitempty,min=1,max=3"`
	Page     int32  `json:"page" validate:"required,min=1"`
	PageSize int32  `json:"page_size" validate:"required,min=1,max=100"`
}

// CouponTemplateListDTO 优惠券模板列表响应DTO
type CouponTemplateListDTO struct {
	TotalCount int64                `json:"total_count"`
	Items      []*CouponTemplateDTO `json:"items"`
}

// ReceiveCouponDTO 领取优惠券DTO
type ReceiveCouponDTO struct {
	UserID           int64 `json:"user_id" validate:"required"`
	CouponTemplateID int64 `json:"coupon_template_id" validate:"required"`
}

// UserCouponDTO 用户优惠券DTO
type UserCouponDTO struct {
	ID               int64              `json:"id"`
	CouponTemplateID int64              `json:"coupon_template_id"`
	UserID           int64              `json:"user_id"`
	CouponCode       string             `json:"coupon_code"`
	Status           int32              `json:"status"`
	OrderSn          *string            `json:"order_sn,omitempty"`
	ReceivedAt       time.Time          `json:"received_at"`
	UsedAt           *time.Time         `json:"used_at,omitempty"`
	ExpiredAt        time.Time          `json:"expired_at"`
	Template         *CouponTemplateDTO `json:"template,omitempty"`
}

// GetUserCouponsDTO 获取用户优惠券列表DTO
type GetUserCouponsDTO struct {
	UserID   int64  `json:"user_id" validate:"required"`
	Status   *int32 `json:"status,omitempty" validate:"omitempty,min=1,max=4"`
	Page     int32  `json:"page" validate:"required,min=1"`
	PageSize int32  `json:"page_size" validate:"required,min=1,max=100"`
}

// UserCouponListDTO 用户优惠券列表响应DTO
type UserCouponListDTO struct {
	TotalCount int64            `json:"total_count"`
	Items      []*UserCouponDTO `json:"items"`
}

// GetAvailableCouponsDTO 获取可用优惠券DTO
type GetAvailableCouponsDTO struct {
	UserID      int64   `json:"user_id" validate:"required"`
	OrderAmount float64 `json:"order_amount" validate:"required,min=0"`
}

// OrderItemDTO 订单商品项DTO
type OrderItemDTO struct {
	GoodsID  int64   `json:"goods_id" validate:"required"`
	Quantity int32   `json:"quantity" validate:"required,min=1"`
	Price    float64 `json:"price" validate:"required,min=0"`
}

// CalculateCouponDiscountDTO 计算优惠券折扣DTO
type CalculateCouponDiscountDTO struct {
	UserID      int64           `json:"user_id" validate:"required"`
	CouponIDs   []int64         `json:"coupon_ids" validate:"required"`
	OrderAmount float64         `json:"order_amount" validate:"required,min=0"`
	OrderItems  []*OrderItemDTO `json:"order_items" validate:"required"`
}

// CouponDiscountResultDTO 优惠券折扣计算结果DTO
type CouponDiscountResultDTO struct {
	OriginalAmount   float64             `json:"original_amount"`
	DiscountAmount   float64             `json:"discount_amount"`
	FinalAmount      float64             `json:"final_amount"`
	AppliedCoupons   []int64             `json:"applied_coupons"`
	RejectedCoupons  []*CouponRejection  `json:"rejected_coupons"`
}

// CouponRejection 优惠券拒绝信息
type CouponRejection struct {
	CouponID int64  `json:"coupon_id"`
	Reason   string `json:"reason"`
}

// UseCouponsDTO 使用优惠券DTO
type UseCouponsDTO struct {
	UserID      int64   `json:"user_id" validate:"required"`
	OrderSn     string  `json:"order_sn" validate:"required"`
	CouponIDs   []int64 `json:"coupon_ids" validate:"required"`
	OrderAmount float64 `json:"order_amount" validate:"required,min=0"`
}

// UseCouponsResultDTO 使用优惠券结果DTO
type UseCouponsResultDTO struct {
	DiscountAmount float64 `json:"discount_amount"`
	UsedCoupons    []int64 `json:"used_coupons"`
}

// ReleaseCouponsDTO 释放优惠券DTO
type ReleaseCouponsDTO struct {
	OrderSn string `json:"order_sn" validate:"required"`
}