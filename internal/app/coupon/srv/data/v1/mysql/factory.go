package mysql

import (
    "context"
    "emshop/internal/app/coupon/srv/data/v1/interfaces"
    "emshop/internal/app/coupon/srv/domain/do"
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
    flashSaleStockLogData  interfaces.FlashSaleStockLogDataInterface
    flashSaleOrderData     interfaces.FlashSaleOrderDataInterface
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
        factory.flashSaleStockLogData = NewFlashSaleStockLogData(factory.db)
        factory.flashSaleOrderData = NewFlashSaleOrderData(factory.db)
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

    // 自动迁移：确保表结构与代码结构体对齐（若表被删除会自动重建）
    if err := f.autoMigrate(db); err != nil {
        log.Errorf("AutoMigrate 失败: %v", err)
        return err
    }

    // 一次性修复/对齐历史列名差异（remaining_count -> sold_count 初始化）
    if err := f.ensureSchema(db); err != nil {
        log.Errorf("Schema 修复失败: %v", err)
        return err
    }
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

// FlashSaleStockLogs 获取库存扣减日志数据访问对象
func (f *dataFactory) FlashSaleStockLogs() interfaces.FlashSaleStockLogDataInterface {
    return f.flashSaleStockLogData
}

// FlashSaleOrders 获取秒杀订单数据访问对象
func (f *dataFactory) FlashSaleOrders() interfaces.FlashSaleOrderDataInterface {
    return f.flashSaleOrderData
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

// autoMigrate 执行GORM自动迁移，创建缺失的表/列/索引
func (f *dataFactory) autoMigrate(db *gorm.DB) error {
    // 仅创建/新增，不做破坏性变更
    return db.AutoMigrate(
        &do.CouponTemplateDO{},
        
        &do.CouponUsageLogDO{},
        &do.CouponConfigDO{},
        &do.FlashSaleActivityDO{},
        &do.FlashSaleOrderDO{},

        // &do.FlashSaleRecordDO{},
        // &do.FlashSaleStockLogDO{},
        // &do.UserCouponDO{},
    )
}

// ensureSchema 对历史不一致做一次性修复
func (f *dataFactory) ensureSchema(db *gorm.DB) error {
    ctx := context.Background()
    mig := db.Migrator()

    // 对 flash_sale_activities 增补 sold_count 并基于 remaining_count 做初始化（如果存在）
    hasSold := mig.HasColumn(&do.FlashSaleActivityDO{}, "sold_count")
    if !hasSold {
        if err := mig.AddColumn(&do.FlashSaleActivityDO{}, "SoldCount"); err != nil {
            return err
        }
    }
    // 回填已售数量（幂等：仅当 sold_count=0 时更新）
    if err := db.WithContext(ctx).Exec(
        "UPDATE flash_sale_activities SET sold_count = GREATEST(0, flash_sale_count - COALESCE(remaining_count, 0)) WHERE sold_count = 0",
    ).Error; err != nil {
        return err
    }

    return nil
}
