package do

import (
	"emshop/pkg/db"
	"time"
)

// PaymentStatus 支付状态定义
type PaymentStatus int32

const (
	PaymentStatusPending   PaymentStatus = 1 // 待支付
	PaymentStatusPaid      PaymentStatus = 2 // 支付成功
	PaymentStatusFailed    PaymentStatus = 3 // 支付失败
	PaymentStatusCancelled PaymentStatus = 4 // 已取消
	PaymentStatusRefunding PaymentStatus = 5 // 退款中
	PaymentStatusRefunded  PaymentStatus = 6 // 已退款
)

// PaymentMethod 支付方式定义
type PaymentMethod int32

const (
	PaymentMethodWechat    PaymentMethod = 1 // 微信支付
	PaymentMethodAlipay    PaymentMethod = 2 // 支付宝
	PaymentMethodUnionPay  PaymentMethod = 3 // 银联支付
	PaymentMethodBank      PaymentMethod = 4 // 网银支付
	PaymentMethodBalance   PaymentMethod = 5 // 余额支付
)

// PaymentOrderDO 支付订单数据对象
type PaymentOrderDO struct {
	db.BaseModel
	PaymentSn    string         `json:"payment_sn" gorm:"column:payment_sn;type:varchar(64);not null;uniqueIndex:idx_payment_sn;comment:支付单号"`
	OrderSn      string         `json:"order_sn" gorm:"column:order_sn;type:varchar(64);not null;index:idx_order_sn;comment:订单号"`
	UserID       int32          `json:"user_id" gorm:"column:user_id;type:int;not null;index:idx_user_id;comment:用户ID"`
	Amount       float64        `json:"amount" gorm:"column:amount;type:decimal(10,2);not null;comment:支付金额"`
	PaymentMethod PaymentMethod `json:"payment_method" gorm:"column:payment_method;type:tinyint;not null;comment:支付方式"`
	PaymentStatus PaymentStatus `json:"payment_status" gorm:"column:payment_status;type:tinyint;not null;default:1;index:idx_status;comment:支付状态"`
	ThirdPartySn *string        `json:"third_party_sn" gorm:"column:third_party_sn;type:varchar(128);comment:第三方支付单号"`
	PaidAt       *time.Time     `json:"paid_at" gorm:"column:paid_at;type:timestamp;comment:支付完成时间"`
	ExpiredAt    time.Time      `json:"expired_at" gorm:"column:expired_at;type:timestamp;not null;index:idx_expired_at;comment:支付过期时间"`
}

// TableName 指定表名
func (PaymentOrderDO) TableName() string {
	return "payment_orders"
}

// PaymentLogDO 支付日志数据对象
type PaymentLogDO struct {
	db.BaseModel
	PaymentSn    string    `json:"payment_sn" gorm:"column:payment_sn;type:varchar(64);not null;index:idx_payment_sn;comment:支付单号"`
	Action       string    `json:"action" gorm:"column:action;type:varchar(32);not null;index:idx_action;comment:操作类型"`
	StatusFrom   *int32    `json:"status_from" gorm:"column:status_from;type:tinyint;comment:状态变更前"`
	StatusTo     *int32    `json:"status_to" gorm:"column:status_to;type:tinyint;comment:状态变更后"`
	Remark       string    `json:"remark" gorm:"column:remark;type:text;comment:备注信息"`
	OperatorType string    `json:"operator_type" gorm:"column:operator_type;type:enum('user','system','admin');not null;default:'system';comment:操作类型"`
	OperatorID   *int32    `json:"operator_id" gorm:"column:operator_id;type:int;comment:操作人ID"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;index:idx_created_at"`
}

// TableName 指定表名
func (PaymentLogDO) TableName() string {
	return "payment_logs"
}

// StockReservationStatus 库存预留状态
type StockReservationStatus int32

const (
	StockReservationStatusReserved StockReservationStatus = 1 // 已预留
	StockReservationStatusConfirmed StockReservationStatus = 2 // 已确认
	StockReservationStatusReleased  StockReservationStatus = 3 // 已释放
)

// StockReservationDO 库存预留记录数据对象
type StockReservationDO struct {
	db.BaseModel
	OrderSn     string                 `json:"order_sn" gorm:"column:order_sn;type:varchar(64);not null;index:idx_order_sn;comment:订单号"`
	GoodsID     int32                  `json:"goods_id" gorm:"column:goods_id;type:int;not null;index:idx_goods_id;comment:商品ID"`
	ReservedNum int32                  `json:"reserved_num" gorm:"column:reserved_num;type:int;not null;comment:预留数量"`
	Status      StockReservationStatus `json:"status" gorm:"column:status;type:tinyint;not null;default:1;index:idx_status;comment:状态"`
	ReservedAt  time.Time              `json:"reserved_at" gorm:"column:reserved_at;type:timestamp;default:CURRENT_TIMESTAMP;index:idx_reserved_at;comment:预留时间"`
	ConfirmedAt *time.Time             `json:"confirmed_at" gorm:"column:confirmed_at;type:timestamp;comment:确认时间"`
	ReleasedAt  *time.Time             `json:"released_at" gorm:"column:released_at;type:timestamp;comment:释放时间"`
}

// TableName 指定表名
func (StockReservationDO) TableName() string {
	return "stock_reservations"
}

// GoodsDetail 商品详情（用于库存操作）
type GoodsDetail struct {
	Goods int32 `json:"goods"`
	Num   int32 `json:"num"`
}

// GoodsDetailList 商品详情列表，实现排序接口
type GoodsDetailList []GoodsDetail

func (g GoodsDetailList) Len() int           { return len(g) }
func (g GoodsDetailList) Less(i, j int) bool { return g[i].Goods < g[j].Goods }
func (g GoodsDetailList) Swap(i, j int)      { g[i], g[j] = g[j], g[i] }

// PaymentOrderDOList 支付订单列表
type PaymentOrderDOList struct {
	TotalCount int64              `json:"totalCount"`
	Items      []*PaymentOrderDO `json:"items"`
}

// PaymentLogDOList 支付日志列表
type PaymentLogDOList struct {
	TotalCount int64           `json:"totalCount"`
	Items      []*PaymentLogDO `json:"items"`
}

// StockReservationDOList 库存预留记录列表
type StockReservationDOList struct {
	TotalCount int64                 `json:"totalCount"`
	Items      []*StockReservationDO `json:"items"`
}