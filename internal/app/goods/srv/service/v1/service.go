package v1

import (
	"context"
	proto "emshop/api/goods/v1"
	dataV1 "emshop/internal/app/goods/srv/data/v1"
	"emshop/internal/app/goods/srv/domain/dto"
	metav1 "emshop/pkg/common/meta/v1"
)

type GoodsSrv interface {
	// 商品列表
	List(ctx context.Context, opts metav1.ListMeta, req *proto.GoodsFilterRequest, orderby []string) (*dto.GoodsDTOList, error)

	// 商品详情
	Get(ctx context.Context, ID uint64) (*dto.GoodsDTO, error)

	// 创建商品
	Create(ctx context.Context, goods *dto.GoodsDTO) error

	// 更新商品
	Update(ctx context.Context, goods *dto.GoodsDTO) error

	// 删除商品
	Delete(ctx context.Context, ID uint64) error

	//批量查询商品
	BatchGet(ctx context.Context, ids []uint64) ([]*dto.GoodsDTO, error)
}

type CategorySrv interface {
	// 分类列表
	List(ctx context.Context, opts metav1.ListMeta, orderby []string) (*dto.CategoryDTOList, error)
	
	// 获取所有分类（树形结构）
	ListAll(ctx context.Context, orderby []string) (*dto.CategoryDTOList, error)
	
	// 获取子分类
	GetSubCategories(ctx context.Context, parentID int32) (*dto.CategoryDTOList, error)
	
	// 分类详情
	Get(ctx context.Context, ID int32) (*dto.CategoryDTO, error)
	
	// 创建分类
	Create(ctx context.Context, category *dto.CategoryDTO) error
	
	// 更新分类
	Update(ctx context.Context, category *dto.CategoryDTO) error
	
	// 删除分类
	Delete(ctx context.Context, ID int32) error
}

type BrandSrv interface {
	// 品牌列表
	List(ctx context.Context, opts metav1.ListMeta, orderby []string) (*dto.BrandDTOList, error)
	
	// 品牌详情
	Get(ctx context.Context, ID int32) (*dto.BrandDTO, error)
	
	// 创建品牌
	Create(ctx context.Context, brand *dto.BrandDTO) error
	
	// 更新品牌
	Update(ctx context.Context, brand *dto.BrandDTO) error
	
	// 删除品牌
	Delete(ctx context.Context, ID int32) error
}

type BannerSrv interface {
	// 轮播图列表
	List(ctx context.Context, orderby []string) (*dto.BannerDTOList, error)
	
	// 轮播图详情
	Get(ctx context.Context, ID int32) (*dto.BannerDTO, error)
	
	// 创建轮播图
	Create(ctx context.Context, banner *dto.BannerDTO) error
	
	// 更新轮播图
	Update(ctx context.Context, banner *dto.BannerDTO) error
	
	// 删除轮播图
	Delete(ctx context.Context, ID int32) error
}

type CategoryBrandSrv interface {
	// 分类品牌关系列表
	List(ctx context.Context, opts metav1.ListMeta, orderby []string) (*dto.CategoryBrandDTOList, error)
	
	// 根据分类ID获取品牌列表
	GetBrandsByCategory(ctx context.Context, categoryID int32) (*dto.BrandDTOList, error)
	
	// 分类品牌关系详情
	Get(ctx context.Context, ID int32) (*dto.CategoryBrandDTO, error)
	
	// 创建分类品牌关系
	Create(ctx context.Context, categoryBrand *dto.CategoryBrandDTO) error
	
	// 更新分类品牌关系
	Update(ctx context.Context, categoryBrand *dto.CategoryBrandDTO) error
	
	// 删除分类品牌关系
	Delete(ctx context.Context, ID int32) error
}

type ServiceFactory interface {
	Goods() GoodsSrv
	Category() CategorySrv
	Brand() BrandSrv
	Banner() BannerSrv
	CategoryBrand() CategoryBrandSrv
}

type service struct {
	factoryManager *dataV1.FactoryManager
}

func NewService(factoryManager *dataV1.FactoryManager) *service {
	return &service{factoryManager: factoryManager}
}

var _ ServiceFactory = &service{}

func (s *service) Goods() GoodsSrv {
	return newGoods(s)
}

func (s *service) Category() CategorySrv {
	return newCategory(s)
}

func (s *service) Brand() BrandSrv {
	return newBrand(s)
}

func (s *service) Banner() BannerSrv {
	return newBanner(s)
}

func (s *service) CategoryBrand() CategoryBrandSrv {
	return newCategoryBrand(s)
}