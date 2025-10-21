package do

import "time"

// FlashSaleStockLogDO Redis 扣减日志落库实体
type FlashSaleStockLogDO struct {
    ID         int64     `gorm:"primaryKey" json:"id"`
    ActivityID int64     `gorm:"column:activity_id;type:bigint;not null;index:idx_activity_user" json:"activity_id"`
    UserID     int64     `gorm:"column:user_id;type:bigint;not null;index:idx_activity_user" json:"user_id"`
    Decr       int32     `gorm:"column:decr;type:int;not null" json:"decr"`
    TS         int64     `gorm:"column:ts;type:bigint;not null;index:idx_ts" json:"ts"`
    RequestID  string    `gorm:"column:request_id;type:varchar(64);index:idx_req" json:"request_id"`
    CreatedAt  time.Time `gorm:"column:created_at;type:datetime;autoCreateTime" json:"created_at"`
}

func (FlashSaleStockLogDO) TableName() string { return "flash_sale_stock_logs" }
