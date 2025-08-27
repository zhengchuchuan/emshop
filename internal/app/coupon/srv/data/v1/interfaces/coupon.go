package interfaces

import (
	"context"
	"time"
	"emshop/internal/app/coupon/srv/domain/do"
	v1 "emshop/pkg/common/meta/v1"
	"gorm.io/gorm"
)

// CouponTemplateDataInterface 优惠券模板数据接口
type CouponTemplateDataInterface interface {
	// 基础CRUD操作
	Create(ctx context.Context, db *gorm.DB, template *do.CouponTemplateDO) error
	Update(ctx context.Context, db *gorm.DB, template *do.CouponTemplateDO) error
	Delete(ctx context.Context, db *gorm.DB, id int64) error
	Get(ctx context.Context, db *gorm.DB, id int64) (*do.CouponTemplateDO, error)
	List(ctx context.Context, db *gorm.DB, status do.CouponStatus, meta v1.ListMeta, orderby []string) (*do.CouponTemplateDOList, error)
	
	// 业务操作
	GetByType(ctx context.Context, db *gorm.DB, couponType do.CouponType) ([]*do.CouponTemplateDO, error)
	GetActiveTemplates(ctx context.Context, db *gorm.DB, currentTime time.Time) ([]*do.CouponTemplateDO, error)
	UpdateUsedCount(ctx context.Context, db *gorm.DB, templateID int64, increment int32) error
	GetAvailableTemplates(ctx context.Context, db *gorm.DB, userID int64, currentTime time.Time) ([]*do.CouponTemplateDO, error)
	CheckTemplateAvailability(ctx context.Context, db *gorm.DB, templateID int64, currentTime time.Time) (bool, error)
}

// UserCouponDataInterface 用户优惠券数据接口
type UserCouponDataInterface interface {
	// 基础CRUD操作
	Create(ctx context.Context, db *gorm.DB, userCoupon *do.UserCouponDO) error
	Update(ctx context.Context, db *gorm.DB, userCoupon *do.UserCouponDO) error
	Delete(ctx context.Context, db *gorm.DB, id int64) error
	Get(ctx context.Context, db *gorm.DB, id int64) (*do.UserCouponDO, error)
	GetByCouponCode(ctx context.Context, db *gorm.DB, couponCode string) (*do.UserCouponDO, error)
	
	// 用户优惠券查询
	GetUserCoupons(ctx context.Context, db *gorm.DB, userID int64, status do.UserCouponStatus, meta v1.ListMeta) (*do.UserCouponDOList, error)
	GetUserAvailableCoupons(ctx context.Context, db *gorm.DB, userID int64, orderAmount float64, currentTime time.Time) ([]*do.UserCouponDO, error)
	GetUserCouponsByTemplate(ctx context.Context, db *gorm.DB, userID int64, templateID int64) ([]*do.UserCouponDO, error)
	
	// 业务操作
	UpdateStatus(ctx context.Context, db *gorm.DB, id int64, status do.UserCouponStatus) error
	UseCoupon(ctx context.Context, db *gorm.DB, id int64, orderSn string, usedAt time.Time) error
	BatchCreate(ctx context.Context, db *gorm.DB, userCoupons []*do.UserCouponDO) error
	CountUserCouponsByTemplate(ctx context.Context, db *gorm.DB, userID int64, templateID int64) (int64, error)
	FindExpiredCoupons(ctx context.Context, db *gorm.DB, beforeTime time.Time) ([]*do.UserCouponDO, error)
}

// CouponUsageLogDataInterface 优惠券使用记录数据接口
type CouponUsageLogDataInterface interface {
	Create(ctx context.Context, db *gorm.DB, log *do.CouponUsageLogDO) error
	GetByUserCoupon(ctx context.Context, db *gorm.DB, userCouponID int64) ([]*do.CouponUsageLogDO, error)
	GetByOrderSn(ctx context.Context, db *gorm.DB, orderSn string) ([]*do.CouponUsageLogDO, error)
	List(ctx context.Context, db *gorm.DB, userID int64, meta v1.ListMeta) (*do.CouponUsageLogDOList, error)
	GetUserUsageStats(ctx context.Context, db *gorm.DB, userID int64, startTime, endTime time.Time) (map[string]interface{}, error)
}

// CouponConfigDataInterface 优惠券配置数据接口
type CouponConfigDataInterface interface {
	Get(ctx context.Context, db *gorm.DB, configKey string) (*do.CouponConfigDO, error)
	Set(ctx context.Context, db *gorm.DB, configKey, configValue, description string) error
	GetAll(ctx context.Context, db *gorm.DB) ([]*do.CouponConfigDO, error)
	Update(ctx context.Context, db *gorm.DB, configKey, configValue string) error
	Delete(ctx context.Context, db *gorm.DB, configKey string) error
}

// DataFactory 优惠券服务数据工厂接口
type DataFactory interface {
	CouponTemplates() CouponTemplateDataInterface
	UserCoupons() UserCouponDataInterface
	CouponUsageLogs() CouponUsageLogDataInterface
	CouponConfigs() CouponConfigDataInterface
	FlashSales() FlashSaleDataInterface
	FlashSaleRecords() FlashSaleRecordDataInterface
	
	// 数据库操作
	DB() *gorm.DB
	Begin() *gorm.DB
	Close() error
}