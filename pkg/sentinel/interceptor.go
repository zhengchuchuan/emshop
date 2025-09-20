package sentinel

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"emshop/pkg/log"

	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// InterceptorConfig 拦截器配置
type InterceptorConfig struct {
	// 是否启用资源名称前缀
	ResourcePrefix string
	// 是否包含方法名
	IncludeMethodName bool
	// 是否包含服务名
	IncludeServiceName bool
	// 降级处理函数
	FallbackFunc func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, err error) (interface{}, error)
	// 监控指标
	EnableMetrics bool
}

// Metrics 监控指标
type Metrics struct {
	blockedCounter   prometheus.Counter
	passedCounter    prometheus.Counter
	totalCounter     prometheus.Counter
	latencyHistogram prometheus.Histogram
}

var (
	globalMetrics *Metrics
	metricsOnce   sync.Once
)

// initMetrics 初始化监控指标
func initMetrics(serviceName string) {
	metricsOnce.Do(func() {
		globalMetrics = &Metrics{
			blockedCounter: prometheus.NewCounter(prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_sentinel_blocked_total", strings.ReplaceAll(serviceName, "-", "_")),
				Help: "Total number of blocked requests by Sentinel",
			}),
			passedCounter: prometheus.NewCounter(prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_sentinel_passed_total", strings.ReplaceAll(serviceName, "-", "_")),
				Help: "Total number of passed requests by Sentinel",
			}),
			totalCounter: prometheus.NewCounter(prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_sentinel_requests_total", strings.ReplaceAll(serviceName, "-", "_")),
				Help: "Total number of requests processed by Sentinel",
			}),
			latencyHistogram: prometheus.NewHistogram(prometheus.HistogramOpts{
				Name:    fmt.Sprintf("%s_sentinel_duration_seconds", strings.ReplaceAll(serviceName, "-", "_")),
				Help:    "Request duration processed by Sentinel",
				Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			}),
		}

		// 安全注册指标，处理重复注册的情况
		registerMetric := func(c prometheus.Collector) {
			if err := prometheus.Register(c); err != nil {
				if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
					log.Errorf("注册Prometheus指标失败: %v", err)
				}
			}
		}

		registerMetric(globalMetrics.blockedCounter)
		registerMetric(globalMetrics.passedCounter)
		registerMetric(globalMetrics.totalCounter)
		registerMetric(globalMetrics.latencyHistogram)

		log.Infof("Sentinel监控指标初始化完成: service=%s", serviceName)
	})
}

// NewUnaryServerInterceptor 创建Unary Server拦截器
func NewUnaryServerInterceptor(config *InterceptorConfig) grpc.UnaryServerInterceptor {
	if config == nil {
		config = &InterceptorConfig{}
	}

	// 初始化监控指标
	if config.EnableMetrics && globalMetrics == nil {
		initMetrics(config.ResourcePrefix)
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 构建资源名称
		resourceName := buildResourceName(config, info.FullMethod)

		if globalMetrics != nil {
			globalMetrics.totalCounter.Inc()
		}

		// 记录开始时间
		startTime := time.Now()

		// 创建Sentinel Entry
		entry, err := api.Entry(resourceName, api.WithTrafficType(base.Inbound))
		if err != nil {
			// 被限流/熔断
			if globalMetrics != nil {
				globalMetrics.blockedCounter.Inc()
			}

			log.Warnf("请求被Sentinel阻断: resource=%s, err=%v", resourceName, err)

			// 执行降级逻辑
			if config.FallbackFunc != nil {
				return config.FallbackFunc(ctx, req, info, err)
			}

			// 返回默认的限流错误
			return nil, status.Error(codes.ResourceExhausted, "服务繁忙，请稍后重试")
		}

		defer func() {
			entry.Exit()

			// 记录延迟
			if globalMetrics != nil {
				globalMetrics.latencyHistogram.Observe(time.Since(startTime).Seconds())
			}
		}()

		if globalMetrics != nil {
			globalMetrics.passedCounter.Inc()
		}

		// 调用实际处理器
		response, handlerErr := handler(ctx, req)

		// 记录异常到Sentinel
		if handlerErr != nil {
			api.TraceError(entry, handlerErr)
		}

		return response, handlerErr
	}
}

