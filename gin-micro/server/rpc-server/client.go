package rpcserver

import (
	"context"

	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	grpcinsecure "google.golang.org/grpc/credentials/insecure"

	"emshop-admin/gin-micro/registry"
	"emshop-admin/gin-micro/server/rpc-server/client-interceptors"
	"emshop-admin/pkg/log"
)



type clientOptions struct {
	endpoint string	// 服务地址
	timeout  time.Duration	// 超时时间, 用于设置请求的超时时间
	discovery     registry.Discovery	// 服务发现接口
	unaryInts     []grpc.UnaryClientInterceptor		// Unary拦截器
	streamInts    []grpc.StreamClientInterceptor	// Stream拦截器
	rpcOpts       []grpc.DialOption	// gRPC客户端选项
	balancerName  string
	log           log.LogHelper
	enableTracing bool
	enableMetrics bool
}

type ClientOption func(o *clientOptions)


func WithEnableTracing(enable bool) ClientOption {
	return func(o *clientOptions) {
		o.enableTracing = enable
	}
}

// 设置地址
func WithEndpoint(endpoint string) ClientOption {
	return func(o *clientOptions) {
		o.endpoint = endpoint
	}
}

// 设置超时时间
func WithClientTimeout(timeout time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.timeout = timeout
	}
}

// 设置服务发现
func WithDiscovery(d registry.Discovery) ClientOption {
	return func(o *clientOptions) {
		o.discovery = d
	}
}

// 设置拦截器
func WithClientUnaryInterceptor(in ...grpc.UnaryClientInterceptor) ClientOption {
	return func(o *clientOptions) {
		o.unaryInts = in
	}
}

// 设置stream拦截器
func WithClientStreamInterceptor(in ...grpc.StreamClientInterceptor) ClientOption {
	return func(o *clientOptions) {
		o.streamInts = in
	}
}

// 设置grpc的dial选项
func WithClientOptions(opts ...grpc.DialOption) ClientOption {
	return func(o *clientOptions) {
		o.rpcOpts = opts
	}
}

// 设置负载均衡器
func WithBalancerName(name string) ClientOption {
	return func(o *clientOptions) {
		o.balancerName = name
	}
}

func DialInsecure(ctx context.Context, opts ...ClientOption) (*grpc.ClientConn, error) {
	return dial(ctx, true, opts...)
}

func Dial(ctx context.Context, opts ...ClientOption) (*grpc.ClientConn, error) {
	return dial(ctx, false, opts...)
}

func dial(ctx context.Context, insecure bool, opts ...ClientOption) (*grpc.ClientConn, error) {
	options := clientOptions{
		timeout:       2000 * time.Millisecond,
		balancerName:  "round_robin",
		enableTracing: true,
	}

	for _, o := range opts {
		o(&options)
	}

	//TODO 客户端默认拦截器
	ints := []grpc.UnaryClientInterceptor{
		clientinterceptors.TimeoutInterceptor(options.timeout),
	}
	if options.enableTracing {
		ints = append(ints, otelgrpc.UnaryClientInterceptor())
	}
	// &&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&
	// if options.enableMetrics {
	// 	ints = append(ints, clientinterceptors.PrometheusInterceptor())
	// }

	streamInts := []grpc.StreamClientInterceptor{}

	if len(options.unaryInts) > 0 {
		ints = append(ints, options.unaryInts...)
	}
	if len(options.streamInts) > 0 {
		streamInts = append(streamInts, options.streamInts...)
	}

	grpcOpts := []grpc.DialOption{
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "` + options.balancerName + `"}`),
		grpc.WithChainUnaryInterceptor(ints...),
		grpc.WithChainStreamInterceptor(streamInts...),
	}

	//TODO 服务发现的选项
	// if options.discovery != nil {
	// 	grpcOpts = append(grpcOpts, grpc.WithResolvers(
	// 		discovery.NewBuilder(
	// 			options.discovery,
	// 			discovery.WithInsecure(insecure),
	// 		),
	// 	))
	// }

	if insecure {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(grpcinsecure.NewCredentials()))
	}

	// 用户自定义的gRPC选项
	if len(options.rpcOpts) > 0 {
		grpcOpts = append(grpcOpts, options.rpcOpts...)
	}

	return grpc.DialContext(ctx, options.endpoint, grpcOpts...)
}
