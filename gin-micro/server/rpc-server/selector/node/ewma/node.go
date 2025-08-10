package ewma

import (
	"container/list"
	"context"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	selector2 "emshop/gin-micro/server/rpc-server/selector"
)

const (
	// tau “cost” 的平均生命周期，在 Tau*ln(2) 后达到半衰期
	tau = int64(time.Millisecond * 600)
	// penalty 如果未收集到统计数据，我们会为端点添加一个大的延迟惩罚
	penalty = uint64(time.Second * 10)
)

var (
	// 编译时接口检查
	_ selector2.WeightedNode        = &Node{}
	_ selector2.WeightedNodeBuilder = &Builder{}
)

// Node EWMA 节点实现，基于指数加权移动平均算法
type Node struct {
	selector2.Node // 嵌入基础节点接口

	// 客户端统计数据
	lag       int64       // 平均延迟
	success   uint64      // 成功率（乘以 1000）
	inflight  int64       // 当前并发请求数
	inflights *list.List  // 正在进行的请求列表，存储请求开始时间
	// 最后收集时间戳
	stamp     int64 // 最后收集的时间戳
	predictTs int64 // 最后预测时间戳
	predict   int64 // 预测延迟
	// 一段时间内的请求数量
	reqs int64 // 请求计数器
	// 最后一次选择时间戳
	lastPick int64 // 最后一次被选择的时间

	errHandler func(err error) (isErr bool) // 错误处理函数
	lk         sync.RWMutex                  // 读写锁，保护并发安全
}

// Builder EWMA 节点构建器
type Builder struct {
	// ErrHandler 错误处理函数，用于判断是否为错误
	ErrHandler func(err error) (isErr bool)
}

// Build 创建一个加权节点
func (b *Builder) Build(n selector2.Node) selector2.WeightedNode {
	s := &Node{
		Node:       n,             // 基础节点
		lag:        0,             // 初始延迟为 0
		success:    1000,          // 初始成功率为 100%（*1000）
		inflight:   1,             // 初始并发数为 1
		inflights:  list.New(),    // 初始化请求列表
		errHandler: b.ErrHandler, // 设置错误处理函数
	}
	return s
}

// health 返回节点的健康度（成功率）
func (n *Node) health() uint64 {
	return atomic.LoadUint64(&n.success)
}

// load 计算节点的当前负载
func (n *Node) load() (load uint64) {
	now := time.Now().UnixNano()
	avgLag := atomic.LoadInt64(&n.lag)      // 获取平均延迟
	lastPredictTs := atomic.LoadInt64(&n.predictTs) // 获取最后预测时间
	// 计算预测间隔，基于平均延迟的 1/5
	predictInterval := avgLag / 5
	// 限制预测间隔在 5ms-200ms 之间
	if predictInterval < int64(time.Millisecond*5) {
		predictInterval = int64(time.Millisecond * 5)
	} else if predictInterval > int64(time.Millisecond*200) {
		predictInterval = int64(time.Millisecond * 200)
	}
	// 如果距离上次预测超过间隔时间，则进行新的预测
	if now-lastPredictTs > predictInterval {
		// 使用 CAS 操作确保只有一个 goroutine 进行预测
		if atomic.CompareAndSwapInt64(&n.predictTs, lastPredictTs, now) {
			var (
				total   int64 // 总延迟
				count   int   // 计数器
				predict int64 // 预测值
			)
			// 遍历当前正在进行的请求
			n.lk.RLock()
			first := n.inflights.Front()
			for first != nil {
				lag := now - first.Value.(int64) // 计算当前请求的延迟
				// 只统计超过平均延迟的请求
				if lag > avgLag {
					count++
					total += lag
				}
				first = first.Next()
			}
			// 如果超过一半的请求都超过平均延迟，则更新预测值
			if count > (n.inflights.Len()/2 + 1) {
				predict = total / int64(count)
			}
			n.lk.RUnlock()
			atomic.StoreInt64(&n.predict, predict) // 存储预测值
		}
	}

	// 如果平均延迟为 0（节点刚启动时没有数据）
	if avgLag == 0 {
		// penalty 是节点刚启动时没有数据时的惩罚值
		// 默认值为 1e9 * 10 = 10秒
		load = penalty * uint64(atomic.LoadInt64(&n.inflight))
	} else {
		// 获取预测延迟
		predict := atomic.LoadInt64(&n.predict)
		// 如果预测延迟大于平均延迟，使用预测值
		if predict > avgLag {
			avgLag = predict
		}
		// 负载 = 平均延迟 * 当前并发数
		load = uint64(avgLag) * uint64(atomic.LoadInt64(&n.inflight))
	}
	return
}

