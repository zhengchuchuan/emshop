package mysql

import (
	"context"
	"time"
	"emshop/internal/app/coupon/srv/domain/do"
	v1 "emshop/pkg/common/meta/v1"
	pkglog "emshop/pkg/log"
	"gorm.io/gorm"
)

type couponUsageLogData struct {
	db *gorm.DB
}

// NewCouponUsageLogData 创建优惠券使用记录数据访问对象
func NewCouponUsageLogData(db *gorm.DB) *couponUsageLogData {
	return &couponUsageLogData{
		db: db,
	}
}

// Create 创建优惠券使用记录
func (culd *couponUsageLogData) Create(ctx context.Context, db *gorm.DB, log *do.CouponUsageLogDO) error {
	if db == nil {
		db = culd.db
	}
	
	if err := db.WithContext(ctx).Create(log).Error; err != nil {
		pkglog.Errorf("创建优惠券使用记录失败: %v", err)
		return err
	}
	return nil
}

// GetByUserCoupon 根据用户优惠券ID获取使用记录
func (culd *couponUsageLogData) GetByUserCoupon(ctx context.Context, db *gorm.DB, userCouponID int64) ([]*do.CouponUsageLogDO, error) {
	if db == nil {
		db = culd.db
	}
	
	var logs []*do.CouponUsageLogDO
	if err := db.WithContext(ctx).
		Where("user_coupon_id = ?", userCouponID).
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		pkglog.Errorf("根据用户优惠券ID查询使用记录失败: %v", err)
		return nil, err
	}
	return logs, nil
}

// GetByOrderSn 根据订单号获取优惠券使用记录
func (culd *couponUsageLogData) GetByOrderSn(ctx context.Context, db *gorm.DB, orderSn string) ([]*do.CouponUsageLogDO, error) {
	if db == nil {
		db = culd.db
	}
	
	var logs []*do.CouponUsageLogDO
	if err := db.WithContext(ctx).
		Where("order_sn = ?", orderSn).
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		pkglog.Errorf("根据订单号查询优惠券使用记录失败: %v", err)
		return nil, err
	}
	return logs, nil
}

// List 获取用户优惠券使用记录列表
func (culd *couponUsageLogData) List(ctx context.Context, db *gorm.DB, userID int64, meta v1.ListMeta) (*do.CouponUsageLogDOList, error) {
	if db == nil {
		db = culd.db
	}
	
	query := db.WithContext(ctx).Model(&do.CouponUsageLogDO{})
	
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	
	// 计算总数
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		pkglog.Errorf("统计优惠券使用记录总数失败: %v", err)
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
	
	var logs []*do.CouponUsageLogDO
	if err := query.Find(&logs).Error; err != nil {
		pkglog.Errorf("查询优惠券使用记录列表失败: %v", err)
		return nil, err
	}
	
	return &do.CouponUsageLogDOList{
		TotalCount: totalCount,
		Items:      logs,
	}, nil
}

// GetUserUsageStats 获取用户使用统计
func (culd *couponUsageLogData) GetUserUsageStats(ctx context.Context, db *gorm.DB, userID int64, startTime, endTime time.Time) (map[string]interface{}, error) {
	if db == nil {
		db = culd.db
	}
	
	var stats []struct {
		TotalCount      int64   `json:"total_count"`
		TotalDiscount   float64 `json:"total_discount"`
		AvgDiscount     float64 `json:"avg_discount"`
	}
	
	if err := db.WithContext(ctx).Model(&do.CouponUsageLogDO{}).
		Select("COUNT(*) as total_count, SUM(discount_amount) as total_discount, AVG(discount_amount) as avg_discount").
		Where("user_id = ? AND created_at BETWEEN ? AND ?", userID, startTime, endTime).
		Find(&stats).Error; err != nil {
		pkglog.Errorf("查询用户优惠券使用统计失败: %v", err)
		return nil, err
	}
	
	result := make(map[string]interface{})
	if len(stats) > 0 {
		result["total_count"] = stats[0].TotalCount
		result["total_discount"] = stats[0].TotalDiscount
		result["avg_discount"] = stats[0].AvgDiscount
	} else {
		result["total_count"] = int64(0)
		result["total_discount"] = float64(0)
		result["avg_discount"] = float64(0)
	}
	
	return result, nil
}