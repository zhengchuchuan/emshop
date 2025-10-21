package mysql

import (
    "context"
    "emshop/internal/app/coupon/srv/domain/do"
    "emshop/pkg/log"
    "gorm.io/gorm"
)

type flashSaleStockLogData struct {
    db *gorm.DB
}

func NewFlashSaleStockLogData(db *gorm.DB) *flashSaleStockLogData {
    return &flashSaleStockLogData{db: db}
}

// BatchCreate 批量写入库存扣减日志
func (d *flashSaleStockLogData) BatchCreate(ctx context.Context, db *gorm.DB, logs []*do.FlashSaleStockLogDO) error {
    if db == nil {
        db = d.db
    }
    if len(logs) == 0 {
        return nil
    }
    if err := db.WithContext(ctx).CreateInBatches(logs, 500).Error; err != nil {
        log.Errorf("批量写入库存扣减日志失败: %v", err)
        return err
    }
    return nil
}

