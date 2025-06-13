package rpc

import (
	"context"
	gpbv1 "emshop/api/goods/v1"
	"emshop/gin-micro/server/rpc-server"
	"emshop/gin-micro/server/rpc-server/client-interceptors"

	"emshop/gin-micro/registry"
)

const goodsserviceName = "discovery:///emshop-goods-srv"

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
