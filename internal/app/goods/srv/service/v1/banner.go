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

type banner struct {
	srv *service
}

func newBanner(srv *service) BannerSrv {
	return &banner{srv: srv}
}

var _ BannerSrv = &banner{}

func (b *banner) List(ctx context.Context, orderby []string) (*dto.BannerDTOList, error) {
	dataFactory := b.srv.factoryManager.GetDataFactory()
	// 获取所有轮播图，不需要分页
	banners, err := dataFactory.Banners().List(ctx, orderby, metav1.ListMeta{Page: 1, PageSize: 100})
	if err != nil {
		log.Errorf("get banners list error: %v", err)
		return nil, err
	}

	ret := &dto.BannerDTOList{
		TotalCount: banners.TotalCount,
		Items: make([]*dto.BannerDTO, 0),
	}

	for _, item := range banners.Items {
		bannerDTO := &dto.BannerDTO{
			BannerDO: *item,
		}
		ret.Items = append(ret.Items, bannerDTO)
	}

	return ret, nil
}

func (b *banner) Get(ctx context.Context, ID int32) (*dto.BannerDTO, error) {
	dataFactory := b.srv.factoryManager.GetDataFactory()
	banner, err := dataFactory.Banners().Get(ctx, uint64(ID))
	if err != nil {
		log.Errorf("get banner error: %v", err)
		return nil, err
	}

	return &dto.BannerDTO{
		BannerDO: *banner,
	}, nil
}

func (b *banner) Create(ctx context.Context, banner *dto.BannerDTO) error {
	dataFactory := b.srv.factoryManager.GetDataFactory()

	bannerDO := &do.BannerDO{
		Image: banner.Image,
		Url:   banner.Url,
		Index: banner.Index,
	}

	err := dataFactory.Banners().Create(ctx, bannerDO)
	if err != nil {
		log.Errorf("create banner error: %v", err)
		return err
	}

	banner.ID = bannerDO.ID
	return nil
}

func (b *banner) Update(ctx context.Context, banner *dto.BannerDTO) error {
	dataFactory := b.srv.factoryManager.GetDataFactory()

	// 检查轮播图是否存在
	existing, err := dataFactory.Banners().Get(ctx, uint64(banner.ID))
	if err != nil {
		log.Errorf("banner not found: %v", err)
		return errors.WithCode(code.ErrBannerNotFound, "轮播图不存在")
	}

	// 更新字段
	existing.Image = banner.Image
	existing.Url = banner.Url
	existing.Index = banner.Index

	err = dataFactory.Banners().Update(ctx, existing)
	if err != nil {
		log.Errorf("update banner error: %v", err)
		return err
	}

	return nil
}

func (b *banner) Delete(ctx context.Context, ID int32) error {
	dataFactory := b.srv.factoryManager.GetDataFactory()

	// 检查轮播图是否存在
	_, err := dataFactory.Banners().Get(ctx, uint64(ID))
	if err != nil {
		log.Errorf("banner not found: %v", err)
		return errors.WithCode(code.ErrBannerNotFound, "轮播图不存在")
	}

	err = dataFactory.Banners().Delete(ctx, uint64(ID))
	if err != nil {
		log.Errorf("delete banner error: %v", err)
		return err
	}

	return nil
}