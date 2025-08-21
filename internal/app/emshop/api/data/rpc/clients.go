package rpc

import (
	"context"
	"time"
	
	gpbv1 "emshop/api/goods/v1"
	ipb "emshop/api/inventory/v1"
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
	inventoryClient ipb.InventoryClient
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

// Fallback地址常量
const clientFallbackInventoryAddress = "127.0.0.1:28055"

// NewUserServiceClient 创建用户服务的 gRPC 客户端
func NewUserServiceClient(r registry.Discovery) upbv1.UserClient {
	log.Infof("Initializing gRPC connection to service: %s", clientUserServiceName)
	conn, err := rpcserver.DialInsecure(
		context.Background(),
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

// NewInventoryServiceClient 创建库存服务的 gRPC 客户端，支持健壮的重试和fallback机制
func NewInventoryServiceClient(r registry.Discovery) ipb.InventoryClient {
	log.Infof("Initializing gRPC connection to service: %s", clientInventoryServiceName)
	
	// 首先尝试服务发现连接，使用更健壮的配置
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	log.Infof("Attempting service discovery connection to: %s", clientInventoryServiceName)
	conn, err := rpcserver.DialInsecure(
		ctx,
		rpcserver.WithEndpoint(clientInventoryServiceName),
		rpcserver.WithDiscovery(r),
		rpcserver.WithClientTimeout(15*time.Second),
		rpcserver.WithClientUnaryInterceptor(clientinterceptors.UnaryTracingInterceptor),
	)
	
	if err != nil {
		log.Warnf("Service discovery connection failed: %v, falling back to direct connection", err)
		// fallback到直连
		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel2()
		
		log.Infof("Attempting direct connection to: %s", clientFallbackInventoryAddress)
		conn, err = rpcserver.DialInsecure(
			ctx2,
			rpcserver.WithEndpoint(clientFallbackInventoryAddress),
			rpcserver.WithClientTimeout(15*time.Second),
			rpcserver.WithClientUnaryInterceptor(clientinterceptors.UnaryTracingInterceptor),
		)
		
		if err != nil {
			log.Fatalf("Both service discovery and direct connection failed: %v", err)
		}
		log.Infof("Successfully connected to inventory service via direct connection fallback")
	} else {
		log.Infof("Successfully connected to inventory service via service discovery")
		// 即使服务发现成功，也要测试连接是否真的可用
		// 如果连接有问题，立即切换到localhost fallback
		log.Infof("Testing service discovery connection...")
		
		testCtx, testCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer testCancel()
		
		testClient := ipb.NewInventoryClient(conn)
		_, testErr := testClient.InvDetail(testCtx, &ipb.GoodsInvInfo{GoodsId: 1})
		
		if testErr != nil {
			log.Warnf("Service discovery connection test failed: %v, switching to localhost fallback", testErr)
			conn.Close()
			
			// 立即尝试localhost连接
			fallbackCtx, fallbackCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer fallbackCancel()
			
			log.Infof("Attempting localhost fallback connection to: %s", clientFallbackInventoryAddress)
			conn, err = rpcserver.DialInsecure(
				fallbackCtx,
				rpcserver.WithEndpoint(clientFallbackInventoryAddress),
				rpcserver.WithClientTimeout(15*time.Second),
				rpcserver.WithClientUnaryInterceptor(clientinterceptors.UnaryTracingInterceptor),
			)
			
			if err != nil {
				log.Fatalf("Localhost fallback connection also failed: %v", err)
			}
			log.Infof("Successfully connected to inventory service via localhost fallback")
		} else {
			log.Infof("Service discovery connection test successful")
		}
	}
	
	return ipb.NewInventoryClient(conn)
}

// NewOrderServiceClient 创建订单服务的 gRPC 客户端
func NewOrderServiceClient(r registry.Discovery) opbv1.OrderClient {
	log.Infof("Initializing gRPC connection to service: %s", clientOrderServiceName)
	conn, err := rpcserver.DialInsecure(
		context.Background(),
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