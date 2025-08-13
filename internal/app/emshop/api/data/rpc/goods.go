package rpc

import (
	"context"
	"time"
	gpbv1 "emshop/api/goods/v1"
	"emshop/gin-micro/server/rpc-server"
	"emshop/gin-micro/server/rpc-server/client-interceptors"
	"emshop/internal/app/emshop/api/data"
	"emshop/gin-micro/registry"
	"emshop/pkg/log"
	"google.golang.org/grpc"
)

const goodsserviceName = "discovery:///emshop-goods-srv"

type goods struct {
	gc gpbv1.GoodsClient
}

func NewGoods(gc gpbv1.GoodsClient) *goods {
	return &goods{gc}
}

func NewGoodsServiceClient(r registry.Discovery) gpbv1.GoodsClient {
	log.Infof("Initializing gRPC connection to service: %s", goodsserviceName)
	conn, err := rpcserver.DialInsecure(
		context.Background(),
		rpcserver.WithEndpoint(goodsserviceName),
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

func (g *goods) GoodsList(ctx context.Context, request *gpbv1.GoodsFilterRequest) (*gpbv1.GoodsListResponse, error) {
	log.Infof("Calling GoodsList gRPC with filter: %+v", request)
	response, err := g.gc.GoodsList(ctx, request)
	if err != nil {
		log.Errorf("GoodsList gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("GoodsList gRPC call successful, total: %d", response.Total)
	return response, nil
}

func (g *goods) CreateGoods(ctx context.Context, info *gpbv1.CreateGoodsInfo) (*gpbv1.GoodsInfoResponse, error) {
	log.Infof("Calling CreateGoods gRPC for goods: %s", info.Name)
	response, err := g.gc.CreateGoods(ctx, info)
	if err != nil {
		log.Errorf("CreateGoods gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("CreateGoods gRPC call successful, goods ID: %d", response.Id)
	return response, nil
}

func (g *goods) SyncGoodsData(ctx context.Context, request *gpbv1.SyncDataRequest) (*gpbv1.SyncDataResponse, error) {
	log.Infof("Calling SyncGoodsData gRPC with request: forceSync=%v, goodsIds=%v", request.ForceSync, request.GoodsIds)
	response, err := g.gc.SyncGoodsData(ctx, request)
	if err != nil {
		log.Errorf("SyncGoodsData gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("SyncGoodsData gRPC call successful, synced=%d, failed=%d", response.SyncedCount, response.FailedCount)
	return response, nil
}

var _ data.GoodsData = &goods{}
