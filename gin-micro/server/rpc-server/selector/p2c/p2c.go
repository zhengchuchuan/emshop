package p2c

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	selector2 "emshop/gin-micro/server/rpc-server/selector"
	"emshop/gin-micro/server/rpc-server/selector/node/ewma"
)

const (
	// forcePick 强制选择间隔时间，3秒
	forcePick = time.Second * 3
	// Name P2C 负载均衡器的名称
	Name = "p2c"
)

// 编译时接口检查
var _ selector2.Balancer = &Balancer{}

// New 创建一个 P2C 选择器
func New() selector2.Selector {
	return NewBuilder().Build()
}

// Balancer P2C 负载均衡器实现
type Balancer struct {
	mu     sync.Mutex // 互斥锁，保护随机数生成器
	r      *rand.Rand // 随机数生成器
	picked int64      // 强制选择标志，用于原子操作
}

// prePick 选择两个不同的节点
func (s *Balancer) prePick(nodes []selector2.WeightedNode) (nodeA selector2.WeightedNode, nodeB selector2.WeightedNode) {
	s.mu.Lock()
	// 随机选择第一个节点
	a := s.r.Intn(len(nodes))
	// 随机选择第二个节点（排除第一个）
	b := s.r.Intn(len(nodes) - 1)
	s.mu.Unlock()
	// 确保两个节点不同
	if b >= a {
		b = b + 1
	}
	nodeA, nodeB = nodes[a], nodes[b]
	return
}

// Pick 选择一个节点
func (s *Balancer) Pick(ctx context.Context, nodes []selector2.WeightedNode) (selector2.WeightedNode, selector2.DoneFunc, error) {
	if len(nodes) == 0 {
		return nil, nil, selector2.ErrNoAvailable
	}
	// 如果只有一个节点，直接返回
	if len(nodes) == 1 {
		done := nodes[0].Pick()
		return nodes[0], done, nil
	}

	// pc: preferred choice (首选节点)
	// upc: unpreferred choice (非首选节点)
	var pc, upc selector2.WeightedNode
	nodeA, nodeB := s.prePick(nodes)
	// 根据权重选择首选节点，权重由服务发布者在发现中设置
	if nodeB.Weight() > nodeA.Weight() {
		pc, upc = nodeB, nodeA
	} else {
		pc, upc = nodeA, nodeB
	}

	// 如果失败节点在 forceGap 期间从未被选中过，则强制选中一次
	// 利用强制机会触发成功率和延迟的更新
	if upc.PickElapsed() > forcePick && atomic.CompareAndSwapInt64(&s.picked, 0, 1) {
		pc = upc
		atomic.StoreInt64(&s.picked, 0)
	}
	// 获取选中节点的完成回调函数
	done := pc.Pick()
	return pc, done, nil
}

// NewBuilder 返回一个带有 P2C 负载均衡器的选择器构建器
func NewBuilder() selector2.Builder {
	return &selector2.DefaultBuilder{
		Balancer: &Builder{},     // 使用 P2C 负载均衡器
		Node:     &ewma.Builder{}, // 使用 EWMA 节点构建器
	}
}

// Builder P2C 负载均衡器构建器
type Builder struct{}

// Build 创建负载均衡器实例
func (b *Builder) Build() selector2.Balancer {
	// 使用当前时间作为随机种子
	return &Balancer{r: rand.New(rand.NewSource(time.Now().UnixNano()))}
}
