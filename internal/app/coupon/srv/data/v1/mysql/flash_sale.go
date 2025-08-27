package mysql

import (
	"context"
	"time"
	"emshop/internal/app/coupon/srv/domain/do"
	v1 "emshop/pkg/common/meta/v1"
	"emshop/pkg/log"
	"gorm.io/gorm"
)

type flashSaleData struct {
	db *gorm.DB
}

// NewFlashSaleData 创建秒杀活动数据访问对象
func NewFlashSaleData(db *gorm.DB) *flashSaleData {
	return &flashSaleData{
		db: db,
	}
}

// Create 创建秒杀活动
func (fsd *flashSaleData) Create(ctx context.Context, db *gorm.DB, activity *do.FlashSaleActivityDO) error {
	if db == nil {
		db = fsd.db
	}
	
	if err := db.WithContext(ctx).Create(activity).Error; err != nil {
		log.Errorf("创建秒杀活动失败: %v", err)
		return err
	}
	return nil
}

// Update 更新秒杀活动
func (fsd *flashSaleData) Update(ctx context.Context, db *gorm.DB, activity *do.FlashSaleActivityDO) error {
	if db == nil {
		db = fsd.db
	}
	
	if err := db.WithContext(ctx).Save(activity).Error; err != nil {
		log.Errorf("更新秒杀活动失败: %v", err)
		return err
	}
	return nil
}

// Delete 删除秒杀活动
func (fsd *flashSaleData) Delete(ctx context.Context, db *gorm.DB, id int64) error {
	if db == nil {
		db = fsd.db
	}
	
	if err := db.WithContext(ctx).Where("id = ?", id).Delete(&do.FlashSaleActivityDO{}).Error; err != nil {
		log.Errorf("删除秒杀活动失败: %v", err)
		return err
	}
	return nil
}

// Get 获取单个秒杀活动
func (fsd *flashSaleData) Get(ctx context.Context, db *gorm.DB, id int64) (*do.FlashSaleActivityDO, error) {
	if db == nil {
		db = fsd.db
	}
	
	var activity do.FlashSaleActivityDO
	if err := db.WithContext(ctx).Where("id = ?", id).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		log.Errorf("获取秒杀活动失败: %v", err)
		return nil, err
	}
	return &activity, nil
}

// List 获取秒杀活动列表
func (fsd *flashSaleData) List(ctx context.Context, db *gorm.DB, status do.FlashSaleStatus, meta v1.ListMeta, orderby []string) (*do.FlashSaleActivityDOList, error) {
	if db == nil {
		db = fsd.db
	}
	
	query := db.WithContext(ctx).Model(&do.FlashSaleActivityDO{})
	
	if status > 0 {
		query = query.Where("status = ?", status)
	}
	
	// 计算总数
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		log.Errorf("统计秒杀活动总数失败: %v", err)
		return nil, err
	}
	
	// 应用分页
	if meta.Page > 0 {
		query = query.Offset((meta.Page - 1) * meta.PageSize)
	}
	if meta.PageSize > 0 {
		query = query.Limit(meta.PageSize)
	}
	
	// 应用排序
	for _, order := range orderby {
		query = query.Order(order)
	}
	
	var activities []*do.FlashSaleActivityDO
	if err := query.Find(&activities).Error; err != nil {
		log.Errorf("查询秒杀活动列表失败: %v", err)
		return nil, err
	}
	
	return &do.FlashSaleActivityDOList{
		TotalCount: totalCount,
		Items:      activities,
	}, nil
}

// GetByStatus 根据状态获取秒杀活动
func (fsd *flashSaleData) GetByStatus(ctx context.Context, db *gorm.DB, status do.FlashSaleStatus) ([]*do.FlashSaleActivityDO, error) {
	if db == nil {
		db = fsd.db
	}
	
	var activities []*do.FlashSaleActivityDO
	if err := db.WithContext(ctx).Where("status = ?", status).Find(&activities).Error; err != nil {
		log.Errorf("根据状态查询秒杀活动失败: %v", err)
		return nil, err
	}
	return activities, nil
}

