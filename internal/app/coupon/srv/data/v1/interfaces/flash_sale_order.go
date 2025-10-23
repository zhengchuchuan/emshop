package interfaces

import (
    "context"
    "emshop/internal/app/coupon/srv/domain/do"
    "gorm.io/gorm"
)

// FlashSaleOrderDataInterface 秒杀订单数据接口
type FlashSaleOrderDataInterface interface {
    Create(ctx context.Context, db *gorm.DB, order *do.FlashSaleOrderDO) error
    GetByRequestID(ctx context.Context, db *gorm.DB, requestID string) (*do.FlashSaleOrderDO, error)
    // MarkCountedByRequestID 将订单状态由 CREATED 原子更新为 COUNTED，用于对 sold_count 进行一次性计数
    // 返回true表示状态成功更新（本次应当进行 sold_count += 1），返回false表示已处理或不存在
    MarkCountedByRequestID(ctx context.Context, db *gorm.DB, requestID string) (bool, error)
}
