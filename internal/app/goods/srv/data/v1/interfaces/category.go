package interfaces

import (
	"context"
	"emshop/internal/app/goods/srv/domain/do"
	metav1 "emshop/pkg/common/meta/v1"
	"gorm.io/gorm"
)

// CategoryStore 分类存储接口
type CategoryStore interface {
	Get(ctx context.Context, db *gorm.DB, ID uint64) (*do.CategoryDO, error)
	List(ctx context.Context, db *gorm.DB, orderby []string, opts metav1.ListMeta) (*do.CategoryDOList, error)
	Create(ctx context.Context, db *gorm.DB, category *do.CategoryDO) error
	Update(ctx context.Context, db *gorm.DB, category *do.CategoryDO) error
	Delete(ctx context.Context, db *gorm.DB, ID uint64) error
	
	// 分类特有方法
	GetByLevel(ctx context.Context, db *gorm.DB, level int) (*do.CategoryDOList, error)
	GetSubCategories(ctx context.Context, db *gorm.DB, parentID uint64) (*do.CategoryDOList, error)
}