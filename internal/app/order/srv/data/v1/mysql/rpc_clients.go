package mysql

import (
	"context"
	"time"

	gpbv1 "emshop/api/goods/v1"
	proto "emshop/api/inventory/v1"
	"emshop/gin-micro/registry/consul"
	"emshop/gin-micro/server/rpc-server"
	"emshop/gin-micro/server/rpc-server/client-interceptors"
	"emshop/internal/app/pkg/options"

	cosulAPI "github.com/hashicorp/consul/api"

	"emshop/gin-micro/registry"
)

const (
	goodsserviceName = "discovery:///emshop-goods-srv"
	ginvserviceName  = "discovery:///emshop-inventory-srv"
)

func NewDiscovery(opts *options.RegistryOptions) registry.Discovery {
	c := cosulAPI.DefaultConfig()
	c.Address = opts.Address
	c.Scheme = opts.Scheme
	cli, err := cosulAPI.NewClient(c)
	if err != nil {
		panic(err)
	}
	r := consul.New(cli, consul.WithHealthCheck(true))
	return r
}

func GetGoodsClient(opts *options.RegistryOptions) gpbv1.GoodsClient {
	discovery := NewDiscovery(opts)
	goodsClient := NewGoodsServiceClient(discovery)
	return goodsClient
}

func NewGoodsServiceClient(r registry.Discovery) gpbv1.GoodsClient {
	// 创建带有重试和超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := rpcserver.DialInsecure(
		ctx,
		rpcserver.WithBalancerName("p2c"),
		rpcserver.WithEndpoint(goodsserviceName),
		rpcserver.WithDiscovery(r),
		rpcserver.WithClientUnaryInterceptor(clientinterceptors.UnaryTracingInterceptor),
	)
	if err != nil {
		panic(err)
	}
	c := gpbv1.NewGoodsClient(conn)
	return c
}

func GetInventoryClient(opts *options.RegistryOptions) proto.InventoryClient {
	discovery := NewDiscovery(opts)
	invClient := NewInventoryServiceClient(discovery)
	return invClient
}

func NewInventoryServiceClient(r registry.Discovery) proto.InventoryClient {
	conn, err := rpcserver.DialInsecure(
		context.Background(),
		rpcserver.WithBalancerName("p2c"),
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
