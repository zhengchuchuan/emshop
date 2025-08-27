package code

const (
	// ErrResourceNotFound - 404: Resource not found.
	ErrResourceNotFound int = iota + 101001

	// ErrInvalidRequest - 400: Invalid request parameters.
	ErrInvalidRequest

	// ErrResourceNotAvailable - 503: Resource not available.
	ErrResourceNotAvailable

	// ErrResourceLimitExceeded - 429: Resource limit exceeded.
	ErrResourceLimitExceeded

	// ErrDatabase - 500: Database operation error.
	ErrDatabase

	// ErrRedis - 500: Redis operation error.
	ErrRedis

	// ErrCouponTemplateNotFound - 404: Coupon template not found.
	ErrCouponTemplateNotFound

	// ErrCouponTemplateInactive - 400: Coupon template is inactive.
	ErrCouponTemplateInactive

	// ErrCouponNotFound - 404: User coupon not found.
	ErrCouponNotFound

	// ErrCouponExpired - 400: Coupon has expired.
	ErrCouponExpired

	// ErrCouponUsed - 400: Coupon has been used.
	ErrCouponUsed

	// ErrCouponNotAvailable - 400: Coupon not available for use.
	ErrCouponNotAvailable

	// ErrCouponLimitExceeded - 400: Coupon usage limit exceeded.
	ErrCouponLimitExceeded

	// ErrFlashSaleNotFound - 404: Flash sale activity not found.
	ErrFlashSaleNotFound

	// ErrFlashSaleNotActive - 400: Flash sale activity is not active.
	ErrFlashSaleNotActive

	// ErrFlashSaleStockEmpty - 400: Flash sale stock is empty.
	ErrFlashSaleStockEmpty

	// ErrFlashSaleUserLimitExceeded - 400: User flash sale participation limit exceeded.
	ErrFlashSaleUserLimitExceeded

	// ErrFlashSaleAlreadyParticipated - 400: User has already participated in this flash sale.
	ErrFlashSaleAlreadyParticipated
)