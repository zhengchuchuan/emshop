package v1

import (
	"context"
	gpb "emshop/api/goods/v1"
	"emshop/internal/app/api/emshop/data"
)

type GoodsSrv interface {
	List(ctx context.Context, request *gpb.GoodsFilterRequest) (*gpb.GoodsListResponse, error)
	Create(ctx context.Context, info *gpb.CreateGoodsInfo) (*gpb.GoodsInfoResponse, error)
	SyncData(ctx context.Context, request *gpb.SyncDataRequest) (*gpb.SyncDataResponse, error)
	Detail(ctx context.Context, request *gpb.GoodInfoRequest) (*gpb.GoodsInfoResponse, error)
	Delete(ctx context.Context, info *gpb.DeleteGoodsInfo) (*gpb.GoodsInfoResponse, error)
	Update(ctx context.Context, info *gpb.CreateGoodsInfo) (*gpb.GoodsInfoResponse, error)
	
	// 分类管理
	CategoryList(ctx context.Context) (*gpb.CategoryListResponse, error)
	CategoryDetail(ctx context.Context, request *gpb.CategoryListRequest) (*gpb.SubCategoryListResponse, error)
	CreateCategory(ctx context.Context, request *gpb.CategoryInfoRequest) (*gpb.CategoryInfoResponse, error)
	UpdateCategory(ctx context.Context, request *gpb.CategoryInfoRequest) (*gpb.CategoryInfoResponse, error)
	DeleteCategory(ctx context.Context, request *gpb.DeleteCategoryRequest) (*gpb.CategoryInfoResponse, error)
	
	// 品牌管理
	BrandList(ctx context.Context, request *gpb.BrandFilterRequest) (*gpb.BrandListResponse, error)
	CreateBrand(ctx context.Context, request *gpb.BrandRequest) (*gpb.BrandInfoResponse, error)
	UpdateBrand(ctx context.Context, request *gpb.BrandRequest) (*gpb.BrandInfoResponse, error)
	DeleteBrand(ctx context.Context, request *gpb.BrandRequest) (*gpb.BrandInfoResponse, error)
	
	// 轮播图管理
	BannerList(ctx context.Context) (*gpb.BannerListResponse, error)
	CreateBanner(ctx context.Context, request *gpb.BannerRequest) (*gpb.BannerResponse, error)
	UpdateBanner(ctx context.Context, request *gpb.BannerRequest) (*gpb.BannerResponse, error)
	DeleteBanner(ctx context.Context, request *gpb.BannerRequest) (*gpb.BannerResponse, error)
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

func (gs *goodsService) Detail(ctx context.Context, request *gpb.GoodInfoRequest) (*gpb.GoodsInfoResponse, error) {
	return gs.data.Goods().GetGoodsDetail(ctx, request)
}

func (gs *goodsService) Delete(ctx context.Context, info *gpb.DeleteGoodsInfo) (*gpb.GoodsInfoResponse, error) {
	_, err := gs.data.Goods().DeleteGoods(ctx, info)
	if err != nil {
		return nil, err
	}
	return &gpb.GoodsInfoResponse{}, nil
}

func (gs *goodsService) Update(ctx context.Context, info *gpb.CreateGoodsInfo) (*gpb.GoodsInfoResponse, error) {
	_, err := gs.data.Goods().UpdateGoods(ctx, info)
	if err != nil {
		return nil, err
	}
	return &gpb.GoodsInfoResponse{}, nil
}

// ==================== 分类管理 ====================

func (gs *goodsService) CategoryList(ctx context.Context) (*gpb.CategoryListResponse, error) {
	return gs.data.Goods().GetAllCategorysList(ctx)
}

func (gs *goodsService) CategoryDetail(ctx context.Context, request *gpb.CategoryListRequest) (*gpb.SubCategoryListResponse, error) {
	return gs.data.Goods().GetSubCategory(ctx, request)
}

func (gs *goodsService) CreateCategory(ctx context.Context, request *gpb.CategoryInfoRequest) (*gpb.CategoryInfoResponse, error) {
	return gs.data.Goods().CreateCategory(ctx, request)
}

func (gs *goodsService) UpdateCategory(ctx context.Context, request *gpb.CategoryInfoRequest) (*gpb.CategoryInfoResponse, error) {
	_, err := gs.data.Goods().UpdateCategory(ctx, request)
	if err != nil {
		return nil, err
	}
	return &gpb.CategoryInfoResponse{}, nil
}

func (gs *goodsService) DeleteCategory(ctx context.Context, request *gpb.DeleteCategoryRequest) (*gpb.CategoryInfoResponse, error) {
	_, err := gs.data.Goods().DeleteCategory(ctx, request)
	if err != nil {
		return nil, err
	}
	return &gpb.CategoryInfoResponse{}, nil
}

// ==================== 品牌管理 ====================

func (gs *goodsService) BrandList(ctx context.Context, request *gpb.BrandFilterRequest) (*gpb.BrandListResponse, error) {
	return gs.data.Goods().BrandList(ctx, request)
}

func (gs *goodsService) CreateBrand(ctx context.Context, request *gpb.BrandRequest) (*gpb.BrandInfoResponse, error) {
	return gs.data.Goods().CreateBrand(ctx, request)
}

func (gs *goodsService) UpdateBrand(ctx context.Context, request *gpb.BrandRequest) (*gpb.BrandInfoResponse, error) {
	_, err := gs.data.Goods().UpdateBrand(ctx, request)
	if err != nil {
		return nil, err
	}
	return &gpb.BrandInfoResponse{}, nil
}

func (gs *goodsService) DeleteBrand(ctx context.Context, request *gpb.BrandRequest) (*gpb.BrandInfoResponse, error) {
	_, err := gs.data.Goods().DeleteBrand(ctx, request)
	if err != nil {
		return nil, err
	}
	return &gpb.BrandInfoResponse{}, nil
}

// ==================== 轮播图管理 ====================

func (gs *goodsService) BannerList(ctx context.Context) (*gpb.BannerListResponse, error) {
	return gs.data.Goods().BannerList(ctx)
}

func (gs *goodsService) CreateBanner(ctx context.Context, request *gpb.BannerRequest) (*gpb.BannerResponse, error) {
	return gs.data.Goods().CreateBanner(ctx, request)
}

func (gs *goodsService) UpdateBanner(ctx context.Context, request *gpb.BannerRequest) (*gpb.BannerResponse, error) {
	_, err := gs.data.Goods().UpdateBanner(ctx, request)
	if err != nil {
		return nil, err
	}
	return &gpb.BannerResponse{}, nil
}

func (gs *goodsService) DeleteBanner(ctx context.Context, request *gpb.BannerRequest) (*gpb.BannerResponse, error) {
	_, err := gs.data.Goods().DeleteBanner(ctx, request)
	if err != nil {
		return nil, err
	}
	return &gpb.BannerResponse{}, nil
}



func NewGoods(data data.DataFactory) *goodsService {
	return &goodsService{data: data}
}

var _ GoodsSrv = &goodsService{}
