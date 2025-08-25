package code

//go:generate codegen -type=int
//go:generate codegen -type=int -doc -output ../../docs/guide/zh-CN/api/error_code_generated.md

// 支付服务相关错误码定义
// 错误码范围: 110001-119999

// 通用支付错误码 110001-110099
const (
	// ErrPaymentNotFound - 404: Payment order not found.
	ErrPaymentNotFound int = iota + 100901

	// ErrPaymentExists - 400: Payment order already exists.
	ErrPaymentExists

	// ErrPaymentStatusInvalid - 400: Payment status invalid.
	ErrPaymentStatusInvalid

	// ErrPaymentAmountInvalid - 400: Payment amount invalid.
	ErrPaymentAmountInvalid

	// ErrPaymentMethodInvalid - 400: Payment method invalid.
	ErrPaymentMethodInvalid

	// ErrPaymentExpired - 400: Payment expired.
	ErrPaymentExpired
)

// 支付订单相关错误码 110101-110199
const (
	// ErrCreatePaymentFailed - 500: Create payment order failed.
	ErrCreatePaymentFailed int = iota + 110101

	// ErrPaymentAlreadyPaid - 400: Payment order already paid.
	ErrPaymentAlreadyPaid

	// ErrPaymentAlreadyCancelled - 400: Payment order already cancelled.
	ErrPaymentAlreadyCancelled

	// ErrPaymentCannotCancel - 400: Payment order cannot be cancelled.
	ErrPaymentCannotCancel

	// ErrPaymentUpdateFailed - 500: Update payment order failed.
	ErrPaymentUpdateFailed
)

// 支付流程相关错误码 110201-110299
const (
	// ErrPaymentProcessFailed - 500: Payment process failed.
	ErrPaymentProcessFailed int = iota + 110201

	// ErrPaymentConfirmFailed - 500: Payment confirmation failed.
	ErrPaymentConfirmFailed

	// ErrPaymentRefundFailed - 500: Payment refund failed.
	ErrPaymentRefundFailed

	// ErrPaymentCallbackInvalid - 400: Payment callback invalid.
	ErrPaymentCallbackInvalid
)

// 库存预留相关错误码 110301-110399
const (
	// ErrStockReservationFailed - 500: Stock reservation failed.
	ErrStockReservationFailed int = iota + 110301

	// ErrStockReservationNotFound - 404: Stock reservation record not found.
	ErrStockReservationNotFound

	// ErrStockReleaseFailed - 500: Stock release failed.
	ErrStockReleaseFailed

	// ErrStockConfirmFailed - 500: Stock confirmation failed.
	ErrStockConfirmFailed

	// ErrStockInsufficientForReservation - 400: Insufficient stock for reservation.
	ErrStockInsufficientForReservation
)