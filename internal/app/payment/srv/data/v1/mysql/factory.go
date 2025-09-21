package mysql

import (
	"emshop/internal/app/payment/srv/data/v1/interfaces"
	"emshop/internal/app/pkg/options"
	gormtrace "emshop/pkg/observability/gormtrace"
	"emshop/pkg/log"
	"fmt"
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	factory      *dataFactory
	factoryOnce sync.Once
)

type dataFactory struct {
	db                     *gorm.DB
	paymentOrderData       interfaces.PaymentOrderDataInterface
	paymentLogData         interfaces.PaymentLogDataInterface
	stockReservationData   interfaces.StockReservationDataInterface
}

// NewDataFactory 创建支付服务数据工厂
func NewDataFactory(mysqlOpts *options.MySQLOptions) (interfaces.DataFactory, error) {
	var err error
	factoryOnce.Do(func() {
		factory = &dataFactory{}
		err = factory.initMySQL(mysqlOpts)
		if err != nil {
			return
		}
		
		factory.paymentOrderData = NewPaymentOrderData(factory.db)
		factory.paymentLogData = NewPaymentLogData(factory.db)
		factory.stockReservationData = NewStockReservationData(factory.db)
	})
	
	return factory, err
}

// initMySQL 初始化MySQL连接
func (f *dataFactory) initMySQL(mysqlOpts *options.MySQLOptions) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		mysqlOpts.Username,
		mysqlOpts.Password,
		mysqlOpts.Host,
		mysqlOpts.Database,
	)
	
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Errorf("MySQL连接失败: %v", err)
		return err
	}

	gormtrace.Enable(db, mysqlOpts.Database)
	
	sqlDB, err := db.DB()
	if err != nil {
		log.Errorf("获取底层数据库连接失败: %v", err)
		return err
	}
	
	// 设置连接池参数
	sqlDB.SetMaxIdleConns(mysqlOpts.MaxIdleConnections)
	sqlDB.SetMaxOpenConns(mysqlOpts.MaxOpenConnections)
	sqlDB.SetConnMaxLifetime(mysqlOpts.MaxConnectionLifetime)
	
	f.db = db
	
	log.Info("MySQL数据库连接成功")
	return nil
}

// PaymentOrders 获取支付订单数据访问对象
func (f *dataFactory) PaymentOrders() interfaces.PaymentOrderDataInterface {
	return f.paymentOrderData
}

// PaymentLogs 获取支付日志数据访问对象
func (f *dataFactory) PaymentLogs() interfaces.PaymentLogDataInterface {
	return f.paymentLogData
}

// StockReservations 获取库存预留数据访问对象
func (f *dataFactory) StockReservations() interfaces.StockReservationDataInterface {
	return f.stockReservationData
}

// DB 获取数据库连接
func (f *dataFactory) DB() *gorm.DB {
	return f.db
}

// Begin 开始事务
func (f *dataFactory) Begin() *gorm.DB {
	return f.db.Begin()
}

// Close 关闭数据库连接
func (f *dataFactory) Close() error {
	sqlDB, err := f.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
