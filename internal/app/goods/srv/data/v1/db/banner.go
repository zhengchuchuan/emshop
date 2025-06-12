package db

import (
	"context"

	"gorm.io/gorm"
	v1 "emshop/internal/app/goods/srv/data/v1"
	"emshop/internal/app/goods/srv/domain/do"
	metav1 "emshop/pkg/common/meta/v1"
)

type banners struct {
	db *gorm.DB
}

func newBanner(factory *mysqlFactory) *banners {
	banners := &banners{
		db: factory.db,
	}
	return banners
}

func (b *banners) List(ctx context.Context, opts metav1.ListMeta, orderby []string) (*do.BannerList, error) {
	//TODO implement me
	panic("implement me")
}

func (b *banners) Create(ctx context.Context, txn *gorm.DB, banner *do.BannerDO) error {
	//TODO implement me
	panic("implement me")
}

func (b *banners) Update(ctx context.Context, txn *gorm.DB, banner *do.BannerDO) error {
	//TODO implement me
	panic("implement me")
}

func (b *banners) Delete(ctx context.Context, ID uint64) error {
	//TODO implement me
	panic("implement me")
}

var _ v1.BannerStore = &banners{}
