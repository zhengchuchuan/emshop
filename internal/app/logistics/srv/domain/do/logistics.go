package do

import "time"

// LogisticsStatus 物流状态枚举
type LogisticsStatus int32

const (
	LogisticsStatusPending     LogisticsStatus = 1 // 待发货
	LogisticsStatusShipped     LogisticsStatus = 2 // 已发货
	LogisticsStatusInTransit   LogisticsStatus = 3 // 运输中
	LogisticsStatusDelivering  LogisticsStatus = 4 // 配送中
	LogisticsStatusDelivered   LogisticsStatus = 5 // 已签收
	LogisticsStatusRejected    LogisticsStatus = 6 // 拒收
	LogisticsStatusReturning   LogisticsStatus = 7 // 退货中
	LogisticsStatusReturned    LogisticsStatus = 8 // 已退货
)

// ShippingMethod 配送方式枚举
type ShippingMethod int32

const (
	ShippingMethodStandard   ShippingMethod = 1 // 标准配送
	ShippingMethodExpress    ShippingMethod = 2 // 急速配送
	ShippingMethodEconomy    ShippingMethod = 3 // 经济配送
	ShippingMethodSelfPickup ShippingMethod = 4 // 自提
)

// LogisticsCompany 物流公司枚举
type LogisticsCompany int32

const (
	CompanyYTO   LogisticsCompany = 1 // 圆通速递
	CompanySTO   LogisticsCompany = 2 // 申通快递
	CompanyZTO   LogisticsCompany = 3 // 中通快递
	CompanyYunda LogisticsCompany = 4 // 韵达速递
	CompanySF    LogisticsCompany = 5 // 顺丰速运
	CompanyJD    LogisticsCompany = 6 // 京东物流
	CompanyEMS   LogisticsCompany = 7 // 中国邮政
)

// LogisticsOrderDO 物流订单数据对象
type LogisticsOrderDO struct {
	ID                    int64     `gorm:"primaryKey;autoIncrement;column:id"`
	LogisticsSn           string    `gorm:"column:logistics_sn;size:64;not null;uniqueIndex"`
	OrderSn               string    `gorm:"column:order_sn;size:64;not null;index"`
	UserID                int32     `gorm:"column:user_id;not null;index"`
	LogisticsCompany      int32     `gorm:"column:logistics_company;not null;index"`
	ShippingMethod        int32     `gorm:"column:shipping_method;not null"`
	TrackingNumber        string    `gorm:"column:tracking_number;size:64;not null;index"`
	LogisticsStatus       int32     `gorm:"column:logistics_status;not null;default:1;index"`
	
	// 发货信息
	SenderName            string    `gorm:"column:sender_name;size:64;not null"`
	SenderPhone           string    `gorm:"column:sender_phone;size:32;not null"`
	SenderAddress         string    `gorm:"column:sender_address;type:text;not null"`
	
	// 收货信息
	ReceiverName          string    `gorm:"column:receiver_name;size:64;not null"`
	ReceiverPhone         string    `gorm:"column:receiver_phone;size:32;not null"`
	ReceiverAddress       string    `gorm:"column:receiver_address;type:text;not null"`
	
	// 时间记录
	ShippedAt             *time.Time `gorm:"column:shipped_at"`
	DeliveredAt           *time.Time `gorm:"column:delivered_at"`
	EstimatedDeliveryAt   *time.Time `gorm:"column:estimated_delivery_at"`
	
	// 费用信息
	ShippingFee           float64   `gorm:"column:shipping_fee;type:decimal(8,2);default:0"`
	InsuranceFee          float64   `gorm:"column:insurance_fee;type:decimal(8,2);default:0"`
	
	// 商品信息（JSON格式）
	ItemsInfo             string    `gorm:"column:items_info;type:text"`
	
	Remark                string    `gorm:"column:remark;type:text"`
	CreatedAt             time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt             time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名
func (LogisticsOrderDO) TableName() string {
	return "logistics_orders"
}

// LogisticsTrackDO 物流轨迹数据对象
type LogisticsTrackDO struct {
	ID             int64     `gorm:"primaryKey;autoIncrement;column:id"`
	LogisticsSn    string    `gorm:"column:logistics_sn;size:64;not null;index"`
	TrackingNumber string    `gorm:"column:tracking_number;size:64;not null;index"`
	Location       string    `gorm:"column:location;size:128;not null"`
	Description    string    `gorm:"column:description;type:text;not null"`
	TrackTime      time.Time `gorm:"column:track_time;not null;index"`
	OperatorName   string    `gorm:"column:operator_name;size:64"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime"`
}

// TableName 指定表名
func (LogisticsTrackDO) TableName() string {
	return "logistics_tracks"
}

// LogisticsCourierDO 物流配送员数据对象
type LogisticsCourierDO struct {
	ID               int64     `gorm:"primaryKey;autoIncrement;column:id"`
	CourierCode      string    `gorm:"column:courier_code;size:32;not null;uniqueIndex"`
	CourierName      string    `gorm:"column:courier_name;size:64;not null"`
	Phone            string    `gorm:"column:phone;size:32;not null"`
	LogisticsCompany int32     `gorm:"column:logistics_company;not null;index"`
	ServiceArea      string    `gorm:"column:service_area;size:128;index"`
	Status           int32     `gorm:"column:status;default:1"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名
func (LogisticsCourierDO) TableName() string {
	return "logistics_couriers"
}

// OrderItem 订单商品项
type OrderItem struct {
	GoodsID  int32   `json:"goods_id"`
	Name     string  `json:"goods_name"`
	Quantity int32   `json:"quantity"`
	Weight   float64 `json:"weight"`
	Volume   float64 `json:"volume"`
}