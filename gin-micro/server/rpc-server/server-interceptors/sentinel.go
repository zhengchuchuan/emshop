package serverinterceptors

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

// SentinelConfig holds the interceptor configuration for Sentinel.
type SentinelConfig struct {
    // ResourcePrefix will be prepended to the resource name (e.g. service name).
    ResourcePrefix string
    // IncludeMethodName indicates whether to include method name in the resource.
    IncludeMethodName bool
    // IncludeServiceName indicates whether to include service name in the resource.
    IncludeServiceName bool
    // FallbackFunc handles downgrade when blocked; if nil, returns ResourceExhausted.
    FallbackFunc func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, err error) (interface{}, error)
    // EnableMetrics enables Prometheus metrics recording.
    EnableMetrics bool
    // ShouldProtect decides whether to apply Sentinel for a given full method name.
    // If nil, protection is applied to all methods.
    ShouldProtect func(fullMethod string) bool
}

// internal metrics
type sentinelMetrics struct {
    blockedCounter   prometheus.Counter
    passedCounter    prometheus.Counter
    totalCounter     prometheus.Counter
    latencyHistogram prometheus.Histogram
}

var (
    globalSentinelMetrics *sentinelMetrics
    metricsOnce           sync.Once
)

func initSentinelMetrics(serviceName string) {
    metricsOnce.Do(func() {
        safeName := strings.ReplaceAll(serviceName, "-", "_")
        globalSentinelMetrics = &sentinelMetrics{
            blockedCounter: prometheus.NewCounter(prometheus.CounterOpts{
                Name: fmt.Sprintf("%s_sentinel_blocked_total", safeName),
                Help: "Total number of blocked requests by Sentinel",
            }),
            passedCounter: prometheus.NewCounter(prometheus.CounterOpts{
                Name: fmt.Sprintf("%s_sentinel_passed_total", safeName),
                Help: "Total number of passed requests by Sentinel",
            }),
            totalCounter: prometheus.NewCounter(prometheus.CounterOpts{
                Name: fmt.Sprintf("%s_sentinel_requests_total", safeName),
                Help: "Total number of requests processed by Sentinel",
            }),
            latencyHistogram: prometheus.NewHistogram(prometheus.HistogramOpts{
                Name:    fmt.Sprintf("%s_sentinel_duration_seconds", safeName),
                Help:    "Request duration processed by Sentinel",
                Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
            }),
        }

        register := func(c prometheus.Collector) {
            if err := prometheus.Register(c); err != nil {
                if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
                    log.Errorf("register Prometheus metric failed: %v", err)
                }
            }
        }

        register(globalSentinelMetrics.blockedCounter)
        register(globalSentinelMetrics.passedCounter)
        register(globalSentinelMetrics.totalCounter)
        register(globalSentinelMetrics.latencyHistogram)

        log.Infof("Sentinel metrics initialized: service=%s", serviceName)
    })
}

// DefaultSentinelServerConfig returns a reasonable default server config.
func DefaultSentinelServerConfig(serviceName string) *SentinelConfig {
    return &SentinelConfig{
        ResourcePrefix:     serviceName,
        IncludeMethodName:  true,
        IncludeServiceName: false,
        EnableMetrics:      true,
        FallbackFunc:       DefaultFallback,
    }
}

// DefaultFallback is a generic downgrade function.
func DefaultFallback(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, err error) (interface{}, error) {
    // treat all block errors as ResourceExhausted
    return nil, status.Error(codes.ResourceExhausted, "系统繁忙，请稍后重试")
}

// UnarySentinelInterceptor creates a Sentinel-based unary server interceptor.
func UnarySentinelInterceptor(cfg *SentinelConfig) grpc.UnaryServerInterceptor {
    if cfg == nil {
        cfg = &SentinelConfig{}
    }
    if cfg.EnableMetrics && globalSentinelMetrics == nil {
        initSentinelMetrics(cfg.ResourcePrefix)
    }

    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        // allow selective protection
        if cfg.ShouldProtect != nil && !cfg.ShouldProtect(info.FullMethod) {
            return handler(ctx, req)
        }
        resourceName := buildResourceName(cfg, info.FullMethod)

        if globalSentinelMetrics != nil {
            globalSentinelMetrics.totalCounter.Inc()
        }
        start := time.Now()

        entry, err := api.Entry(resourceName, api.WithTrafficType(base.Inbound))
        if err != nil {
            if globalSentinelMetrics != nil {
                globalSentinelMetrics.blockedCounter.Inc()
            }
            log.Warnf("request blocked by Sentinel: resource=%s, err=%v", resourceName, err)
            if cfg.FallbackFunc != nil {
                return cfg.FallbackFunc(ctx, req, info, err)
            }
            return nil, status.Error(codes.ResourceExhausted, "服务繁忙，请稍后重试")
        }
        defer func() {
            entry.Exit()
            if globalSentinelMetrics != nil {
                globalSentinelMetrics.latencyHistogram.Observe(time.Since(start).Seconds())
            }
        }()

        if globalSentinelMetrics != nil {
            globalSentinelMetrics.passedCounter.Inc()
        }

        resp, hErr := handler(ctx, req)
        if hErr != nil {
            api.TraceError(entry, hErr)
        }
        return resp, hErr
    }
}

// StreamSentinelInterceptor creates a Sentinel-based stream server interceptor.
func StreamSentinelInterceptor(cfg *SentinelConfig) grpc.StreamServerInterceptor {
    if cfg == nil {
        cfg = &SentinelConfig{}
    }
    return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
        if cfg.ShouldProtect != nil && !cfg.ShouldProtect(info.FullMethod) {
            return handler(srv, stream)
        }
        resourceName := buildResourceName(cfg, info.FullMethod)
        entry, err := api.Entry(resourceName, api.WithTrafficType(base.Inbound))
        if err != nil {
            log.Warnf("stream blocked by Sentinel: resource=%s, err=%v", resourceName, err)
            return status.Error(codes.ResourceExhausted, "服务繁忙，请稍后重试")
        }
        defer entry.Exit()
        hErr := handler(srv, stream)
        if hErr != nil {
            api.TraceError(entry, hErr)
        }
        return hErr
    }
}

func buildResourceName(cfg *SentinelConfig, fullMethod string) string {
    parts := strings.Split(fullMethod, "/")
    if len(parts) < 3 {
        return fullMethod
    }
    serviceName := parts[1]
    methodName := parts[2]

    var b strings.Builder
    if cfg.ResourcePrefix != "" {
        b.WriteString(cfg.ResourcePrefix)
        b.WriteString(":")
    }
    if cfg.IncludeServiceName {
        b.WriteString(serviceName)
        if cfg.IncludeMethodName {
            b.WriteString(".")
        }
    }
    if cfg.IncludeMethodName {
        b.WriteString(methodName)
    }
    if !cfg.IncludeServiceName && !cfg.IncludeMethodName {
        return fullMethod
    }
    return b.String()
}
