package interfaces

import (
	"context"
	"emshop/internal/app/order/srv/domain/do"

	"gorm.io/gorm"

	metav1 "emshop/pkg/common/meta/v1"
)

// ShopCartStore 购物车存储接口
type ShopCartStore interface {
	List(ctx context.Context, db *gorm.DB, userID uint64, checked bool, meta metav1.ListMeta, orderby []string) (*do.ShoppingCartDOList, error)
	Create(ctx context.Context, db *gorm.DB, cartItem *do.ShoppingCartDO) error
	Get(ctx context.Context, db *gorm.DB, userID, goodsID uint64) (*do.ShoppingCartDO, error)
	UpdateNum(ctx context.Context, db *gorm.DB, cartItem *do.ShoppingCartDO) error
	Delete(ctx context.Context, db *gorm.DB, ID uint64) error
	ClearCheck(ctx context.Context, db *gorm.DB, userID uint64) error

	DeleteByGoodsIDs(ctx context.Context, db *gorm.DB, userID uint64, goodsIDs []int32) error
}