package interfaces

import (
	"context"
	"emshop/internal/app/goods/srv/domain/do"
	metav1 "emshop/pkg/common/meta/v1"
	"gorm.io/gorm"
)

// BannerStore 轮播图存储接口
type BannerStore interface {
	Get(ctx context.Context, db *gorm.DB, ID uint64) (*do.BannerDO, error)
	List(ctx context.Context, db *gorm.DB, orderby []string, opts metav1.ListMeta) (*do.BannerDOList, error)
	Create(ctx context.Context, db *gorm.DB, banner *do.BannerDO) error
	Update(ctx context.Context, db *gorm.DB, banner *do.BannerDO) error
	Delete(ctx context.Context, db *gorm.DB, ID uint64) error
}