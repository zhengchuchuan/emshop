package goods

import (
	"context"
	gpbv1 "emshop/api/goods/v1"
	"emshop/internal/app/emshop/admin/data"
	"emshop/pkg/log"
)

// GoodsSrv 管理员商品服务接口
type GoodsSrv interface {
	// 商品管理
	GetGoodsList(ctx context.Context, request *gpbv1.GoodsFilterRequest) (*gpbv1.GoodsListResponse, error)
	CreateGoods(ctx context.Context, info *gpbv1.CreateGoodsInfo) (*gpbv1.GoodsInfoResponse, error)
	UpdateGoods(ctx context.Context, info *gpbv1.CreateGoodsInfo) (*gpbv1.GoodsInfoResponse, error)
	DeleteGoods(ctx context.Context, id uint64) (*gpbv1.GoodsInfoResponse, error)
	GetGoodsDetail(ctx context.Context, id uint64) (*gpbv1.GoodsInfoResponse, error)
	SyncGoodsData(ctx context.Context, request *gpbv1.SyncDataRequest) (*gpbv1.SyncDataResponse, error)
	
	// 分类管理
	GetAllCategoriesList(ctx context.Context) (*gpbv1.CategoryListResponse, error)
	CreateCategory(ctx context.Context, request *gpbv1.CategoryInfoRequest) (*gpbv1.CategoryInfoResponse, error)
	UpdateCategory(ctx context.Context, request *gpbv1.CategoryInfoRequest) (*gpbv1.CategoryInfoResponse, error)
	DeleteCategory(ctx context.Context, id uint64) (*gpbv1.CategoryInfoResponse, error)
	
	// 品牌管理
	GetBrandsList(ctx context.Context, request *gpbv1.BrandFilterRequest) (*gpbv1.BrandListResponse, error)
	CreateBrand(ctx context.Context, request *gpbv1.BrandRequest) (*gpbv1.BrandInfoResponse, error)
	UpdateBrand(ctx context.Context, request *gpbv1.BrandRequest) (*gpbv1.BrandInfoResponse, error)
	DeleteBrand(ctx context.Context, id uint64) (*gpbv1.BrandInfoResponse, error)
	
	// 轮播图管理
	GetBannersList(ctx context.Context) (*gpbv1.BannerListResponse, error)
	CreateBanner(ctx context.Context, request *gpbv1.BannerRequest) (*gpbv1.BannerResponse, error)
	UpdateBanner(ctx context.Context, request *gpbv1.BannerRequest) (*gpbv1.BannerResponse, error)
	DeleteBanner(ctx context.Context, id uint64) (*gpbv1.BannerResponse, error)
}

type goodsService struct {
	data data.DataFactory
}

func NewGoodsService(data data.DataFactory) GoodsSrv {
	return &goodsService{data: data}
}

// ==================== 商品管理 ====================

func (g *goodsService) GetGoodsList(ctx context.Context, request *gpbv1.GoodsFilterRequest) (*gpbv1.GoodsListResponse, error) {
	log.Infof("Admin GetGoodsList called")
	return g.data.Goods().GoodsList(ctx, request)
}

func (g *goodsService) CreateGoods(ctx context.Context, info *gpbv1.CreateGoodsInfo) (*gpbv1.GoodsInfoResponse, error) {
	log.Infof("Admin CreateGoods called for: %s", info.Name)
	// 管理员创建商品可以添加额外的业务逻辑，如审核流程、权限检查等
	return g.data.Goods().CreateGoods(ctx, info)
}

func (g *goodsService) UpdateGoods(ctx context.Context, info *gpbv1.CreateGoodsInfo) (*gpbv1.GoodsInfoResponse, error) {
	log.Infof("Admin UpdateGoods called for ID: %d", info.Id)
	return g.data.Goods().UpdateGoods(ctx, info)
}

func (g *goodsService) DeleteGoods(ctx context.Context, id uint64) (*gpbv1.GoodsInfoResponse, error) {
	log.Infof("Admin DeleteGoods called for ID: %d", id)
	deleteInfo := &gpbv1.DeleteGoodsInfo{Id: int32(id)}
	return g.data.Goods().DeleteGoods(ctx, deleteInfo)
}