// GetActiveActivities 获取当前进行中的秒杀活动
func (fsd *flashSaleData) GetActiveActivities(ctx context.Context, db *gorm.DB, currentTime time.Time) ([]*do.FlashSaleActivityDO, error) {
	if db == nil {
		db = fsd.db
	}
	
	var activities []*do.FlashSaleActivityDO
	if err := db.WithContext(ctx).Where("status = ? AND start_time <= ? AND end_time >= ?", 
		do.FlashSaleStatusActive, currentTime, currentTime).Find(&activities).Error; err != nil {
		log.Errorf("查询当前进行中的秒杀活动失败: %v", err)
		return nil, err
	}
	return activities, nil
}

// GetUpcomingActivities 获取即将开始的秒杀活动
func (fsd *flashSaleData) GetUpcomingActivities(ctx context.Context, db *gorm.DB, currentTime time.Time, limit int) ([]*do.FlashSaleActivityDO, error) {
	if db == nil {
		db = fsd.db
	}
	
	var activities []*do.FlashSaleActivityDO
	query := db.WithContext(ctx).Where("status = ? AND start_time > ?", do.FlashSaleStatusPending, currentTime).
		Order("start_time ASC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Find(&activities).Error; err != nil {
		log.Errorf("查询即将开始的秒杀活动失败: %v", err)
		return nil, err
	}
	return activities, nil
}

// GetByCouponTemplate 根据优惠券模板获取秒杀活动
func (fsd *flashSaleData) GetByCouponTemplate(ctx context.Context, db *gorm.DB, templateID int64) ([]*do.FlashSaleActivityDO, error) {
	if db == nil {
		db = fsd.db
	}
	
	var activities []*do.FlashSaleActivityDO
	if err := db.WithContext(ctx).Where("coupon_template_id = ?", templateID).Find(&activities).Error; err != nil {
		log.Errorf("根据优惠券模板查询秒杀活动失败: %v", err)
		return nil, err
	}
	return activities, nil
}

// UpdateSoldCount 更新已售数量
func (fsd *flashSaleData) UpdateSoldCount(ctx context.Context, db *gorm.DB, id int64, increment int32) error {
	if db == nil {
		db = fsd.db
	}
	
	if err := db.WithContext(ctx).Model(&do.FlashSaleActivityDO{}).
		Where("id = ?", id).
		UpdateColumn("sold_count", gorm.Expr("sold_count + ?", increment)).Error; err != nil {
		log.Errorf("更新秒杀活动已售数量失败: %v", err)
		return err
	}
	return nil
}

// UpdateStatus 更新秒杀活动状态
func (fsd *flashSaleData) UpdateStatus(ctx context.Context, db *gorm.DB, id int64, status do.FlashSaleStatus) error {
	if db == nil {
		db = fsd.db
	}
	
	if err := db.WithContext(ctx).Model(&do.FlashSaleActivityDO{}).
		Where("id = ?", id).Update("status", status).Error; err != nil {
		log.Errorf("更新秒杀活动状态失败: %v", err)
		return err
	}
	return nil
}

// CheckStock 检查库存信息
func (fsd *flashSaleData) CheckStock(ctx context.Context, db *gorm.DB, id int64) (*do.FlashSaleStockInfo, error) {
	if db == nil {
		db = fsd.db
	}
	
	var activity do.FlashSaleActivityDO
	if err := db.WithContext(ctx).Where("id = ?", id).First(&activity).Error; err != nil {
		log.Errorf("查询秒杀活动库存失败: %v", err)
		return nil, err
	}
	
	return &do.FlashSaleStockInfo{
		FlashSaleID:    activity.ID,
		TotalStock:     activity.FlashSaleCount,
		RemainingStock: activity.FlashSaleCount - activity.SoldCount,
		SoldCount:      activity.SoldCount,
	}, nil
}

// IncrementSoldCount 增加已售数量
func (fsd *flashSaleData) IncrementSoldCount(ctx context.Context, db *gorm.DB, id int64) error {
	return fsd.UpdateSoldCount(ctx, db, id, 1)
}

// DecrementSoldCount 减少已售数量
func (fsd *flashSaleData) DecrementSoldCount(ctx context.Context, db *gorm.DB, id int64) error {
	return fsd.UpdateSoldCount(ctx, db, id, -1)
}