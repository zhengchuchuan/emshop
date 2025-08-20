package v1

import (
	"context"
	"emshop/internal/app/goods/srv/domain/do"
	"emshop/internal/app/goods/srv/domain/dto"
	v12 "emshop/pkg/common/meta/v1"
	"emshop/pkg/errors"

	"google.golang.org/protobuf/types/known/emptypb"
	proto "emshop/api/goods/v1"
	v1 "emshop/internal/app/goods/srv/service/v1"
	"emshop/pkg/log"
)

type goodsServer struct {
	proto.UnimplementedGoodsServer
	srv v1.ServiceFactory
}

// Common error codes for validation
const (
	ErrInvalidParameter = 100400 // HTTP 400 Bad Request equivalent
)

// Validation functions
func validateGoodsFilterRequest(request *proto.GoodsFilterRequest) error {
	if request.Pages < 0 {
		return errors.WithCode(ErrInvalidParameter, "pages must be non-negative")
	}
	if request.PagePerNums < 0 {
		return errors.WithCode(ErrInvalidParameter, "pagePerNums must be non-negative")
	}
	if request.PriceMin < 0 || request.PriceMax < 0 {
		return errors.WithCode(ErrInvalidParameter, "price range must be non-negative")
	}
	if request.PriceMin > 0 && request.PriceMax > 0 && request.PriceMin > request.PriceMax {
		return errors.WithCode(ErrInvalidParameter, "priceMin cannot be greater than priceMax")
	}
	return nil
}

func validateCreateGoodsInfo(info *proto.CreateGoodsInfo) error {
	if info.Name == "" {
		return errors.WithCode(ErrInvalidParameter, "goods name is required")
	}
	if info.GoodsSn == "" {
		return errors.WithCode(ErrInvalidParameter, "goods SN is required")
	}
	if info.CategoryId <= 0 {
		return errors.WithCode(ErrInvalidParameter, "invalid category ID")
	}
	if info.BrandId <= 0 {
		return errors.WithCode(ErrInvalidParameter, "invalid brand ID")
	}
	if info.ShopPrice < 0 {
		return errors.WithCode(ErrInvalidParameter, "shop price must be non-negative")
	}
	if info.MarketPrice < 0 {
		return errors.WithCode(ErrInvalidParameter, "market price must be non-negative")
	}
	return nil
}

func validateCategoryInfoRequest(request *proto.CategoryInfoRequest) error {
	if request.Name == "" {
		return errors.WithCode(ErrInvalidParameter, "category name is required")
	}
	if request.Level < 1 || request.Level > 3 {
		return errors.WithCode(ErrInvalidParameter, "category level must be between 1 and 3")
	}
	if request.Level > 1 && request.ParentCategory <= 0 {
		return errors.WithCode(ErrInvalidParameter, "parent category is required for non-root categories")
	}
	return nil
}

func validateBrandRequest(request *proto.BrandRequest) error {
	if request.Name == "" {
		return errors.WithCode(ErrInvalidParameter, "brand name is required")
	}
	return nil
}

func validateBannerRequest(request *proto.BannerRequest) error {
	if request.Image == "" {
		return errors.WithCode(ErrInvalidParameter, "banner image is required")
	}
	if request.Url == "" {
		return errors.WithCode(ErrInvalidParameter, "banner URL is required")
	}
	return nil
}

func (gs *goodsServer) GoodsList(ctx context.Context, request *proto.GoodsFilterRequest) (*proto.GoodsListResponse, error) {
	// Input validation
	if err := validateGoodsFilterRequest(request); err != nil {
		log.Errorf("invalid goods filter request: %v", err)
		return nil, err
	}

	list, err := gs.srv.Goods().List(ctx, v12.ListMeta{Page: int(request.Pages), PageSize: int(request.PagePerNums)}, request, []string{})
	if err != nil {
		log.Errorf("get goods list error: %v", err.Error())
		return nil, err
	}
	var ret proto.GoodsListResponse
	ret.Total = int32(list.TotalCount)
	for _, item := range list.Items {
		ret.Data = append(ret.Data, ModelToResponse(item))
	}
	return &ret, nil
}

