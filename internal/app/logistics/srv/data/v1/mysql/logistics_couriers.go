package mysql

import (
	"context"
	"emshop/internal/app/logistics/srv/data/v1/interfaces"
	"emshop/internal/app/logistics/srv/domain/do"
	"emshop/internal/app/pkg/code"
	"emshop/pkg/errors"
	"gorm.io/gorm"
)

type logisticsCouriersRepo struct{}

// NewLogisticsCouriersRepo 创建配送员仓储实例
func NewLogisticsCouriersRepo() interfaces.LogisticsCouriersRepo {
	return &logisticsCouriersRepo{}
}

// Create 创建配送员
func (r *logisticsCouriersRepo) Create(ctx context.Context, db *gorm.DB, courier *do.LogisticsCourierDO) error {
	if err := db.WithContext(ctx).Create(courier).Error; err != nil {
		return errors.WithCode(code.ErrConnectDB, "创建配送员失败: %v", err)
	}
	return nil
}

// GetByCourierCode 根据配送员编号查询
func (r *logisticsCouriersRepo) GetByCourierCode(ctx context.Context, db *gorm.DB, courierCode string) (*do.LogisticsCourierDO, error) {
	var courier do.LogisticsCourierDO
	err := db.WithContext(ctx).Where("courier_code = ? AND status = 1", courierCode).First(&courier).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrLogisticsCourierNotFound, "配送员不存在")
		}
		return nil, errors.WithCode(code.ErrConnectDB, "查询配送员失败: %v", err)
	}
	return &courier, nil
}

// GetByCompany 根据物流公司查询配送员
func (r *logisticsCouriersRepo) GetByCompany(ctx context.Context, db *gorm.DB, company int32) ([]*do.LogisticsCourierDO, error) {
	var couriers []*do.LogisticsCourierDO
	err := db.WithContext(ctx).Where("logistics_company = ? AND status = 1", company).Find(&couriers).Error
	if err != nil {
		return nil, errors.WithCode(code.ErrConnectDB, "查询配送员失败: %v", err)
	}
	return couriers, nil
}

// GetByServiceArea 根据服务区域查询配送员
func (r *logisticsCouriersRepo) GetByServiceArea(ctx context.Context, db *gorm.DB, serviceArea string) ([]*do.LogisticsCourierDO, error) {
	var couriers []*do.LogisticsCourierDO
	err := db.WithContext(ctx).Where("service_area LIKE ? AND status = 1", "%"+serviceArea+"%").Find(&couriers).Error
	if err != nil {
		return nil, errors.WithCode(code.ErrConnectDB, "查询配送员失败: %v", err)
	}
	return couriers, nil
}

// GetRandomCourier 随机获取配送员
func (r *logisticsCouriersRepo) GetRandomCourier(ctx context.Context, db *gorm.DB, company int32) (*do.LogisticsCourierDO, error) {
	var courier do.LogisticsCourierDO
	err := db.WithContext(ctx).Where("logistics_company = ? AND status = 1", company).
		Order("RAND()").First(&courier).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrLogisticsCourierNotFound, "没有可用的配送员")
		}
		return nil, errors.WithCode(code.ErrConnectDB, "查询配送员失败: %v", err)
	}
	return &courier, nil
}

// Update 更新配送员信息
func (r *logisticsCouriersRepo) Update(ctx context.Context, db *gorm.DB, courier *do.LogisticsCourierDO) error {
	result := db.WithContext(ctx).Save(courier)
	if result.Error != nil {
		return errors.WithCode(code.ErrConnectDB, "更新配送员信息失败: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.WithCode(code.ErrLogisticsCourierNotFound, "配送员不存在")
	}
	return nil
}

// List 分页查询配送员
func (r *logisticsCouriersRepo) List(ctx context.Context, db *gorm.DB, offset, limit int, company *int32, area *string) ([]*do.LogisticsCourierDO, int64, error) {
	var couriers []*do.LogisticsCourierDO
	var total int64
	
	query := db.WithContext(ctx).Model(&do.LogisticsCourierDO{}).Where("status = 1")
	
	if company != nil {
		query = query.Where("logistics_company = ?", *company)
	}
	if area != nil && *area != "" {
		query = query.Where("service_area LIKE ?", "%"+*area+"%")
	}
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.WithCode(code.ErrConnectDB, "查询配送员总数失败: %v", err)
	}
	
	// 分页查询
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&couriers).Error; err != nil {
		return nil, 0, errors.WithCode(code.ErrConnectDB, "查询配送员失败: %v", err)
	}
	
	return couriers, total, nil
}