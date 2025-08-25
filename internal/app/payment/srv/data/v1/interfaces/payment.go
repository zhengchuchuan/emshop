package interfaces

import (
	"context"
	"time"
	"emshop/internal/app/payment/srv/domain/do"
	v1 "emshop/pkg/common/meta/v1"
	"gorm.io/gorm"
)

// PaymentOrderDataInterface 支付订单数据接口
type PaymentOrderDataInterface interface {
	// 基础CRUD操作
	Create(ctx context.Context, db *gorm.DB, payment *do.PaymentOrderDO) error
	Update(ctx context.Context, db *gorm.DB, payment *do.PaymentOrderDO) error
	Delete(ctx context.Context, db *gorm.DB, paymentSn string) error
	Get(ctx context.Context, db *gorm.DB, paymentSn string) (*do.PaymentOrderDO, error)
	GetByOrderSn(ctx context.Context, db *gorm.DB, orderSn string) (*do.PaymentOrderDO, error)
	List(ctx context.Context, db *gorm.DB, userID int32, meta v1.ListMeta, orderby []string) (*do.PaymentOrderDOList, error)

	// 业务操作
	UpdateStatus(ctx context.Context, db *gorm.DB, paymentSn string, status do.PaymentStatus) error
	UpdatePaidInfo(ctx context.Context, db *gorm.DB, paymentSn string, thirdPartySn *string, paidAt *time.Time) error
	FindExpiredPayments(ctx context.Context, db *gorm.DB, beforeTime time.Time) ([]*do.PaymentOrderDO, error)
	CountByStatus(ctx context.Context, db *gorm.DB, status do.PaymentStatus) (int64, error)
}

// PaymentLogDataInterface 支付日志数据接口
type PaymentLogDataInterface interface {
	Create(ctx context.Context, db *gorm.DB, log *do.PaymentLogDO) error
	List(ctx context.Context, db *gorm.DB, paymentSn string, meta v1.ListMeta) (*do.PaymentLogDOList, error)
	GetByAction(ctx context.Context, db *gorm.DB, paymentSn string, action string) ([]*do.PaymentLogDO, error)
}

// StockReservationDataInterface 库存预留数据接口
type StockReservationDataInterface interface {
	// 基础操作
	Create(ctx context.Context, db *gorm.DB, reservation *do.StockReservationDO) error
	Update(ctx context.Context, db *gorm.DB, reservation *do.StockReservationDO) error
	Get(ctx context.Context, db *gorm.DB, orderSn string, goodsID int32) (*do.StockReservationDO, error)
	GetByOrderSn(ctx context.Context, db *gorm.DB, orderSn string) ([]*do.StockReservationDO, error)
	
	// 批量操作
	BatchCreate(ctx context.Context, db *gorm.DB, reservations []*do.StockReservationDO) error
	BatchUpdateStatus(ctx context.Context, db *gorm.DB, orderSn string, status do.StockReservationStatus) error
	
	// 业务操作
	UpdateStatus(ctx context.Context, db *gorm.DB, orderSn string, goodsID int32, status do.StockReservationStatus) error
	FindExpiredReservations(ctx context.Context, db *gorm.DB, beforeTime time.Time) ([]*do.StockReservationDO, error)
	CountByStatus(ctx context.Context, db *gorm.DB, status do.StockReservationStatus) (int64, error)
}

// DataFactory 支付服务数据工厂接口
type DataFactory interface {
	PaymentOrders() PaymentOrderDataInterface
	PaymentLogs() PaymentLogDataInterface
	StockReservations() StockReservationDataInterface
	
	// 数据库操作
	DB() *gorm.DB
	Begin() *gorm.DB
	Close() error
}