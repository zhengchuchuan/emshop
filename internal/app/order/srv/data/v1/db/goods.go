package db

import (
	"context"
	gpbv1 "emshop/api/goods/v1"
	"emshop/gin-micro/registry/consul"
	"emshop/gin-micro/server/rpc-server"
	"emshop/gin-micro/server/rpc-server/client-interceptors"
	"emshop/internal/app/pkg/options"

	cosulAPI "github.com/hashicorp/consul/api"

	"emshop/gin-micro/registry"
)

const goodsserviceName = "discovery:///emshop-goods-srv"

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
	conn, err := rpcserver.DialInsecure(
		context.Background(),
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
