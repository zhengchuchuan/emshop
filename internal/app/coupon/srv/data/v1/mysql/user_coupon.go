package mysql

import (
	"context"
	"time"
	"emshop/internal/app/coupon/srv/domain/do"
	v1 "emshop/pkg/common/meta/v1"
	"emshop/pkg/log"
	"gorm.io/gorm"
)

type userCouponData struct {
	db *gorm.DB
}

// NewUserCouponData 创建用户优惠券数据访问对象
func NewUserCouponData(db *gorm.DB) *userCouponData {
	return &userCouponData{
		db: db,
	}
}

// Create 创建用户优惠券
func (ucd *userCouponData) Create(ctx context.Context, db *gorm.DB, userCoupon *do.UserCouponDO) error {
	if db == nil {
		db = ucd.db
	}
	
	if err := db.WithContext(ctx).Create(userCoupon).Error; err != nil {
		log.Errorf("创建用户优惠券失败: %v", err)
		return err
	}
	return nil
}

// Update 更新用户优惠券
func (ucd *userCouponData) Update(ctx context.Context, db *gorm.DB, userCoupon *do.UserCouponDO) error {
	if db == nil {
		db = ucd.db
	}
	
	if err := db.WithContext(ctx).Save(userCoupon).Error; err != nil {
		log.Errorf("更新用户优惠券失败: %v", err)
		return err
	}
	return nil
}

// Delete 删除用户优惠券
func (ucd *userCouponData) Delete(ctx context.Context, db *gorm.DB, id int64) error {
	if db == nil {
		db = ucd.db
	}
	
	if err := db.WithContext(ctx).Where("id = ?", id).Delete(&do.UserCouponDO{}).Error; err != nil {
		log.Errorf("删除用户优惠券失败: %v", err)
		return err
	}
	return nil
}

// Get 获取单个用户优惠券
func (ucd *userCouponData) Get(ctx context.Context, db *gorm.DB, id int64) (*do.UserCouponDO, error) {
	if db == nil {
		db = ucd.db
	}
	
	var userCoupon do.UserCouponDO
	if err := db.WithContext(ctx).Where("id = ?", id).First(&userCoupon).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		log.Errorf("获取用户优惠券失败: %v", err)
		return nil, err
	}
	return &userCoupon, nil
}

// GetByCouponCode 根据优惠券码获取用户优惠券
func (ucd *userCouponData) GetByCouponCode(ctx context.Context, db *gorm.DB, couponCode string) (*do.UserCouponDO, error) {
	if db == nil {
		db = ucd.db
	}
	
	var userCoupon do.UserCouponDO
	if err := db.WithContext(ctx).Where("coupon_code = ?", couponCode).First(&userCoupon).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		log.Errorf("根据优惠券码获取用户优惠券失败: %v", err)
		return nil, err
	}
	return &userCoupon, nil
}

// GetUserCoupons 获取用户优惠券列表
func (ucd *userCouponData) GetUserCoupons(ctx context.Context, db *gorm.DB, userID int64, status do.UserCouponStatus, meta v1.ListMeta) (*do.UserCouponDOList, error) {
	if db == nil {
		db = ucd.db
	}
	
	query := db.WithContext(ctx).Model(&do.UserCouponDO{}).Where("user_id = ?", userID)
	
	if status > 0 {
		query = query.Where("status = ?", status)
	}
	
	// 计算总数
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		log.Errorf("统计用户优惠券总数失败: %v", err)
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
	
	var userCoupons []*do.UserCouponDO
	if err := query.Find(&userCoupons).Error; err != nil {
		log.Errorf("查询用户优惠券列表失败: %v", err)
		return nil, err
	}
	
	return &do.UserCouponDOList{
		TotalCount: totalCount,
		Items:      userCoupons,
	}, nil
}

