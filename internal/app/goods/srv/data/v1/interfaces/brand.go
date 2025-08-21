package interfaces

import (
	"context"
	"emshop/internal/app/goods/srv/domain/do"
	metav1 "emshop/pkg/common/meta/v1"
	"gorm.io/gorm"
)

// BrandsStore 品牌存储接口
type BrandsStore interface {
	Get(ctx context.Context, db *gorm.DB, ID uint64) (*do.BrandsDO, error)
	List(ctx context.Context, db *gorm.DB, orderby []string, opts metav1.ListMeta) (*do.BrandsDOList, error)
	Create(ctx context.Context, db *gorm.DB, brand *do.BrandsDO) error
	Update(ctx context.Context, db *gorm.DB, brand *do.BrandsDO) error
	Delete(ctx context.Context, db *gorm.DB, ID uint64) error
}