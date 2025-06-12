package v1

import (
	"context"
	"gorm.io/gorm"
	"emshop/internal/app/goods/srv/domain/do"
	metav1 "emshop/pkg/common/meta/v1"
)

type BannerStore interface {
	List(ctx context.Context, opts metav1.ListMeta, orderby []string) (*do.BannerList, error)
	Create(ctx context.Context, txn *gorm.DB, banner *do.BannerDO) error
	Update(ctx context.Context, txn *gorm.DB, banner *do.BannerDO) error
	Delete(ctx context.Context, ID uint64) error
}
