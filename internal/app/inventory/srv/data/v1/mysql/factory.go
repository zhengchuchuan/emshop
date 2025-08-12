package mysql

import (
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"emshop/internal/app/inventory/srv/data/v1/interfaces"
	"emshop/internal/app/pkg/code"
	"emshop/internal/app/pkg/options"
	errors2 "emshop/pkg/errors"
	"os"
	"sync"
	"time"

	"gorm.io/driver/mysql"
)

// DataFactory 数据工厂接口
type DataFactory interface {
	// 主存储接口
	Inventorys() interfaces.InventoryStore

	// 事务支持
	Begin() *gorm.DB
	
	// 关闭连接
	Close() error
}

var (
	factory DataFactory
	once    sync.Once
)

// mysqlFactory MySQL数据工厂实现
type mysqlFactory struct {
	db *gorm.DB
}

func (mf *mysqlFactory) Begin() *gorm.DB {
	return mf.db.Begin()
}

func (mf *mysqlFactory) Close() error {
	sqlDB, err := mf.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (mf *mysqlFactory) Inventorys() interfaces.InventoryStore {
	return newInventorys(mf)
}

var _ DataFactory = &mysqlFactory{}

// NewMySQLFactory 创建MySQL数据工厂
func NewMySQLFactory(mysqlOpts *options.MySQLOptions) (DataFactory, error) {
	if mysqlOpts == nil && factory == nil {
		return nil, fmt.Errorf("failed to get mysql store factory")
	}

	var err error
	once.Do(func() {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			mysqlOpts.Username,
			mysqlOpts.Password,
			mysqlOpts.Host,
			mysqlOpts.Port,
			mysqlOpts.Database)

		// gorm打印日志集成,使用的标准库的logger,没有使用自己封装的
		newLogger := logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的目标，前缀和日志包含的内容——译者注）
			logger.Config{
				SlowThreshold:             time.Second,                         // 慢 SQL 阈值
				LogLevel:                  logger.LogLevel(mysqlOpts.LogLevel), // 日志级别
				IgnoreRecordNotFoundError: true,                                // 忽略ErrRecordNotFound（记录未找到）错误
				Colorful:                  false,                               // 禁用彩色打印
			},
		)
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: newLogger,
		})
		if err != nil {
			return
		}

		sqlDB, _ := db.DB()
		factory = &mysqlFactory{
			db: db,
		}

		sqlDB.SetMaxOpenConns(mysqlOpts.MaxOpenConnections)
		sqlDB.SetMaxIdleConns(mysqlOpts.MaxIdleConnections)
		sqlDB.SetConnMaxLifetime(mysqlOpts.MaxConnectionLifetime)
	})

	if factory == nil || err != nil {
		return nil, errors2.WithCode(code.ErrConnectDB, "failed to get mysql store factory")
	}
	return factory, nil
}