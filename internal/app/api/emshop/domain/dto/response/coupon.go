package response

import (
	cpbv1 "emshop/api/coupon/v1"
	"time"
)

// CouponTemplateResponse 优惠券模板响应
type CouponTemplateResponse struct {
	ID             int64   `json:"id"`               // 模板ID
	Name           string  `json:"name"`             // 优惠券名称
	DiscountType   int32   `json:"discount_type"`    // 折扣类型
	DiscountValue  float64 `json:"discount_value"`   // 折扣值
	MinOrderAmount float64 `json:"min_order_amount"` // 最小订单金额
	ValidDays      int32   `json:"valid_days"`       // 有效天数
	TotalCount     int32   `json:"total_count"`      // 总发放数量
	ReceivedCount  int32   `json:"received_count"`   // 已领取数量
	PerUserLimit   int32   `json:"per_user_limit"`   // 每用户限领数量
	Description    string  `json:"description"`      // 使用说明
	Status         int32   `json:"status"`           // 状态
}

// FromProto 从protobuf转换
func (r *CouponTemplateResponse) FromProto(pb *cpbv1.CouponTemplateResponse) {
	r.ID = pb.Id
	r.Name = pb.Name
	r.DiscountType = pb.DiscountType
	r.DiscountValue = pb.DiscountValue
	r.MinOrderAmount = pb.MinOrderAmount
	r.ValidDays = pb.ValidDays
	r.TotalCount = pb.TotalCount
	r.ReceivedCount = pb.UsedCount // 这里使用UsedCount作为已领取数量
	r.PerUserLimit = pb.PerUserLimit
	r.Description = pb.Description
	r.Status = pb.Status
}

// CouponTemplateListResponse 优惠券模板列表响应
type CouponTemplateListResponse struct {
	Total int64                     `json:"total"` // 总数量
	Items []*CouponTemplateResponse `json:"items"` // 模板列表
}

// FromProto 从protobuf转换
func (r *CouponTemplateListResponse) FromProto(pb *cpbv1.ListCouponTemplatesResponse) {
	r.Total = pb.TotalCount
	r.Items = make([]*CouponTemplateResponse, len(pb.Items))
	for i, item := range pb.Items {
		r.Items[i] = &CouponTemplateResponse{}
		r.Items[i].FromProto(item)
	}
}

// UserCouponResponse 用户优惠券响应
type UserCouponResponse struct {
	ID         int64                   `json:"id"`          // 用户优惠券ID
	CouponCode string                  `json:"coupon_code"` // 优惠券码
	Status     int32                   `json:"status"`      // 状态
	ReceivedAt time.Time               `json:"received_at"` // 领取时间
	ExpiredAt  time.Time               `json:"expired_at"`  // 过期时间
	UsedAt     *time.Time              `json:"used_at"`     // 使用时间，可为空
	OrderSN    *string                 `json:"order_sn"`    // 关联订单号，可为空
	Template   *CouponTemplateResponse `json:"template"`    // 优惠券模板信息
}

// FromProto 从protobuf转换
func (r *UserCouponResponse) FromProto(pb *cpbv1.UserCouponResponse) {
	r.ID = pb.Id
	r.CouponCode = pb.CouponCode
	r.Status = pb.Status
	r.ReceivedAt = time.Unix(pb.ReceivedAt, 0)
	r.ExpiredAt = time.Unix(pb.ExpiredAt, 0)

	if pb.UsedAt != nil {
		usedAt := time.Unix(*pb.UsedAt, 0)
		r.UsedAt = &usedAt
	}

	if pb.OrderSn != nil {
		r.OrderSN = pb.OrderSn
	}

	if pb.Template != nil {
		r.Template = &CouponTemplateResponse{}
		r.Template.FromProto(pb.Template)
	}
}

// UserCouponListResponse 用户优惠券列表响应
type UserCouponListResponse struct {
	Total int64                 `json:"total"` // 总数量
	Items []*UserCouponResponse `json:"items"` // 优惠券列表
}

// FromProto 从protobuf转换
func (r *UserCouponListResponse) FromProto(pb *cpbv1.ListUserCouponsResponse) {
	r.Total = pb.TotalCount
	r.Items = make([]*UserCouponResponse, len(pb.Items))
	for i, item := range pb.Items {
		r.Items[i] = &UserCouponResponse{}
		r.Items[i].FromProto(item)
	}
}

