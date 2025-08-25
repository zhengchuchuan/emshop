package code



const (
	// ErrLogisticsOrderNotFound - 404: Logistics order not found.
	ErrLogisticsOrderNotFound int = iota + 1001001

	// ErrLogisticsOrderExists - 400: Logistics order already exists.
	ErrLogisticsOrderExists 

	// ErrCreateLogisticsOrderFailed - 500: Create logistics order failed.
	ErrCreateLogisticsOrderFailed 

	// ErrLogisticsOrderStatusInvalid - 400: Invalid logistics order status.
	ErrLogisticsOrderStatusInvalid 

	// ErrLogisticsOrderCannotCancel - 400: Logistics order cannot be cancelled.
	ErrLogisticsOrderCannotCancel 

	// ErrLogisticsOrderCannotUpdate - 400: Logistics order cannot be updated.
	ErrLogisticsOrderCannotUpdate

	// ErrLogisticsTrackNotFound - 404: Logistics track not found.
	ErrLogisticsTrackNotFound 

	// ErrCreateLogisticsTrackFailed - 500: Create logistics track failed.
	ErrCreateLogisticsTrackFailed

	// ErrLogisticsTrackInvalid - 400: Invalid logistics track data.
	ErrLogisticsTrackInvalid 

	// ErrLogisticsCourierNotFound - 404: Logistics courier not found.
	ErrLogisticsCourierNotFound 

	// ErrLogisticsCourierExists - 400: Logistics courier already exists.
	ErrLogisticsCourierExists 

	// ErrLogisticsCourierUnavailable - 500: Logistics courier unavailable.
	ErrLogisticsCourierUnavailable 

	// ErrShippingFeeCalculationFailed - 500: Shipping fee calculation failed.
	ErrShippingFeeCalculationFailed 

	// ErrInvalidShippingAddress - 400: Invalid shipping address.
	ErrInvalidShippingAddress 

	// ErrInvalidShippingMethod - 400: Invalid shipping method.
	ErrInvalidShippingMethod 

	// ErrInvalidGoodsWeight - 400: Invalid goods weight.
	ErrInvalidGoodsWeight 

	// ErrLogisticsStatusTransitionInvalid - 400: Invalid logistics status transition.
	ErrLogisticsStatusTransitionInvalid 

	// ErrLogisticsAlreadyShipped - 400: Logistics order already shipped.
	ErrLogisticsAlreadyShipped 

	// ErrLogisticsAlreadyDelivered - 400: Logistics order already delivered.
	ErrLogisticsAlreadyDelivered 

	// ErrLogisticsNotShipped - 400: Logistics order not shipped yet.
	ErrLogisticsNotShipped 


	// ErrLogisticsCompanyNotSupported - 400: Logistics company not supported.
	ErrLogisticsCompanyNotSupported

	// ErrLogisticsCompanyServiceUnavailable - 500: Logistics company service unavailable.
	ErrLogisticsCompanyServiceUnavailable 


	// ErrLogisticsServiceUnavailable - 500: Logistics service unavailable.
	ErrLogisticsServiceUnavailable 

	// ErrLogisticsDataIntegrityError - 500: Logistics data integrity error.
	ErrLogisticsDataIntegrityError 
	// ErrLogisticsOperationTimeout - 500: Logistics operation timeout.
	ErrLogisticsOperationTimeout 
)

