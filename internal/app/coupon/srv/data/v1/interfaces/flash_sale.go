package interfaces

import (
	"context"
	"time"
	"emshop/internal/app/coupon/srv/domain/do"
	v1 "emshop/pkg/common/meta/v1"
	"gorm.io/gorm"
)

// FlashSaleDataInterface 秒杀活动数据接口
type FlashSaleDataInterface interface {
	// 基础CRUD操作
	Create(ctx context.Context, db *gorm.DB, activity *do.FlashSaleActivityDO) error
	Update(ctx context.Context, db *gorm.DB, activity *do.FlashSaleActivityDO) error
	Delete(ctx context.Context, db *gorm.DB, id int64) error
	Get(ctx context.Context, db *gorm.DB, id int64) (*do.FlashSaleActivityDO, error)
	List(ctx context.Context, db *gorm.DB, status do.FlashSaleStatus, meta v1.ListMeta, orderby []string) (*do.FlashSaleActivityDOList, error)
	
	// 业务查询操作
	GetByStatus(ctx context.Context, db *gorm.DB, status do.FlashSaleStatus) ([]*do.FlashSaleActivityDO, error)
	GetActiveActivities(ctx context.Context, db *gorm.DB, currentTime time.Time) ([]*do.FlashSaleActivityDO, error)
	GetUpcomingActivities(ctx context.Context, db *gorm.DB, currentTime time.Time, limit int) ([]*do.FlashSaleActivityDO, error)
	GetByCouponTemplate(ctx context.Context, db *gorm.DB, templateID int64) ([]*do.FlashSaleActivityDO, error)
	
	// 库存和状态更新操作
	UpdateSoldCount(ctx context.Context, db *gorm.DB, id int64, increment int32) error
	UpdateStatus(ctx context.Context, db *gorm.DB, id int64, status do.FlashSaleStatus) error
	CheckStock(ctx context.Context, db *gorm.DB, id int64) (*do.FlashSaleStockInfo, error)
	IncrementSoldCount(ctx context.Context, db *gorm.DB, id int64) error
	DecrementSoldCount(ctx context.Context, db *gorm.DB, id int64) error
}

// FlashSaleRecordDataInterface 秒杀记录数据接口
type FlashSaleRecordDataInterface interface {
	// 基础CRUD操作
	Create(ctx context.Context, db *gorm.DB, record *do.FlashSaleRecordDO) error
	Update(ctx context.Context, db *gorm.DB, record *do.FlashSaleRecordDO) error
	Get(ctx context.Context, db *gorm.DB, id int64) (*do.FlashSaleRecordDO, error)
	GetByFlashSaleAndUser(ctx context.Context, db *gorm.DB, flashSaleID int64, userID int64) (*do.FlashSaleRecordDO, error)
	
	// 查询操作
	GetUserRecords(ctx context.Context, db *gorm.DB, userID int64, meta v1.ListMeta) (*do.FlashSaleRecordDOList, error)
	GetFlashSaleRecords(ctx context.Context, db *gorm.DB, flashSaleID int64, meta v1.ListMeta) (*do.FlashSaleRecordDOList, error)
	GetUserFlashSaleHistory(ctx context.Context, db *gorm.DB, userID int64, flashSaleID int64) ([]*do.FlashSaleRecordDO, error)
	
	// 统计操作
	CountUserParticipation(ctx context.Context, db *gorm.DB, userID int64, flashSaleID int64) (int64, error)
	CountSuccessfulParticipation(ctx context.Context, db *gorm.DB, flashSaleID int64) (int64, error)
	GetFlashSaleStatistics(ctx context.Context, db *gorm.DB, flashSaleID int64) (*do.FlashSaleStatistics, error)
	
	// 业务操作
	UpdateStatus(ctx context.Context, db *gorm.DB, id int64, status do.FlashSaleRecordStatus) error
	UpdateUserCouponID(ctx context.Context, db *gorm.DB, id int64, userCouponID int64) error
	BatchCreate(ctx context.Context, db *gorm.DB, records []*do.FlashSaleRecordDO) error
}