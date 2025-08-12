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

type brand struct {
	srv *service
}

func newBrand(srv *service) BrandSrv {
	return &brand{srv: srv}
}

var _ BrandSrv = &brand{}

func (b *brand) List(ctx context.Context, opts metav1.ListMeta, orderby []string) (*dto.BrandDTOList, error) {
	dataFactory := b.srv.factoryManager.GetDataFactory()
	brands, err := dataFactory.Brands().List(ctx, orderby, opts)
	if err != nil {
		log.Errorf("get brands list error: %v", err)
		return nil, err
	}

	ret := &dto.BrandDTOList{
		TotalCount: brands.TotalCount,
		Items: make([]*dto.BrandDTO, 0),
	}

	for _, item := range brands.Items {
		brandDTO := &dto.BrandDTO{
			BrandsDO: *item,
		}
		ret.Items = append(ret.Items, brandDTO)
	}

	return ret, nil
}

func (b *brand) Get(ctx context.Context, ID int32) (*dto.BrandDTO, error) {
	dataFactory := b.srv.factoryManager.GetDataFactory()
	brand, err := dataFactory.Brands().Get(ctx, uint64(ID))
	if err != nil {
		log.Errorf("get brand error: %v", err)
		return nil, err
	}

	return &dto.BrandDTO{
		BrandsDO: *brand,
	}, nil
}

func (b *brand) Create(ctx context.Context, brand *dto.BrandDTO) error {
	dataFactory := b.srv.factoryManager.GetDataFactory()

	// 检查品牌名称是否已存在
	existingBrands, err := dataFactory.Brands().List(ctx, []string{}, metav1.ListMeta{Page: 1, PageSize: 1})
	if err == nil {
		for _, existing := range existingBrands.Items {
			if existing.Name == brand.Name {
				log.Errorf("brand name already exists: %s", brand.Name)
				return errors.WithCode(code.ErrBrandNotFound, "品牌名称已存在")
			}
		}
	}

	brandDO := &do.BrandsDO{
		Name: brand.Name,
		Logo: brand.Logo,
	}

	err = dataFactory.Brands().Create(ctx, brandDO)
	if err != nil {
		log.Errorf("create brand error: %v", err)
		return err
	}

	brand.ID = brandDO.ID
	return nil
}

func (b *brand) Update(ctx context.Context, brand *dto.BrandDTO) error {
	dataFactory := b.srv.factoryManager.GetDataFactory()

	// 检查品牌是否存在
	existing, err := dataFactory.Brands().Get(ctx, uint64(brand.ID))
	if err != nil {
		log.Errorf("brand not found: %v", err)
		return errors.WithCode(code.ErrBrandNotFound, "品牌不存在")
	}

	// 检查品牌名称是否已被其他品牌使用
	if brand.Name != existing.Name {
		existingBrands, err := dataFactory.Brands().List(ctx, []string{}, metav1.ListMeta{Page: 1, PageSize: 100})
		if err == nil {
			for _, existingBrand := range existingBrands.Items {
				if existingBrand.Name == brand.Name && existingBrand.ID != brand.ID {
					log.Errorf("brand name already exists: %s", brand.Name)
					return errors.WithCode(code.ErrBrandNotFound, "品牌名称已存在")
				}
			}
		}
	}

	// 更新字段
	existing.Name = brand.Name
	existing.Logo = brand.Logo

	err = dataFactory.Brands().Update(ctx, existing)
	if err != nil {
		log.Errorf("update brand error: %v", err)
		return err
	}

	return nil
}

func (b *brand) Delete(ctx context.Context, ID int32) error {
	dataFactory := b.srv.factoryManager.GetDataFactory()

	// 检查品牌是否存在
	_, err := dataFactory.Brands().Get(ctx, uint64(ID))
	if err != nil {
		log.Errorf("brand not found: %v", err)
		return errors.WithCode(code.ErrBrandNotFound, "品牌不存在")
	}

	// TODO: 检查是否有商品在使用该品牌
	// TODO: 检查是否有分类品牌关系在使用该品牌

	err = dataFactory.Brands().Delete(ctx, uint64(ID))
	if err != nil {
		log.Errorf("delete brand error: %v", err)
		return err
	}

	return nil
}