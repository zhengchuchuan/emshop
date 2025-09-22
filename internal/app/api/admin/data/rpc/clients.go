package rpc

import (
	"context"
	"time"
	
	gpbv1 "emshop/api/goods/v1"
	ipbv1 "emshop/api/inventory/v1"
	opbv1 "emshop/api/order/v1"
	upbv1 "emshop/api/user/v1"
	uoppbv1 "emshop/api/userop/v1"
	"emshop/gin-micro/registry"
	rpcserver "emshop/gin-micro/server/rpc-server"
	clientinterceptors "emshop/gin-micro/server/rpc-server/client-interceptors"
	"emshop/pkg/log"
	"google.golang.org/grpc"
)

// grpcClients 集中管理所有的 gRPC 客户端
type grpcClients struct {
	userClient      upbv1.UserClient
	goodsClient     gpbv1.GoodsClient
	inventoryClient ipbv1.InventoryClient
	orderClient     opbv1.OrderClient
	userOpClient    uoppbv1.UserOpClient
}

// newGrpcClients 创建并初始化所有 gRPC 客户端
func newGrpcClients(discovery registry.Discovery) *grpcClients {
	return &grpcClients{
		userClient:      NewUserServiceClient(discovery),
		goodsClient:     NewGoodsServiceClient(discovery),
		inventoryClient: NewInventoryServiceClient(discovery),
		orderClient:     NewOrderServiceClient(discovery),
		userOpClient:    NewUserOpServiceClient(discovery),
	}
}

// 服务名称常量 - 集中定义所有服务名称
const (
	clientUserServiceName      = "discovery:///emshop-user-srv"
	clientGoodsServiceName     = "discovery:///emshop-goods-srv"
	clientInventoryServiceName = "discovery:///emshop-inventory-srv"
	clientOrderServiceName     = "discovery:///emshop-order-srv"
	clientUseropServiceName    = "discovery:///emshop-userop-srv"
)

// NewUserServiceClient 创建用户服务的 gRPC 客户端
func NewUserServiceClient(r registry.Discovery) upbv1.UserClient {
    log.Infof("Initializing gRPC connection to service: %s", clientUserServiceName)
    conn, err := rpcserver.DialInsecure(
        context.Background(),
        rpcserver.WithBalancerName("selector"),
        rpcserver.WithEndpoint(clientUserServiceName),
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
	c := upbv1.NewUserClient(conn)
	return c
}

// NewGoodsServiceClient 创建商品服务的 gRPC 客户端
func NewGoodsServiceClient(r registry.Discovery) gpbv1.GoodsClient {
    log.Infof("Initializing gRPC connection to service: %s", clientGoodsServiceName)
    conn, err := rpcserver.DialInsecure(
        context.Background(),
        rpcserver.WithBalancerName("selector"),
        rpcserver.WithEndpoint(clientGoodsServiceName),
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
	c := gpbv1.NewGoodsClient(conn)
	return c
}

// NewInventoryServiceClient 创建库存服务的 gRPC 客户端
func NewInventoryServiceClient(r registry.Discovery) ipbv1.InventoryClient {
    log.Infof("Initializing gRPC connection to service: %s", clientInventoryServiceName)
    conn, err := rpcserver.DialInsecure(
        context.Background(),
        rpcserver.WithBalancerName("selector"),
        rpcserver.WithEndpoint(clientInventoryServiceName),
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
	c := ipbv1.NewInventoryClient(conn)
	return c
}

// NewOrderServiceClient 创建订单服务的 gRPC 客户端
func NewOrderServiceClient(r registry.Discovery) opbv1.OrderClient {
    log.Infof("Initializing gRPC connection to service: %s", clientOrderServiceName)
    conn, err := rpcserver.DialInsecure(
        context.Background(),
        rpcserver.WithBalancerName("selector"),
        rpcserver.WithEndpoint(clientOrderServiceName),
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
	c := opbv1.NewOrderClient(conn)
	return c
}

// NewUserOpServiceClient 创建用户操作服务的 gRPC 客户端
func NewUserOpServiceClient(r registry.Discovery) uoppbv1.UserOpClient {
    log.Infof("Initializing gRPC connection to service: %s", clientUseropServiceName)
    conn, err := rpcserver.DialInsecure(
        context.Background(),
        rpcserver.WithBalancerName("selector"),
        rpcserver.WithEndpoint(clientUseropServiceName),
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
