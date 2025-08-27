package mysql

import (
	"context"
	"emshop/internal/app/coupon/srv/domain/do"
	v1 "emshop/pkg/common/meta/v1"
	"emshop/pkg/log"
	"gorm.io/gorm"
)

type flashSaleRecordData struct {
	db *gorm.DB
}

// NewFlashSaleRecordData 创建秒杀记录数据访问对象
func NewFlashSaleRecordData(db *gorm.DB) *flashSaleRecordData {
	return &flashSaleRecordData{
		db: db,
	}
}

// Create 创建秒杀记录
func (fsrd *flashSaleRecordData) Create(ctx context.Context, db *gorm.DB, record *do.FlashSaleRecordDO) error {
	if db == nil {
		db = fsrd.db
	}
	
	if err := db.WithContext(ctx).Create(record).Error; err != nil {
		log.Errorf("创建秒杀记录失败: %v", err)
		return err
	}
	return nil
}

// Update 更新秒杀记录
func (fsrd *flashSaleRecordData) Update(ctx context.Context, db *gorm.DB, record *do.FlashSaleRecordDO) error {
	if db == nil {
		db = fsrd.db
	}
	
	if err := db.WithContext(ctx).Save(record).Error; err != nil {
		log.Errorf("更新秒杀记录失败: %v", err)
		return err
	}
	return nil
}

// Get 获取单个秒杀记录
func (fsrd *flashSaleRecordData) Get(ctx context.Context, db *gorm.DB, id int64) (*do.FlashSaleRecordDO, error) {
	if db == nil {
		db = fsrd.db
	}
	
	var record do.FlashSaleRecordDO
	if err := db.WithContext(ctx).Where("id = ?", id).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		log.Errorf("获取秒杀记录失败: %v", err)
		return nil, err
	}
	return &record, nil
}

// GetByFlashSaleAndUser 根据秒杀活动和用户获取记录
func (fsrd *flashSaleRecordData) GetByFlashSaleAndUser(ctx context.Context, db *gorm.DB, flashSaleID int64, userID int64) (*do.FlashSaleRecordDO, error) {
	if db == nil {
		db = fsrd.db
	}
	
	var record do.FlashSaleRecordDO
	if err := db.WithContext(ctx).Where("flash_sale_id = ? AND user_id = ?", flashSaleID, userID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		log.Errorf("根据秒杀活动和用户获取记录失败: %v", err)
		return nil, err
	}
	return &record, nil
}

// GetUserRecords 获取用户秒杀记录列表
func (fsrd *flashSaleRecordData) GetUserRecords(ctx context.Context, db *gorm.DB, userID int64, meta v1.ListMeta) (*do.FlashSaleRecordDOList, error) {
	if db == nil {
		db = fsrd.db
	}
	
	query := db.WithContext(ctx).Model(&do.FlashSaleRecordDO{}).Where("user_id = ?", userID)
	
	// 计算总数
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		log.Errorf("统计用户秒杀记录总数失败: %v", err)
		return nil, err
	}
	
	// 应用分页
	if meta.Page > 0 {
		query = query.Offset((meta.Page - 1) * meta.PageSize)
	}
	if meta.PageSize > 0 {
		query = query.Limit(meta.PageSize)
	}
	
	query = query.Order("created_at DESC")
	
	var records []*do.FlashSaleRecordDO
	if err := query.Find(&records).Error; err != nil {
		log.Errorf("查询用户秒杀记录列表失败: %v", err)
		return nil, err
	}
	
	return &do.FlashSaleRecordDOList{
		TotalCount: totalCount,
		Items:      records,
	}, nil
}

// GetFlashSaleRecords 获取秒杀活动记录列表
func (fsrd *flashSaleRecordData) GetFlashSaleRecords(ctx context.Context, db *gorm.DB, flashSaleID int64, meta v1.ListMeta) (*do.FlashSaleRecordDOList, error) {
	if db == nil {
		db = fsrd.db
	}
	
	query := db.WithContext(ctx).Model(&do.FlashSaleRecordDO{}).Where("flash_sale_id = ?", flashSaleID)
	
	// 计算总数
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		log.Errorf("统计秒杀活动记录总数失败: %v", err)
		return nil, err
	}
	
	// 应用分页
	if meta.Page > 0 {
		query = query.Offset((meta.Page - 1) * meta.PageSize)
	}
	if meta.PageSize > 0 {
		query = query.Limit(meta.PageSize)
	}
	
	query = query.Order("created_at DESC")
	
	var records []*do.FlashSaleRecordDO
	if err := query.Find(&records).Error; err != nil {
		log.Errorf("查询秒杀活动记录列表失败: %v", err)
		return nil, err
	}
	
	return &do.FlashSaleRecordDOList{
		TotalCount: totalCount,
		Items:      records,
	}, nil
}

