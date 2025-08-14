package srv

import (
	pb "emshop/api/userop/v1"
	controllerv1 "emshop/internal/app/userop/srv/controller/v1"
	servicev1 "emshop/internal/app/userop/srv/service/v1"
	"emshop/pkg/log"
	"google.golang.org/grpc"
)

// RegisterGRPCServer 注册gRPC服务
func RegisterGRPCServer(s *grpc.Server, srv servicev1.Service) {
	// 创建控制器
	controller := controllerv1.NewUserOpController(srv)
	
	// 注册服务
	pb.RegisterUserOpServer(s, controller)
	
	log.Info("userop gRPC server registered successfully")
}