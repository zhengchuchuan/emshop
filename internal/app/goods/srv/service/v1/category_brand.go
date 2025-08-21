package v1

import (
	"context"

	"emshop/internal/app/goods/srv/domain/do"
	"emshop/internal/app/goods/srv/domain/dto"
	"emshop/internal/app/pkg/code"
	metav1 "emshop/pkg/common/meta/v1"
	"emshop/pkg/errors"
	"emshop/pkg/log"
)

type categoryBrand struct {
	srv *service
}

func newCategoryBrand(srv *service) CategoryBrandSrv {
	return &categoryBrand{srv: srv}
}

var _ CategoryBrandSrv = &categoryBrand{}

func (cb *categoryBrand) List(ctx context.Context, opts metav1.ListMeta, orderby []string) (*dto.CategoryBrandDTOList, error) {
	dataFactory := cb.srv.factoryManager.GetDataFactory()
	categoryBrands, err := dataFactory.CategoryBrands().List(ctx, dataFactory.DB(), orderby, opts)
	if err != nil {
		log.Errorf("get category brands list error: %v", err)
		return nil, err
	}

	ret := &dto.CategoryBrandDTOList{
		TotalCount: categoryBrands.TotalCount,
		Items: make([]*dto.CategoryBrandDTO, 0),
	}

	for _, item := range categoryBrands.Items {
		// 获取关联的分类和品牌信息
		category, err := dataFactory.Categorys().Get(ctx, dataFactory.DB(), uint64(item.CategoryID))
		if err != nil {
			log.Warnf("get category error: %v", err)
			continue
		}

		brand, err := dataFactory.Brands().Get(ctx, dataFactory.DB(), uint64(item.BrandsID))
		if err != nil {
			log.Warnf("get brand error: %v", err)
			continue
		}

		categoryBrandDTO := &dto.CategoryBrandDTO{
			GoodsCategoryBrandDO: *item,
		}
		
		// 设置关联信息
		categoryBrandDTO.Category = *category
		categoryBrandDTO.Brands = *brand
		
		ret.Items = append(ret.Items, categoryBrandDTO)
	}

	return ret, nil
}

func (cb *categoryBrand) GetBrandsByCategory(ctx context.Context, categoryID int32) (*dto.BrandDTOList, error) {
	dataFactory := cb.srv.factoryManager.GetDataFactory()
	
	// 先检查分类是否存在
	_, err := dataFactory.Categorys().Get(ctx, dataFactory.DB(), uint64(categoryID))
	if err != nil {
		log.Errorf("category not found: %v", err)
		return nil, errors.WithCode(code.ErrCategoryNotFound, "分类不存在")
	}

	// 获取该分类下的所有品牌关系
	categoryBrands, err := dataFactory.CategoryBrands().GetByCategory(ctx, dataFactory.DB(), uint64(categoryID))
	if err != nil {
		log.Errorf("get brands by category error: %v", err)
		return nil, err
	}

	ret := &dto.BrandDTOList{
		TotalCount: int64(len(categoryBrands.Items)),
		Items: make([]*dto.BrandDTO, 0),
	}

	for _, item := range categoryBrands.Items {
		brand, err := dataFactory.Brands().Get(ctx, dataFactory.DB(), uint64(item.BrandsID))
		if err != nil {
			log.Warnf("get brand error: %v", err)
			continue
		}

		brandDTO := &dto.BrandDTO{
			BrandsDO: *brand,
		}
		ret.Items = append(ret.Items, brandDTO)
	}

	return ret, nil
}

func (cb *categoryBrand) Get(ctx context.Context, ID int32) (*dto.CategoryBrandDTO, error) {
	dataFactory := cb.srv.factoryManager.GetDataFactory()
	categoryBrand, err := dataFactory.CategoryBrands().Get(ctx, dataFactory.DB(), uint64(ID))
	if err != nil {
		log.Errorf("get category brand error: %v", err)
		return nil, err
	}

	// 获取关联的分类和品牌信息
	category, err := dataFactory.Categorys().Get(ctx, dataFactory.DB(), uint64(categoryBrand.CategoryID))
	if err != nil {
		log.Errorf("get category error: %v", err)
		return nil, err
	}

	brand, err := dataFactory.Brands().Get(ctx, dataFactory.DB(), uint64(categoryBrand.BrandsID))
	if err != nil {
		log.Errorf("get brand error: %v", err)
		return nil, err
	}

	categoryBrandDTO := &dto.CategoryBrandDTO{
		GoodsCategoryBrandDO: *categoryBrand,
	}
	categoryBrandDTO.Category = *category
	categoryBrandDTO.Brands = *brand

	return categoryBrandDTO, nil
}

