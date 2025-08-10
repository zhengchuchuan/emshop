package selector

import (
	"context"

	"emshop/pkg/errors"
)

// ErrNoAvailable 表示没有可用节点错误
var ErrNoAvailable = errors.New("no_available_node")

// Selector 是节点选择负载均衡器接口
type Selector interface {
	Rebalancer

	// Select 选择节点
	// 如果 err == nil，则 selected 和 done 必须不为空
	Select(ctx context.Context) (selected Node, done DoneFunc, err error)
}

// Rebalancer 是节点重平衡器接口
type Rebalancer interface {
	// Apply 当节点发生任何变化时应用所有节点
	Apply(nodes []Node)
}

// Builder 构建选择器接口
type Builder interface {
	Build() Selector
}

// Node 是节点接口
type Node interface {
	// Scheme 返回服务节点协议
	Scheme() string

	// Address 返回同一服务下的唯一地址
	Address() string

	// ServiceName 返回服务名称
	ServiceName() string

	// InitialWeight 返回调度权重的初始值
	// 如果未设置则返回 nil
	InitialWeight() *int64

	// Version 返回服务节点版本
	Version() string

	// Metadata 返回与服务实例关联的键值对元数据
	// 包括版本、命名空间、区域、协议等
	Metadata() map[string]string
}

// DoneInfo 是 RPC 调用完成时的回调信息
type DoneInfo struct {
	// Err 响应错误
	Err error
	// ReplyMD 响应元数据
	ReplyMD ReplyMD

	// BytesSent 表示是否有字节发送到服务器
	BytesSent bool
	// BytesReceived 表示是否从服务器接收到字节
	BytesReceived bool
}

// ReplyMD 是响应元数据接口
type ReplyMD interface {
	Get(key string) string
}

// DoneFunc 是 RPC 调用完成时的回调函数
type DoneFunc func(ctx context.Context, di DoneInfo)