// GetUserFlashSaleHistory 获取用户在特定秒杀活动的历史记录
func (fsrd *flashSaleRecordData) GetUserFlashSaleHistory(ctx context.Context, db *gorm.DB, userID int64, flashSaleID int64) ([]*do.FlashSaleRecordDO, error) {
	if db == nil {
		db = fsrd.db
	}
	
	var records []*do.FlashSaleRecordDO
	if err := db.WithContext(ctx).
		Where("user_id = ? AND flash_sale_id = ?", userID, flashSaleID).
		Order("created_at DESC").
		Find(&records).Error; err != nil {
		log.Errorf("获取用户秒杀历史记录失败: %v", err)
		return nil, err
	}
	return records, nil
}

// CountUserParticipation 统计用户参与次数
func (fsrd *flashSaleRecordData) CountUserParticipation(ctx context.Context, db *gorm.DB, userID int64, flashSaleID int64) (int64, error) {
	if db == nil {
		db = fsrd.db
	}
	
	var count int64
	if err := db.WithContext(ctx).Model(&do.FlashSaleRecordDO{}).
		Where("user_id = ? AND flash_sale_id = ?", userID, flashSaleID).
		Count(&count).Error; err != nil {
		log.Errorf("统计用户参与次数失败: %v", err)
		return 0, err
	}
	return count, nil
}

// CountSuccessfulParticipation 统计成功参与次数
func (fsrd *flashSaleRecordData) CountSuccessfulParticipation(ctx context.Context, db *gorm.DB, flashSaleID int64) (int64, error) {
	if db == nil {
		db = fsrd.db
	}
	
	var count int64
	if err := db.WithContext(ctx).Model(&do.FlashSaleRecordDO{}).
		Where("flash_sale_id = ? AND status = ?", flashSaleID, do.FlashSaleRecordStatusSuccess).
		Count(&count).Error; err != nil {
		log.Errorf("统计成功参与次数失败: %v", err)
		return 0, err
	}
	return count, nil
}

// GetFlashSaleStatistics 获取秒杀统计信息
func (fsrd *flashSaleRecordData) GetFlashSaleStatistics(ctx context.Context, db *gorm.DB, flashSaleID int64) (*do.FlashSaleStatistics, error) {
	if db == nil {
		db = fsrd.db
	}
	
	var stats struct {
		TotalParticipants   int64 `json:"total_participants"`
		SuccessParticipants int64 `json:"success_participants"`
	}
	
	if err := db.WithContext(ctx).Model(&do.FlashSaleRecordDO{}).
		Select("COUNT(DISTINCT user_id) as total_participants, COUNT(CASE WHEN status = ? THEN 1 END) as success_participants", do.FlashSaleRecordStatusSuccess).
		Where("flash_sale_id = ?", flashSaleID).
		Find(&stats).Error; err != nil {
		log.Errorf("获取秒杀统计信息失败: %v", err)
		return nil, err
	}
	
	successRate := float64(0)
	if stats.TotalParticipants > 0 {
		successRate = float64(stats.SuccessParticipants) / float64(stats.TotalParticipants) * 100
	}
	
	return &do.FlashSaleStatistics{
		FlashSaleID:           flashSaleID,
		TotalParticipants:     stats.TotalParticipants,
		SuccessParticipants:   stats.SuccessParticipants,
		SuccessRate:           successRate,
		PeakQPS:               0, // 需要从监控系统获取
		AverageResponseTime:   0, // 需要从监控系统获取
	}, nil
}

// UpdateStatus 更新秒杀记录状态
func (fsrd *flashSaleRecordData) UpdateStatus(ctx context.Context, db *gorm.DB, id int64, status do.FlashSaleRecordStatus) error {
	if db == nil {
		db = fsrd.db
	}
	
	if err := db.WithContext(ctx).Model(&do.FlashSaleRecordDO{}).
		Where("id = ?", id).Update("status", status).Error; err != nil {
		log.Errorf("更新秒杀记录状态失败: %v", err)
		return err
	}
	return nil
}

// UpdateUserCouponID 更新用户优惠券ID
func (fsrd *flashSaleRecordData) UpdateUserCouponID(ctx context.Context, db *gorm.DB, id int64, userCouponID int64) error {
	if db == nil {
		db = fsrd.db
	}
	
	if err := db.WithContext(ctx).Model(&do.FlashSaleRecordDO{}).
		Where("id = ?", id).Update("user_coupon_id", userCouponID).Error; err != nil {
		log.Errorf("更新秒杀记录用户优惠券ID失败: %v", err)
		return err
	}
	return nil
}

// BatchCreate 批量创建秒杀记录
func (fsrd *flashSaleRecordData) BatchCreate(ctx context.Context, db *gorm.DB, records []*do.FlashSaleRecordDO) error {
	if db == nil {
		db = fsrd.db
	}
	
	if err := db.WithContext(ctx).CreateInBatches(records, 100).Error; err != nil {
		log.Errorf("批量创建秒杀记录失败: %v", err)
		return err
	}
	return nil
}