package rpc

import (
	"context"
	gpbv1 "emshop/api/goods/v1"
	"emshop/internal/app/api/admin/data"
	"emshop/pkg/log"

	"google.golang.org/protobuf/types/known/emptypb"
)


type goods struct {
	gc gpbv1.GoodsClient
}

func NewGoods(gc gpbv1.GoodsClient) *goods {
	return &goods{gc}
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

func (g *goods) GetGoodsDetail(ctx context.Context, request *gpbv1.GoodInfoRequest) (*gpbv1.GoodsInfoResponse, error) {
	log.Infof("Calling GetGoodsDetail gRPC for goods ID: %d", request.Id)
	response, err := g.gc.GetGoodsDetail(ctx, request)
	if err != nil {
		log.Errorf("GetGoodsDetail gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("GetGoodsDetail gRPC call successful, goods name: %s", response.Name)
	return response, nil
}

func (g *goods) DeleteGoods(ctx context.Context, info *gpbv1.DeleteGoodsInfo) (*gpbv1.GoodsInfoResponse, error) {
	log.Infof("Calling DeleteGoods gRPC for goods ID: %d", info.Id)
	_, err := g.gc.DeleteGoods(ctx, info)
	if err != nil {
		log.Errorf("DeleteGoods gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("DeleteGoods gRPC call successful, goods ID: %d", info.Id)
	return &gpbv1.GoodsInfoResponse{}, nil
}

func (g *goods) UpdateGoods(ctx context.Context, info *gpbv1.CreateGoodsInfo) (*gpbv1.GoodsInfoResponse, error) {
	log.Infof("Calling UpdateGoods gRPC for goods ID: %d", info.Id)
	_, err := g.gc.UpdateGoods(ctx, info)
	if err != nil {
		log.Errorf("UpdateGoods gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("UpdateGoods gRPC call successful, goods ID: %d", info.Id)
	return &gpbv1.GoodsInfoResponse{}, nil
}

// ==================== 分类管理 ====================

func (g *goods) GetCategoriesList(ctx context.Context) (*gpbv1.CategoryListResponse, error) {
	log.Infof("Calling GetCategoriesList gRPC")
	response, err := g.gc.GetAllCategorysList(ctx, &emptypb.Empty{})
	if err != nil {
		log.Errorf("GetCategoriesList gRPC call failed: %v", err)
		return nil, err
	}
	
	log.Infof("GetCategoriesList gRPC call successful, total: %d", response.Total)
	return response, nil
}

func (g *goods) GetCategoriesByLevel(ctx context.Context, level int32) (*gpbv1.CategoryListResponse, error) {
	log.Infof("Calling GetCategoriesByLevel gRPC with level: %d", level)
	
	// 直接在admin数据层实现，调用goods服务的gRPC方法获取所有分类
	// 然后在这里按层级过滤
	response, err := g.gc.GetAllCategorysList(ctx, &emptypb.Empty{})
	if err != nil {
		log.Errorf("GetCategoriesByLevel gRPC call failed: %v", err)
		return nil, err
	}
	
	log.Infof("GetAllCategorysList returned %d total categories", len(response.Data))
	
	// 过滤出指定层级的分类
	var filteredCategories []*gpbv1.CategoryInfoResponse
	for _, category := range response.Data {
		// log.Debugf("Category %d: ID=%d, Name=%s, Level=%d", i, category.Id, category.Name, category.Level)
		if category.Level == level {
			filteredCategories = append(filteredCategories, category)
		}
	}
	
	filteredResponse := &gpbv1.CategoryListResponse{
		Total: int32(len(filteredCategories)),
		Data:  filteredCategories,
	}
	
	log.Infof("GetCategoriesByLevel completed, found %d categories at level %d", len(filteredCategories), level)
	return filteredResponse, nil
}

func (g *goods) GetCategoryTree(ctx context.Context) (*gpbv1.CategoryTreeResponse, error) {
	log.Infof("Calling GetCategoryTree gRPC")
	response, err := g.gc.GetCategoryTree(ctx, &emptypb.Empty{})
	if err != nil {
		log.Errorf("GetCategoryTree gRPC call failed: %v", err)
		return nil, err
	}
	
	log.Infof("GetCategoryTree completed, got %d root categories with %d total", 
		len(response.Categories), response.Stats.TotalCount)
	return response, nil
}

func (g *goods) GetSubCategory(ctx context.Context, request *gpbv1.CategoryListRequest) (*gpbv1.SubCategoryListResponse, error) {
	log.Infof("Calling GetSubCategory gRPC for category ID: %d", request.Id)
	response, err := g.gc.GetSubCategory(ctx, request)
	if err != nil {
		log.Errorf("GetSubCategory gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("GetSubCategory gRPC call successful")
	return response, nil
}

func (g *goods) CreateCategory(ctx context.Context, request *gpbv1.CategoryInfoRequest) (*gpbv1.CategoryInfoResponse, error) {
	log.Infof("Calling CreateCategory gRPC: %s", request.Name)
	response, err := g.gc.CreateCategory(ctx, request)
	if err != nil {
		log.Errorf("CreateCategory gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("CreateCategory gRPC call successful, category ID: %d", response.Id)
	return response, nil
}

func (g *goods) UpdateCategory(ctx context.Context, request *gpbv1.CategoryInfoRequest) (*gpbv1.CategoryInfoResponse, error) {
	log.Infof("Calling UpdateCategory gRPC for category ID: %d", request.Id)
	_, err := g.gc.UpdateCategory(ctx, request)
	if err != nil {
		log.Errorf("UpdateCategory gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("UpdateCategory gRPC call successful, category ID: %d", request.Id)
	return &gpbv1.CategoryInfoResponse{}, nil
}

func (g *goods) DeleteCategory(ctx context.Context, request *gpbv1.DeleteCategoryRequest) (*gpbv1.CategoryInfoResponse, error) {
	log.Infof("Calling DeleteCategory gRPC for category ID: %d", request.Id)
	_, err := g.gc.DeleteCategory(ctx, request)
	if err != nil {
		log.Errorf("DeleteCategory gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("DeleteCategory gRPC call successful, category ID: %d", request.Id)
	return &gpbv1.CategoryInfoResponse{}, nil
}

// ==================== 品牌管理 ====================

func (g *goods) BrandList(ctx context.Context, request *gpbv1.BrandFilterRequest) (*gpbv1.BrandListResponse, error) {
	log.Infof("Calling BrandList gRPC with pages: %d", request.Pages)
	response, err := g.gc.BrandList(ctx, request)
	if err != nil {
		log.Errorf("BrandList gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("BrandList gRPC call successful, total: %d", response.Total)
	return response, nil
}

func (g *goods) CreateBrand(ctx context.Context, request *gpbv1.BrandRequest) (*gpbv1.BrandInfoResponse, error) {
	log.Infof("Calling CreateBrand gRPC: %s", request.Name)
	response, err := g.gc.CreateBrand(ctx, request)
	if err != nil {
		log.Errorf("CreateBrand gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("CreateBrand gRPC call successful, brand ID: %d", response.Id)
	return response, nil
}

func (g *goods) UpdateBrand(ctx context.Context, request *gpbv1.BrandRequest) (*gpbv1.BrandInfoResponse, error) {
	log.Infof("Calling UpdateBrand gRPC for brand ID: %d", request.Id)
	_, err := g.gc.UpdateBrand(ctx, request)
	if err != nil {
		log.Errorf("UpdateBrand gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("UpdateBrand gRPC call successful, brand ID: %d", request.Id)
	return &gpbv1.BrandInfoResponse{}, nil
}

func (g *goods) DeleteBrand(ctx context.Context, request *gpbv1.BrandRequest) (*gpbv1.BrandInfoResponse, error) {
	log.Infof("Calling DeleteBrand gRPC for brand ID: %d", request.Id)
	_, err := g.gc.DeleteBrand(ctx, request)
	if err != nil {
		log.Errorf("DeleteBrand gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("DeleteBrand gRPC call successful, brand ID: %d", request.Id)
	return &gpbv1.BrandInfoResponse{}, nil
}

// ==================== 轮播图管理 ====================

func (g *goods) BannerList(ctx context.Context) (*gpbv1.BannerListResponse, error) {
	log.Infof("Calling BannerList gRPC")
	response, err := g.gc.BannerList(ctx, &emptypb.Empty{})
	if err != nil {
		log.Errorf("BannerList gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("BannerList gRPC call successful, total: %d", response.Total)
	return response, nil
}

func (g *goods) CreateBanner(ctx context.Context, request *gpbv1.BannerRequest) (*gpbv1.BannerResponse, error) {
	log.Infof("Calling CreateBanner gRPC: %s", request.Url)
	response, err := g.gc.CreateBanner(ctx, request)
	if err != nil {
		log.Errorf("CreateBanner gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("CreateBanner gRPC call successful, banner ID: %d", response.Id)
	return response, nil
}

func (g *goods) UpdateBanner(ctx context.Context, request *gpbv1.BannerRequest) (*gpbv1.BannerResponse, error) {
	log.Infof("Calling UpdateBanner gRPC for banner ID: %d", request.Id)
	_, err := g.gc.UpdateBanner(ctx, request)
	if err != nil {
		log.Errorf("UpdateBanner gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("UpdateBanner gRPC call successful, banner ID: %d", request.Id)
	return &gpbv1.BannerResponse{}, nil
}

func (g *goods) DeleteBanner(ctx context.Context, request *gpbv1.BannerRequest) (*gpbv1.BannerResponse, error) {
	log.Infof("Calling DeleteBanner gRPC for banner ID: %d", request.Id)
	_, err := g.gc.DeleteBanner(ctx, request)
	if err != nil {
		log.Errorf("DeleteBanner gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("DeleteBanner gRPC call successful, banner ID: %d", request.Id)
	return &gpbv1.BannerResponse{}, nil
}

// ==================== 批量操作 ====================

func (g *goods) BatchDeleteGoods(ctx context.Context, request *gpbv1.BatchDeleteGoodsRequest) (*gpbv1.BatchOperationResponse, error) {
	log.Infof("Calling BatchDeleteGoods gRPC for %d items", len(request.Ids))
	return g.gc.BatchDeleteGoods(ctx, request)
}

func (g *goods) BatchUpdateGoodsStatus(ctx context.Context, request *gpbv1.BatchUpdateGoodsStatusRequest) (*gpbv1.BatchOperationResponse, error) {
	log.Infof("Calling BatchUpdateGoodsStatus gRPC for %d items", len(request.Ids))
	return g.gc.BatchUpdateGoodsStatus(ctx, request)
}

var _ data.GoodsData = &goods{}