func (g *goodsService) GetGoodsDetail(ctx context.Context, id uint64) (*gpbv1.GoodsInfoResponse, error) {
	log.Infof("Admin GetGoodsDetail called for ID: %d", id)
	request := &gpbv1.GoodInfoRequest{Id: int32(id)}
	return g.data.Goods().GetGoodsDetail(ctx, request)
}

func (g *goodsService) SyncGoodsData(ctx context.Context, request *gpbv1.SyncDataRequest) (*gpbv1.SyncDataResponse, error) {
	log.Infof("Admin SyncGoodsData called")
	return g.data.Goods().SyncGoodsData(ctx, request)
}

// ==================== 分类管理 ====================

func (g *goodsService) GetAllCategoriesList(ctx context.Context) (*gpbv1.CategoryListResponse, error) {
	log.Infof("Admin GetAllCategoriesList called")
	return g.data.Goods().GetAllCategorysList(ctx)
}

func (g *goodsService) CreateCategory(ctx context.Context, request *gpbv1.CategoryInfoRequest) (*gpbv1.CategoryInfoResponse, error) {
	log.Infof("Admin CreateCategory called: %s", request.Name)
	return g.data.Goods().CreateCategory(ctx, request)
}

func (g *goodsService) UpdateCategory(ctx context.Context, request *gpbv1.CategoryInfoRequest) (*gpbv1.CategoryInfoResponse, error) {
	log.Infof("Admin UpdateCategory called for ID: %d", request.Id)
	return g.data.Goods().UpdateCategory(ctx, request)
}

func (g *goodsService) DeleteCategory(ctx context.Context, id uint64) (*gpbv1.CategoryInfoResponse, error) {
	log.Infof("Admin DeleteCategory called for ID: %d", id)
	request := &gpbv1.DeleteCategoryRequest{Id: int32(id)}
	return g.data.Goods().DeleteCategory(ctx, request)
}

// ==================== 品牌管理 ====================

func (g *goodsService) GetBrandsList(ctx context.Context, request *gpbv1.BrandFilterRequest) (*gpbv1.BrandListResponse, error) {
	log.Infof("Admin GetBrandsList called")
	return g.data.Goods().BrandList(ctx, request)
}

func (g *goodsService) CreateBrand(ctx context.Context, request *gpbv1.BrandRequest) (*gpbv1.BrandInfoResponse, error) {
	log.Infof("Admin CreateBrand called: %s", request.Name)
	return g.data.Goods().CreateBrand(ctx, request)
}

func (g *goodsService) UpdateBrand(ctx context.Context, request *gpbv1.BrandRequest) (*gpbv1.BrandInfoResponse, error) {
	log.Infof("Admin UpdateBrand called for ID: %d", request.Id)
	return g.data.Goods().UpdateBrand(ctx, request)
}

func (g *goodsService) DeleteBrand(ctx context.Context, id uint64) (*gpbv1.BrandInfoResponse, error) {
	log.Infof("Admin DeleteBrand called for ID: %d", id)
	request := &gpbv1.BrandRequest{Id: int32(id)}
	return g.data.Goods().DeleteBrand(ctx, request)
}

// ==================== 轮播图管理 ====================

func (g *goodsService) GetBannersList(ctx context.Context) (*gpbv1.BannerListResponse, error) {
	log.Infof("Admin GetBannersList called")
	return g.data.Goods().BannerList(ctx)
}

func (g *goodsService) CreateBanner(ctx context.Context, request *gpbv1.BannerRequest) (*gpbv1.BannerResponse, error) {
	log.Infof("Admin CreateBanner called")
	return g.data.Goods().CreateBanner(ctx, request)
}

func (g *goodsService) UpdateBanner(ctx context.Context, request *gpbv1.BannerRequest) (*gpbv1.BannerResponse, error) {
	log.Infof("Admin UpdateBanner called for ID: %d", request.Id)
	return g.data.Goods().UpdateBanner(ctx, request)
}

func (g *goodsService) DeleteBanner(ctx context.Context, id uint64) (*gpbv1.BannerResponse, error) {
	log.Infof("Admin DeleteBanner called for ID: %d", id)
	request := &gpbv1.BannerRequest{Id: int32(id)}
	return g.data.Goods().DeleteBanner(ctx, request)
}