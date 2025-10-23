package mysql

import (
    "context"
    "errors"
    "emshop/internal/app/coupon/srv/data/v1/interfaces"
    "emshop/internal/app/coupon/srv/domain/do"
    "emshop/pkg/log"
    "gorm.io/gorm"
)

type flashSaleOrderData struct { db *gorm.DB }

func NewFlashSaleOrderData(db *gorm.DB) *flashSaleOrderData { return &flashSaleOrderData{db: db} }

func (d *flashSaleOrderData) Create(ctx context.Context, db *gorm.DB, order *do.FlashSaleOrderDO) error {
    if db == nil { db = d.db }
    if err := db.WithContext(ctx).Create(order).Error; err != nil {
        log.Errorf("创建秒杀订单失败: %v", err)
        return err
    }
    return nil
}

func (d *flashSaleOrderData) GetByRequestID(ctx context.Context, db *gorm.DB, requestID string) (*do.FlashSaleOrderDO, error) {
    if db == nil { db = d.db }
    var row do.FlashSaleOrderDO
    if err := db.WithContext(ctx).Where("request_id = ?", requestID).First(&row).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
        log.Errorf("根据request_id查询秒杀订单失败: %v", err)
        return nil, err
    }
    // 保护性校验（极端情况下防止脏数据或误判）
    if row.ID == 0 || row.RequestID == "" {
        return nil, nil
    }
    return &row, nil
}

// MarkCountedByRequestID 原子将订单状态由 CREATED 更新为 COUNTED，作为对 sold_count 计数的幂等闸门
func (d *flashSaleOrderData) MarkCountedByRequestID(ctx context.Context, db *gorm.DB, requestID string) (bool, error) {
    if db == nil { db = d.db }
    res := db.WithContext(ctx).Model(&do.FlashSaleOrderDO{}).
        Where("request_id = ? AND status = ?", requestID, "CREATED").
        Update("status", "COUNTED")
    if res.Error != nil {
        log.Errorf("标记订单已计数失败: %v", res.Error)
        return false, res.Error
    }
    return res.RowsAffected > 0, nil
}

var _ interfaces.FlashSaleOrderDataInterface = (*flashSaleOrderData)(nil)
