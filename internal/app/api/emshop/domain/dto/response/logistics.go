package response

import (
	lpbv1 "emshop/api/logistics/v1"
	"time"
)

// LogisticsInfoResponse 物流信息响应
type LogisticsInfoResponse struct {
	LogisticsSN      string     `json:"logistics_sn"`      // 物流单号
	OrderSN          string     `json:"order_sn"`          // 订单号
	TrackingNumber   string     `json:"tracking_number"`   // 快递单号
	LogisticsCompany int32      `json:"logistics_company"` // 物流公司
	ShippingMethod   int32      `json:"shipping_method"`   // 配送方式
	LogisticsStatus  int32      `json:"logistics_status"`  // 物流状态
	SenderName       string     `json:"sender_name"`       // 发件人姓名
	SenderPhone      string     `json:"sender_phone"`      // 发件人电话
	SenderAddress    string     `json:"sender_address"`    // 发件地址
	ReceiverName     string     `json:"receiver_name"`     // 收件人姓名
	ReceiverPhone    string     `json:"receiver_phone"`    // 收件人电话
	ReceiverAddress  string     `json:"receiver_address"`  // 收件地址
	ShippingFee      float64    `json:"shipping_fee"`      // 运费
	EstimatedAt      *time.Time `json:"estimated_at"`      // 预计送达时间
	ShippedAt        *time.Time `json:"shipped_at"`        // 发货时间
	DeliveredAt      *time.Time `json:"delivered_at"`      // 送达时间
	StatusText       string     `json:"status_text"`       // 状态描述
	CompanyName      string     `json:"company_name"`      // 物流公司名称
}

// FromProto 从protobuf转换
func (r *LogisticsInfoResponse) FromProto(pb *lpbv1.GetLogisticsInfoResponse) {
	r.LogisticsSN = pb.LogisticsSn
	r.OrderSN = pb.OrderSn
	r.TrackingNumber = pb.TrackingNumber
	r.LogisticsCompany = pb.LogisticsCompany
	r.ShippingMethod = pb.ShippingMethod
	r.LogisticsStatus = pb.LogisticsStatus
	r.SenderName = pb.SenderName
	r.SenderPhone = pb.SenderPhone
	r.SenderAddress = pb.SenderAddress
	r.ReceiverName = pb.ReceiverName
	r.ReceiverPhone = pb.ReceiverPhone
	r.ReceiverAddress = pb.ReceiverAddress
	r.ShippingFee = pb.ShippingFee

	if pb.EstimatedDeliveryAt != 0 {
		estimatedAt := time.Unix(pb.EstimatedDeliveryAt, 0)
		r.EstimatedAt = &estimatedAt
	}

	if pb.ShippedAt != 0 {
		shippedAt := time.Unix(pb.ShippedAt, 0)
		r.ShippedAt = &shippedAt
	}

	if pb.DeliveredAt != 0 {
		deliveredAt := time.Unix(pb.DeliveredAt, 0)
		r.DeliveredAt = &deliveredAt
	}

	// 设置状态和公司名称描述
	r.StatusText = getLogisticsStatusText(pb.LogisticsStatus)
	r.CompanyName = getLogisticsCompanyName(pb.LogisticsCompany)
}

// LogisticsTrackResponse 物流轨迹响应
type LogisticsTrackResponse struct {
	TrackTime   time.Time `json:"track_time"`  // 轨迹时间
	Location    string    `json:"location"`    // 所在位置
	Description string    `json:"description"` // 描述信息
	Status      int32     `json:"status"`      // 当前状态
}

// FromProto 从protobuf转换
func (r *LogisticsTrackResponse) FromProto(pb *lpbv1.LogisticsTrack) {
	r.TrackTime = time.Unix(pb.TrackTime, 0)
	r.Location = pb.Location
	r.Description = pb.Description
	r.Status = 1 // LogisticsTrack没有Status字段，使用默认值
}

// LogisticsTracksResponse 物流轨迹列表响应
type LogisticsTracksResponse struct {
	LogisticsSN string                    `json:"logistics_sn"` // 物流单号
	OrderSN     string                    `json:"order_sn"`     // 订单号
	Tracks      []*LogisticsTrackResponse `json:"tracks"`       // 轨迹列表
}