// GetUserAvailableCoupons 获取用户可用优惠券
func (ucd *userCouponData) GetUserAvailableCoupons(ctx context.Context, db *gorm.DB, userID int64, orderAmount float64, currentTime time.Time) ([]*do.UserCouponDO, error) {
	if db == nil {
		db = ucd.db
	}
	
	var userCoupons []*do.UserCouponDO
	if err := db.WithContext(ctx).
		Joins("JOIN coupon_templates ON user_coupons.coupon_template_id = coupon_templates.id").
		Where("user_coupons.user_id = ? AND user_coupons.status = ? AND user_coupons.expired_at > ? AND coupon_templates.min_order_amount <= ?",
			userID, do.UserCouponStatusUnused, currentTime, orderAmount).
		Find(&userCoupons).Error; err != nil {
		log.Errorf("查询用户可用优惠券失败: %v", err)
		return nil, err
	}
	return userCoupons, nil
}

// GetUserCouponsByTemplate 根据模板获取用户优惠券
func (ucd *userCouponData) GetUserCouponsByTemplate(ctx context.Context, db *gorm.DB, userID int64, templateID int64) ([]*do.UserCouponDO, error) {
	if db == nil {
		db = ucd.db
	}
	
	var userCoupons []*do.UserCouponDO
	if err := db.WithContext(ctx).
		Where("user_id = ? AND coupon_template_id = ?", userID, templateID).
		Find(&userCoupons).Error; err != nil {
		log.Errorf("根据模板查询用户优惠券失败: %v", err)
		return nil, err
	}
	return userCoupons, nil
}

// UpdateStatus 更新用户优惠券状态
func (ucd *userCouponData) UpdateStatus(ctx context.Context, db *gorm.DB, id int64, status do.UserCouponStatus) error {
	if db == nil {
		db = ucd.db
	}
	
	if err := db.WithContext(ctx).Model(&do.UserCouponDO{}).
		Where("id = ?", id).Update("status", status).Error; err != nil {
		log.Errorf("更新用户优惠券状态失败: %v", err)
		return err
	}
	return nil
}

// UseCoupon 使用优惠券
func (ucd *userCouponData) UseCoupon(ctx context.Context, db *gorm.DB, id int64, orderSn string, usedAt time.Time) error {
	if db == nil {
		db = ucd.db
	}
	
	if err := db.WithContext(ctx).Model(&do.UserCouponDO{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":   do.UserCouponStatusUsed,
			"order_sn": orderSn,
			"used_at":  usedAt,
		}).Error; err != nil {
		log.Errorf("使用优惠券失败: %v", err)
		return err
	}
	return nil
}

// BatchCreate 批量创建用户优惠券
func (ucd *userCouponData) BatchCreate(ctx context.Context, db *gorm.DB, userCoupons []*do.UserCouponDO) error {
	if db == nil {
		db = ucd.db
	}
	
	if err := db.WithContext(ctx).CreateInBatches(userCoupons, 100).Error; err != nil {
		log.Errorf("批量创建用户优惠券失败: %v", err)
		return err
	}
	return nil
}

// CountUserCouponsByTemplate 统计用户某模板的优惠券数量
func (ucd *userCouponData) CountUserCouponsByTemplate(ctx context.Context, db *gorm.DB, userID int64, templateID int64) (int64, error) {
	if db == nil {
		db = ucd.db
	}
	
	var count int64
	if err := db.WithContext(ctx).Model(&do.UserCouponDO{}).
		Where("user_id = ? AND coupon_template_id = ?", userID, templateID).
		Count(&count).Error; err != nil {
		log.Errorf("统计用户优惠券模板数量失败: %v", err)
		return 0, err
	}
	return count, nil
}

// FindExpiredCoupons 查找过期的优惠券
func (ucd *userCouponData) FindExpiredCoupons(ctx context.Context, db *gorm.DB, beforeTime time.Time) ([]*do.UserCouponDO, error) {
	if db == nil {
		db = ucd.db
	}
	
	var userCoupons []*do.UserCouponDO
	if err := db.WithContext(ctx).
		Where("status = ? AND expired_at < ?", do.UserCouponStatusUnused, beforeTime).
		Find(&userCoupons).Error; err != nil {
		log.Errorf("查找过期优惠券失败: %v", err)
		return nil, err
	}
	return userCoupons, nil
}