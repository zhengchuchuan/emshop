package mysql

import (
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"emshop/internal/app/goods/srv/data/v1/interfaces"
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
	Goods() interfaces.GoodsStore
	Categorys() interfaces.CategoryStore
	Brands() interfaces.BrandsStore
	Banners() interfaces.BannerStore
	CategoryBrands() interfaces.GoodsCategoryBrandStore

	// 搜索引擎接口
	Search() SearchFactory

	// 事务支持
	Begin() *gorm.DB
	
	// DB连接访问
	DB() *gorm.DB
	
	// 关闭连接
	Close() error
}

var (
	factory DataFactory
	once    sync.Once
)


// SearchFactory 搜索工厂接口
type SearchFactory interface {
	Goods() interfaces.GoodsSearchStore
}
// mysqlFactory MySQL数据工厂实现
type mysqlFactory struct {
	db           *gorm.DB
	searchFactory SearchFactory
	
	// DAO单例
	goodsDAO         interfaces.GoodsStore
	categoryDAO      interfaces.CategoryStore
	brandDAO         interfaces.BrandsStore
	bannerDAO        interfaces.BannerStore
	categoryBrandDAO interfaces.GoodsCategoryBrandStore
}

func (mf *mysqlFactory) Begin() *gorm.DB {
	return mf.db.Begin()
}

func (mf *mysqlFactory) DB() *gorm.DB {
	return mf.db
}

func (mf *mysqlFactory) Close() error {
	sqlDB, err := mf.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (mf *mysqlFactory) Goods() interfaces.GoodsStore {
	return mf.goodsDAO
}

func (mf *mysqlFactory) Categorys() interfaces.CategoryStore {
	return mf.categoryDAO
}

func (mf *mysqlFactory) Brands() interfaces.BrandsStore {
	return mf.brandDAO
}

func (mf *mysqlFactory) Banners() interfaces.BannerStore {
	return mf.bannerDAO
}

func (mf *mysqlFactory) CategoryBrands() interfaces.GoodsCategoryBrandStore {
	return mf.categoryBrandDAO
}

func (mf *mysqlFactory) Search() SearchFactory {
	return mf.searchFactory
}

var _ DataFactory = &mysqlFactory{}

// NewMySQLFactory 创建MySQL数据工厂
func NewMySQLFactory(mysqlOpts *options.MySQLOptions, searchFactory SearchFactory) (DataFactory, error) {
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
		// 创建临时变量来构建factory
		tempFactory := &mysqlFactory{
			db:            db,
			searchFactory: searchFactory,
		}
		
		// 创建DAO实例
		tempFactory.goodsDAO = newGoods()
		tempFactory.categoryDAO = newCategorys()  // 无状态DAO
		tempFactory.brandDAO = newBrands()        // 无状态DAO
		tempFactory.bannerDAO = newBanner()       // 无状态DAO
		tempFactory.categoryBrandDAO = newCategoryBrands()     // 无状态DAO
		
		factory = tempFactory

		sqlDB.SetMaxOpenConns(mysqlOpts.MaxOpenConnections)
		sqlDB.SetMaxIdleConns(mysqlOpts.MaxIdleConnections)
		sqlDB.SetConnMaxLifetime(mysqlOpts.MaxConnectionLifetime)
	})

	if err != nil {
		return nil, errors2.WithCode(code.ErrConnectDB, "failed to get mysql store factory")
	}
	return factory, nil
}