// Package discovery 实现了基于服务注册中心的gRPC动态服务发现解析器
package discovery

import (
	"context"
	"errors"
	"strings"
	"time"

	"google.golang.org/grpc/resolver"
	"emshop/gin-micro/registry"
)

// name 定义服务发现解析器的协议名称
const name = "discovery"

// Option 是builder配置选项函数类型
type Option func(o *builder)

// WithTimeout 设置服务发现的超时时间选项
func WithTimeout(timeout time.Duration) Option {
	return func(b *builder) {
		b.timeout = timeout
	}
}

// WithInsecure 设置是否启用不安全连接选项
func WithInsecure(insecure bool) Option {
	return func(b *builder) {
		b.insecure = insecure
	}
}

// builder 服务发现解析器构建器
// 实现gRPC的resolver.Builder接口，用于创建动态服务发现解析器
type builder struct {
	discoverer registry.Discovery // 服务发现接口实现（如consul、etcd等）
	timeout    time.Duration      // 创建watcher的超时时间
	insecure   bool               // 是否允许不安全连接
}

// NewBuilder 创建一个新的服务发现解析器构建器
// d: 服务注册中心发现接口实现（consul、etcd、zookeeper等）
// opts: 可选配置参数
// 返回实现了resolver.Builder接口的构建器实例
func NewBuilder(d registry.Discovery, opts ...Option) resolver.Builder {
	b := &builder{
		discoverer: d,                   // 传递注册的服务发现实现
		timeout:    time.Second * 10,    // 默认10秒超时
		insecure:   false,               // 默认使用安全连接
	}
	for _, o := range opts {
		o(b)
	}
	return b
}

// Build 实现resolver.Builder接口，创建服务发现解析器实例
// target: gRPC目标地址，包含服务名称信息
// cc: gRPC客户端连接，用于更新服务实例状态
// opts: gRPC构建选项
// 返回resolver.Resolver接口实现和可能的错误
func (b *builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	var (
		err error
		w   registry.Watcher // 服务监听器
	)
	// 创建异步通道等待watcher创建完成
	done := make(chan struct{}, 1)
	ctx, cancel := context.WithCancel(context.Background())
	
	// 在goroutine中创建服务监听器，避免阻塞
	go func() {
		// 从target URL路径中提取服务名称并创建watcher
		serviceName := strings.TrimPrefix(target.URL.Path, "/")
		w, err = b.discoverer.Watch(ctx, serviceName)
		close(done) // 通知创建完成
	}()
	
	// 等待watcher创建完成或超时
	select {
	case <-done:
		// watcher创建完成
	case <-time.After(b.timeout):
		// 创建超时
		err = errors.New("discovery create watcher overtime")
	}
	
	if err != nil {
		cancel()
		return nil, err
	}
	
	// 创建服务发现解析器实例
	r := &discoveryResolver{
		w:        w,           // 服务监听器
		cc:       cc,          // gRPC客户端连接
		ctx:      ctx,         // 上下文
		cancel:   cancel,      // 取消函数
		insecure: b.insecure,  // 安全连接配置
	}
	
	// 启动后台监听协程，持续监听服务变化
	go r.watch()
	return r, nil
}

// Scheme 返回服务发现解析器的协议方案名称
// 实现resolver.Builder接口的必需方法
func (*builder) Scheme() string {
	return name
}
