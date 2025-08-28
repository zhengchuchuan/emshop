package request

import (
	cpbv1 "emshop/api/coupon/v1"
	"fmt"
)

// ReceiveCouponRequest 领取优惠券请求
type ReceiveCouponRequest struct {
	TemplateID int64 `json:"template_id" binding:"required,min=1" form:"template_id"`
}

// ToProto 转换为protobuf请求
func (r *ReceiveCouponRequest) ToProto(userID int64) *cpbv1.ReceiveCouponRequest {
	return &cpbv1.ReceiveCouponRequest{
		UserId:           userID,
		CouponTemplateId: r.TemplateID,
	}
}

// Validate 验证请求参数
func (r *ReceiveCouponRequest) Validate() error {
	if r.TemplateID <= 0 {
		return &ValidationError{Field: "template_id", Message: "优惠券模板ID必须大于0"}
	}
	return nil
}

// GetUserCouponsRequest 获取用户优惠券列表请求
type GetUserCouponsRequest struct {
	Status   *int32 `form:"status" json:"status"`                                     // 状态筛选，可选
	Page     int32  `form:"page" json:"page" binding:"required,min=1"`                // 页码
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

// Validate 验证请求参数
func (r *GetUserCouponsRequest) Validate() error {
	if r.Page <= 0 {
		return &ValidationError{Field: "page", Message: "页码必须大于0"}
	}
	if r.PageSize <= 0 || r.PageSize > 50 {
		return &ValidationError{Field: "pageSize", Message: "页大小必须在1-50之间"}
	}
	return nil
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

// Validate 验证请求参数
func (r *GetAvailableCouponsRequest) Validate() error {
	if r.OrderAmount <= 0 {
		return &ValidationError{Field: "order_amount", Message: "订单金额必须大于0"}
	}
	return nil
}

// CalculateCouponDiscountRequest 计算优惠券折扣请求
type CalculateCouponDiscountRequest struct {
	CouponIDs   []int64     `json:"coupon_ids" binding:"required,min=1"`      // 优惠券ID列表
	OrderAmount float64     `json:"order_amount" binding:"required,min=0.01"` // 订单金额
	OrderItems  []OrderItem `json:"order_items" binding:"required,dive"`      // 订单商品明细
}

// OrderItem 订单商品项
type OrderItem struct {
	GoodsID  int64   `json:"goods_id" binding:"required,min=1"` // 商品ID
	Quantity int32   `json:"quantity" binding:"required,min=1"` // 数量
	Price    float64 `json:"price" binding:"required,min=0.01"` // 单价
}

// ToProto 转换为protobuf请求
func (r *CalculateCouponDiscountRequest) ToProto(userID int64) *cpbv1.CalculateCouponDiscountRequest {
	orderItems := make([]*cpbv1.CouponOrderItem, len(r.OrderItems))
	for i, item := range r.OrderItems {
		orderItems[i] = &cpbv1.CouponOrderItem{
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

// Validate 验证请求参数
func (r *CalculateCouponDiscountRequest) Validate() error {
	if len(r.CouponIDs) == 0 {
		return &ValidationError{Field: "coupon_ids", Message: "优惠券ID列表不能为空"}
	}
	if r.OrderAmount <= 0 {
		return &ValidationError{Field: "order_amount", Message: "订单金额必须大于0"}
	}
	if len(r.OrderItems) == 0 {
		return &ValidationError{Field: "order_items", Message: "订单商品明细不能为空"}
	}
	for i, item := range r.OrderItems {
		if item.GoodsID <= 0 {
			return &ValidationError{Field: "order_items", Message: fmt.Sprintf("第%d个商品ID无效", i+1)}
		}
		if item.Quantity <= 0 {
			return &ValidationError{Field: "order_items", Message: fmt.Sprintf("第%d个商品数量必须大于0", i+1)}
		}
		if item.Price <= 0 {
			return &ValidationError{Field: "order_items", Message: fmt.Sprintf("第%d个商品价格必须大于0", i+1)}
		}
	}
	return nil
}

// ListCouponTemplatesRequest 获取优惠券模板列表请求
type ListCouponTemplatesRequest struct {
	Page     int32 `form:"page" json:"page" binding:"required,min=1"`                // 页码
	PageSize int32 `form:"pageSize" json:"pageSize" binding:"required,min=1,max=50"` // 页大小
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

// Validate 验证请求参数
func (r *ListCouponTemplatesRequest) Validate() error {
	if r.Page <= 0 {
		return &ValidationError{Field: "page", Message: "页码必须大于0"}
	}
	if r.PageSize <= 0 || r.PageSize > 50 {
		return &ValidationError{Field: "pageSize", Message: "页大小必须在1-50之间"}
	}
	return nil
}
