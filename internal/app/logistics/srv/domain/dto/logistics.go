package dto

import "time"

// CreateLogisticsOrderDTO 创建物流订单请求
type CreateLogisticsOrderDTO struct {
	OrderSn          string      `json:"order_sn" binding:"required"`
	UserID           int32       `json:"user_id" binding:"required"`
	LogisticsCompany int32       `json:"logistics_company" binding:"required"`
	ShippingMethod   int32       `json:"shipping_method" binding:"required"`
	
	// 发货信息
	SenderName       string      `json:"sender_name" binding:"required"`
	SenderPhone      string      `json:"sender_phone" binding:"required"`
	SenderAddress    string      `json:"sender_address" binding:"required"`
	
	// 收货信息
	ReceiverName     string      `json:"receiver_name" binding:"required"`
	ReceiverPhone    string      `json:"receiver_phone" binding:"required"`
	ReceiverAddress  string      `json:"receiver_address" binding:"required"`
	
	// 商品信息
	Items            []OrderItemDTO `json:"items"`
	
	Remark           string      `json:"remark"`
}

// OrderItemDTO 订单商品项
type OrderItemDTO struct {
	GoodsID  int32   `json:"goods_id"`
	Name     string  `json:"goods_name"`
	Quantity int32   `json:"quantity"`
	Weight   float64 `json:"weight"`  // 重量(kg)
	Volume   float64 `json:"volume"`  // 体积(cm³)
}

// LogisticsOrderDTO 物流订单响应
type LogisticsOrderDTO struct {
	LogisticsSn         string    `json:"logistics_sn"`
	TrackingNumber      string    `json:"tracking_number"`
	ShippingFee         float64   `json:"shipping_fee"`
	EstimatedDeliveryAt time.Time `json:"estimated_delivery_at"`
}

// GetLogisticsInfoDTO 查询物流信息请求
type GetLogisticsInfoDTO struct {
	LogisticsSn    *string `json:"logistics_sn,omitempty"`
	OrderSn        *string `json:"order_sn,omitempty"`
	TrackingNumber *string `json:"tracking_number,omitempty"`
}

// LogisticsInfoDTO 物流信息响应
type LogisticsInfoDTO struct {
	LogisticsSn         string     `json:"logistics_sn"`
	OrderSn             string     `json:"order_sn"`
	TrackingNumber      string     `json:"tracking_number"`
	LogisticsCompany    int32      `json:"logistics_company"`
	ShippingMethod      int32      `json:"shipping_method"`
	LogisticsStatus     int32      `json:"logistics_status"`
	
	SenderName          string     `json:"sender_name"`
	SenderPhone         string     `json:"sender_phone"`
	SenderAddress       string     `json:"sender_address"`
	
	ReceiverName        string     `json:"receiver_name"`
	ReceiverPhone       string     `json:"receiver_phone"`
	ReceiverAddress     string     `json:"receiver_address"`
	
	ShippingFee         float64    `json:"shipping_fee"`
	ShippedAt           *time.Time `json:"shipped_at,omitempty"`
	DeliveredAt         *time.Time `json:"delivered_at,omitempty"`
	EstimatedDeliveryAt *time.Time `json:"estimated_delivery_at,omitempty"`
	
	Remark              string     `json:"remark"`
}

// GetLogisticsTracksDTO 查询物流轨迹请求
type GetLogisticsTracksDTO struct {
	LogisticsSn    *string `json:"logistics_sn,omitempty"`
	TrackingNumber *string `json:"tracking_number,omitempty"`
}

// LogisticsTrackDTO 物流轨迹
type LogisticsTrackDTO struct {
	Location     string    `json:"location"`
	Description  string    `json:"description"`
	TrackTime    time.Time `json:"track_time"`
	OperatorName string    `json:"operator_name"`
}

// LogisticsTracksDTO 物流轨迹响应
type LogisticsTracksDTO struct {
	LogisticsSn    string              `json:"logistics_sn"`
	TrackingNumber string              `json:"tracking_number"`
	Tracks         []LogisticsTrackDTO `json:"tracks"`
}

// UpdateLogisticsStatusDTO 更新物流状态请求
type UpdateLogisticsStatusDTO struct {
	LogisticsSn string `json:"logistics_sn" binding:"required"`
	NewStatus   int32  `json:"new_status" binding:"required"`
	Remark      string `json:"remark"`
}

// SimulateShipmentDTO 模拟发货请求
type SimulateShipmentDTO struct {
	LogisticsSn  string `json:"logistics_sn" binding:"required"`
	CourierName  string `json:"courier_name"`
	CourierPhone string `json:"courier_phone"`
}

// SimulateDeliveryDTO 模拟签收请求
type SimulateDeliveryDTO struct {
	LogisticsSn     string `json:"logistics_sn" binding:"required"`
	ReceiverName    string `json:"receiver_name"`
	DeliveryRemark  string `json:"delivery_remark"`
}

// CalculateShippingFeeDTO 计算运费请求
type CalculateShippingFeeDTO struct {
	SenderAddress   string  `json:"sender_address" binding:"required"`
	ReceiverAddress string  `json:"receiver_address" binding:"required"`
	ShippingMethod  int32   `json:"shipping_method" binding:"required"`
	TotalWeight     float64 `json:"total_weight"`
	TotalVolume     float64 `json:"total_volume"`
	GoodsValue      float64 `json:"goods_value"`
	NeedInsurance   bool    `json:"need_insurance"`
}

// ShippingFeeDTO 运费计算响应
type ShippingFeeDTO struct {
	ShippingFee    float64 `json:"shipping_fee"`
	InsuranceFee   float64 `json:"insurance_fee"`
	TotalFee       float64 `json:"total_fee"`
	EstimatedDays  int32   `json:"estimated_days"`
}

// LogisticsCompanyDTO 物流公司信息
type LogisticsCompanyDTO struct {
	CompanyID   int32  `json:"company_id"`
	CompanyName string `json:"company_name"`
	CompanyCode string `json:"company_code"`
}

// CourierDTO 配送员信息
type CourierDTO struct {
	CourierCode      string `json:"courier_code"`
	CourierName      string `json:"courier_name"`
	Phone            string `json:"phone"`
	LogisticsCompany int32  `json:"logistics_company"`
	ServiceArea      string `json:"service_area"`
}

// GetCouriersDTO 获取配送员请求
type GetCouriersDTO struct {
	LogisticsCompany *int32  `json:"logistics_company,omitempty"`
	ServiceArea      *string `json:"service_area,omitempty"`
}