package mysql

import (
	"context"
	"emshop/internal/app/payment/srv/data/v1/interfaces"
	"emshop/internal/app/payment/srv/domain/do"
	"emshop/internal/app/pkg/code"
	"emshop/pkg/errors"
	"emshop/pkg/log"
	"time"

	"gorm.io/gorm"
)

type stockReservationData struct {
	db *gorm.DB
}

// NewStockReservationData 创建库存预留数据访问对象
func NewStockReservationData(db *gorm.DB) interfaces.StockReservationDataInterface {
	return &stockReservationData{db: db}
}

// Create 创建库存预留记录
func (s *stockReservationData) Create(ctx context.Context, db *gorm.DB, reservation *do.StockReservationDO) error {
	if db == nil {
		db = s.db
	}
	
	if err := db.WithContext(ctx).Create(reservation).Error; err != nil {
		log.Errorf("创建库存预留记录失败: %v", err)
		return errors.WithCode(code.ErrStockReservationFailed, "创建库存预留记录失败")
	}
	
	return nil
}

// Update 更新库存预留记录
func (s *stockReservationData) Update(ctx context.Context, db *gorm.DB, reservation *do.StockReservationDO) error {
	if db == nil {
		db = s.db
	}
	
	if err := db.WithContext(ctx).Save(reservation).Error; err != nil {
		log.Errorf("更新库存预留记录失败: %v", err)
		return errors.WithCode(code.ErrConnectDB, "更新库存预留记录失败")
	}
	
	return nil
}

// Get 获取单个库存预留记录
func (s *stockReservationData) Get(ctx context.Context, db *gorm.DB, orderSn string, goodsID int32) (*do.StockReservationDO, error) {
	if db == nil {
		db = s.db
	}
	
	var reservation do.StockReservationDO
	if err := db.WithContext(ctx).Where("order_sn = ? AND goods_id = ?", orderSn, goodsID).First(&reservation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrStockReservationNotFound, "库存预留记录不存在")
		}
		log.Errorf("查询库存预留记录失败: %v", err)
		return nil, errors.WithCode(code.ErrConnectDB, "查询库存预留记录失败")
	}
	
	return &reservation, nil
}

// GetByOrderSn 根据订单号获取所有库存预留记录
func (s *stockReservationData) GetByOrderSn(ctx context.Context, db *gorm.DB, orderSn string) ([]*do.StockReservationDO, error) {
	if db == nil {
		db = s.db
	}
	
	var reservations []*do.StockReservationDO
	if err := db.WithContext(ctx).Where("order_sn = ?", orderSn).Find(&reservations).Error; err != nil {
		log.Errorf("查询订单库存预留记录失败: %v", err)
		return nil, errors.WithCode(code.ErrConnectDB, "查询订单库存预留记录失败")
	}
	
	return reservations, nil
}

// BatchCreate 批量创建库存预留记录
func (s *stockReservationData) BatchCreate(ctx context.Context, db *gorm.DB, reservations []*do.StockReservationDO) error {
	if db == nil {
		db = s.db
	}
	
	if len(reservations) == 0 {
		return nil
	}
	
	if err := db.WithContext(ctx).CreateInBatches(reservations, 100).Error; err != nil {
		log.Errorf("批量创建库存预留记录失败: %v", err)
		return errors.WithCode(code.ErrStockReservationFailed, "批量创建库存预留记录失败")
	}
	
	return nil
}

// BatchUpdateStatus 批量更新订单的库存预留状态
func (s *stockReservationData) BatchUpdateStatus(ctx context.Context, db *gorm.DB, orderSn string, status do.StockReservationStatus) error {
	if db == nil {
		db = s.db
	}
	
	updates := map[string]interface{}{
		"status": status,
	}
	
	// 根据状态更新相应的时间字段
	now := time.Now()
	switch status {
	case do.StockReservationStatusConfirmed:
		updates["confirmed_at"] = &now
	case do.StockReservationStatusReleased:
		updates["released_at"] = &now
	}
	
	if err := db.WithContext(ctx).Model(&do.StockReservationDO{}).
		Where("order_sn = ?", orderSn).
		Updates(updates).Error; err != nil {
		log.Errorf("批量更新库存预留状态失败: %v", err)
		return errors.WithCode(code.ErrConnectDB, "批量更新库存预留状态失败")
	}
	
	return nil
}

// UpdateStatus 更新单个商品的库存预留状态
func (s *stockReservationData) UpdateStatus(ctx context.Context, db *gorm.DB, orderSn string, goodsID int32, status do.StockReservationStatus) error {
	if db == nil {
		db = s.db
	}
	
	updates := map[string]interface{}{
		"status": status,
	}
	
	// 根据状态更新相应的时间字段
	now := time.Now()
	switch status {
	case do.StockReservationStatusConfirmed:
		updates["confirmed_at"] = &now
	case do.StockReservationStatusReleased:
		updates["released_at"] = &now
	}
	
	if err := db.WithContext(ctx).Model(&do.StockReservationDO{}).
		Where("order_sn = ? AND goods_id = ?", orderSn, goodsID).
		Updates(updates).Error; err != nil {
		log.Errorf("更新库存预留状态失败: %v", err)
		return errors.WithCode(code.ErrConnectDB, "更新库存预留状态失败")
	}
	
	return nil
}

// FindExpiredReservations 查找过期的库存预留记录
func (s *stockReservationData) FindExpiredReservations(ctx context.Context, db *gorm.DB, beforeTime time.Time) ([]*do.StockReservationDO, error) {
	if db == nil {
		db = s.db
	}
	
	var reservations []*do.StockReservationDO
	if err := db.WithContext(ctx).
		Where("status = ? AND reserved_at < ?", do.StockReservationStatusReserved, beforeTime).
		Find(&reservations).Error; err != nil {
		log.Errorf("查找过期库存预留记录失败: %v", err)
		return nil, errors.WithCode(code.ErrConnectDB, "查找过期库存预留记录失败")
	}
	
	return reservations, nil
}

// CountByStatus 按状态统计库存预留记录数量
func (s *stockReservationData) CountByStatus(ctx context.Context, db *gorm.DB, status do.StockReservationStatus) (int64, error) {
	if db == nil {
		db = s.db
	}
	
	var count int64
	if err := db.WithContext(ctx).Model(&do.StockReservationDO{}).Where("status = ?", status).Count(&count).Error; err != nil {
		log.Errorf("统计库存预留记录数量失败: %v", err)
		return 0, errors.WithCode(code.ErrConnectDB, "统计库存预留记录数量失败")
	}
	
	return count, nil
}