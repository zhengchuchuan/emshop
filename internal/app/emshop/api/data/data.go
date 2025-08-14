package data

import (
	"context"
	gpb "emshop/api/goods/v1"
)

type GoodsData interface {
	GoodsList(ctx context.Context, request *gpb.GoodsFilterRequest) (*gpb.GoodsListResponse, error)
	CreateGoods(ctx context.Context, info *gpb.CreateGoodsInfo) (*gpb.GoodsInfoResponse, error)
	SyncGoodsData(ctx context.Context, request *gpb.SyncDataRequest) (*gpb.SyncDataResponse, error)
	GetGoodsDetail(ctx context.Context, request *gpb.GoodInfoRequest) (*gpb.GoodsInfoResponse, error)
	DeleteGoods(ctx context.Context, info *gpb.DeleteGoodsInfo) (*gpb.GoodsInfoResponse, error)
	UpdateGoods(ctx context.Context, info *gpb.CreateGoodsInfo) (*gpb.GoodsInfoResponse, error)
}

type DataFactory interface {
	Goods() GoodsData
	Users() UserData
}
