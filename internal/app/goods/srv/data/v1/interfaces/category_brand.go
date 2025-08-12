package interfaces

import (
	"context"
	"emshop/internal/app/goods/srv/domain/do"
	metav1 "emshop/pkg/common/meta/v1"
	"gorm.io/gorm"
)

// GoodsCategoryBrandStore 分类品牌关系存储接口
type GoodsCategoryBrandStore interface {
	Get(ctx context.Context, ID uint64) (*do.GoodsCategoryBrandDO, error)
	List(ctx context.Context, orderby []string, opts metav1.ListMeta) (*do.GoodsCategoryBrandDOList, error)
	Create(ctx context.Context, categoryBrand *do.GoodsCategoryBrandDO) error
	CreateInTxn(ctx context.Context, txn *gorm.DB, categoryBrand *do.GoodsCategoryBrandDO) error
	Update(ctx context.Context, categoryBrand *do.GoodsCategoryBrandDO) error
	UpdateInTxn(ctx context.Context, txn *gorm.DB, categoryBrand *do.GoodsCategoryBrandDO) error
	Delete(ctx context.Context, ID uint64) error
	DeleteInTxn(ctx context.Context, txn *gorm.DB, ID uint64) error
	
	// 分类品牌关系特有方法
	GetByCategory(ctx context.Context, categoryID uint64) (*do.GoodsCategoryBrandDOList, error)
	GetByBrand(ctx context.Context, brandID uint64) (*do.GoodsCategoryBrandDOList, error)
}