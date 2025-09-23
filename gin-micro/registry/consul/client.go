package consul

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"emshop/gin-micro/registry"
	"emshop/pkg/log"

	"github.com/hashicorp/consul/api"
)

// Consul客户端配置
type Client struct {
	cli    *api.Client         // Consul API客户端
	ctx    context.Context     // 上下文
	cancel context.CancelFunc  // 取消函数
	resolver ServiceResolver	// 解析服务入口端点
	healthcheckInterval int	// 健康检查时间间隔(秒)
	heartbeat bool			// 是否启用心跳
    deregisterCriticalServiceAfter int // 严重错误服务自动注销时间间隔(秒)
    serviceChecks api.AgentServiceChecks 	// 用户自定义检查项: 内置了TCP检查和TTL检查, 可以额外添加http或者grpc
    checkTimeout int // 健康检查超时时间(秒)
}

// 创建Consul客户端
func NewClient(cli *api.Client) *Client {
    c := &Client{
        cli:                            cli,
        resolver:                       defaultResolver,
        healthcheckInterval:            10,
        heartbeat:                      true,
        deregisterCriticalServiceAfter: 600,
        checkTimeout:                   5,
    }
    c.ctx, c.cancel = context.WithCancel(context.Background())
    return c
}

// 默认的服务解析器，将Consul服务条目转换为服务实例
func defaultResolver(_ context.Context, entries []*api.ServiceEntry) []*registry.ServiceInstance {
	// 初始化服务实例列表
	services := make([]*registry.ServiceInstance, 0, len(entries))
	// 遍历所有服务条目
	for _, entry := range entries {
		var version string
		// 从标签中提取版本信息
		// "Tags": ["primary", "v1"],
		for _, tag := range entry.Service.Tags {
			ss := strings.SplitN(tag, "=", 2)
			if len(ss) == 2 && ss[0] == "version" {
				version = ss[1]
			}
		}
		// 提取服务端点信息
		// 在混合云或跨网络架构中，节点可能同时拥有多个 IP（如内网和公网）
		endpoints := make([]string, 0)
		// 从标签地址中提取端点（跳过网络地址）
		// "TaggedAddresses": {
		// 	"lan": {
		// 		"Address": "127.0.0.1",
		// 		"Port": 8000
		// 	},
		// 	"wan": {
		// 		"Address": "198.18.0.1",
		// 		"Port": 80
		// 	}
    	// }
		for scheme, addr := range entry.Service.TaggedAddresses {
			if scheme == "lan_ipv4" || scheme == "wan_ipv4" || scheme == "lan_ipv6" || scheme == "wan_ipv6" {
				continue
			}
			endpoints = append(endpoints, addr.Address)
		}
		// 如果没有端点，使用默认地址和端口
		if len(endpoints) == 0 && entry.Service.Address != "" && entry.Service.Port != 0 {
			endpoints = append(endpoints, fmt.Sprintf("http://%s:%d", entry.Service.Address, entry.Service.Port))
		}
		// 创建服务实例
		services = append(services, &registry.ServiceInstance{
			ID:        entry.Service.ID,
			Name:      entry.Service.Service,
			Metadata:  entry.Service.Meta,
			Version:   version,
			Endpoints: endpoints,
		})
	}

	return services
}

// ServiceResolver 用于解析服务端点
type ServiceResolver func(ctx context.Context, entries []*api.ServiceEntry) []*registry.ServiceInstance

// Service 从Consul获取服务列表
//	@param ctx 
//	@param service 要查询的服务名
//	@param index 用于长轮询的索引，0表示立即返回
//	@param passingOnly 是否只返回健康检查通过的实例
//	@return []*registry.ServiceInstance 
//	@return uint64 
//	@return error 
func (c *Client) Service(ctx context.Context, service string, index uint64, passingOnly bool) ([]*registry.ServiceInstance, uint64, error) {
	opts := &api.QueryOptions{
		WaitIndex: index,
		WaitTime:  time.Second * 55,
	}
	opts = opts.WithContext(ctx)
	entries, meta, err := c.cli.Health().Service(service, "", passingOnly, opts)
	if err != nil {
		return nil, 0, err
	}
	return c.resolver(ctx, entries), meta.LastIndex, nil
}

