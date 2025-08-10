package wrr

import (
	"context"
	"sync"

	selector2 "emshop/gin-micro/server/rpc-server/selector"
	"emshop/gin-micro/server/rpc-server/selector/node/direct"
)

const (
	// Name 加权轮询负载均衡器的名称
	Name = "wrr"
)

// 编译时接口检查
var _ selector2.Balancer = &Balancer{}

// Balancer 加权轮询负载均衡器实现
type Balancer struct {
	mu            sync.Mutex           // 互斥锁，保护并发安全
	currentWeight map[string]float64   // 当前权重映射，键为节点地址
}

// New 创建一个加权轮询选择器
func New() selector2.Selector {
	return NewBuilder().Build()
}

// Pick 从加权节点中选择一个
func (p *Balancer) Pick(_ context.Context, nodes []selector2.WeightedNode) (selector2.WeightedNode, selector2.DoneFunc, error) {
	if len(nodes) == 0 {
		return nil, nil, selector2.ErrNoAvailable
	}
	var totalWeight float64              // 总权重
	var selected selector2.WeightedNode  // 选中的节点
	var selectWeight float64             // 选中节点的当前权重

	// 基于 Nginx WRR 负载均衡算法实现
	// 参考: http://blog.csdn.net/zhangskd/article/details/50194069
	p.mu.Lock()
	for _, node := range nodes {
		// 累计总权重
		totalWeight += node.Weight()
		// 获取节点当前权重
		cwt := p.currentWeight[node.Address()]
		// 当前权重 += 有效权重
		cwt += node.Weight()
		p.currentWeight[node.Address()] = cwt
		// 选择当前权重最大的节点
		if selected == nil || selectWeight < cwt {
			selectWeight = cwt
			selected = node
		}
	}
	// 选中节点的当前权重减去总权重
	p.currentWeight[selected.Address()] = selectWeight - totalWeight
	p.mu.Unlock()

	// 获取节点的完成回调函数
	d := selected.Pick()
	return selected, d, nil
}

// NewBuilder 返回一个带有加权轮询负载均衡器的选择器构建器
func NewBuilder() selector2.Builder {
	return &selector2.DefaultBuilder{
		Balancer: &Builder{},       // 使用 WRR 负载均衡器
		Node:     &direct.Builder{}, // 使用直接节点构建器
	}
}

// Builder 加权轮询负载均衡器构建器
type Builder struct{}

// Build 创建负载均衡器实例
func (b *Builder) Build() selector2.Balancer {
	return &Balancer{currentWeight: make(map[string]float64)}
}
