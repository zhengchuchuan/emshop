package rpcserver

import (
	"net"
	"net/url"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
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