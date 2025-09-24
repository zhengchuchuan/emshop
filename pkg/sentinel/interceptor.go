package sentinel

import (
    clientints "emshop/gin-micro/server/rpc-server/client-interceptors"
    serverints "emshop/gin-micro/server/rpc-server/server-interceptors"
    "google.golang.org/grpc"
)

// InterceptorConfig 是对 gin-micro Sentinel 配置的别名，保留原有 API。
type InterceptorConfig = serverints.SentinelConfig

// DefaultServerInterceptorConfig 兼容旧函数，转调到 gin-micro。
func DefaultServerInterceptorConfig(serviceName string) *InterceptorConfig {
    return serverints.DefaultSentinelServerConfig(serviceName)
}

// DefaultClientInterceptorConfig 兼容旧函数，保持原语义。
func DefaultClientInterceptorConfig(serviceName string) *InterceptorConfig {
    cfg := serverints.DefaultSentinelServerConfig(serviceName)
    cfg.IncludeServiceName = true
    cfg.EnableMetrics = false
    return cfg
}

// NewUnaryServerInterceptor 转调到 gin-micro 的实现。
func NewUnaryServerInterceptor(config *InterceptorConfig) grpc.UnaryServerInterceptor {
    return serverints.UnarySentinelInterceptor(config)
}

// NewStreamServerInterceptor 转调到 gin-micro 的实现。
func NewStreamServerInterceptor(config *InterceptorConfig) grpc.StreamServerInterceptor {
    return serverints.StreamSentinelInterceptor(config)
}

// NewUnaryClientInterceptor 转调到 gin-micro 的实现（客户端）。
func NewUnaryClientInterceptor(config *InterceptorConfig) grpc.UnaryClientInterceptor {
    return clientints.UnarySentinelClientInterceptor(config)
}

// DefaultFallbackFunc 已在各业务层自定义（见 fallback.go），若仍需通用可在业务处引用。
// 为保持兼容，这里不再保留内部默认实现；如确需可在调用处传入自定义 FallbackFunc。
