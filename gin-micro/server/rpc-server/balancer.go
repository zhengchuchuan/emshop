package rpcserver

import (
	"sync"

	"emshop/gin-micro/registry"
	"emshop/gin-micro/server/rpc-server/selector"
	"emshop/gin-micro/server/rpc-server/selector/p2c"
	"emshop/gin-micro/server/rpc-server/selector/random"
	"emshop/gin-micro/server/rpc-server/selector/wrr"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/metadata"
)

const (
	// selectorName 是“全局选择器”映射到的 gRPC 策略名
	// 通过 selector.SetGlobalSelector() 设置其具体算法
	selectorName = "selector"

	// 直接选择内置算法的策略名（客户端可按名选择）
	p2cName    = "p2c"
	wrrName    = "wrr"
	randomName = "random"
)

var (
	// 编译时接口检查
	_ base.PickerBuilder = &balancerBuilder{}
	_ balancer.Picker    = &balancerPicker{}

	regOnce     sync.Once
	regMu       sync.Mutex
	regNamesSet = map[string]struct{}{}
)

// InitBuilder 初始化并注册内置负载均衡器（幂等）
// - 注册 selector（使用全局选择器；若未设置则默认 p2c）
// - 注册 p2c/wrr/random 三种算法的独立策略名，便于客户端按名选择
func InitBuilder() {
	regOnce.Do(func() {
		// selector: 使用全局选择器（若未设置则回落到 p2c）
		global := selector.GlobalSelector()
		if global == nil {
			global = p2c.NewBuilder()
		}
		registerBalancer(selectorName, global)

		// 直接按算法名注册内置负载均衡器，方便客户端明确选择
		registerBalancer(p2cName, p2c.NewBuilder())
		registerBalancer(wrrName, wrr.NewBuilder())
		registerBalancer(randomName, random.NewBuilder())
	})
}

// RegisterBalancer 允许外部以自定义名称注册一个选择器（幂等）
// 若重复注册相同名称将被忽略，以避免 gRPC 重复注册造成崩溃
func RegisterBalancer(name string, builder selector.Builder) {
	registerBalancer(name, builder)
}

func registerBalancer(name string, builder selector.Builder) {
	if builder == nil || name == "" {
		return
	}
	regMu.Lock()
	if _, ok := regNamesSet[name]; ok {
		regMu.Unlock()
		return
	}
	regNamesSet[name] = struct{}{}
	regMu.Unlock()

	b := base.NewBalancerBuilder(
		name,
		&balancerBuilder{
			builder: builder,
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
	selector.Node                  // 基础节点接口
	subConn       balancer.SubConn // gRPC 子连接
}
