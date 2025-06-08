package rpcserver

import (
	"net"
	"net/url"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"

	"emshop-admin/pkg/host"
)

type Server struct {
	*grpc.Server

	address string 	// 服务地址
	unaryInts  []grpc.UnaryServerInterceptor	// Unary拦截器
	streamInts []grpc.StreamServerInterceptor	// Stream拦截器
	grpcOpts   []grpc.ServerOption				// gRPC服务器选项
	lis        net.Listener						// 监听器

	health   	*health.Server					// 健康检查服务
	endpoint 	*url.URL						// 服务地址
}

// 函数选项模式
type ServerOption func(o *Server)


func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		address: ":0",
		health:  health.NewServer(),
		//timeout: 1 * time.Second,
	}
	// 根据传入函数设置参数
	for _, o := range opts {
		o(srv)
	}

	//不设置拦截器的情况下，自动默认加上一些必须的拦截器，如 crash，tracing
	// unaryInts := []grpc.UnaryServerInterceptor{
	// 	srvintc.UnaryCrashInterceptor,
	// 	otelgrpc.UnaryServerInterceptor(),
	// }

	// if srv.enableMetrics {
	// 	unaryInts = append(unaryInts, srvintc.UnaryPrometheusInterceptor)
	// }

	// if srv.timeout > 0 {
	// 	unaryInts = append(unaryInts, srvintc.UnaryTimeoutInterceptor(srv.timeout))
	// }

	// if len(srv.unaryInts) > 0 {
	// 	unaryInts = append(unaryInts, srv.unaryInts...)
	// }

	//把传入的拦截器转换成grpc的ServerOption
	grpcOpts := []grpc.ServerOption{grpc.ChainUnaryInterceptor(srv.unaryInts...)}
	// grpcOpts := []grpc.ServerOption{grpc.ChainUnaryInterceptor(unaryInts...)}

	//把用户自己传入的grpc.ServerOption放在一起
	srv.Server = grpc.NewServer(grpcOpts...)

	//解析address
	err := srv.listenAndEndpoint()
	if err != nil {
		panic(err)
	}

	return srv
}




func WithAddress(address string) ServerOption {
	return func(s *Server) {
		s.address = address
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

// func WithMetrics(metric bool) ServerOption {
// 	return func(s *Server) {
// 		s.enableMetrics = metric
// 	}
// }

// func WithTimeout(timeout time.Duration) ServerOption {
// 	return func(s *Server) {
// 		s.timeout = timeout
// 	}
// }


// 完成ip和端口的提取
func (s *Server) listenAndEndpoint() error {
	if s.lis == nil {
		lis, err := net.Listen("tcp", s.address)
		if err != nil {
			return err
		}
		s.lis = lis
	}
	// 完成ip和port的提取
	addr, err := host.Extract(s.address, s.lis)
	if err != nil {
		_ = s.lis.Close()
		return err
	}
	s.endpoint = &url.URL{Scheme: "grpc", Host: addr}
	return nil
}