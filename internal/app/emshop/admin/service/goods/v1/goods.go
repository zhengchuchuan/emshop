package goods

import (
	"context"
	gpbv1 "emshop/api/goods/v1"
	ipbv1 "emshop/api/inventory/v1"
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
	GetCategoriesList(ctx context.Context) (*gpbv1.CategoryListResponse, error)
	GetCategoriesByLevel(ctx context.Context, level int32) (*gpbv1.CategoryListResponse, error)
	GetCategoryTree(ctx context.Context) (*gpbv1.CategoryTreeResponse, error)
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
	
	// 库存管理
	GetGoodsInventory(ctx context.Context, goodsId int32) (*ipbv1.GoodsInvInfo, error)
	SetGoodsInventory(ctx context.Context, request *ipbv1.GoodsInvInfo) error
	BatchSetGoodsInventory(ctx context.Context, inventories []*ipbv1.GoodsInvInfo) error
	
	// 批量操作
	BatchDeleteGoods(ctx context.Context, request *gpbv1.BatchDeleteGoodsRequest) (*gpbv1.BatchOperationResponse, error)
	BatchUpdateGoodsStatus(ctx context.Context, request *gpbv1.BatchUpdateGoodsStatusRequest) (*gpbv1.BatchOperationResponse, error)
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
	
	// 获取商品列表
	goodsResp, err := g.data.Goods().GoodsList(ctx, request)
	if err != nil {
		return nil, err
	}
	
	// 批量获取库存信息
	if len(goodsResp.Data) > 0 {
		goodsIds := make([]int32, 0, len(goodsResp.Data))
		for _, goods := range goodsResp.Data {
			goodsIds = append(goodsIds, goods.Id)
		}
		
		inventoryMap, err := g.data.Inventory().BatchGetInventory(ctx, goodsIds)
		if err != nil {
			log.Errorf("Failed to get inventory info: %v", err)
			// 库存获取失败不影响商品列表返回，设置默认库存为0
			for _, goods := range goodsResp.Data {
				goods.Stocks = 0
			}
		} else {
			// 设置库存信息
			for _, goods := range goodsResp.Data {
				if inv, exists := inventoryMap[goods.Id]; exists {
					goods.Stocks = inv.Num
				} else {
					goods.Stocks = 0
				}
			}
		}
	}
	
	return goodsResp, nil
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
	
	// 获取商品详情
	request := &gpbv1.GoodInfoRequest{Id: int32(id)}
	goodsResp, err := g.data.Goods().GetGoodsDetail(ctx, request)
	if err != nil {
		log.Errorf("Failed to get goods detail for ID %d: %v", id, err)
		return nil, err
	}
	log.Infof("Goods detail retrieved successfully for ID %d, name: %s", id, goodsResp.Name)
	
	// 获取库存信息
	log.Infof("Calling inventory service for goods ID: %d", id)
	inv, err := g.data.Inventory().GetInventory(ctx, int32(id))
	if err != nil {
		log.Errorf("Failed to get inventory for goods %d: %v", id, err)
		goodsResp.Stocks = 0 // 库存获取失败设置为0
		log.Warnf("Set stocks to 0 for goods %d due to inventory error", id)
	} else {
		log.Infof("Inventory retrieved successfully for goods %d: %d stocks", id, inv.Num)
		goodsResp.Stocks = inv.Num
		log.Infof("Set goods %d stocks to %d", id, goodsResp.Stocks)
	}
	
	log.Infof("Final response for goods %d: stocks=%d", id, goodsResp.Stocks)
	return goodsResp, nil
}

func (g *goodsService) SyncGoodsData(ctx context.Context, request *gpbv1.SyncDataRequest) (*gpbv1.SyncDataResponse, error) {
	log.Infof("Admin SyncGoodsData called")
	return g.data.Goods().SyncGoodsData(ctx, request)
}

// ==================== 分类管理 ====================

func (g *goodsService) GetCategoriesList(ctx context.Context) (*gpbv1.CategoryListResponse, error) {
	log.Infof("Admin GetCategoriesList called")
	return g.data.Goods().GetCategoriesList(ctx)
}

func (g *goodsService) GetCategoriesByLevel(ctx context.Context, level int32) (*gpbv1.CategoryListResponse, error) {
	log.Infof("Admin GetCategoriesByLevel called with level: %d", level)
	return g.data.Goods().GetCategoriesByLevel(ctx, level)
}

func (g *goodsService) GetCategoryTree(ctx context.Context) (*gpbv1.CategoryTreeResponse, error) {
	log.Infof("Admin GetCategoryTree called")
	return g.data.Goods().GetCategoryTree(ctx)
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

// ==================== 库存管理 ====================

func (g *goodsService) GetGoodsInventory(ctx context.Context, goodsId int32) (*ipbv1.GoodsInvInfo, error) {
	log.Infof("Admin GetGoodsInventory called for goods ID: %d", goodsId)
	return g.data.Inventory().GetInventory(ctx, goodsId)
}

func (g *goodsService) SetGoodsInventory(ctx context.Context, request *ipbv1.GoodsInvInfo) error {
	log.Infof("Admin SetGoodsInventory called for goods ID: %d, num: %d", request.GoodsId, request.Num)
	return g.data.Inventory().SetInventory(ctx, request)
}

func (g *goodsService) BatchSetGoodsInventory(ctx context.Context, inventories []*ipbv1.GoodsInvInfo) error {
	log.Infof("Admin BatchSetGoodsInventory called for %d items", len(inventories))
	return g.data.Inventory().BatchSetInventory(ctx, inventories)
}

// ==================== 批量操作 ====================

func (g *goodsService) BatchDeleteGoods(ctx context.Context, request *gpbv1.BatchDeleteGoodsRequest) (*gpbv1.BatchOperationResponse, error) {
	log.Infof("Admin BatchDeleteGoods called for %d items", len(request.Ids))
	return g.data.Goods().BatchDeleteGoods(ctx, request)
}

func (g *goodsService) BatchUpdateGoodsStatus(ctx context.Context, request *gpbv1.BatchUpdateGoodsStatusRequest) (*gpbv1.BatchOperationResponse, error) {
	log.Infof("Admin BatchUpdateGoodsStatus called for %d items", len(request.Ids))
	return g.data.Goods().BatchUpdateGoodsStatus(ctx, request)
}