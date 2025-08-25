package mysql

import (
	"context"
	"emshop/internal/app/logistics/srv/data/v1/interfaces"
	"emshop/internal/app/logistics/srv/domain/do"
	"emshop/internal/app/pkg/code"
	"emshop/pkg/errors"
	"gorm.io/gorm"
)

type logisticsTracksRepo struct{}

// NewLogisticsTracksRepo 创建物流轨迹仓储实例
func NewLogisticsTracksRepo() interfaces.LogisticsTracksRepo {
	return &logisticsTracksRepo{}
}

// Create 创建轨迹记录
func (r *logisticsTracksRepo) Create(ctx context.Context, db *gorm.DB, track *do.LogisticsTrackDO) error {
	if err := db.WithContext(ctx).Create(track).Error; err != nil {
		return errors.WithCode(code.ErrConnectDB, "创建物流轨迹失败: %v", err)
	}
	return nil
}

// BatchCreate 批量创建轨迹记录
func (r *logisticsTracksRepo) BatchCreate(ctx context.Context, db *gorm.DB, tracks []*do.LogisticsTrackDO) error {
	if len(tracks) == 0 {
		return nil
	}
	
	if err := db.WithContext(ctx).CreateInBatches(tracks, 100).Error; err != nil {
		return errors.WithCode(code.ErrConnectDB, "批量创建物流轨迹失败: %v", err)
	}
	return nil
}

// GetByLogisticsSn 根据物流单号查询轨迹
func (r *logisticsTracksRepo) GetByLogisticsSn(ctx context.Context, db *gorm.DB, logisticsSn string) ([]*do.LogisticsTrackDO, error) {
	var tracks []*do.LogisticsTrackDO
	err := db.WithContext(ctx).Where("logistics_sn = ?", logisticsSn).
		Order("track_time ASC").Find(&tracks).Error
	if err != nil {
		return nil, errors.WithCode(code.ErrConnectDB, "查询物流轨迹失败: %v", err)
	}
	return tracks, nil
}

// GetByTrackingNumber 根据快递单号查询轨迹
func (r *logisticsTracksRepo) GetByTrackingNumber(ctx context.Context, db *gorm.DB, trackingNumber string) ([]*do.LogisticsTrackDO, error) {
	var tracks []*do.LogisticsTrackDO
	err := db.WithContext(ctx).Where("tracking_number = ?", trackingNumber).
		Order("track_time ASC").Find(&tracks).Error
	if err != nil {
		return nil, errors.WithCode(code.ErrConnectDB, "查询物流轨迹失败: %v", err)
	}
	return tracks, nil
}

// GetLatest 获取最新轨迹
func (r *logisticsTracksRepo) GetLatest(ctx context.Context, db *gorm.DB, logisticsSn string) (*do.LogisticsTrackDO, error) {
	var track do.LogisticsTrackDO
	err := db.WithContext(ctx).Where("logistics_sn = ?", logisticsSn).
		Order("track_time DESC").First(&track).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrLogisticsTrackNotFound, "物流轨迹不存在")
		}
		return nil, errors.WithCode(code.ErrConnectDB, "查询最新物流轨迹失败: %v", err)
	}
	return &track, nil
}