// Register 注册服务实例
//	@param _ 上下文
//	@param svc 服务实例
//	@param enableHealthCheck 是否启用TCP健康检查
//	@return error
func (c *Client) Register(_ context.Context, svc *registry.ServiceInstance, enableHealthCheck bool) error {
    // 初始化地址映射和检查地址列表
    addresses := make(map[string]api.ServiceAddress, len(svc.Endpoints))
    tcpCheckAddresses := make([]string, 0, len(svc.Endpoints))
    grpcCheckAddresses := make([]string, 0, len(svc.Endpoints))
	// 解析所有服务端点
    for _, endpoint := range svc.Endpoints {
        raw, err := url.Parse(endpoint)
        if err != nil {
            return err
        }
        addr := raw.Hostname()
        port, _ := strconv.ParseUint(raw.Port(), 10, 16)

        // 根据不同协议收集健康检查地址
        if raw.Scheme == "grpc" {
            grpcCheckAddresses = append(grpcCheckAddresses, net.JoinHostPort(addr, strconv.FormatUint(port, 10)))
        } else {
            tcpCheckAddresses = append(tcpCheckAddresses, net.JoinHostPort(addr, strconv.FormatUint(port, 10)))
        }
        addresses[raw.Scheme] = api.ServiceAddress{Address: endpoint, Port: int(port)}
    }
	// 创建服务注册配置,协议无关
	asr := &api.AgentServiceRegistration{
		ID:              svc.ID,
		Name:            svc.Name,
		Meta:            svc.Metadata,
		Tags:            []string{fmt.Sprintf("version=%s", svc.Version)},
		TaggedAddresses: addresses,
	}
	// 设置主地址和端口
    if len(tcpCheckAddresses) > 0 || len(grpcCheckAddresses) > 0 {
        // 选任一地址作为主注册地址
        pick := ""
        if len(grpcCheckAddresses) > 0 {
            pick = grpcCheckAddresses[0]
        } else if len(tcpCheckAddresses) > 0 {
            pick = tcpCheckAddresses[0]
        }
        host, portRaw, _ := net.SplitHostPort(pick)
        port, _ := strconv.ParseInt(portRaw, 10, 32)
        asr.Address = host
        asr.Port = int(port)
    }
    // 如果启用健康检查，添加TCP检查
    if enableHealthCheck {
        // 为非 gRPC 端点添加 TCP 健康检查
        for _, address := range tcpCheckAddresses {
            asr.Checks = append(asr.Checks, &api.AgentServiceCheck{
                TCP:                            address,
                Interval:                       fmt.Sprintf("%ds", c.healthcheckInterval),
                DeregisterCriticalServiceAfter: fmt.Sprintf("%ds", c.deregisterCriticalServiceAfter),
                Timeout:                        fmt.Sprintf("%ds", c.checkTimeout),
            })
        }
        // 为 gRPC 端点添加 gRPC 健康检查（依赖服务端注册了 grpc health 服务）
        for _, address := range grpcCheckAddresses {
            asr.Checks = append(asr.Checks, &api.AgentServiceCheck{
                GRPC:                           address,
                GRPCUseTLS:                     false,
                Interval:                       fmt.Sprintf("%ds", c.healthcheckInterval),
                DeregisterCriticalServiceAfter: fmt.Sprintf("%ds", c.deregisterCriticalServiceAfter),
                Timeout:                        fmt.Sprintf("%ds", c.checkTimeout),
            })
        }
    }
	// 如果启用心跳，添加TTL检查
	if c.heartbeat {
		asr.Checks = append(asr.Checks, &api.AgentServiceCheck{
			CheckID:                        "service:" + svc.ID,
			TTL:                            fmt.Sprintf("%ds", c.healthcheckInterval*2),		// 如果此时间内没有收到心跳,就认为不健康
			DeregisterCriticalServiceAfter: fmt.Sprintf("%ds", c.deregisterCriticalServiceAfter),
		})
	}

	// 添加用户自定义检查
	asr.Checks = append(asr.Checks, c.serviceChecks...)

	err := c.cli.Agent().ServiceRegister(asr)
	if err != nil {
		return err
	}
	// 如果启用心跳，启动心跳协程
	if c.heartbeat {
		go func() {
			// 避免心跳比注册更糟到达consul
			time.Sleep(time.Second)
			// 立即心跳,使得服务注册后立马能健康检查通过
			err = c.cli.Agent().UpdateTTL("service:"+svc.ID, "pass", "pass")
			if err != nil {
				log.Errorf("[Consul]update ttl heartbeat to consul failed!err:=%v", err)
			}
			// 定时发送心跳
			ticker := time.NewTicker(time.Second * time.Duration(c.healthcheckInterval))
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					err = c.cli.Agent().UpdateTTL("service:"+svc.ID, "pass", "pass")
					if err != nil {
						log.Errorf("[Consul]update ttl heartbeat to consul failed!err:=%v", err)
					}
				case <-c.ctx.Done():
					return
				}
			}
		}()
	}
	return nil
}

// Deregister 根据服务ID从Consul注销服务
func (c *Client) Deregister(_ context.Context, serviceID string) error {
	c.cancel()
	return c.cli.Agent().ServiceDeregister(serviceID)
}
