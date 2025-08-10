package selector

import (
	"strconv"

	"emshop/gin-micro/registry"
)

// DefaultNode 默认节点实现
type DefaultNode struct {
	scheme   string            // 协议方案（如 grpc、http 等）
	addr     string            // 节点地址
	weight   *int64            // 初始权重
	version  string            // 服务版本
	name     string            // 服务名称
	metadata map[string]string // 元数据
}

// Scheme 返回节点协议方案
func (n *DefaultNode) Scheme() string {
	return n.scheme
}

// Address 返回节点地址
func (n *DefaultNode) Address() string {
	return n.addr
}

// ServiceName 返回服务名称
func (n *DefaultNode) ServiceName() string {
	return n.name
}

// InitialWeight 返回节点初始权重
func (n *DefaultNode) InitialWeight() *int64 {
	return n.weight
}

// Version 返回节点版本
func (n *DefaultNode) Version() string {
	return n.version
}

// Metadata 返回节点元数据
func (n *DefaultNode) Metadata() map[string]string {
	return n.metadata
}

// NewNode 创建新节点
func NewNode(scheme, addr string, ins *registry.ServiceInstance) Node {
	n := &DefaultNode{
		scheme: scheme, // 设置协议方案
		addr:   addr,   // 设置地址
	}
	// 如果提供了服务实例信息，则填充相关属性
	if ins != nil {
		n.name = ins.Name         // 设置服务名称
		n.version = ins.Version   // 设置版本
		n.metadata = ins.Metadata // 设置元数据
		// 尝试从元数据中获取权重信息
		if str, ok := ins.Metadata["weight"]; ok {
			if weight, err := strconv.ParseInt(str, 10, 64); err == nil {
				n.weight = &weight // 设置权重
			}
		}
	}
	return n
}
