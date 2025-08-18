package rpc

import (
	"context"
	"time"
	ipbv1 "emshop/api/inventory/v1"
	"emshop/gin-micro/server/rpc-server"
	"emshop/gin-micro/server/rpc-server/client-interceptors"
	"emshop/internal/app/emshop/admin/data"
	"emshop/gin-micro/registry"
	"emshop/pkg/log"
	"google.golang.org/grpc"
)

const inventoryServiceName = "discovery:///emshop-inventory-srv"

type inventory struct {
	ic ipbv1.InventoryClient
}

func NewInventory(ic ipbv1.InventoryClient) *inventory {
	return &inventory{ic}
}

func NewInventoryServiceClient(r registry.Discovery) ipbv1.InventoryClient {
	log.Infof("Initializing gRPC connection to service: %s", inventoryServiceName)
	conn, err := rpcserver.DialInsecure(
		context.Background(),
		rpcserver.WithEndpoint(inventoryServiceName),
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

// 库存相关方法可以根据需要添加

var _ data.InventoryData = &inventory{}