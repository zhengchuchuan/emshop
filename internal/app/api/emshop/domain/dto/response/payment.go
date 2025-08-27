package response

import (
	"time"
	ppbv1 "emshop/api/payment/v1"
)

// CreatePaymentResponse 创建支付订单响应
type CreatePaymentResponse struct {
	PaymentSN  string     `json:"payment_sn"`  // 支付单号
	PaymentURL *string    `json:"payment_url"` // 模拟支付链接（可选）
	ExpiredAt  time.Time  `json:"expired_at"`  // 支付过期时间
}

// FromProto 从protobuf转换
func (r *CreatePaymentResponse) FromProto(pb *ppbv1.CreatePaymentResponse) {
	r.PaymentSN = pb.PaymentSn
	r.PaymentURL = pb.PaymentUrl
	r.ExpiredAt = time.Unix(pb.ExpiredAt, 0)
}

// PaymentStatusResponse 支付状态响应
type PaymentStatusResponse struct {
	PaymentSN     string     `json:"payment_sn"`     // 支付单号
	OrderSN       string     `json:"order_sn"`       // 订单号
	PaymentStatus int32      `json:"payment_status"` // 支付状态
	Amount        float64    `json:"amount"`         // 支付金额
	PaymentMethod int32      `json:"payment_method"` // 支付方式
	PaidAt        *time.Time `json:"paid_at"`        // 支付时间（可选）
	ExpiredAt     time.Time  `json:"expired_at"`     // 过期时间
	StatusText    string     `json:"status_text"`    // 状态描述
}

// FromProto 从protobuf转换
func (r *PaymentStatusResponse) FromProto(pb *ppbv1.PaymentStatusResponse) {
	r.PaymentSN = pb.PaymentSn
	r.OrderSN = pb.OrderSn
	r.PaymentStatus = pb.PaymentStatus
	r.Amount = pb.Amount
	r.PaymentMethod = pb.PaymentMethod
	r.ExpiredAt = time.Unix(pb.ExpiredAt, 0)
	
	if pb.PaidAt != nil {
		paidAt := time.Unix(*pb.PaidAt, 0)
		r.PaidAt = &paidAt
	}
	
	// 设置状态描述
	r.StatusText = getPaymentStatusText(pb.PaymentStatus)
}

// getPaymentStatusText 获取支付状态描述
func getPaymentStatusText(status int32) string {
	switch status {
	case 1:
		return "待支付"
	case 2:
		return "支付成功"
	case 3:
		return "支付失败"
	case 4:
		return "支付取消"
	case 5:
		return "支付过期"
	default:
		return "未知状态"
	}
}

// SimulatePaymentResponse 模拟支付响应（通常为空，只关心成功失败）
type SimulatePaymentResponse struct {
	Success bool   `json:"success"` // 是否成功
	Message string `json:"message"` // 响应消息
}