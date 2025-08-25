package dto

import (
	"emshop/internal/app/payment/srv/domain/do"
	"time"
)

// CreatePaymentDTO 创建支付订单DTO
type CreatePaymentDTO struct {
	OrderSn        string             `json:"order_sn" binding:"required"`
	UserID         int32              `json:"user_id" binding:"required"`
	Amount         float64            `json:"amount" binding:"required"`
	PaymentMethod  do.PaymentMethod   `json:"payment_method" binding:"required"`
	ExpiredMinutes int32              `json:"expired_minutes"` // 可选，默认15分钟
}

// PaymentDTO 支付订单DTO
type PaymentDTO struct {
	do.PaymentOrderDO
}

// PaymentLogDTO 支付日志DTO
type PaymentLogDTO struct {
	do.PaymentLogDO
}

// PaymentStatusDTO 支付状态查询DTO
type PaymentStatusDTO struct {
	PaymentSn     string           `json:"payment_sn"`
	OrderSn       string           `json:"order_sn"`
	PaymentStatus do.PaymentStatus `json:"payment_status"`
	Amount        float64          `json:"amount"`
	PaymentMethod do.PaymentMethod `json:"payment_method"`
	PaidAt        *time.Time       `json:"paid_at"`
	ExpiredAt     time.Time        `json:"expired_at"`
}

// ConfirmPaymentDTO 确认支付DTO
type ConfirmPaymentDTO struct {
	PaymentSn    string  `json:"payment_sn" binding:"required"`
	ThirdPartySn *string `json:"third_party_sn"`
}

// RefundPaymentDTO 退款DTO
type RefundPaymentDTO struct {
	PaymentSn    string  `json:"payment_sn" binding:"required"`
	RefundAmount float64 `json:"refund_amount" binding:"required"`
	Reason       *string `json:"reason"`
}

// StockReservationDTO 库存预留DTO
type StockReservationDTO struct {
	do.StockReservationDO
}

// ReserveStockDTO 预留库存请求DTO
type ReserveStockDTO struct {
	OrderSn   string            `json:"order_sn" binding:"required"`
	GoodsInfo []do.GoodsDetail `json:"goods_info" binding:"required"`
}

// ReleaseReservedDTO 释放预留库存请求DTO
type ReleaseReservedDTO struct {
	OrderSn   string            `json:"order_sn" binding:"required"`
	GoodsInfo []do.GoodsDetail `json:"goods_info" binding:"required"`
}