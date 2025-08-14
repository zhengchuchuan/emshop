// Package consul 实现了基于Consul的服务注册与发现
package consul

import (
	"sync"
	"sync/atomic"

	"emshop/gin-micro/registry"
)

// serviceSet 服务集合，管理具有相同名称的服务实例和监听器
type serviceSet struct {
	serviceName string                     // 服务名称
	watcher     map[*watcher]struct{}       // 监听器集合
	services    *atomic.Value              // 原子存储的服务实例列表
	lock        sync.RWMutex               // 读写锁
}

// broadcast 广播服务更新给所有监听器
func (s *serviceSet) broadcast(ss []*registry.ServiceInstance) {
	// 原子操作，保证线程安全
	s.services.Store(ss)
	s.lock.RLock()
	defer s.lock.RUnlock()
	// 通知所有监听器服务发生变化
	for k := range s.watcher {
		select {
		case k.event <- struct{}{}:
		default:
		}
	}
}