func (gs *goodsServer) BatchGetGoods(ctx context.Context, info *proto.BatchGoodsIdInfo) (*proto.GoodsListResponse, error) {
	// Input validation
	if len(info.Id) == 0 {
		return nil, errors.WithCode(ErrInvalidParameter, "goods IDs are required")
	}
	if len(info.Id) > 100 {
		return nil, errors.WithCode(ErrInvalidParameter, "too many goods IDs, maximum 100 allowed")
	}

	var ids []uint64
	for _, id := range info.Id {
		if id <= 0 {
			return nil, errors.WithCode(ErrInvalidParameter, "invalid goods ID")
		}
		ids = append(ids, uint64(id))
	}
	
	get, err := gs.srv.Goods().BatchGet(ctx, ids)
	if err != nil {
		log.Errorf("batch get goods error: %v", err)
		return nil, err
	}
	var ret proto.GoodsListResponse
	for _, item := range get {
		ret.Data = append(ret.Data, ModelToResponse(item))
	}
	return &ret, nil
}

func (gs *goodsServer) CreateGoods(ctx context.Context, info *proto.CreateGoodsInfo) (*proto.GoodsInfoResponse, error) {
	// Input validation
	if err := validateCreateGoodsInfo(info); err != nil {
		log.Errorf("invalid create goods info: %v", err)
		return nil, err
	}

	// 构建商品DTO
	goodsDTO := &dto.GoodsDTO{}
	goodsDTO.Name = info.Name
	goodsDTO.GoodsSn = info.GoodsSn
	goodsDTO.CategoryID = info.CategoryId
	goodsDTO.BrandsID = info.BrandId
	goodsDTO.MarketPrice = info.MarketPrice
	goodsDTO.ShopPrice = info.ShopPrice
	goodsDTO.GoodsBrief = info.GoodsBrief
	goodsDTO.GoodsDesc = info.GoodsDesc
	goodsDTO.ShipFree = info.ShipFree
	goodsDTO.Images = info.Images
	goodsDTO.DescImages = info.DescImages
	goodsDTO.GoodsFrontImage = info.GoodsFrontImage
	goodsDTO.IsNew = info.IsNew
	goodsDTO.IsHot = info.IsHot
	goodsDTO.OnSale = info.OnSale

	// 创建商品（业务层会验证分类和品牌）
	createdGoods, err := gs.srv.Goods().Create(ctx, goodsDTO)
	if err != nil {
		log.Errorf("create goods error: %v", err)
		return nil, err
	}

	return &proto.GoodsInfoResponse{
		Id: createdGoods.ID,
	}, nil
}

