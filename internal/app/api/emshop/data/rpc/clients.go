package rpc

import (
	"context"
	"time"

	cpbv1 "emshop/api/coupon/v1"
	gpbv1 "emshop/api/goods/v1"
	ipb "emshop/api/inventory/v1"
	lpbv1 "emshop/api/logistics/v1"
	opbv1 "emshop/api/order/v1"
	ppbv1 "emshop/api/payment/v1"
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
	couponClient    cpbv1.CouponClient
	paymentClient   ppbv1.PaymentClient
	logisticsClient lpbv1.LogisticsClient
}

// newGrpcClients 创建并初始化所有 gRPC 客户端
func newGrpcClients(discovery registry.Discovery) *grpcClients {
	return &grpcClients{
		userClient:      NewUserServiceClient(discovery),
		goodsClient:     NewGoodsServiceClient(discovery),
		inventoryClient: NewInventoryServiceClient(discovery),
		orderClient:     NewOrderServiceClient(discovery),
		userOpClient:    NewUserOpServiceClient(discovery),
		couponClient:    NewCouponServiceClient(discovery),
		paymentClient:   NewPaymentServiceClient(discovery),
		logisticsClient: NewLogisticsServiceClient(discovery),
	}
}

// 服务名称常量 - 集中定义所有服务名称
const (
	clientUserServiceName      = "discovery:///emshop-user-srv"
	clientGoodsServiceName     = "discovery:///emshop-goods-srv"
	clientInventoryServiceName = "discovery:///emshop-inventory-srv"
	clientOrderServiceName     = "discovery:///emshop-order-srv"
	clientUseropServiceName    = "discovery:///emshop-userop-srv"
	clientCouponServiceName    = "discovery:///emshop-coupon-srv"
	clientPaymentServiceName   = "discovery:///emshop-payment-srv"
	clientLogisticsServiceName = "discovery:///emshop-logistics-srv"
)

// 移除直连fallback逻辑，统一通过服务发现

// NewUserServiceClient 创建用户服务的 gRPC 客户端
func NewUserServiceClient(r registry.Discovery) upbv1.UserClient {
	log.Infof("Initializing gRPC connection to service: %s", clientUserServiceName)
	conn, err := rpcserver.DialInsecure(
		context.Background(),
		rpcserver.WithBalancerName("p2c"),
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
		rpcserver.WithBalancerName("p2c"),
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

// NewInventoryServiceClient 创建库存服务的 gRPC 客户端，仅使用 Consul 服务发现
func NewInventoryServiceClient(r registry.Discovery) ipb.InventoryClient {
	log.Infof("Initializing gRPC connection to service: %s", clientInventoryServiceName)
	conn, err := rpcserver.DialInsecure(
		context.Background(),
		rpcserver.WithBalancerName("p2c"),
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
	return ipb.NewInventoryClient(conn)
}

// NewOrderServiceClient 创建订单服务的 gRPC 客户端
func NewOrderServiceClient(r registry.Discovery) opbv1.OrderClient {
	log.Infof("Initializing gRPC connection to service: %s", clientOrderServiceName)
	conn, err := rpcserver.DialInsecure(
		context.Background(),
		rpcserver.WithBalancerName("p2c"),
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
		rpcserver.WithBalancerName("p2c"),
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

// NewCouponServiceClient 创建优惠券服务的 gRPC 客户端
func NewCouponServiceClient(r registry.Discovery) cpbv1.CouponClient {
	log.Infof("Initializing gRPC connection to service: %s", clientCouponServiceName)
	conn, err := rpcserver.DialInsecure(
		context.Background(),
		rpcserver.WithBalancerName("p2c"),
		rpcserver.WithEndpoint(clientCouponServiceName),
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
	c := cpbv1.NewCouponClient(conn)
	return c
}

// NewPaymentServiceClient 创建支付服务的 gRPC 客户端
func NewPaymentServiceClient(r registry.Discovery) ppbv1.PaymentClient {
	log.Infof("Initializing gRPC connection to service: %s", clientPaymentServiceName)
	conn, err := rpcserver.DialInsecure(
		context.Background(),
		rpcserver.WithEndpoint(clientPaymentServiceName),
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
	c := ppbv1.NewPaymentClient(conn)
	return c
}

// NewLogisticsServiceClient 创建物流服务的 gRPC 客户端
func NewLogisticsServiceClient(r registry.Discovery) lpbv1.LogisticsClient {
	log.Infof("Initializing gRPC connection to service: %s", clientLogisticsServiceName)
	conn, err := rpcserver.DialInsecure(
		context.Background(),
		rpcserver.WithBalancerName("p2c"),
		rpcserver.WithEndpoint(clientLogisticsServiceName),
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
	c := lpbv1.NewLogisticsClient(conn)
	return c
}
