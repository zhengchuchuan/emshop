package db

import (
	"context"

	proto "emshop/api/inventory/v1"
	"emshop/gin-micro/server/rpc-server"
	"emshop/gin-micro/server/rpc-server/client-interceptors"
	"emshop/internal/app/pkg/options"

	"emshop/gin-micro/registry"
)

const ginvserviceName = "discovery:///emshop-inventory-srv"

func GetInventoryClient(opts *options.RegistryOptions) proto.InventoryClient {
	discovery := NewDiscovery(opts)
	invClient := NewInventoryServiceClient(discovery)
	return invClient
}

func NewInventoryServiceClient(r registry.Discovery) proto.InventoryClient {
	conn, err := rpcserver.DialInsecure(
		context.Background(),
		rpcserver.WithEndpoint(ginvserviceName),
		rpcserver.WithDiscovery(r),
		rpcserver.WithClientUnaryInterceptor(clientinterceptors.UnaryTracingInterceptor),
	)
	if err != nil {
		panic(err)
	}
	c := proto.NewInventoryClient(conn)
	return c
}
