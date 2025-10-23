package do

import (
    "time"
)

// FlashSaleOrderDO 秒杀订单（在优惠券服务内直接生成的轻量订单）
type FlashSaleOrderDO struct {
    ID          int64     `gorm:"primaryKey" json:"id"`
    OrderSn     string    `gorm:"column:order_sn;type:varchar(64);uniqueIndex:uk_order_sn" json:"order_sn"`
    RequestID   string    `gorm:"column:request_id;type:varchar(64);uniqueIndex:uk_req_id" json:"request_id"`
    UserID      int64     `gorm:"column:user_id;type:bigint;index:idx_user" json:"user_id"`
    FlashSaleID int64     `gorm:"column:flash_sale_id;type:bigint;index:idx_flash" json:"flash_sale_id"`
    CouponID    int64     `gorm:"column:coupon_id;type:bigint;index:idx_coupon" json:"coupon_id"`
    Status      string    `gorm:"column:status;type:varchar(32);index:idx_status" json:"status"`
    Amount      float64   `gorm:"column:amount;type:decimal(10,2);default:0" json:"amount"`
    CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (FlashSaleOrderDO) TableName() string { return "flash_sale_orders" }

