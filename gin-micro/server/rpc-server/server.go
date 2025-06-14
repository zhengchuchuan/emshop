package rpcserver

import (
	"context"
	"net"
	"net/url"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	apimd "emshop/api/metadata"
	srvintc "emshop/gin-micro/server/rpc-server/server-interceptors"
	"emshop/pkg/host"
	"emshop/pkg/log"
)

type Server struct {
	*grpc.Server

	address string 	// 服务地址
	unaryInts  []grpc.UnaryServerInterceptor	// Unary拦截器
	streamInts []grpc.StreamServerInterceptor	// Stream拦截器
	grpcOpts   []grpc.ServerOption				// gRPC服务器选项
	lis        net.Listener						// 监听器

	timeout    time.Duration					// 超时时间, 用于设置请求的超时时间

	health   	*health.Server					// 健康检查服务
	metadata *apimd.Server						// 元数据服务
	endpoint 	*url.URL						// 服务地址

	enableMetrics bool							// 是否开启prometheus 
}

// 函数选项模式
type ServerOption func(o *Server)


func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		address: ":0",
		health:  health.NewServer(),
		timeout: 1 * time.Second,
	}
	// 根据传入函数设置参数
	for _, o := range opts {
		o(srv)
	}

	//不设置拦截器的情况下，自动默认加上一些必须的拦截器，如 crash，tracing
	unaryInts := []grpc.UnaryServerInterceptor{
		srvintc.UnaryCrashInterceptor,
		otelgrpc.UnaryServerInterceptor(),
	}
	// 如果用户传入了自定义的Unary拦截器，则添加到unaryInts中

	// prometheus拦截器
	if srv.enableMetrics {
		unaryInts = append(unaryInts, srvintc.UnaryPrometheusInterceptor)
	}

	if srv.timeout > 0 {
		unaryInts = append(unaryInts, srvintc.UnaryTimeoutInterceptor(srv.timeout))
	}

	if len(srv.unaryInts) > 0 {
		unaryInts = append(unaryInts, srv.unaryInts...)
	}

	//把传入的拦截器转换成grpc的ServerOption
	grpcOpts := []grpc.ServerOption{grpc.ChainUnaryInterceptor(srv.unaryInts...)}

	//把用户自己传入的grpc.ServerOption放在一起
	srv.Server = grpc.NewServer(grpcOpts...)

	//注册metadata的Server
	srv.metadata = apimd.NewServer(srv.Server)

	//解析address
	err := srv.listenAndEndpoint()
	if err != nil {
		panic(err)
	}

	// 注册健康检查服务
	grpc_health_v1.RegisterHealthServer(srv.Server, srv.health)

	// 直接使用kratos的metadata服务
	// 这个服务会自动注册到grpc的Server中
	// 可以支持用户直接通过grpc的一个接口查看当前支持的所有的rpc服务
	apimd.RegisterMetadataServer(srv.Server, srv.metadata)
	// 注册反射服务,允许客户端（如 grpcurl、Postman、grpcui 等工具）在不知道 proto 文件的情况下，动态查询服务支持的所有 RPC 方法和消息类型
	reflection.Register(srv.Server)

	return srv
}




func WithAddress(address string) ServerOption {
	return func(s *Server) {
		s.address = address
	}
}

func WithMetrics(metric bool) ServerOption {
	return func(s *Server) {
		s.enableMetrics = metric
	}
}

func WithTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.timeout = timeout
	}
}

func WithLis(lis net.Listener) ServerOption {
	return func(s *Server) {
		s.lis = lis
	}
}

func WithUnaryInterceptor(in ...grpc.UnaryServerInterceptor) ServerOption {
	return func(s *Server) {
		s.unaryInts = in
	}
}

func WithStreamInterceptor(in ...grpc.StreamServerInterceptor) ServerOption {
	return func(s *Server) {
		s.streamInts = in
	}
}

func WithOptions(opts ...grpc.ServerOption) ServerOption {
	return func(s *Server) {
		s.grpcOpts = opts
	}
}

// 完成ip和端口的提取
func (s *Server) listenAndEndpoint() error {
	if s.lis == nil {
		lis, err := net.Listen("tcp", s.address)
		if err != nil {
			return err
		}
		s.lis = lis
	}
	addr, err := host.Extract(s.address, s.lis)
	if err != nil {
		_ = s.lis.Close()
		return err
	}
	s.endpoint = &url.URL{Scheme: "grpc", Host: addr}
	return nil
}

func (s *Server) Endpoint() *url.URL {
	return s.endpoint
}

func (s *Server) Address() string {
	return s.address
}


// 启动grpc的服务
func (s *Server) Start(ctx context.Context) error {
	log.Infof("[grpc] server listening on: %s", s.lis.Addr().String())
	s.health.Resume()
	return s.Serve(s.lis)
}

func (s *Server) Stop(ctx context.Context) error {
	//设置服务的状态为not_serving，防止接收新的请求过来
	s.health.Shutdown()
	s.GracefulStop()
	log.Infof("[grpc] server stopped")
	return nil
}
