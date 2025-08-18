package rpc

import (
	"context"
	"time"
	uoppbv1 "emshop/api/userop/v1"
	"emshop/gin-micro/server/rpc-server"
	"emshop/gin-micro/server/rpc-server/client-interceptors"
	"emshop/internal/app/emshop/admin/data"
	"emshop/gin-micro/registry"
	"emshop/pkg/log"
	"google.golang.org/grpc"
)

const useropServiceName = "discovery:///emshop-userop-srv"

type userop struct {
	uopc uoppbv1.UserOpClient
}

func NewUserOp(uopc uoppbv1.UserOpClient) *userop {
	return &userop{uopc}
}

func NewUserOpServiceClient(r registry.Discovery) uoppbv1.UserOpClient {
	log.Infof("Initializing gRPC connection to service: %s", useropServiceName)
	conn, err := rpcserver.DialInsecure(
		context.Background(),
		rpcserver.WithEndpoint(useropServiceName),
		rpcserver.WithDiscovery(r),
		rpcserver.WithClientTimeout(10*time.Second),
		rpcserver.WithClientOptions(grpc.WithNoProxy()),
		rpcserver.WithClientUnaryInterceptor(clientinterceptors.UnaryTracingInterceptor),
	)
	if err != nil {
		log.Errorf("Failed to create gRPC connection: %v", err)
		panic(err)
	}
	log.Info("gRPC connection established successfully")
	c := uoppbv1.NewUserOpClient(conn)
	return c
}

// 用户操作相关方法可以根据需要添加

var _ data.UserOpData = &userop{}