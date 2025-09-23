package consul

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"emshop/gin-micro/registry"

	"github.com/hashicorp/consul/api"
)

// 编译时检查接口实现
var (
	_ registry.Registrar = &Registry{}
	_ registry.Discovery = &Registry{}
)

// Consul注册器选项
type Option func(*Registry)

// 设置注册器健康检查选项
func WithHealthCheck(enable bool) Option {
	return func(o *Registry) {
		o.enableHealthCheck = enable
	}
}

// 启用或禁用心跳检查
func WithHeartbeat(enable bool) Option {
	return func(o *Registry) {
		if o.cli != nil {
			o.cli.heartbeat = enable
		}
	}
}

// 设置服务端点解析函数选项
func WithServiceResolver(fn ServiceResolver) Option {
	return func(o *Registry) {
		if o.cli != nil {
			o.cli.resolver = fn
		}
	}
}

// 设置健康检查间隔时间(秒)
func WithHealthCheckInterval(interval int) Option {
	return func(o *Registry) {
		if o.cli != nil {
			o.cli.healthcheckInterval = interval
		}
	}
}

// 设置严重错误服务自动注销时间(秒)
func WithDeregisterCriticalServiceAfter(interval int) Option {
    return func(o *Registry) {
        if o.cli != nil {
            o.cli.deregisterCriticalServiceAfter = interval
        }
    }
}

// WithCheckTimeout sets the health check timeout (seconds) for gRPC/TCP checks
func WithCheckTimeout(seconds int) Option {
    return func(o *Registry) {
        if o.cli != nil {
            o.cli.checkTimeout = seconds
        }
    }
}

// 设置自定义服务检查
func WithServiceCheck(checks ...*api.AgentServiceCheck) Option {
	return func(o *Registry) {
		if o.cli != nil {
			o.cli.serviceChecks = checks
		}
	}
}

// Config Consul注册器配置
type Config struct {
	*api.Config
}

// Registry Consul注册器实现
type Registry struct {
	cli               *Client                  // Consul客户端
	enableHealthCheck bool                     // 是否启用健康检查
	registry          map[string]*serviceSet   // 服务集合映射
	lock              sync.RWMutex             // 读写锁
}

// 函数选项 创建Consul注册器实例
func New(apiClient *api.Client, opts ...Option) *Registry {
	r := &Registry{
		cli:               NewClient(apiClient),
		registry:          make(map[string]*serviceSet),
		enableHealthCheck: true,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// 注册服务到Consul
func (r *Registry) Register(ctx context.Context, svc *registry.ServiceInstance) error {
	return r.cli.Register(ctx, svc, r.enableHealthCheck)
}

// 从Consul注销服务
func (r *Registry) Deregister(ctx context.Context, svc *registry.ServiceInstance) error {
	return r.cli.Deregister(ctx, svc.ID)
}

// 根据服务名称获取服务实例列表
func (r *Registry) GetService(ctx context.Context, name string) ([]*registry.ServiceInstance, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	set := r.registry[name]

	// 从远程Consul获取服务实例
	getRemote := func() []*registry.ServiceInstance {
		services, _, err := r.cli.Service(ctx, name, 0, true)
		if err == nil && len(services) > 0 {
			return services
		}
		return nil
	}

	if set == nil {
		if s := getRemote(); len(s) > 0 {
			return s, nil
		}
		return nil, fmt.Errorf("service %s not resolved in registry", name)
	}
	ss, _ := set.services.Load().([]*registry.ServiceInstance)
	if ss == nil {
		if s := getRemote(); len(s) > 0 {
			return s, nil
		}
		return nil, fmt.Errorf("service %s not found in registry", name)
	}
	return ss, nil
}

// ListServices 返回所有已注册的服务列表
func (r *Registry) ListServices() (allServices map[string][]*registry.ServiceInstance, err error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	allServices = make(map[string][]*registry.ServiceInstance)
	for name, set := range r.registry {
		var services []*registry.ServiceInstance
		ss, _ := set.services.Load().([]*registry.ServiceInstance)
		if ss == nil {
			continue
		}
		services = append(services, ss...)
		allServices[name] = services
	}
	return
}

// Watch 监听指定服务名称的服务变化
func (r *Registry) Watch(ctx context.Context, name string) (registry.Watcher, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	set, ok := r.registry[name]
	// 如果服务集合不存在，创建新的服务集合
	if !ok {
		set = &serviceSet{
			watcher:     make(map[*watcher]struct{}),
			services:    &atomic.Value{},
			serviceName: name,
		}
		r.registry[name] = set
	}

	// 初始化服务监听器
	w := &watcher{
		event: make(chan struct{}, 1),
	}
	w.ctx, w.cancel = context.WithCancel(context.Background())
	w.set = set
	set.lock.Lock()
	set.watcher[w] = struct{}{}
	set.lock.Unlock()
	ss, _ := set.services.Load().([]*registry.ServiceInstance)
	if len(ss) > 0 {
		// 如果服务有值，需要推送给监听器
		// 否则初始数据可能在监听期间永远阻塞
		w.event <- struct{}{}
	}

	if !ok {
		err := r.resolve(set)
		if err != nil {
			return nil, err
		}
	}
	return w, nil
}

// resolve 解析并监听服务变化
func (r *Registry) resolve(ss *serviceSet) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	services, idx, err := r.cli.Service(ctx, ss.serviceName, 0, true)
	cancel()
	if err != nil {
		return err
	} else if len(services) > 0 {
		ss.broadcast(services)
	}
	// 启动后台协程持续监听服务变化
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
			tmpService, tmpIdx, err := r.cli.Service(ctx, ss.serviceName, idx, true)
			cancel()
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			// 有服务变化
			if len(tmpService) != 0 && tmpIdx != idx {
				services = tmpService
				ss.broadcast(services)
			}
			idx = tmpIdx
		}
	}()

	return nil
}
