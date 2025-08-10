package rpcserver

import (
	"emshop/gin-micro/registry"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/metadata"
	"emshop/gin-micro/server/rpc-server/selector"
)

const (
	// balancerName 是负载均衡器的名称
	balancerName = "selector"
)

var (
	// 编译时接口检查
	_ base.PickerBuilder = &balancerBuilder{}
	_ balancer.Picker    = &balancerPicker{}
)

// InitBuilder 初始化并注册负载均衡器构建器
func InitBuilder() {
	b := base.NewBalancerBuilder(
		balancerName,
		&balancerBuilder{
			builder: selector.GlobalSelector(),
		},
		base.Config{HealthCheck: true},
	)
	balancer.Register(b)
}

// balancerBuilder 负载均衡器构建器
type balancerBuilder struct {
	builder selector.Builder // selector构建器
}

// Build 创建 gRPC Picker
func (b *balancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	if len(info.ReadySCs) == 0 {
		// 阻塞 RPC 直到通过 UpdateState() 提供新的 picker
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	// 将就绪的子连接转换为 selector 节点
	nodes := make([]selector.Node, 0, len(info.ReadySCs))
	for conn, info := range info.ReadySCs {
		ins, _ := info.Address.Attributes.Value("rawServiceInstance").(*registry.ServiceInstance)
		nodes = append(nodes, &grpcNode{
			Node:    selector.NewNode("grpc", info.Address.Addr, ins),
			subConn: conn,
		})
	}
	// 创建 picker 并应用节点
	p := &balancerPicker{
		selector: b.builder.Build(),
	}
	p.selector.Apply(nodes)
	return p
}

// balancerPicker 是 gRPC picker 实现
type balancerPicker struct {
	selector selector.Selector // 节点选择器
}

// Pick 选择服务实例
func (p *balancerPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	// 使用 selector 选择节点
	n, done, err := p.selector.Select(info.Ctx)
	if err != nil {
		return balancer.PickResult{}, err
	}

	// 返回选中的子连接和完成回调
	return balancer.PickResult{
		SubConn: n.(*grpcNode).subConn,
		Done: func(di balancer.DoneInfo) {
			// 将 gRPC DoneInfo 转换为 selector DoneInfo
			done(info.Ctx, selector.DoneInfo{
				Err:           di.Err,
				BytesSent:     di.BytesSent,
				BytesReceived: di.BytesReceived,
				ReplyMD:       Trailer(di.Trailer),
			})
		},
	}, nil
}

// Trailer 是 gRPC 响应尾部元数据
type Trailer metadata.MD

// Get 获取 gRPC trailer 值
func (t Trailer) Get(k string) string {
	v := metadata.MD(t).Get(k)
	if len(v) > 0 {
		return v[0]
	}
	return ""
}

// grpcNode gRPC 节点包装器，包含 selector 节点和 gRPC 子连接
type grpcNode struct {
	selector.Node            // 基础节点接口
	subConn       balancer.SubConn // gRPC 子连接
}
