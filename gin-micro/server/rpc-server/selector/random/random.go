package random

import (
	"context"
	"math/rand"

	selector2 "emshop/gin-micro/server/rpc-server/selector"
	"emshop/gin-micro/server/rpc-server/selector/node/direct"
)

const (
	// Name 随机负载均衡器的名称
	Name = "random"
)

// 编译时接口检查
var _ selector2.Balancer = &Balancer{}

// Balancer 随机负载均衡器实现
type Balancer struct{}

// New 创建一个随机选择器
func New() selector2.Selector {
	return NewBuilder().Build()
}

// Pick 从加权节点中随机选择一个
func (p *Balancer) Pick(_ context.Context, nodes []selector2.WeightedNode) (selector2.WeightedNode, selector2.DoneFunc, error) {
	if len(nodes) == 0 {
		return nil, nil, selector2.ErrNoAvailable
	}
	// 随机选择一个节点索引
	cur := rand.Intn(len(nodes))
	selected := nodes[cur]
	// 获取节点的完成回调函数
	d := selected.Pick()
	return selected, d, nil
}

// NewBuilder 返回一个带有随机负载均衡器的选择器构建器
func NewBuilder() selector2.Builder {
	return &selector2.DefaultBuilder{
		Balancer: &Builder{},       // 使用随机负载均衡器
		Node:     &direct.Builder{}, // 使用直接节点构建器
	}
}

// Builder 随机负载均衡器构建器
type Builder struct{}

// Build 创建负载均衡器实例
func (b *Builder) Build() selector2.Balancer {
	return &Balancer{}
}
