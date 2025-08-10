package direct

import (
	"context"
	"sync/atomic"
	"time"

	selector2 "emshop/gin-micro/server/rpc-server/selector"
)

const (
	// defaultWeight 默认权重值
	defaultWeight = 100
)

var (
	// 编译时接口检查
	_ selector2.WeightedNode        = &Node{}
	_ selector2.WeightedNodeBuilder = &Builder{}
)

// Node 直接节点实现，不进行复杂的负载计算
type Node struct {
	selector2.Node          // 嵌入基础节点接口

	// lastPick 最后一次被选择的时间戳
	lastPick int64
}

// Builder 直接节点构建器
type Builder struct{}

// Build 创建节点
func (*Builder) Build(n selector2.Node) selector2.WeightedNode {
	return &Node{Node: n, lastPick: 0}
}

// Pick 选择该节点，返回空的完成回调函数
func (n *Node) Pick() selector2.DoneFunc {
	now := time.Now().UnixNano()
	// 记录选择时间
	atomic.StoreInt64(&n.lastPick, now)
	// 返回空的回调函数，不进行任何统计
	return func(ctx context.Context, di selector2.DoneInfo) {}
}

// Weight 返回节点的有效权重
func (n *Node) Weight() float64 {
	// 尝试获取初始权重
	if n.InitialWeight() != nil {
		return float64(*n.InitialWeight())
	}
	// 如果没有设置初始权重，返回默认值
	return defaultWeight
}

// PickElapsed 返回从上次选择以来经过的时间
func (n *Node) PickElapsed() time.Duration {
	return time.Duration(time.Now().UnixNano() - atomic.LoadInt64(&n.lastPick))
}

// Raw 返回原始节点
func (n *Node) Raw() selector2.Node {
	return n.Node
}