// ReceiveCouponResponse 领取优惠券响应
type ReceiveCouponResponse struct {
	ID            int64     `json:"id"`             // 用户优惠券ID
	CouponCode    string    `json:"coupon_code"`    // 优惠券码
	TemplateName  string    `json:"template_name"`  // 优惠券名称
	DiscountValue float64   `json:"discount_value"` // 折扣值
	ExpiredAt     time.Time `json:"expired_at"`     // 过期时间
	Status        int32     `json:"status"`         // 状态
}

// FromProto 从protobuf转换
func (r *ReceiveCouponResponse) FromProto(pb *cpbv1.UserCouponResponse) {
	r.ID = pb.Id
	r.CouponCode = pb.CouponCode
	r.Status = pb.Status
	r.ExpiredAt = time.Unix(pb.ExpiredAt, 0)

	if pb.Template != nil {
		r.TemplateName = pb.Template.Name
		r.DiscountValue = pb.Template.DiscountValue
	}
}

// AvailableCouponResponse 可用优惠券响应
type AvailableCouponResponse struct {
	ID             int64                   `json:"id"`              // 用户优惠券ID
	CouponCode     string                  `json:"coupon_code"`     // 优惠券码
	Template       *CouponTemplateResponse `json:"template"`        // 优惠券模板信息
	CanUse         bool                    `json:"can_use"`         // 是否可以使用
	DiscountAmount float64                 `json:"discount_amount"` // 预计优惠金额
}

// FromProto 从protobuf转换
func (r *AvailableCouponResponse) FromProto(pb *cpbv1.UserCouponResponse) {
	r.ID = pb.Id
	r.CouponCode = pb.CouponCode
	r.CanUse = true // API层查询的都是可用的

	if pb.Template != nil {
		r.Template = &CouponTemplateResponse{}
		r.Template.FromProto(pb.Template)
		// 简单计算预计优惠金额（实际应该通过计算接口获取）
		r.DiscountAmount = pb.Template.DiscountValue
	}
}

// AvailableCouponsResponse 可用优惠券列表响应
type AvailableCouponsResponse struct {
	AvailableCoupons []*AvailableCouponResponse `json:"available_coupons"` // 可用优惠券列表
}

// FromProto 从protobuf转换
func (r *AvailableCouponsResponse) FromProto(pb *cpbv1.ListUserCouponsResponse) {
	r.AvailableCoupons = make([]*AvailableCouponResponse, len(pb.Items))
	for i, item := range pb.Items {
		r.AvailableCoupons[i] = &AvailableCouponResponse{}
		r.AvailableCoupons[i].FromProto(item)
	}
}

// CouponDiscountResponse 优惠券折扣计算响应
type CouponDiscountResponse struct {
	OriginalAmount  float64            `json:"original_amount"`  // 原始金额
	DiscountAmount  float64            `json:"discount_amount"`  // 优惠金额
	FinalAmount     float64            `json:"final_amount"`     // 最终金额
	AppliedCoupons  []int64            `json:"applied_coupons"`  // 已应用的优惠券ID
	RejectedCoupons []*CouponRejection `json:"rejected_coupons"` // 被拒绝的优惠券
}

// CouponRejection 优惠券拒绝信息
type CouponRejection struct {
	CouponID int64  `json:"coupon_id"` // 优惠券ID
	Reason   string `json:"reason"`    // 拒绝原因
}

// FromProto 从protobuf转换
func (r *CouponDiscountResponse) FromProto(pb *cpbv1.CalculateCouponDiscountResponse) {
	r.OriginalAmount = pb.OriginalAmount
	r.DiscountAmount = pb.DiscountAmount
	r.FinalAmount = pb.FinalAmount
	r.AppliedCoupons = pb.AppliedCoupons

	r.RejectedCoupons = make([]*CouponRejection, len(pb.RejectedCoupons))
	for i, rejection := range pb.RejectedCoupons {
		r.RejectedCoupons[i] = &CouponRejection{
			CouponID: rejection.CouponId,
			Reason:   rejection.Reason,
		}
	}
}
