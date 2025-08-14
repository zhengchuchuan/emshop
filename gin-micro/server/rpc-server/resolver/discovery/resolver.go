package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"time"

	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"emshop/gin-micro/registry"
	"emshop/pkg/log"
)

// discoveryResolver 基于服务注册中心的动态服务发现解析器
// 实现gRPC的resolver.Resolver接口，提供实时的服务实例发现和更新
type discoveryResolver struct {
	w  registry.Watcher      // 服务监听器，用于监听服务实例变化
	cc resolver.ClientConn   // gRPC客户端连接，用于通知连接状态更新

	ctx    context.Context     // 上下文，用于控制生命周期
	cancel context.CancelFunc  // 取消函数，用于停止监听

	insecure bool // 是否允许不安全连接
}

// watch 持续监听服务实例变化的后台协程
// 通过watcher.Next()阻塞等待服务变化事件，并更新gRPC连接状态
func (r *discoveryResolver) watch() {
	for {
		// 检查上下文是否已取消
		select {
		case <-r.ctx.Done():
			return
		default:
		}
		
		// 阻塞等待下一个服务实例变化事件
		ins, err := r.w.Next()
		if err != nil {
			// 如果是上下文取消，则正常退出
			if errors.Is(err, context.Canceled) {
				return
			}
			// 记录错误并短暂等待后重试
			log.Errorf("[resolver] Failed to watch discovery endpoint: %v", err)
			time.Sleep(time.Second)
			continue
		}
		
		// 更新服务实例到gRPC连接管理器
		r.update(ins)
	}
}

// update 处理服务实例变化，更新gRPC连接状态
// ins: 从服务注册中心获取的最新服务实例列表
func (r *discoveryResolver) update(ins []*registry.ServiceInstance) {
	// 存储解析后的gRPC地址列表
	addrs := make([]resolver.Address, 0)
	// 用于去重的端点集合
	endpoints := make(map[string]struct{})
	
	// 遍历所有服务实例，解析有效的gRPC端点
	for _, in := range ins {
		// 从服务实例的端点列表中解析出gRPC协议的地址
		endpoint, err := ParseEndpoint(in.Endpoints, "grpc", !r.insecure)
		if err != nil {
			log.Errorf("[resolver] Failed to parse discovery endpoint: %v", err)
			continue
		}
		if endpoint == "" {
			continue
		}
		
		// 过滤重复的端点地址
		if _, ok := endpoints[endpoint]; ok {
			continue
		}
		endpoints[endpoint] = struct{}{}
		
		// 构建gRPC resolver地址对象
		addr := resolver.Address{
			ServerName: in.Name,                         // 服务名称
			Attributes: parseAttributes(in.Metadata),    // 解析元数据为属性
			Addr:       endpoint,                        // 端点地址
		}
		// 将原始服务实例信息作为属性保存，供负载均衡器使用
		addr.Attributes = addr.Attributes.WithValue("rawServiceInstance", in)
		addrs = append(addrs, addr)
	}
	
	// 如果没有有效的端点，拒绝更新并记录警告
	if len(addrs) == 0 {
		log.Warnf("[resolver] Zero endpoint found,refused to write, instances: %v", ins)
		return
	}
	
	// 更新gRPC客户端连接状态，传递最新的地址列表
	err := r.cc.UpdateState(resolver.State{Addresses: addrs})
	if err != nil {
		log.Errorf("[resolver] failed to update state: %s", err)
	}
	
	// 记录成功更新的服务实例信息
	b, _ := json.Marshal(ins)
	log.Infof("[resolver] update instances: %s", b)
}

// Close 关闭服务发现解析器，清理资源
// 实现resolver.Resolver接口的必需方法
func (r *discoveryResolver) Close() {
	// 取消上下文，停止watch协程
	r.cancel()
	// 停止服务监听器
	err := r.w.Stop()
	if err != nil {
		log.Errorf("[resolver] failed to watch stop: %s", err)
	}
}

// ResolveNow 立即触发服务解析
// 实现resolver.Resolver接口的必需方法，对于基于watcher的实现通常为空
func (r *discoveryResolver) ResolveNow(options resolver.ResolveNowOptions) {}

// parseAttributes 将服务元数据映射转换为gRPC属性对象
// md: 服务实例的元数据键值对
// 返回包含所有元数据的gRPC属性对象
func parseAttributes(md map[string]string) *attributes.Attributes {
	var a *attributes.Attributes
	for k, v := range md {
		if a == nil {
			// 创建第一个属性
			a = attributes.New(k, v)
		} else {
			// 添加后续属性
			a = a.WithValue(k, v)
		}
	}
	return a
}

// NewEndpoint 创建一个新的端点URL
// scheme: 协议方案（如"grpc"、"http"）
// host: 主机地址（如"127.0.0.1:9000"）
// isSecure: 是否为安全连接
// 返回构造的URL对象
func NewEndpoint(scheme, host string, isSecure bool) *url.URL {
	var query string
	if isSecure {
		query = "isSecure=true"
	}
	return &url.URL{Scheme: scheme, Host: host, RawQuery: query}
}

// ParseEndpoint 从端点列表中解析出匹配指定协议和安全性的端点地址
// endpoints: 服务实例的端点URL列表
// scheme: 目标协议方案（如"grpc"）
// isSecure: 是否要求安全连接
// 返回匹配的主机地址或空字符串及可能的错误
func ParseEndpoint(endpoints []string, scheme string, isSecure bool) (string, error) {
	for _, e := range endpoints {
		// 解析端点URL
		u, err := url.Parse(e)
		if err != nil {
			return "", err
		}
		// 检查协议方案是否匹配
		if u.Scheme == scheme {
			// 检查安全性设置是否匹配
			if IsSecure(u) == isSecure {
				return u.Host, nil
			}
		}
	}
	return "", nil
}

// IsSecure 解析端点URL的安全性设置
// u: 要检查的URL对象
// 返回该端点是否配置为安全连接
func IsSecure(u *url.URL) bool {
	ok, err := strconv.ParseBool(u.Query().Get("isSecure"))
	if err != nil {
		return false
	}
	return ok
}
