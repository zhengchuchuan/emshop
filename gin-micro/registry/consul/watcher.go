// Package consul 实现了基于Consul的服务注册与发现
package consul

import (
	"context"

	"emshop/gin-micro/registry"
)

// watcher 服务监听器，用于监听服务实例的变化
type watcher struct {
	event chan struct{}      // 事件通道
	set   *serviceSet        // 所属的服务集合

	// 用于取消监听
	ctx    context.Context
	cancel context.CancelFunc
}

// Next 等待并返回下一个服务实例列表变化
func (w *watcher) Next() (services []*registry.ServiceInstance, err error) {
	// 等待事件或上下文取消
	select {
	case <-w.ctx.Done():
		err = w.ctx.Err()
	// 阻塞直到有事件发生
	case <-w.event:
	}

	// 从服务集合中加载服务实例列表
	ss, ok := w.set.services.Load().([]*registry.ServiceInstance)

	if ok {
		// services初始为nil,复制服务实例列表
		services = append(services, ss...)
	}
	return
}

// Stop 停止监听并从服务集合中移除自身
func (w *watcher) Stop() error {
	// 取消上下文
	w.cancel()
	// 从服务集合中移除监听器
	w.set.lock.Lock()
	defer w.set.lock.Unlock()
	delete(w.set.watcher, w)
	return nil
}
