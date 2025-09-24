package clientinterceptors

import (
    "context"

    servint "emshop/gin-micro/server/rpc-server/server-interceptors"
    "emshop/pkg/log"

    "github.com/alibaba/sentinel-golang/api"
    "github.com/alibaba/sentinel-golang/core/base"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// UnarySentinelClientInterceptor creates a Sentinel-based unary client interceptor.
// It reuses the server-side SentinelConfig for simplicity.
func UnarySentinelClientInterceptor(cfg *servint.SentinelConfig) grpc.UnaryClientInterceptor {
    if cfg == nil {
        cfg = servint.DefaultSentinelServerConfig("")
    }
    return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
        if cfg.ShouldProtect != nil && !cfg.ShouldProtect(method) {
            return invoker(ctx, method, req, reply, cc, opts...)
        }
        resourceName := buildClientResourceName(cfg, method)
        entry, err := api.Entry(resourceName, api.WithTrafficType(base.Outbound))
        if err != nil {
            log.Warnf("client request blocked by Sentinel: resource=%s, err=%v", resourceName, err)
            if cfg.FallbackFunc != nil {
                // For client, info is not available; pass nil to keep signature
                if _, fbErr := cfg.FallbackFunc(ctx, req, nil, err); fbErr != nil {
                    return fbErr
                }
            }
            return status.Error(codes.ResourceExhausted, "服务调用繁忙，请稍后重试")
        }
        defer entry.Exit()

        callErr := invoker(ctx, method, req, reply, cc, opts...)
        if callErr != nil {
            api.TraceError(entry, callErr)
        }
        return callErr
    }
}

func buildClientResourceName(cfg *servint.SentinelConfig, method string) string {
    if cfg.ResourcePrefix != "" {
        return cfg.ResourcePrefix + ":client:" + method
    }
    return method
}
