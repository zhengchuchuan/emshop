package initialize

import (
	"emshop/internal/app/inventory/srv/global"
	"emshop/internal/app/inventory/srv/pkg/redis_lock"
	"fmt"
)

// InitRedis 初始化Redis分布式锁
func InitRedis() {
	opts := global.Config.RedisOptions
	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	global.RedisLockManager = redis_lock.NewRedisLockManager(addr)
}