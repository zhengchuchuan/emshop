package request

import (
	ppbv1 "emshop/api/payment/v1"
)

// CreatePaymentRequest 创建支付订单请求
type CreatePaymentRequest struct {
	OrderSN        string  `json:"order_sn" binding:"required"`                                  // 订单号
	Amount         float64 `json:"amount" binding:"required,min=0.01"`                           // 支付金额
	PaymentMethod  int32   `json:"payment_method" binding:"required,oneof=1 2 3"`                // 支付方式：1-支付宝，2-微信，3-银行卡
	ExpiredMinutes *int32  `json:"expired_minutes,omitempty" binding:"omitempty,min=1,max=1440"` // 支付过期时间(分钟)
}

// ToProto 转换为protobuf请求
func (r *CreatePaymentRequest) ToProto(userID int32) *ppbv1.CreatePaymentRequest {
	return &ppbv1.CreatePaymentRequest{
		OrderSn:        r.OrderSN,
		UserId:         userID,
		Amount:         r.Amount,
		PaymentMethod:  r.PaymentMethod,
		ExpiredMinutes: r.ExpiredMinutes,
	}
}

// SimulatePaymentRequest 模拟支付请求
type SimulatePaymentRequest struct {
	ThirdPartySN *string `json:"third_party_sn,omitempty"` // 第三方支付单号（模拟）
}

// ToProto 转换为protobuf请求
func (r *SimulatePaymentRequest) ToProto(paymentSN string) *ppbv1.SimulatePaymentRequest {
	return &ppbv1.SimulatePaymentRequest{
		PaymentSn:    paymentSN,
		ThirdPartySn: r.ThirdPartySN,
	}
}