// NewUnaryClientInterceptor 创建Unary Client拦截器
func NewUnaryClientInterceptor(config *InterceptorConfig) grpc.UnaryClientInterceptor {
	if config == nil {
		config = &InterceptorConfig{}
	}

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// 构建资源名称
		resourceName := buildClientResourceName(config, method)

		// 创建Sentinel Entry
		entry, err := api.Entry(resourceName, api.WithTrafficType(base.Outbound))
		if err != nil {
			log.Warnf("客户端请求被Sentinel阻断: resource=%s, err=%v", resourceName, err)

			// 执行降级逻辑
			if config.FallbackFunc != nil {
				_, fallbackErr := config.FallbackFunc(ctx, req, nil, err)
				if fallbackErr != nil {
					return fallbackErr
				}
				// 如果降级成功，返回原始的限流错误
			}

			return status.Error(codes.ResourceExhausted, "服务调用繁忙，请稍后重试")
		}

		defer entry.Exit()

		// 调用实际方法
		callErr := invoker(ctx, method, req, reply, cc, opts...)

		// 记录异常到Sentinel
		if callErr != nil {
			api.TraceError(entry, callErr)
		}

		return callErr
	}
}

// NewStreamServerInterceptor 创建Stream Server拦截器
func NewStreamServerInterceptor(config *InterceptorConfig) grpc.StreamServerInterceptor {
	if config == nil {
		config = &InterceptorConfig{}
	}

	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// 构建资源名称
		resourceName := buildResourceName(config, info.FullMethod)

		// 创建Sentinel Entry
		entry, err := api.Entry(resourceName, api.WithTrafficType(base.Inbound))
		if err != nil {
			log.Warnf("流式请求被Sentinel阻断: resource=%s, err=%v", resourceName, err)
			return status.Error(codes.ResourceExhausted, "服务繁忙，请稍后重试")
		}

		defer entry.Exit()

		// 调用实际处理器
		handlerErr := handler(srv, stream)

		// 记录异常到Sentinel
		if handlerErr != nil {
			api.TraceError(entry, handlerErr)
		}

		return handlerErr
	}
}

// buildResourceName 构建资源名称
func buildResourceName(config *InterceptorConfig, fullMethod string) string {
	parts := strings.Split(fullMethod, "/")
	if len(parts) < 3 {
		return fullMethod
	}

	serviceName := parts[1]
	methodName := parts[2]

	var resourceName strings.Builder

	// 添加前缀
	if config.ResourcePrefix != "" {
		resourceName.WriteString(config.ResourcePrefix)
		resourceName.WriteString(":")
	}

	// 添加服务名
	if config.IncludeServiceName {
		resourceName.WriteString(serviceName)
		if config.IncludeMethodName {
			resourceName.WriteString(".")
		}
	}

	// 添加方法名
	if config.IncludeMethodName {
		resourceName.WriteString(methodName)
	}

	// 如果都不包含，使用完整方法名
	if !config.IncludeServiceName && !config.IncludeMethodName {
		return fullMethod
	}

	return resourceName.String()
}

// buildClientResourceName 构建客户端资源名称
func buildClientResourceName(config *InterceptorConfig, method string) string {
	resourceName := method
	if config.ResourcePrefix != "" {
		resourceName = fmt.Sprintf("%s:client:%s", config.ResourcePrefix, method)
	}
	return resourceName
}

// DefaultServerInterceptorConfig 默认服务端拦截器配置
func DefaultServerInterceptorConfig(serviceName string) *InterceptorConfig {
	return &InterceptorConfig{
		ResourcePrefix:     serviceName,
		IncludeMethodName:  true,
		IncludeServiceName: false,
		EnableMetrics:      true,
		FallbackFunc:       DefaultFallbackFunc,
	}
}

// DefaultClientInterceptorConfig 默认客户端拦截器配置
func DefaultClientInterceptorConfig(serviceName string) *InterceptorConfig {
	return &InterceptorConfig{
		ResourcePrefix:     serviceName,
		IncludeMethodName:  true,
		IncludeServiceName: true,
		EnableMetrics:      false,
		FallbackFunc:       nil,
	}
}

// DefaultFallbackFunc 默认降级函数
func DefaultFallbackFunc(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, err error) (interface{}, error) {
	log.Warnf("执行默认降级逻辑: method=%s, err=%v", info.FullMethod, err)

	// 可以根据不同的错误类型返回不同的降级响应
	if err != nil {
		// 检查是否是Sentinel的阻断错误
		if blockErr, ok := err.(*base.BlockError); ok {
			log.Warnf("Sentinel阻断: type=%v, resource=%s", blockErr.BlockType(), blockErr.TriggeredRule())
			return nil, status.Error(codes.ResourceExhausted, "系统繁忙，请稍后重试")
		}
		
		// 检查其他类型的Sentinel错误
		switch err.(type) {
		case *base.BlockError:
			return nil, status.Error(codes.ResourceExhausted, "系统繁忙，请稍后重试")
		}
	}

	// 其他错误直接返回
	return nil, err
}