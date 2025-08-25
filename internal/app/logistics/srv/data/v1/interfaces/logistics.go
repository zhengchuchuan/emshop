package interfaces

import (
	"context"
	"emshop/internal/app/logistics/srv/domain/do"
	"time"
	"gorm.io/gorm"
)

// LogisticsOrdersRepo 物流订单数据访问接口
type LogisticsOrdersRepo interface {
	// 创建物流订单
	Create(ctx context.Context, db *gorm.DB, order *do.LogisticsOrderDO) error
	
	// 根据物流单号查询
	GetByLogisticsSn(ctx context.Context, db *gorm.DB, logisticsSn string) (*do.LogisticsOrderDO, error)
	
	// 根据订单号查询
	GetByOrderSn(ctx context.Context, db *gorm.DB, orderSn string) (*do.LogisticsOrderDO, error)
	
	// 根据快递单号查询
	GetByTrackingNumber(ctx context.Context, db *gorm.DB, trackingNumber string) (*do.LogisticsOrderDO, error)
	
	// 更新物流状态
	UpdateStatus(ctx context.Context, db *gorm.DB, logisticsSn string, status int32) error
	
	// 更新发货信息
	UpdateShipmentInfo(ctx context.Context, db *gorm.DB, logisticsSn string, shippedAt *time.Time) error
	
	// 更新签收信息
	UpdateDeliveryInfo(ctx context.Context, db *gorm.DB, logisticsSn string, deliveredAt *time.Time) error
	
	// 分页查询物流订单
	List(ctx context.Context, db *gorm.DB, offset, limit int, userID *int32) ([]*do.LogisticsOrderDO, int64, error)
	
	// 根据状态查询物流订单
	FindByStatus(ctx context.Context, db *gorm.DB, status int32) ([]*do.LogisticsOrderDO, error)
}

// LogisticsTracksRepo 物流轨迹数据访问接口
type LogisticsTracksRepo interface {
	// 创建轨迹记录
	Create(ctx context.Context, db *gorm.DB, track *do.LogisticsTrackDO) error
	
	// 批量创建轨迹记录
	BatchCreate(ctx context.Context, db *gorm.DB, tracks []*do.LogisticsTrackDO) error
	
	// 根据物流单号查询轨迹
	GetByLogisticsSn(ctx context.Context, db *gorm.DB, logisticsSn string) ([]*do.LogisticsTrackDO, error)
	
	// 根据快递单号查询轨迹
	GetByTrackingNumber(ctx context.Context, db *gorm.DB, trackingNumber string) ([]*do.LogisticsTrackDO, error)
	
	// 获取最新轨迹
	GetLatest(ctx context.Context, db *gorm.DB, logisticsSn string) (*do.LogisticsTrackDO, error)
}

// LogisticsCouriersRepo 配送员数据访问接口
type LogisticsCouriersRepo interface {
	// 创建配送员
	Create(ctx context.Context, db *gorm.DB, courier *do.LogisticsCourierDO) error
	
	// 根据配送员编号查询
	GetByCourierCode(ctx context.Context, db *gorm.DB, courierCode string) (*do.LogisticsCourierDO, error)
	
	// 根据物流公司查询配送员
	GetByCompany(ctx context.Context, db *gorm.DB, company int32) ([]*do.LogisticsCourierDO, error)
	
	// 根据服务区域查询配送员
	GetByServiceArea(ctx context.Context, db *gorm.DB, serviceArea string) ([]*do.LogisticsCourierDO, error)
	
	// 随机获取配送员
	GetRandomCourier(ctx context.Context, db *gorm.DB, company int32) (*do.LogisticsCourierDO, error)
	
	// 更新配送员信息
	Update(ctx context.Context, db *gorm.DB, courier *do.LogisticsCourierDO) error
	
	// 分页查询配送员
	List(ctx context.Context, db *gorm.DB, offset, limit int, company *int32, area *string) ([]*do.LogisticsCourierDO, int64, error)
}

// DataFactory 数据访问工厂接口
type DataFactory interface {
	// 获取数据库连接
	DB() *gorm.DB
	
	// 开始事务
	Begin() *gorm.DB
	
	// 获取仓储接口
	LogisticsOrders() LogisticsOrdersRepo
	LogisticsTracks() LogisticsTracksRepo
	LogisticsCouriers() LogisticsCouriersRepo
}