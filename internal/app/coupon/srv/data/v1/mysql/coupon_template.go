package mysql

import (
	"context"
	"time"
	"emshop/internal/app/coupon/srv/domain/do"
	v1 "emshop/pkg/common/meta/v1"
	"emshop/pkg/log"
	"gorm.io/gorm"
)

type couponTemplateData struct {
	db *gorm.DB
}

// NewCouponTemplateData 创建优惠券模板数据访问对象
func NewCouponTemplateData(db *gorm.DB) *couponTemplateData {
	return &couponTemplateData{
		db: db,
	}
}

// Create 创建优惠券模板
func (ctd *couponTemplateData) Create(ctx context.Context, db *gorm.DB, template *do.CouponTemplateDO) error {
	if db == nil {
		db = ctd.db
	}
	
	if err := db.WithContext(ctx).Create(template).Error; err != nil {
		log.Errorf("创建优惠券模板失败: %v", err)
		return err
	}
	return nil
}

// Update 更新优惠券模板
func (ctd *couponTemplateData) Update(ctx context.Context, db *gorm.DB, template *do.CouponTemplateDO) error {
	if db == nil {
		db = ctd.db
	}
	
	if err := db.WithContext(ctx).Save(template).Error; err != nil {
		log.Errorf("更新优惠券模板失败: %v", err)
		return err
	}
	return nil
}

// Delete 删除优惠券模板
func (ctd *couponTemplateData) Delete(ctx context.Context, db *gorm.DB, id int64) error {
	if db == nil {
		db = ctd.db
	}
	
	if err := db.WithContext(ctx).Where("id = ?", id).Delete(&do.CouponTemplateDO{}).Error; err != nil {
		log.Errorf("删除优惠券模板失败: %v", err)
		return err
	}
	return nil
}

// Get 获取单个优惠券模板
func (ctd *couponTemplateData) Get(ctx context.Context, db *gorm.DB, id int64) (*do.CouponTemplateDO, error) {
	if db == nil {
		db = ctd.db
	}
	
	var template do.CouponTemplateDO
	if err := db.WithContext(ctx).Where("id = ?", id).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		log.Errorf("获取优惠券模板失败: %v", err)
		return nil, err
	}
	return &template, nil
}

// List 获取优惠券模板列表
func (ctd *couponTemplateData) List(ctx context.Context, db *gorm.DB, status do.CouponStatus, meta v1.ListMeta, orderby []string) (*do.CouponTemplateDOList, error) {
	if db == nil {
		db = ctd.db
	}
	
	query := db.WithContext(ctx).Model(&do.CouponTemplateDO{})
	
	if status > 0 {
		query = query.Where("status = ?", status)
	}
	
	// 计算总数
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		log.Errorf("统计优惠券模板总数失败: %v", err)
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
	
	var templates []*do.CouponTemplateDO
	if err := query.Find(&templates).Error; err != nil {
		log.Errorf("查询优惠券模板列表失败: %v", err)
		return nil, err
	}
	
	return &do.CouponTemplateDOList{
		TotalCount: totalCount,
		Items:      templates,
	}, nil
}

// GetByType 根据类型获取优惠券模板
func (ctd *couponTemplateData) GetByType(ctx context.Context, db *gorm.DB, couponType do.CouponType) ([]*do.CouponTemplateDO, error) {
	if db == nil {
		db = ctd.db
	}
	
	var templates []*do.CouponTemplateDO
	if err := db.WithContext(ctx).Where("type = ? AND status = ?", couponType, do.CouponStatusActive).Find(&templates).Error; err != nil {
		log.Errorf("根据类型查询优惠券模板失败: %v", err)
		return nil, err
	}
	return templates, nil
}

// GetActiveTemplates 获取当前有效的优惠券模板
func (ctd *couponTemplateData) GetActiveTemplates(ctx context.Context, db *gorm.DB, currentTime time.Time) ([]*do.CouponTemplateDO, error) {
	if db == nil {
		db = ctd.db
	}
	
	var templates []*do.CouponTemplateDO
	if err := db.WithContext(ctx).Where("status = ? AND valid_start_time <= ? AND valid_end_time >= ?", 
		do.CouponStatusActive, currentTime, currentTime).Find(&templates).Error; err != nil {
		log.Errorf("查询当前有效优惠券模板失败: %v", err)
		return nil, err
	}
	return templates, nil
}

// UpdateUsedCount 更新已使用数量
func (ctd *couponTemplateData) UpdateUsedCount(ctx context.Context, db *gorm.DB, templateID int64, increment int32) error {
	if db == nil {
		db = ctd.db
	}
	
	if err := db.WithContext(ctx).Model(&do.CouponTemplateDO{}).
		Where("id = ?", templateID).
		UpdateColumn("used_count", gorm.Expr("used_count + ?", increment)).Error; err != nil {
		log.Errorf("更新优惠券模板已使用数量失败: %v", err)
		return err
	}
	return nil
}

// GetAvailableTemplates 获取用户可领取的优惠券模板
func (ctd *couponTemplateData) GetAvailableTemplates(ctx context.Context, db *gorm.DB, userID int64, currentTime time.Time) ([]*do.CouponTemplateDO, error) {
	if db == nil {
		db = ctd.db
	}
	
	var templates []*do.CouponTemplateDO
	if err := db.WithContext(ctx).Where(`
		status = ? AND 
		valid_start_time <= ? AND 
		valid_end_time >= ? AND 
		used_count < total_count AND
		id NOT IN (
			SELECT coupon_template_id FROM user_coupons 
			WHERE user_id = ? 
			GROUP BY coupon_template_id 
			HAVING COUNT(*) >= (SELECT per_user_limit FROM coupon_templates WHERE id = coupon_template_id)
		)
	`, do.CouponStatusActive, currentTime, currentTime, userID).Find(&templates).Error; err != nil {
		log.Errorf("查询用户可领取优惠券模板失败: %v", err)
		return nil, err
	}
	return templates, nil
}

// CheckTemplateAvailability 检查优惠券模板是否可用
func (ctd *couponTemplateData) CheckTemplateAvailability(ctx context.Context, db *gorm.DB, templateID int64, currentTime time.Time) (bool, error) {
	if db == nil {
		db = ctd.db
	}
	
	var count int64
    if err := db.WithContext(ctx).Model(&do.CouponTemplateDO{}).
        Where("id = ? AND status = ? AND valid_start_time <= ? AND valid_end_time >= ? AND (total_count = 0 OR used_count < total_count)",
        templateID, do.CouponStatusActive, currentTime, currentTime).Count(&count).Error; err != nil {
        log.Errorf("检查优惠券模板可用性失败: %v", err)
        return false, err
    }
	return count > 0, nil
}
