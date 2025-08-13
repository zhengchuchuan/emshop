package v1

import (
	"context"
	gpb "emshop/api/goods/v1"
	"emshop/internal/app/emshop/api/data"


)

type GoodsSrv interface {
	List(ctx context.Context, request *gpb.GoodsFilterRequest) (*gpb.GoodsListResponse, error)
	Create(ctx context.Context, info *gpb.CreateGoodsInfo) (*gpb.GoodsInfoResponse, error)
	SyncData(ctx context.Context, request *gpb.SyncDataRequest) (*gpb.SyncDataResponse, error)
}

type goodsService struct {
	data data.DataFactory
}

func (gs *goodsService) List(ctx context.Context, request *gpb.GoodsFilterRequest) (*gpb.GoodsListResponse, error) {
	return gs.data.Goods().GoodsList(ctx, request)
}

func (gs *goodsService) Create(ctx context.Context, info *gpb.CreateGoodsInfo) (*gpb.GoodsInfoResponse, error) {
	return gs.data.Goods().CreateGoods(ctx, info)
}

func (gs *goodsService) SyncData(ctx context.Context, request *gpb.SyncDataRequest) (*gpb.SyncDataResponse, error) {
	return gs.data.Goods().SyncGoodsData(ctx, request)
}



func NewGoods(data data.DataFactory) *goodsService {
	return &goodsService{data: data}
}

var _ GoodsSrv = &goodsService{}
