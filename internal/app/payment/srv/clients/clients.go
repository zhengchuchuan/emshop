package clients

import (
	"context"
	"time"
	
	opbv1 "emshop/api/order/v1"
	ipbv1 "emshop/api/inventory/v1"
	lpbv1 "emshop/api/logistics/v1"
	"emshop/gin-micro/registry"
	rpcserver "emshop/gin-micro/server/rpc-server"
	clientinterceptors "emshop/gin-micro/server/rpc-server/client-interceptors"
	"emshop/pkg/log"
	"google.golang.org/grpc"
)

// ServiceClients 支付服务需要调用的其他服务客户端
type ServiceClients struct {
	orderClient     opbv1.OrderClient
	inventoryClient ipbv1.InventoryClient
	logisticsClient lpbv1.LogisticsClient
}

// NewServiceClients 创建支付服务的客户端集合
func NewServiceClients(discovery registry.Discovery) *ServiceClients {
	return &ServiceClients{
		orderClient:     NewOrderServiceClient(discovery),
		inventoryClient: NewInventoryServiceClient(discovery),
		logisticsClient: NewLogisticsServiceClient(discovery),
	}
}

// GetOrderClient 获取订单服务客户端
func (sc *ServiceClients) GetOrderClient() opbv1.OrderClient {
	return sc.orderClient
}

// GetInventoryClient 获取库存服务客户端
func (sc *ServiceClients) GetInventoryClient() ipbv1.InventoryClient {
	return sc.inventoryClient
}

// GetLogisticsClient 获取物流服务客户端
func (sc *ServiceClients) GetLogisticsClient() lpbv1.LogisticsClient {
	return sc.logisticsClient
}

// 服务名称常量
const (
	orderServiceName     = "discovery:///emshop-order-srv"
	inventoryServiceName = "discovery:///emshop-inventory-srv"
	logisticsServiceName = "discovery:///emshop-logistics-srv"
)

// NewOrderServiceClient 创建订单服务的 gRPC 客户端
func NewOrderServiceClient(r registry.Discovery) opbv1.OrderClient {
	log.Infof("Initializing gRPC connection to order service: %s", orderServiceName)
	conn, err := rpcserver.DialInsecure(
		context.Background(),
		rpcserver.WithEndpoint(orderServiceName),
		rpcserver.WithDiscovery(r),
		rpcserver.WithClientTimeout(10*time.Second),
		rpcserver.WithClientOptions(grpc.WithNoProxy()),
		rpcserver.WithClientUnaryInterceptor(clientinterceptors.UnaryTracingInterceptor),
	)
	if err != nil {
		log.Errorf("Failed to create gRPC connection to order service: %v", err)
		panic(err)
	}
	log.Info("Order service gRPC connection established successfully")
	return opbv1.NewOrderClient(conn)
}

// NewInventoryServiceClient 创建库存服务的 gRPC 客户端
func NewInventoryServiceClient(r registry.Discovery) ipbv1.InventoryClient {
	log.Infof("Initializing gRPC connection to inventory service: %s", inventoryServiceName)
	conn, err := rpcserver.DialInsecure(
		context.Background(),
		rpcserver.WithEndpoint(inventoryServiceName),
		rpcserver.WithDiscovery(r),
		rpcserver.WithClientTimeout(10*time.Second),
		rpcserver.WithClientOptions(grpc.WithNoProxy()),
		rpcserver.WithClientUnaryInterceptor(clientinterceptors.UnaryTracingInterceptor),
	)
	if err != nil {
		log.Errorf("Failed to create gRPC connection to inventory service: %v", err)
		panic(err)
	}
	log.Info("Inventory service gRPC connection established successfully")
	return ipbv1.NewInventoryClient(conn)
}

// NewLogisticsServiceClient 创建物流服务的 gRPC 客户端
func NewLogisticsServiceClient(r registry.Discovery) lpbv1.LogisticsClient {
	log.Infof("Initializing gRPC connection to logistics service: %s", logisticsServiceName)
	conn, err := rpcserver.DialInsecure(
		context.Background(),
		rpcserver.WithEndpoint(logisticsServiceName),
		rpcserver.WithDiscovery(r),
		rpcserver.WithClientTimeout(10*time.Second),
		rpcserver.WithClientOptions(grpc.WithNoProxy()),
		rpcserver.WithClientUnaryInterceptor(clientinterceptors.UnaryTracingInterceptor),
	)
	if err != nil {
		log.Errorf("Failed to create gRPC connection to logistics service: %v", err)
		panic(err)
	}
	log.Info("Logistics service gRPC connection established successfully")
	return lpbv1.NewLogisticsClient(conn)
}