func (cb *categoryBrand) Create(ctx context.Context, categoryBrand *dto.CategoryBrandDTO) error {
	dataFactory := cb.srv.factoryManager.GetDataFactory()

	// 验证分类是否存在
	_, err := dataFactory.Categorys().Get(ctx, dataFactory.DB(), uint64(categoryBrand.CategoryID))
	if err != nil {
		log.Errorf("category not found: %v", err)
		return errors.WithCode(code.ErrCategoryNotFound, "分类不存在")
	}

	// 验证品牌是否存在
	_, err = dataFactory.Brands().Get(ctx, dataFactory.DB(), uint64(categoryBrand.BrandsID))
	if err != nil {
		log.Errorf("brand not found: %v", err)
		return errors.WithCode(code.ErrBrandNotFound, "品牌不存在")
	}

	// 检查关系是否已存在
	existingRelations, err := dataFactory.CategoryBrands().GetByCategory(ctx, dataFactory.DB(), uint64(categoryBrand.CategoryID))
	if err == nil {
		for _, existing := range existingRelations.Items {
			if existing.BrandsID == categoryBrand.BrandsID {
				log.Errorf("category brand relation already exists: category=%d, brand=%d", categoryBrand.CategoryID, categoryBrand.BrandsID)
				return errors.WithCode(code.ErrCategoryBrandNotFound, "分类品牌关系已存在")
			}
		}
	}

	categoryBrandDO := &do.GoodsCategoryBrandDO{
		CategoryID: categoryBrand.CategoryID,
		BrandsID:   categoryBrand.BrandsID,
	}

	err = dataFactory.CategoryBrands().Create(ctx, dataFactory.DB(), categoryBrandDO)
	if err != nil {
		log.Errorf("create category brand error: %v", err)
		return err
	}

	categoryBrand.ID = categoryBrandDO.ID
	return nil
}

func (cb *categoryBrand) Update(ctx context.Context, categoryBrand *dto.CategoryBrandDTO) error {
	dataFactory := cb.srv.factoryManager.GetDataFactory()

	// 检查关系是否存在
	existing, err := dataFactory.CategoryBrands().Get(ctx, dataFactory.DB(), uint64(categoryBrand.ID))
	if err != nil {
		log.Errorf("category brand not found: %v", err)
		return errors.WithCode(code.ErrCategoryBrandNotFound, "分类品牌关系不存在")
	}

	// 验证分类是否存在
	_, err = dataFactory.Categorys().Get(ctx, dataFactory.DB(), uint64(categoryBrand.CategoryID))
	if err != nil {
		log.Errorf("category not found: %v", err)
		return errors.WithCode(code.ErrCategoryNotFound, "分类不存在")
	}

	// 验证品牌是否存在
	_, err = dataFactory.Brands().Get(ctx, dataFactory.DB(), uint64(categoryBrand.BrandsID))
	if err != nil {
		log.Errorf("brand not found: %v", err)
		return errors.WithCode(code.ErrBrandNotFound, "品牌不存在")
	}

	// 检查新的关系是否与其他记录冲突（如果改变了分类或品牌）
	if existing.CategoryID != categoryBrand.CategoryID || existing.BrandsID != categoryBrand.BrandsID {
		existingRelations, err := dataFactory.CategoryBrands().GetByCategory(ctx, dataFactory.DB(), uint64(categoryBrand.CategoryID))
		if err == nil {
			for _, existingRelation := range existingRelations.Items {
				if existingRelation.BrandsID == categoryBrand.BrandsID && existingRelation.ID != categoryBrand.ID {
					log.Errorf("category brand relation already exists: category=%d, brand=%d", categoryBrand.CategoryID, categoryBrand.BrandsID)
					return errors.WithCode(code.ErrCategoryBrandNotFound, "分类品牌关系已存在")
				}
			}
		}
	}

	// 更新字段
	existing.CategoryID = categoryBrand.CategoryID
	existing.BrandsID = categoryBrand.BrandsID

	err = dataFactory.CategoryBrands().Update(ctx, dataFactory.DB(), existing)
	if err != nil {
		log.Errorf("update category brand error: %v", err)
		return err
	}

	return nil
}

func (cb *categoryBrand) Delete(ctx context.Context, ID int32) error {
	dataFactory := cb.srv.factoryManager.GetDataFactory()

	// 检查关系是否存在
	_, err := dataFactory.CategoryBrands().Get(ctx, dataFactory.DB(), uint64(ID))
	if err != nil {
		log.Errorf("category brand not found: %v", err)
		return errors.WithCode(code.ErrCategoryBrandNotFound, "分类品牌关系不存在")
	}

	err = dataFactory.CategoryBrands().Delete(ctx, dataFactory.DB(), uint64(ID))
	if err != nil {
		log.Errorf("delete category brand error: %v", err)
		return err
	}

	return nil
}