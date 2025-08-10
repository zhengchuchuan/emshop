package selector

import (
	"context"
	"time"
)

// Balancer 是负载均衡器接口
type Balancer interface {
	// Pick 从节点列表中选择一个节点
	Pick(ctx context.Context, nodes []WeightedNode) (selected WeightedNode, done DoneFunc, err error)
}

// BalancerBuilder 构建负载均衡器接口
type BalancerBuilder interface {
	// Build 创建负载均衡器实例
	Build() Balancer
}

// WeightedNode 实时计算调度权重的节点接口
type WeightedNode interface {
	Node

	// Raw 返回原始节点
	Raw() Node

	// Weight 返回运行时计算的权重
	Weight() float64

	// Pick 选择该节点，返回完成回调函数
	Pick() DoneFunc

	// PickElapsed 返回从上次选择以来经过的时间
	PickElapsed() time.Duration
}

// WeightedNodeBuilder 是加权节点构建器接口
type WeightedNodeBuilder interface {
	// Build 根据给定节点构建加权节点
	Build(Node) WeightedNode
}