// FromProto 从protobuf转换
func (r *LogisticsTracksResponse) FromProto(pb *lpbv1.GetLogisticsTracksResponse) {
	r.LogisticsSN = pb.LogisticsSn
	r.OrderSN = "" // GetLogisticsTracksResponse没有OrderSn字段，需要从请求上下文获取
	r.Tracks = make([]*LogisticsTrackResponse, len(pb.Tracks))
	for i, track := range pb.Tracks {
		r.Tracks[i] = &LogisticsTrackResponse{}
		r.Tracks[i].FromProto(track)
	}
}

// ShippingFeeResponse 运费计算响应
type ShippingFeeResponse struct {
	ShippingFee         float64    `json:"shipping_fee"`          // 运费
	EstimatedDeliveryAt *time.Time `json:"estimated_delivery_at"` // 预计送达时间
	CompanyName         string     `json:"company_name"`          // 物流公司名称
	ShippingMethodName  string     `json:"shipping_method_name"`  // 配送方式名称
}

// FromProto 从protobuf转换
func (r *ShippingFeeResponse) FromProto(pb *lpbv1.CalculateShippingFeeResponse) {
	r.ShippingFee = pb.ShippingFee

	// 根据EstimatedDays计算预计送达时间
	if pb.EstimatedDays > 0 {
		estimatedAt := time.Now().AddDate(0, 0, int(pb.EstimatedDays))
		r.EstimatedDeliveryAt = &estimatedAt
	}

	// CalculateShippingFeeResponse没有这些字段，需要从请求参数获取
	r.CompanyName = ""
	r.ShippingMethodName = ""
}

// LogisticsCompanyResponse 物流公司响应
type LogisticsCompanyResponse struct {
	ID          int32  `json:"id"`           // 公司ID
	Name        string `json:"name"`         // 公司名称
	Code        string `json:"code"`         // 公司代码
	PhoneNumber string `json:"phone_number"` // 客服电话
	Website     string `json:"website"`      // 官网地址
}

// FromProto 从protobuf转换
func (r *LogisticsCompanyResponse) FromProto(pb *lpbv1.LogisticsCompany) {
	r.ID = pb.CompanyId
	r.Name = pb.CompanyName
	r.Code = pb.CompanyCode
	r.PhoneNumber = "" // LogisticsCompany没有PhoneNumber字段
	r.Website = ""     // LogisticsCompany没有Website字段
}

// LogisticsCompaniesResponse 物流公司列表响应
type LogisticsCompaniesResponse struct {
	Companies []*LogisticsCompanyResponse `json:"companies"` // 公司列表
}

// FromProto 从protobuf转换
func (r *LogisticsCompaniesResponse) FromProto(pb *lpbv1.LogisticsCompaniesResponse) {
	r.Companies = make([]*LogisticsCompanyResponse, len(pb.Companies))
	for i, company := range pb.Companies {
		r.Companies[i] = &LogisticsCompanyResponse{}
		r.Companies[i].FromProto(company)
	}
}

// getLogisticsStatusText 获取物流状态描述
func getLogisticsStatusText(status int32) string {
	switch status {
	case 1:
		return "待发货"
	case 2:
		return "已发货"
	case 3:
		return "运输中"
	case 4:
		return "派送中"
	case 5:
		return "已签收"
	case 6:
		return "异常"
	case 7:
		return "退回"
	default:
		return "未知状态"
	}
}

// getLogisticsCompanyName 获取物流公司名称（简单映射，实际应该从配置或数据库获取）
func getLogisticsCompanyName(company int32) string {
	switch company {
	case 1:
		return "顺丰快递"
	case 2:
		return "圆通快递"
	case 3:
		return "中通快递"
	case 4:
		return "申通快递"
	case 5:
		return "韵达快递"
	case 6:
		return "百世汇通"
	case 7:
		return "德邦快递"
	case 8:
		return "京东物流"
	default:
		return "未知快递公司"
	}
}

