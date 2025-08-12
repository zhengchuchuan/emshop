package initialize

import (
	"emshop/internal/app/inventory/srv/domain/do"
	"emshop/internal/app/inventory/srv/global"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB 初始化数据库
func InitDB() {
	opts := global.Config.MySQLOptions
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		opts.Username,
		opts.Password,
		opts.Host,
		opts.Port,
		opts.Database)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.LogLevel(opts.LogLevel),
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	var err error
	global.DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %v", err))
	}

	sqlDB, err := global.DB.DB()
	if err != nil {
		panic(fmt.Sprintf("failed to get underlying sql.DB: %v", err))
	}

	sqlDB.SetMaxOpenConns(opts.MaxOpenConnections)
	sqlDB.SetMaxIdleConns(opts.MaxIdleConnections)
	sqlDB.SetConnMaxLifetime(opts.MaxConnectionLifetime)

	// 自动迁移表结构
	err = global.DB.AutoMigrate(
		&do.InventoryDO{},
		&do.InventoryNewDO{},
		&do.StockSellDetailDO{},
		&do.DeliveryDO{},
	)
	if err != nil {
		panic(fmt.Sprintf("failed to migrate database: %v", err))
	}
}