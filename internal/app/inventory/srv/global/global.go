package global

import (
	"emshop/internal/app/inventory/srv/config"
	"emshop/internal/app/inventory/srv/data/v1"
	"emshop/internal/app/inventory/srv/pkg/redis_lock"

	"gorm.io/gorm"
)

var (
	// Config 全局配置
	Config *config.Config

	// DB 数据库连接
	DB *gorm.DB

	// FactoryManager 数据工厂管理器
	FactoryManager *v1.FactoryManager

	// RedisLockManager Redis分布式锁管理器
	RedisLockManager *redis_lock.RedisLockManager
)