// Pick 选择该节点，返回完成回调函数
func (n *Node) Pick() selector2.DoneFunc {
	now := time.Now().UnixNano()
	atomic.StoreInt64(&n.lastPick, now)   // 记录最后选择时间
	atomic.AddInt64(&n.inflight, 1)       // 增加并发计数
	atomic.AddInt64(&n.reqs, 1)           // 增加请求计数
	n.lk.Lock()
	e := n.inflights.PushBack(now)       // 将请求开始时间加入列表
	n.lk.Unlock()
	// 返回完成回调函数
	return func(ctx context.Context, di selector2.DoneInfo) {
		// 从正在进行的请求列表中移除
		n.lk.Lock()
		n.inflights.Remove(e)
		n.lk.Unlock()
		// 减少并发计数
		atomic.AddInt64(&n.inflight, -1)

		now := time.Now().UnixNano()
		// 获取移动平均比率 w
		stamp := atomic.SwapInt64(&n.stamp, now)
		td := now - stamp // 时间差
		if td < 0 {
			td = 0
		}
		// 计算指数衰减因子
		w := math.Exp(float64(-td) / float64(tau))

		// 计算本次请求的延迟
		start := e.Value.(int64)
		lag := now - start
		if lag < 0 {
			lag = 0
		}
		// 获取旧的平均延迟
		oldLag := atomic.LoadInt64(&n.lag)
		if oldLag == 0 {
			// 如果是第一次请求，直接使用当前值
			w = 0.0
		}
		// 使用 EWMA 算法更新平均延迟
		lag = int64(float64(oldLag)*w + float64(lag)*(1.0-w))
		atomic.StoreInt64(&n.lag, lag)

		// 默认成功率为 100%（*1000）
		success := uint64(1000)
		// 检查是否有错误
		if di.Err != nil {
			if n.errHandler != nil {
				// 使用自定义错误处理函数
				if n.errHandler(di.Err) {
					success = 0 // 设置为失败
				}
			} else if errors.Is(context.DeadlineExceeded, di.Err) || errors.Is(context.Canceled, di.Err) ||
				errors.IsServiceUnavailable(di.Err) || errors.IsGatewayTimeout(di.Err) {
				// 对于特定的错误类型，设置为失败
				success = 0
			}
		}
		// 使用 EWMA 算法更新成功率
		oldSuc := atomic.LoadUint64(&n.success)
		success = uint64(float64(oldSuc)*w + float64(success)*(1.0-w))
		atomic.StoreUint64(&n.success, success)
	}
}

// Weight 返回节点的有效权重
func (n *Node) Weight() (weight float64) {
	// 权重 = 健康度 * 时间单位 / 负载
	weight = float64(n.health()*uint64(time.Second)) / float64(n.load())
	return
}

// PickElapsed 返回从上次选择以来经过的时间
func (n *Node) PickElapsed() time.Duration {
	return time.Duration(time.Now().UnixNano() - atomic.LoadInt64(&n.lastPick))
}

// Raw 返回原始节点
func (n *Node) Raw() selector2.Node {
	return n.Node
}
