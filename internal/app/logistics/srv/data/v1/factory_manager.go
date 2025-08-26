package v1

import (
	"emshop/internal/app/logistics/srv/data/v1/interfaces"
	"emshop/internal/app/logistics/srv/data/v1/mysql"
	"emshop/internal/app/pkg/options"
	"emshop/pkg/log"

	mysqlDriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// FactoryManager 数据访问工厂管理器
type FactoryManager struct {
	dataFactory interfaces.DataFactory
	db          *gorm.DB
}

// NewFactoryManager 创建工厂管理器实例
func NewFactoryManager(mysqlOpts *options.MySQLOptions) (*FactoryManager, error) {
	// 创建数据库连接
	db, err := createDBConnection(mysqlOpts)
	if err != nil {
		return nil, err
	}

	// 创建数据工厂
	dataFactory := mysql.NewDataFactory(db)

	return &FactoryManager{
		dataFactory: dataFactory,
		db:          db,
	}, nil
}

// GetDataFactory 获取数据工厂实例
func (fm *FactoryManager) GetDataFactory() interfaces.DataFactory {
	return fm.dataFactory
}

// GetDB 获取数据库连接
func (fm *FactoryManager) GetDB() *gorm.DB {
	return fm.db
}

// Close 关闭数据库连接
func (fm *FactoryManager) Close() error {
	sqlDB, err := fm.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// createDBConnection 创建数据库连接
func createDBConnection(opts *options.MySQLOptions) (*gorm.DB, error) {
	// 构建DSN
	dsn := opts.DSN()

	// 设置GORM日志级别
	var logLevel logger.LogLevel
	switch opts.LogLevel {
	case 1:
		logLevel = logger.Silent
	case 2:
		logLevel = logger.Error
	case 3:
		logLevel = logger.Warn
	case 4:
		logLevel = logger.Info
	default:
		logLevel = logger.Info
	}

	// 创建数据库连接
	db, err := gorm.Open(mysqlDriver.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Errorf("failed to connect to MySQL: %v", err)
		return nil, err
	}

	// 设置连接池参数
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(opts.MaxIdleConnections)
	sqlDB.SetMaxOpenConns(opts.MaxOpenConnections)
	sqlDB.SetConnMaxLifetime(opts.MaxConnectionLifetime)

	log.Infof("Successfully connected to MySQL database: %s", opts.Database)
	return db, nil
}