package interfaces

import (
	"context"
	"emshop/internal/app/goods/srv/domain/do"
	metav1 "emshop/pkg/common/meta/v1"
	"gorm.io/gorm"
)

// BrandsStore 品牌存储接口
type BrandsStore interface {
	Get(ctx context.Context, ID uint64) (*do.BrandsDO, error)
	List(ctx context.Context, orderby []string, opts metav1.ListMeta) (*do.BrandsDOList, error)
	Create(ctx context.Context, brand *do.BrandsDO) error
	CreateInTxn(ctx context.Context, txn *gorm.DB, brand *do.BrandsDO) error
	Update(ctx context.Context, brand *do.BrandsDO) error
	UpdateInTxn(ctx context.Context, txn *gorm.DB, brand *do.BrandsDO) error
	Delete(ctx context.Context, ID uint64) error
	DeleteInTxn(ctx context.Context, txn *gorm.DB, ID uint64) error
}