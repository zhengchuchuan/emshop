package interfaces

import (
    "context"
    "emshop/internal/app/coupon/srv/domain/do"
    "gorm.io/gorm"
)

// FlashSaleStockLogDataInterface 扣减日志数据接口
type FlashSaleStockLogDataInterface interface {
    BatchCreate(ctx context.Context, db *gorm.DB, logs []*do.FlashSaleStockLogDO) error
}

