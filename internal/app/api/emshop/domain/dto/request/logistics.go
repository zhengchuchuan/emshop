package request

import (
	lpbv1 "emshop/api/logistics/v1"
)

// GetLogisticsInfoRequest 获取物流信息请求
type GetLogisticsInfoRequest struct {
	OrderSN        *string `form:"order_sn" json:"order_sn"`               // 订单号
	LogisticsSN    *string `form:"logistics_sn" json:"logistics_sn"`       // 物流单号  
	TrackingNumber *string `form:"tracking_number" json:"tracking_number"` // 快递单号
}

// ToProto 转换为protobuf请求
func (r *GetLogisticsInfoRequest) ToProto() *lpbv1.GetLogisticsInfoRequest {
	req := &lpbv1.GetLogisticsInfoRequest{}
	
	// 根据提供的参数设置查询条件（oneof字段）
	if r.OrderSN != nil {
		req.Query = &lpbv1.GetLogisticsInfoRequest_OrderSn{
			OrderSn: *r.OrderSN,
		}
	} else if r.LogisticsSN != nil {
		req.Query = &lpbv1.GetLogisticsInfoRequest_LogisticsSn{
			LogisticsSn: *r.LogisticsSN,
		}
	} else if r.TrackingNumber != nil {
		req.Query = &lpbv1.GetLogisticsInfoRequest_TrackingNumber{
			TrackingNumber: *r.TrackingNumber,
		}
	}
	
	return req
}

// Validate 验证请求参数
func (r *GetLogisticsInfoRequest) Validate() error {
	if r.OrderSN == nil && r.LogisticsSN == nil && r.TrackingNumber == nil {
		return &ValidationError{Field: "query", Message: "必须提供order_sn、logistics_sn或tracking_number中的一个"}
	}
	return nil
}

// GetLogisticsTracksRequest 获取物流轨迹请求
type GetLogisticsTracksRequest struct {
	OrderSN     *string `form:"order_sn" json:"order_sn"`           // 订单号
	LogisticsSN *string `form:"logistics_sn" json:"logistics_sn"`   // 物流单号
}

// ToProto 转换为protobuf请求  
func (r *GetLogisticsTracksRequest) ToProto() *lpbv1.GetLogisticsTracksRequest {
	req := &lpbv1.GetLogisticsTracksRequest{}
	
	// 根据提供的参数设置查询条件（oneof字段）
	// GetLogisticsTracksRequest 只支持 LogisticsSn 和 TrackingNumber
	if r.LogisticsSN != nil {
		req.Query = &lpbv1.GetLogisticsTracksRequest_LogisticsSn{
			LogisticsSn: *r.LogisticsSN,
		}
	}
	
	return req
}

// Validate 验证请求参数
func (r *GetLogisticsTracksRequest) Validate() error {
	if r.LogisticsSN == nil {
		return &ValidationError{Field: "logistics_sn", Message: "必须提供logistics_sn"}
	}
	return nil
}

// CalculateShippingFeeRequest 计算运费请求
type CalculateShippingFeeRequest struct {
	ReceiverAddress   string          `json:"receiver_address" binding:"required,min=5"`    // 收货地址
	Items             []LogisticsItem `json:"items" binding:"required,dive"`               // 商品列表
	LogisticsCompany  int32           `json:"logistics_company" binding:"required,min=1"`  // 物流公司
	ShippingMethod    int32           `json:"shipping_method" binding:"required,min=1"`    // 配送方式
}

// LogisticsItem 物流商品项
type LogisticsItem struct {
	GoodsID   int32   `json:"goods_id" binding:"required,min=1"`   // 商品ID
	GoodsName string  `json:"goods_name" binding:"required,min=1"` // 商品名称
	Quantity  int32   `json:"quantity" binding:"required,min=1"`   // 数量
	Weight    float64 `json:"weight" binding:"required,min=0"`     // 重量(kg)
	Volume    float64 `json:"volume" binding:"required,min=0"`     // 体积(cm³)
}

// ToProto 转换为protobuf请求
func (r *CalculateShippingFeeRequest) ToProto() *lpbv1.CalculateShippingFeeRequest {
	// 计算总重量和体积
	var totalWeight, totalVolume, goodsValue float64
	for _, item := range r.Items {
		totalWeight += item.Weight * float64(item.Quantity)
		totalVolume += item.Volume * float64(item.Quantity)
		// 这里可能需要从商品服务获取价格来计算goodsValue
	}
	
	return &lpbv1.CalculateShippingFeeRequest{
		SenderAddress:   "",                // 需要从配置或请求中获取
		ReceiverAddress: r.ReceiverAddress,
		ShippingMethod:  r.ShippingMethod,
		TotalWeight:     totalWeight,
		TotalVolume:     totalVolume,
		GoodsValue:      goodsValue,
		NeedInsurance:   false, // 可以作为参数传入
	}
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}