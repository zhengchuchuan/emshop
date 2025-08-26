package mysql

import (
	"context"
	"emshop/internal/app/logistics/srv/data/v1/interfaces"
	"emshop/internal/app/logistics/srv/domain/do"
	"emshop/internal/app/pkg/code"
	"emshop/pkg/errors"
	"time"
	"gorm.io/gorm"
)

type logisticsOrdersRepo struct{}

// NewLogisticsOrdersRepo 创建物流订单仓储实例
func NewLogisticsOrdersRepo() interfaces.LogisticsOrdersRepo {
	return &logisticsOrdersRepo{}
}

// Create 创建物流订单
func (r *logisticsOrdersRepo) Create(ctx context.Context, db *gorm.DB, order *do.LogisticsOrderDO) error {
	if err := db.WithContext(ctx).Create(order).Error; err != nil {
		return errors.WithCode(code.ErrConnectDB, "创建物流订单失败: %v", err)
	}
	return nil
}

// GetByLogisticsSn 根据物流单号查询
func (r *logisticsOrdersRepo) GetByLogisticsSn(ctx context.Context, db *gorm.DB, logisticsSn string) (*do.LogisticsOrderDO, error) {
	var order do.LogisticsOrderDO
	err := db.WithContext(ctx).Where("logistics_sn = ?", logisticsSn).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrLogisticsOrderNotFound, "物流订单不存在")
		}
		return nil, errors.WithCode(code.ErrConnectDB, "查询物流订单失败: %v", err)
	}
	return &order, nil
}

// GetByOrderSn 根据订单号查询
func (r *logisticsOrdersRepo) GetByOrderSn(ctx context.Context, db *gorm.DB, orderSn string) (*do.LogisticsOrderDO, error) {
	var order do.LogisticsOrderDO
	err := db.WithContext(ctx).Where("order_sn = ?", orderSn).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrLogisticsOrderNotFound, "物流订单不存在")
		}
		return nil, errors.WithCode(code.ErrConnectDB, "查询物流订单失败: %v", err)
	}
	return &order, nil
}

// GetByTrackingNumber 根据快递单号查询
func (r *logisticsOrdersRepo) GetByTrackingNumber(ctx context.Context, db *gorm.DB, trackingNumber string) (*do.LogisticsOrderDO, error) {
	var order do.LogisticsOrderDO
	err := db.WithContext(ctx).Where("tracking_number = ?", trackingNumber).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrLogisticsOrderNotFound, "物流订单不存在")
		}
		return nil, errors.WithCode(code.ErrConnectDB, "查询物流订单失败: %v", err)
	}
	return &order, nil
}

// UpdateStatus 更新物流状态
func (r *logisticsOrdersRepo) UpdateStatus(ctx context.Context, db *gorm.DB, logisticsSn string, status int32) error {
	result := db.WithContext(ctx).Model(&do.LogisticsOrderDO{}).
		Where("logistics_sn = ?", logisticsSn).
		Update("logistics_status", status)
	
	if result.Error != nil {
		return errors.WithCode(code.ErrConnectDB, "更新物流状态失败: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.WithCode(code.ErrLogisticsOrderNotFound, "物流订单不存在")
	}
	return nil
}

// UpdateShipmentInfo 更新发货信息
func (r *logisticsOrdersRepo) UpdateShipmentInfo(ctx context.Context, db *gorm.DB, logisticsSn string, shippedAt *time.Time) error {
	updates := map[string]interface{}{
		"logistics_status": do.LogisticsStatusShipped,
	}
	if shippedAt != nil {
		updates["shipped_at"] = shippedAt
	}
	
	result := db.WithContext(ctx).Model(&do.LogisticsOrderDO{}).
		Where("logistics_sn = ?", logisticsSn).
		Updates(updates)
	
	if result.Error != nil {
		return errors.WithCode(code.ErrConnectDB, "更新发货信息失败: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.WithCode(code.ErrLogisticsOrderNotFound, "物流订单不存在")
	}
	return nil
}

// UpdateDeliveryInfo 更新签收信息
func (r *logisticsOrdersRepo) UpdateDeliveryInfo(ctx context.Context, db *gorm.DB, logisticsSn string, deliveredAt *time.Time) error {
	updates := map[string]interface{}{
		"logistics_status": do.LogisticsStatusDelivered,
	}
	if deliveredAt != nil {
		updates["delivered_at"] = deliveredAt
	}
	
	result := db.WithContext(ctx).Model(&do.LogisticsOrderDO{}).
		Where("logistics_sn = ?", logisticsSn).
		Updates(updates)
	
	if result.Error != nil {
		return errors.WithCode(code.ErrConnectDB, "更新签收信息失败: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.WithCode(code.ErrLogisticsOrderNotFound, "物流订单不存在")
	}
	return nil
}

// List 分页查询物流订单
func (r *logisticsOrdersRepo) List(ctx context.Context, db *gorm.DB, offset, limit int, userID *int32) ([]*do.LogisticsOrderDO, int64, error) {
	var orders []*do.LogisticsOrderDO
	var total int64
	
	query := db.WithContext(ctx).Model(&do.LogisticsOrderDO{})
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.WithCode(code.ErrConnectDB, "查询物流订单总数失败: %v", err)
	}
	
	// 分页查询
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, 0, errors.WithCode(code.ErrConnectDB, "查询物流订单失败: %v", err)
	}
	
	return orders, total, nil
}

// FindByStatus 根据状态查询物流订单
func (r *logisticsOrdersRepo) FindByStatus(ctx context.Context, db *gorm.DB, status int32) ([]*do.LogisticsOrderDO, error) {
	var orders []*do.LogisticsOrderDO
	err := db.WithContext(ctx).Where("logistics_status = ?", status).Find(&orders).Error
	if err != nil {
		return nil, errors.WithCode(code.ErrConnectDB, "根据状态查询物流订单失败: %v", err)
	}
	return orders, nil
}

// Update 更新物流订单 (通用更新方法)
func (r *logisticsOrdersRepo) Update(ctx context.Context, db *gorm.DB, order *do.LogisticsOrderDO) error {
	err := db.WithContext(ctx).Save(order).Error
	if err != nil {
		return errors.WithCode(code.ErrConnectDB, "更新物流订单失败: %v", err)
	}
	return nil
}