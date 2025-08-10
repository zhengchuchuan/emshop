package selector

import (
	"context"
	"sync/atomic"
)

// Default 组合式选择器实现
type Default struct {
	NodeBuilder WeightedNodeBuilder // 节点构建器
	Balancer    Balancer           // 负载均衡器

	nodes atomic.Value // 原子存储的节点列表，类型为 []WeightedNode
}

// Select 选择一个节点
func (d *Default) Select(ctx context.Context) (selected Node, done DoneFunc, err error) {
	var (
		candidates []WeightedNode // 候选节点列表
	)
	// 从原子值中加载节点列表
	nodes, ok := d.nodes.Load().([]WeightedNode)
	if !ok {
		return nil, nil, ErrNoAvailable
	}
	candidates = nodes

	// 检查是否有可用节点
	if len(candidates) == 0 {
		return nil, nil, ErrNoAvailable
	}
	// 使用负载均衡器选择节点
	wn, done, err := d.Balancer.Pick(ctx, candidates)
	if err != nil {
		return nil, nil, err
	}
	// 将选中的节点信息存储到上下文中
	p, ok := FromPeerContext(ctx)
	if ok {
		p.Node = wn.Raw()
	}
	return wn.Raw(), done, nil
}

// Apply 更新节点信息
func (d *Default) Apply(nodes []Node) {
	// 将普通节点转换为加权节点
	weightedNodes := make([]WeightedNode, 0, len(nodes))
	for _, n := range nodes {
		weightedNodes = append(weightedNodes, d.NodeBuilder.Build(n))
	}
	// TODO: 不要删除未变化的节点
	d.nodes.Store(weightedNodes)
}

// DefaultBuilder 默认选择器构建器
type DefaultBuilder struct {
	Node     WeightedNodeBuilder // 节点构建器
	Balancer BalancerBuilder     // 负载均衡器构建器
}

// Build 创建选择器实例
func (db *DefaultBuilder) Build() Selector {
	return &Default{
		NodeBuilder: db.Node,             // 设置节点构建器
		Balancer:    db.Balancer.Build(), // 创建负载均衡器实例
	}
}
