package v1

import (
	"context"
	"emshop/internal/app/goods/srv/domain/dto"
	v12 "emshop/pkg/common/meta/v1"

	"google.golang.org/protobuf/types/known/emptypb"
	proto "emshop/api/goods/v1"
	v1 "emshop/internal/app/goods/srv/service/v1"
	"emshop/pkg/log"
)

type goodsServer struct {
	proto.UnimplementedGoodsServer
	srv v1.ServiceFactory
}

func (gs *goodsServer) GoodsList(ctx context.Context, request *proto.GoodsFilterRequest) (*proto.GoodsListResponse, error) {
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
	var ids []uint64
	for _, id := range info.Id {
		ids = append(ids, uint64(id))
	}
	get, err := gs.srv.Goods().BatchGet(ctx, ids)
	if err != nil {
		return nil, err
	}
	var ret proto.GoodsListResponse
	for _, item := range get {
		ret.Data = append(ret.Data, ModelToResponse(item))
	}
	return &ret, nil
}

func (gs *goodsServer) CreateGoods(ctx context.Context, info *proto.CreateGoodsInfo) (*proto.GoodsInfoResponse, error) {
	// 构建商品DTO
	goodsDTO := &dto.GoodsDTO{}
	goodsDTO.Name = info.Name
	goodsDTO.GoodsSn = info.GoodsSn
	goodsDTO.CategoryID = info.CategoryId
	goodsDTO.BrandsID = info.BrandId
	goodsDTO.MarketPrice = info.MarketPrice
	goodsDTO.ShopPrice = info.ShopPrice
	goodsDTO.GoodsBrief = info.GoodsBrief
	goodsDTO.ShipFree = info.ShipFree
	goodsDTO.Images = info.Images
	goodsDTO.DescImages = info.DescImages
	goodsDTO.GoodsFrontImage = info.GoodsFrontImage
	goodsDTO.IsNew = info.IsNew
	goodsDTO.IsHot = info.IsHot
	goodsDTO.OnSale = info.OnSale

	// 创建商品（业务层会验证分类和品牌）
	err := gs.srv.Goods().Create(ctx, goodsDTO)
	if err != nil {
		log.Errorf("create goods error: %v", err)
		return nil, err
	}

	return &proto.GoodsInfoResponse{
		Id: goodsDTO.ID,
	}, nil
}

func (gs *goodsServer) DeleteGoods(ctx context.Context, info *proto.DeleteGoodsInfo) (*emptypb.Empty, error) {
	err := gs.srv.Goods().Delete(ctx, uint64(info.Id))
	if err != nil {
		log.Errorf("delete goods error: %v", err)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (gs *goodsServer) UpdateGoods(ctx context.Context, info *proto.CreateGoodsInfo) (*emptypb.Empty, error) {
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
	goods, err := gs.srv.Goods().Get(ctx, uint64(request.Id))
	if err != nil {
		log.Errorf("get goods detail error: %v", err)
		return nil, err
	}

	return ModelToResponse(goods), nil
}

func (gs *goodsServer) GetAllCategorysList(ctx context.Context, empty *emptypb.Empty) (*proto.CategoryListResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (gs *goodsServer) GetSubCategory(ctx context.Context, request *proto.CategoryListRequest) (*proto.SubCategoryListResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (gs *goodsServer) CreateCategory(ctx context.Context, request *proto.CategoryInfoRequest) (*proto.CategoryInfoResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (gs *goodsServer) DeleteCategory(ctx context.Context, request *proto.DeleteCategoryRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (gs *goodsServer) UpdateCategory(ctx context.Context, request *proto.CategoryInfoRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (gs *goodsServer) BrandList(ctx context.Context, request *proto.BrandFilterRequest) (*proto.BrandListResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (gs *goodsServer) CreateBrand(ctx context.Context, request *proto.BrandRequest) (*proto.BrandInfoResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (gs *goodsServer) DeleteBrand(ctx context.Context, request *proto.BrandRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (gs *goodsServer) UpdateBrand(ctx context.Context, request *proto.BrandRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (gs *goodsServer) BannerList(ctx context.Context, empty *emptypb.Empty) (*proto.BannerListResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (gs *goodsServer) CreateBanner(ctx context.Context, request *proto.BannerRequest) (*proto.BannerResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (gs *goodsServer) DeleteBanner(ctx context.Context, request *proto.BannerRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (gs *goodsServer) UpdateBanner(ctx context.Context, request *proto.BannerRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (gs *goodsServer) CategoryBrandList(ctx context.Context, request *proto.CategoryBrandFilterRequest) (*proto.CategoryBrandListResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (gs *goodsServer) GetCategoryBrandList(ctx context.Context, request *proto.CategoryInfoRequest) (*proto.BrandListResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (gs *goodsServer) CreateCategoryBrand(ctx context.Context, request *proto.CategoryBrandRequest) (*proto.CategoryBrandResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (gs *goodsServer) DeleteCategoryBrand(ctx context.Context, request *proto.CategoryBrandRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (gs *goodsServer) UpdateCategoryBrand(ctx context.Context, request *proto.CategoryBrandRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
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