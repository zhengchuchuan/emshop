package mysql

import (
	"context"
	"emshop/internal/app/coupon/srv/domain/do"
	"emshop/pkg/log"
	"gorm.io/gorm"
)

type couponConfigData struct {
	db *gorm.DB
}

// NewCouponConfigData 创建优惠券配置数据访问对象
func NewCouponConfigData(db *gorm.DB) *couponConfigData {
	return &couponConfigData{
		db: db,
	}
}

// Get 获取配置项
func (ccd *couponConfigData) Get(ctx context.Context, db *gorm.DB, configKey string) (*do.CouponConfigDO, error) {
	if db == nil {
		db = ccd.db
	}
	
	var config do.CouponConfigDO
	if err := db.WithContext(ctx).Where("config_key = ?", configKey).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		log.Errorf("获取优惠券配置失败: %v", err)
		return nil, err
	}
	return &config, nil
}

// Set 设置配置项
func (ccd *couponConfigData) Set(ctx context.Context, db *gorm.DB, configKey, configValue, description string) error {
	if db == nil {
		db = ccd.db
	}
	
	config := &do.CouponConfigDO{
		ConfigKey:   configKey,
		ConfigValue: configValue,
		Description: description,
	}
	
	// 使用ON DUPLICATE KEY UPDATE语义
	if err := db.WithContext(ctx).
		Where("config_key = ?", configKey).
		Assign(map[string]interface{}{
			"config_value": configValue,
			"description":  description,
		}).
		FirstOrCreate(config).Error; err != nil {
		log.Errorf("设置优惠券配置失败: %v", err)
		return err
	}
	return nil
}

// GetAll 获取所有配置项
func (ccd *couponConfigData) GetAll(ctx context.Context, db *gorm.DB) ([]*do.CouponConfigDO, error) {
	if db == nil {
		db = ccd.db
	}
	
	var configs []*do.CouponConfigDO
	if err := db.WithContext(ctx).Find(&configs).Error; err != nil {
		log.Errorf("获取所有优惠券配置失败: %v", err)
		return nil, err
	}
	return configs, nil
}

// Update 更新配置项
func (ccd *couponConfigData) Update(ctx context.Context, db *gorm.DB, configKey, configValue string) error {
	if db == nil {
		db = ccd.db
	}
	
	if err := db.WithContext(ctx).Model(&do.CouponConfigDO{}).
		Where("config_key = ?", configKey).
		Update("config_value", configValue).Error; err != nil {
		log.Errorf("更新优惠券配置失败: %v", err)
		return err
	}
	return nil
}

// Delete 删除配置项
func (ccd *couponConfigData) Delete(ctx context.Context, db *gorm.DB, configKey string) error {
	if db == nil {
		db = ccd.db
	}
	
	if err := db.WithContext(ctx).Where("config_key = ?", configKey).Delete(&do.CouponConfigDO{}).Error; err != nil {
		log.Errorf("删除优惠券配置失败: %v", err)
		return err
	}
	return nil
}