package mysql

import (
	"emshop/internal/app/coupon/srv/data/v1/interfaces"
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
	couponTemplateData     interfaces.CouponTemplateDataInterface
	userCouponData         interfaces.UserCouponDataInterface
	couponUsageLogData     interfaces.CouponUsageLogDataInterface
	couponConfigData       interfaces.CouponConfigDataInterface
	flashSaleData          interfaces.FlashSaleDataInterface
	flashSaleRecordData    interfaces.FlashSaleRecordDataInterface
}

// NewDataFactory 创建优惠券服务数据工厂
func NewDataFactory(mysqlOpts *options.MySQLOptions) (interfaces.DataFactory, error) {
	var err error
	factoryOnce.Do(func() {
		factory = &dataFactory{}
		err = factory.initMySQL(mysqlOpts)
		if err != nil {
			return
		}
		
		factory.couponTemplateData = NewCouponTemplateData(factory.db)
		factory.userCouponData = NewUserCouponData(factory.db)
		factory.couponUsageLogData = NewCouponUsageLogData(factory.db)
		factory.couponConfigData = NewCouponConfigData(factory.db)
		factory.flashSaleData = NewFlashSaleData(factory.db)
		factory.flashSaleRecordData = NewFlashSaleRecordData(factory.db)
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
	
	log.Info("优惠券服务MySQL数据库连接成功")
	return nil
}

// CouponTemplates 获取优惠券模板数据访问对象
func (f *dataFactory) CouponTemplates() interfaces.CouponTemplateDataInterface {
	return f.couponTemplateData
}

// UserCoupons 获取用户优惠券数据访问对象
func (f *dataFactory) UserCoupons() interfaces.UserCouponDataInterface {
	return f.userCouponData
}

// CouponUsageLogs 获取优惠券使用记录数据访问对象
func (f *dataFactory) CouponUsageLogs() interfaces.CouponUsageLogDataInterface {
	return f.couponUsageLogData
}

// CouponConfigs 获取优惠券配置数据访问对象
func (f *dataFactory) CouponConfigs() interfaces.CouponConfigDataInterface {
	return f.couponConfigData
}

// FlashSales 获取秒杀活动数据访问对象
func (f *dataFactory) FlashSales() interfaces.FlashSaleDataInterface {
	return f.flashSaleData
}

// FlashSaleRecords 获取秒杀记录数据访问对象
func (f *dataFactory) FlashSaleRecords() interfaces.FlashSaleRecordDataInterface {
	return f.flashSaleRecordData
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
