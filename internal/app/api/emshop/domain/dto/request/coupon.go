package request

import (
	cpbv1 "emshop/api/coupon/v1"
)

// ReceiveCouponRequest 领取优惠券请求
type ReceiveCouponRequest struct {
	CouponTemplateID int64 `json:"coupon_template_id" binding:"required,min=1" form:"coupon_template_id"`
}

// ToProto 转换为protobuf请求
func (r *ReceiveCouponRequest) ToProto(userID int64) *cpbv1.ReceiveCouponRequest {
	return &cpbv1.ReceiveCouponRequest{
		UserId:           userID,
		CouponTemplateId: r.CouponTemplateID,
	}
}

// GetUserCouponsRequest 获取用户优惠券列表请求
type GetUserCouponsRequest struct {
	Status   *int32 `form:"status" json:"status"`                                 // 状态筛选，可选
	Page     int32  `form:"page" json:"page" binding:"required,min=1"`            // 页码
	PageSize int32  `form:"pageSize" json:"pageSize" binding:"required,min=1,max=50"` // 页大小
}

// ToProto 转换为protobuf请求
func (r *GetUserCouponsRequest) ToProto(userID int64) *cpbv1.GetUserCouponsRequest {
	return &cpbv1.GetUserCouponsRequest{
		UserId:   userID,
		Status:   r.Status,
		Page:     r.Page,
		PageSize: r.PageSize,
	}
}

// GetAvailableCouponsRequest 获取可用优惠券请求
type GetAvailableCouponsRequest struct {
	OrderAmount float64 `form:"order_amount" json:"order_amount" binding:"required,min=0.01"` // 订单金额
}

// ToProto 转换为protobuf请求
func (r *GetAvailableCouponsRequest) ToProto(userID int64) *cpbv1.GetAvailableCouponsRequest {
	return &cpbv1.GetAvailableCouponsRequest{
		UserId:      userID,
		OrderAmount: r.OrderAmount,
	}
}

// CalculateCouponDiscountRequest 计算优惠券折扣请求
type CalculateCouponDiscountRequest struct {
	CouponIDs   []int64     `json:"coupon_ids" binding:"required,min=1"`          // 优惠券ID列表
	OrderAmount float64     `json:"order_amount" binding:"required,min=0.01"`     // 订单金额
	OrderItems  []OrderItem `json:"order_items" binding:"required,dive"`         // 订单商品明细
}

// OrderItem 订单商品项
type OrderItem struct {
	GoodsID  int64   `json:"goods_id" binding:"required,min=1"`  // 商品ID
	Quantity int32   `json:"quantity" binding:"required,min=1"`  // 数量
	Price    float64 `json:"price" binding:"required,min=0.01"`  // 单价
}

// ToProto 转换为protobuf请求
func (r *CalculateCouponDiscountRequest) ToProto(userID int64) *cpbv1.CalculateCouponDiscountRequest {
	orderItems := make([]*cpbv1.OrderItem, len(r.OrderItems))
	for i, item := range r.OrderItems {
		orderItems[i] = &cpbv1.OrderItem{
			GoodsId:  item.GoodsID,
			Quantity: item.Quantity,
			Price:    item.Price,
		}
	}

	return &cpbv1.CalculateCouponDiscountRequest{
		UserId:      userID,
		CouponIds:   r.CouponIDs,
		OrderAmount: r.OrderAmount,
		OrderItems:  orderItems,
	}
}

// ListCouponTemplatesRequest 获取优惠券模板列表请求
type ListCouponTemplatesRequest struct {
	Page     int32 `form:"page" json:"page" binding:"required,min=1"`                         // 页码
	PageSize int32 `form:"pageSize" json:"pageSize" binding:"required,min=1,max=50"`        // 页大小
}

// ToProto 转换为protobuf请求
func (r *ListCouponTemplatesRequest) ToProto() *cpbv1.ListCouponTemplatesRequest {
	// 只显示状态为1（有效）的优惠券模板
	status := int32(1)
	return &cpbv1.ListCouponTemplatesRequest{
		Status:   &status,
		Page:     r.Page,
		PageSize: r.PageSize,
	}
}