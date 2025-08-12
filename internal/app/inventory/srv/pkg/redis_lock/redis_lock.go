package redis_lock

import (
	"fmt"
	"sync"
	"time"

	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
)

// RedisLockManager Redis分布式锁管理器
type RedisLockManager struct {
	client *goredislib.Client
	rs     *redsync.Redsync
}

// NewRedisLockManager 创建Redis分布式锁管理器
func NewRedisLockManager(addr string) *RedisLockManager {
	client := goredislib.NewClient(&goredislib.Options{
		Addr: addr,
	})
	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	return &RedisLockManager{
		client: client,
		rs:     rs,
	}
}

// GetMutex 获取分布式互斥锁
func (rlm *RedisLockManager) GetMutex(name string) *redsync.Mutex {
	return rlm.rs.NewMutex(name)
}

// TestRedisLock 测试Redis分布式锁
func TestRedisLock(addr string) {
	manager := NewRedisLockManager(addr)
	gNum := 2
	mutexname := "test_lock"

	var wg sync.WaitGroup
	wg.Add(gNum)

	for i := 0; i < gNum; i++ {
		go func(id int) {
			defer wg.Done()
			mutex := manager.GetMutex(mutexname)

			fmt.Printf("协程%d 开始获取锁\n", id)
			if err := mutex.Lock(); err != nil {
				fmt.Printf("协程%d 获取锁失败: %v\n", id, err)
				return
			}

			fmt.Printf("协程%d 获取锁成功\n", id)
			time.Sleep(time.Second * 3)

			fmt.Printf("协程%d 开始释放锁\n", id)
			if ok, err := mutex.Unlock(); !ok || err != nil {
				fmt.Printf("协程%d 释放锁失败: %v\n", id, err)
				return
			}
			fmt.Printf("协程%d 释放锁成功\n", id)
		}(i)
	}
	wg.Wait()
}