func (gs *goodsServer) DeleteGoods(ctx context.Context, info *proto.DeleteGoodsInfo) (*emptypb.Empty, error) {
	// Input validation
	if info.Id <= 0 {
		return nil, errors.WithCode(ErrInvalidParameter, "invalid goods ID")
	}

	err := gs.srv.Goods().Delete(ctx, uint64(info.Id))
	if err != nil {
		log.Errorf("delete goods error: %v", err)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (gs *goodsServer) UpdateGoods(ctx context.Context, info *proto.CreateGoodsInfo) (*emptypb.Empty, error) {
	// Input validation
	if info.Id <= 0 {
		return nil, errors.WithCode(ErrInvalidParameter, "invalid goods ID")
	}
	if err := validateCreateGoodsInfo(info); err != nil {
		log.Errorf("invalid update goods info: %v", err)
		return nil, err
	}

	// 构建商品DTO
	goodsDTO := &dto.GoodsDTO{}
	goodsDTO.ID = info.Id
	goodsDTO.Name = info.Name
	goodsDTO.GoodsSn = info.GoodsSn
	goodsDTO.CategoryID = info.CategoryId
	goodsDTO.BrandsID = info.BrandId
	goodsDTO.MarketPrice = info.MarketPrice
	goodsDTO.ShopPrice = info.ShopPrice
	goodsDTO.GoodsBrief = info.GoodsBrief
	goodsDTO.GoodsDesc = info.GoodsDesc
	goodsDTO.ShipFree = info.ShipFree
	goodsDTO.Images = info.Images
	goodsDTO.DescImages = info.DescImages
	goodsDTO.GoodsFrontImage = info.GoodsFrontImage
	goodsDTO.IsNew = info.IsNew
	goodsDTO.IsHot = info.IsHot
	goodsDTO.OnSale = info.OnSale

	// 更新商品
	err := gs.srv.Goods().Update(ctx, goodsDTO)
	if err != nil {
		log.Errorf("update goods error: %v", err)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (gs *goodsServer) GetGoodsDetail(ctx context.Context, request *proto.GoodInfoRequest) (*proto.GoodsInfoResponse, error) {
	// Input validation
	if request.Id <= 0 {
		return nil, errors.WithCode(ErrInvalidParameter, "invalid goods ID")
	}

	goods, err := gs.srv.Goods().Get(ctx, uint64(request.Id))
	if err != nil {
		log.Errorf("get goods detail error: %v", err)
		return nil, err
	}

	return ModelToResponse(goods), nil
}

func (gs *goodsServer) GetAllCategorysList(ctx context.Context, empty *emptypb.Empty) (*proto.CategoryListResponse, error) {
	categories, err := gs.srv.Category().ListAll(ctx, []string{})
	if err != nil {
		log.Errorf("get all categories error: %v", err)
		return nil, err
	}

	// 首先扁平化所有嵌套的分类数据
	allCategories := flattenCategories(categories.Items)
	
	log.Infof("GetAllCategorysList: flattened %d categories from %d root categories", len(allCategories), len(categories.Items))

	response := &proto.CategoryListResponse{
		Total: int32(len(allCategories)), // 使用扁平化后的总数
	}

	// 构建树状结构的分类数据
	categoryMap := make(map[int32][]*proto.CategoryInfoResponse)
	var rootCategories []*proto.CategoryInfoResponse

	// 处理所有扁平化后的分类
	for _, item := range allCategories {
		categoryInfo := &proto.CategoryInfoResponse{
			Id:             item.ID,
			Name:           item.Name,
			Level:          item.Level,
			IsTab:          item.IsTab,
			ParentCategory: item.ParentCategoryID,
		}

		if item.ParentCategoryID == 0 {
			rootCategories = append(rootCategories, categoryInfo)
		} else {
			categoryMap[item.ParentCategoryID] = append(categoryMap[item.ParentCategoryID], categoryInfo)
		}
	}

	// 将所有分类（包括层级关系）添加到响应中
	response.Data = rootCategories
	for _, children := range categoryMap {
		response.Data = append(response.Data, children...)
	}

	log.Infof("GetAllCategorysList: returning %d total categories in response", len(response.Data))
	return response, nil
}

func (gs *goodsServer) GetSubCategory(ctx context.Context, request *proto.CategoryListRequest) (*proto.SubCategoryListResponse, error) {
	subCategories, err := gs.srv.Category().GetSubCategories(ctx, request.Id)
	if err != nil {
		log.Errorf("get sub categories error: %v", err)
		return nil, err
	}

	response := &proto.SubCategoryListResponse{}
	if subCategories.ParentInfo != nil {
		response.Info = &proto.CategoryInfoResponse{
			Id:             subCategories.ParentInfo.ID,
			Name:           subCategories.ParentInfo.Name,
			Level:          subCategories.ParentInfo.Level,
			IsTab:          subCategories.ParentInfo.IsTab,
			ParentCategory: subCategories.ParentInfo.ParentCategoryID,
		}
	}

	for _, item := range subCategories.Items {
		response.SubCategorys = append(response.SubCategorys, &proto.CategoryInfoResponse{
			Id:             item.ID,
			Name:           item.Name,
			Level:          item.Level,
			IsTab:          item.IsTab,
			ParentCategory: item.ParentCategoryID,
		})
	}

	return response, nil
}

func (gs *goodsServer) GetCategoryTree(ctx context.Context, empty *emptypb.Empty) (*proto.CategoryTreeResponse, error) {
	categories, err := gs.srv.Category().ListAll(ctx, []string{})
	if err != nil {
		log.Errorf("get category tree error: %v", err)
		return nil, err
	}
	
	// 构建树形结构
	treeNodes := buildCategoryTreeNodes(categories.Items)
	
	// 计算统计信息
	stats := calculateCategoryStats(categories.Items)
	
	response := &proto.CategoryTreeResponse{
		Categories: treeNodes,
		Stats:      stats,
	}
	
	log.Infof("GetCategoryTree: returning %d root categories with %d total categories", 
		len(treeNodes), stats.TotalCount)
	
	return response, nil
}

// buildCategoryTreeNodes 将DTO结构转换为protobuf树形结构
func buildCategoryTreeNodes(categories []*dto.CategoryDTO) []*proto.CategoryTreeNode {
	var nodes []*proto.CategoryTreeNode
	
	for _, category := range categories {
		node := &proto.CategoryTreeNode{
			Id:             category.ID,
			Name:           category.Name,
			ParentCategory: category.ParentCategoryID,
			Level:          category.Level,
			IsTab:          category.IsTab,
		}
		
		// 递归构建子节点
		if len(category.SubCategories) > 0 {
			node.Children = buildCategoryTreeNodes(category.SubCategories)
		}
		
		nodes = append(nodes, node)
	}
	
	return nodes
}

// calculateCategoryStats 计算分类统计信息
func calculateCategoryStats(categories []*dto.CategoryDTO) *proto.CategoryStatistics {
	stats := &proto.CategoryStatistics{}
	
	// 递归统计各级分类
	calculateStatsRecursive(categories, stats, 1)
	
	return stats
}

func calculateStatsRecursive(categories []*dto.CategoryDTO, stats *proto.CategoryStatistics, currentDepth int32) {
	for _, category := range categories {
		stats.TotalCount++
		
		// 统计各级别分类数量
		switch category.Level {
		case 1:
			stats.Level1Count++
		case 2:
			stats.Level2Count++
		case 3:
			stats.Level3Count++
		}
		
		// 更新最大深度
		if currentDepth > stats.MaxDepth {
			stats.MaxDepth = currentDepth
		}
		
		// 递归处理子分类
		if len(category.SubCategories) > 0 {
			calculateStatsRecursive(category.SubCategories, stats, currentDepth+1)
		}
	}
}

// flattenCategories 递归将嵌套的分类结构扁平化
func flattenCategories(categories []*dto.CategoryDTO) []*dto.CategoryDTO {
	var flattened []*dto.CategoryDTO
	
	for _, category := range categories {
		// 添加当前分类
		flattened = append(flattened, category)
		
		// 递归添加子分类
		if len(category.SubCategories) > 0 {
			subFlattened := flattenCategories(category.SubCategories)
			flattened = append(flattened, subFlattened...)
		}
	}
	
	return flattened
}

func (gs *goodsServer) CreateCategory(ctx context.Context, request *proto.CategoryInfoRequest) (*proto.CategoryInfoResponse, error) {
	// Input validation
	if err := validateCategoryInfoRequest(request); err != nil {
		log.Errorf("invalid create category request: %v", err)
		return nil, err
	}

	categoryDTO := &dto.CategoryDTO{
		CategoryDO: do.CategoryDO{
			Name:             request.Name,
			ParentCategoryID: request.ParentCategory,
			Level:            request.Level,
			IsTab:            request.IsTab,
		},
	}

	err := gs.srv.Category().Create(ctx, categoryDTO)
	if err != nil {
		log.Errorf("create category error: %v", err)
		return nil, err
	}

	return &proto.CategoryInfoResponse{Id: categoryDTO.ID}, nil
}

func (gs *goodsServer) DeleteCategory(ctx context.Context, request *proto.DeleteCategoryRequest) (*emptypb.Empty, error) {
	err := gs.srv.Category().Delete(ctx, request.Id)
	if err != nil {
		log.Errorf("delete category error: %v", err)
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (gs *goodsServer) UpdateCategory(ctx context.Context, request *proto.CategoryInfoRequest) (*emptypb.Empty, error) {
	// 先获取现有分类信息
	existing, err := gs.srv.Category().Get(ctx, request.Id)
	if err != nil {
		log.Errorf("get existing category error: %v", err)
		return nil, err
	}

	// 只更新提供的字段，保留原有的ParentCategoryID和Level
	categoryDTO := &dto.CategoryDTO{
		CategoryDO: do.CategoryDO{
			Name:             request.Name,
			ParentCategoryID: existing.ParentCategoryID, // 保持原有值
			Level:            existing.Level,            // 保持原有值
			IsTab:            request.IsTab,
			Url:              existing.Url, // 保持原有值
		},
	}
	categoryDTO.ID = request.Id

	err = gs.srv.Category().Update(ctx, categoryDTO)
	if err != nil {
		log.Errorf("update category error: %v", err)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (gs *goodsServer) BrandList(ctx context.Context, request *proto.BrandFilterRequest) (*proto.BrandListResponse, error) {
	// 如果Pages或PagePerNums为nil，表示不分页，返回所有数据
	listMeta := v12.ListMeta{}
	if request.Pages == nil && request.PagePerNums == nil {
		// 设置一个很大的PageSize以获取所有数据，Page设置为1
		listMeta.Page = 1
		listMeta.PageSize = 10000 // 足够大的数字来获取所有品牌数据
	} else {
		// 正常分页逻辑
		page := 1
		pageSize := 10
		
		if request.Pages != nil {
			page = int(*request.Pages)
			if page <= 0 {
				page = 1
			}
		}
		
		if request.PagePerNums != nil {
			pageSize = int(*request.PagePerNums)
			if pageSize <= 0 {
				pageSize = 10
			}
		}
		
		listMeta.Page = page
		listMeta.PageSize = pageSize
	}
	
	brands, err := gs.srv.Brand().List(ctx, listMeta, []string{})
	if err != nil {
		log.Errorf("get brands list error: %v", err)
		return nil, err
	}

	response := &proto.BrandListResponse{
		Total: int32(brands.TotalCount),
	}

	for _, item := range brands.Items {
		response.Data = append(response.Data, &proto.BrandInfoResponse{
			Id:   item.ID,
			Name: item.Name,
			Logo: item.Logo,
		})
	}

	return response, nil
}

func (gs *goodsServer) CreateBrand(ctx context.Context, request *proto.BrandRequest) (*proto.BrandInfoResponse, error) {
	// Input validation
	if err := validateBrandRequest(request); err != nil {
		log.Errorf("invalid create brand request: %v", err)
		return nil, err
	}

	brandDTO := &dto.BrandDTO{
		BrandsDO: do.BrandsDO{
			Name: request.Name,
			Logo: request.Logo,
		},
	}

	err := gs.srv.Brand().Create(ctx, brandDTO)
	if err != nil {
		log.Errorf("create brand error: %v", err)
		return nil, err
	}

	return &proto.BrandInfoResponse{Id: brandDTO.ID}, nil
}

func (gs *goodsServer) DeleteBrand(ctx context.Context, request *proto.BrandRequest) (*emptypb.Empty, error) {
	err := gs.srv.Brand().Delete(ctx, request.Id)
	if err != nil {
		log.Errorf("delete brand error: %v", err)
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (gs *goodsServer) UpdateBrand(ctx context.Context, request *proto.BrandRequest) (*emptypb.Empty, error) {
	brandDTO := &dto.BrandDTO{
		BrandsDO: do.BrandsDO{
			Name: request.Name,
			Logo: request.Logo,
		},
	}
	brandDTO.ID = request.Id

	err := gs.srv.Brand().Update(ctx, brandDTO)
	if err != nil {
		log.Errorf("update brand error: %v", err)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (gs *goodsServer) BannerList(ctx context.Context, empty *emptypb.Empty) (*proto.BannerListResponse, error) {
	banners, err := gs.srv.Banner().List(ctx, []string{})
	if err != nil {
		log.Errorf("get banners list error: %v", err)
		return nil, err
	}

	response := &proto.BannerListResponse{
		Total: int32(banners.TotalCount),
	}

	for _, item := range banners.Items {
		response.Data = append(response.Data, &proto.BannerResponse{
			Id:    item.ID,
			Image: item.Image,
			Url:   item.Url,
			Index: item.Index,
		})
	}

	return response, nil
}

func (gs *goodsServer) CreateBanner(ctx context.Context, request *proto.BannerRequest) (*proto.BannerResponse, error) {
	// Input validation
	if err := validateBannerRequest(request); err != nil {
		log.Errorf("invalid create banner request: %v", err)
		return nil, err
	}

	bannerDTO := &dto.BannerDTO{
		BannerDO: do.BannerDO{
			Image: request.Image,
			Url:   request.Url,
			Index: request.Index,
		},
	}

	err := gs.srv.Banner().Create(ctx, bannerDTO)
	if err != nil {
		log.Errorf("create banner error: %v", err)
		return nil, err
	}

	return &proto.BannerResponse{Id: bannerDTO.ID}, nil
}

func (gs *goodsServer) DeleteBanner(ctx context.Context, request *proto.BannerRequest) (*emptypb.Empty, error) {
	err := gs.srv.Banner().Delete(ctx, request.Id)
	if err != nil {
		log.Errorf("delete banner error: %v", err)
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (gs *goodsServer) UpdateBanner(ctx context.Context, request *proto.BannerRequest) (*emptypb.Empty, error) {
	bannerDTO := &dto.BannerDTO{
		BannerDO: do.BannerDO{
			Image: request.Image,
			Url:   request.Url,
			Index: request.Index,
		},
	}
	bannerDTO.ID = request.Id

	err := gs.srv.Banner().Update(ctx, bannerDTO)
	if err != nil {
		log.Errorf("update banner error: %v", err)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (gs *goodsServer) CategoryBrandList(ctx context.Context, request *proto.CategoryBrandFilterRequest) (*proto.CategoryBrandListResponse, error) {
	categoryBrands, err := gs.srv.CategoryBrand().List(ctx, v12.ListMeta{Page: int(request.Pages), PageSize: int(request.PagePerNums)}, []string{})
	if err != nil {
		log.Errorf("get category brands list error: %v", err)
		return nil, err
	}

	response := &proto.CategoryBrandListResponse{
		Total: int32(categoryBrands.TotalCount),
	}

	for _, item := range categoryBrands.Items {
		response.Data = append(response.Data, &proto.CategoryBrandResponse{
			Id: item.ID,
			Category: &proto.CategoryInfoResponse{
				Id:             item.Category.ID,
				Name:           item.Category.Name,
				Level:          item.Category.Level,
				IsTab:          item.Category.IsTab,
				ParentCategory: item.Category.ParentCategoryID,
			},
			Brand: &proto.BrandInfoResponse{
				Id:   item.Brands.ID,
				Name: item.Brands.Name,
				Logo: item.Brands.Logo,
			},
		})
	}

	return response, nil
}

func (gs *goodsServer) GetCategoryBrandList(ctx context.Context, request *proto.CategoryInfoRequest) (*proto.BrandListResponse, error) {
	brands, err := gs.srv.CategoryBrand().GetBrandsByCategory(ctx, request.Id)
	if err != nil {
		log.Errorf("get brands by category error: %v", err)
		return nil, err
	}

	response := &proto.BrandListResponse{
		Total: int32(brands.TotalCount),
	}

	for _, item := range brands.Items {
		response.Data = append(response.Data, &proto.BrandInfoResponse{
			Id:   item.ID,
			Name: item.Name,
			Logo: item.Logo,
		})
	}

	return response, nil
}

func (gs *goodsServer) CreateCategoryBrand(ctx context.Context, request *proto.CategoryBrandRequest) (*proto.CategoryBrandResponse, error) {
	categoryBrandDTO := &dto.CategoryBrandDTO{
		GoodsCategoryBrandDO: do.GoodsCategoryBrandDO{
			CategoryID: request.CategoryId,
			BrandsID:   request.BrandId,
		},
	}

	err := gs.srv.CategoryBrand().Create(ctx, categoryBrandDTO)
	if err != nil {
		log.Errorf("create category brand error: %v", err)
		return nil, err
	}

	return &proto.CategoryBrandResponse{Id: categoryBrandDTO.ID}, nil
}

func (gs *goodsServer) DeleteCategoryBrand(ctx context.Context, request *proto.CategoryBrandRequest) (*emptypb.Empty, error) {
	err := gs.srv.CategoryBrand().Delete(ctx, request.Id)
	if err != nil {
		log.Errorf("delete category brand error: %v", err)
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (gs *goodsServer) UpdateCategoryBrand(ctx context.Context, request *proto.CategoryBrandRequest) (*emptypb.Empty, error) {
	categoryBrandDTO := &dto.CategoryBrandDTO{
		GoodsCategoryBrandDO: do.GoodsCategoryBrandDO{
			CategoryID: request.CategoryId,
			BrandsID:   request.BrandId,
		},
	}
	categoryBrandDTO.ID = request.Id

	err := gs.srv.CategoryBrand().Update(ctx, categoryBrandDTO)
	if err != nil {
		log.Errorf("update category brand error: %v", err)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (gs *goodsServer) SyncGoodsData(ctx context.Context, request *proto.SyncDataRequest) (*proto.SyncDataResponse, error) {
	log.Infof("sync goods data request: forceSync=%v, goodsIds=%v", request.ForceSync, request.GoodsIds)

	// 转换goodsIds到uint64
	var goodsIds []uint64
	for _, id := range request.GoodsIds {
		goodsIds = append(goodsIds, uint64(id))
	}

	result, err := gs.srv.DataSync().SyncGoodsData(ctx, request.ForceSync, goodsIds)
	if err != nil {
		log.Errorf("sync goods data error: %v", err)
		return &proto.SyncDataResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &proto.SyncDataResponse{
		Success:     true,
		Message:     "Data sync completed successfully",
		SyncedCount: int32(result.SyncedCount),
		FailedCount: int32(result.FailedCount),
		Errors:      result.Errors,
	}, nil
}

func NewGoodsServer(srv v1.ServiceFactory) *goodsServer {
	return &goodsServer{srv: srv}
}

func ModelToResponse(goods *dto.GoodsDTO) *proto.GoodsInfoResponse {
	return &proto.GoodsInfoResponse{
		Id:              goods.ID,
		CategoryId:      goods.CategoryID,
		Name:            goods.Name,
		GoodsSn:         goods.GoodsSn,
		ClickNum:        goods.ClickNum,
		SoldNum:         goods.SoldNum,
		FavNum:          goods.FavNum,
		MarketPrice:     goods.MarketPrice,
		ShopPrice:       goods.ShopPrice,
		GoodsBrief:      goods.GoodsBrief,
		ShipFree:        goods.ShipFree,
		GoodsFrontImage: goods.GoodsFrontImage,
		IsNew:           goods.IsNew,
		IsHot:           goods.IsHot,
		OnSale:          goods.OnSale,
		DescImages:      goods.DescImages,
		Images:          goods.Images,
		Category: &proto.CategoryBriefInfoResponse{
			Id:   goods.Category.ID,
			Name: goods.Category.Name,
		},
		Brand: &proto.BrandInfoResponse{
			Id:   goods.Brands.ID,
			Name: goods.Brands.Name,
			Logo: goods.Brands.Logo,
		},
	}
}

