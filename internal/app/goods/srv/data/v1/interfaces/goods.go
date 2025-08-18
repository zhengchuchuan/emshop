package interfaces

import (
	"context"
	"emshop/internal/app/goods/srv/domain/do"
	metav1 "emshop/pkg/common/meta/v1"
	"gorm.io/gorm"
)

// GoodsStore 商品主存储接口
type GoodsStore interface {
	Get(ctx context.Context, ID uint64) (*do.GoodsDO, error)
	ListByIDs(ctx context.Context, ids []uint64, orderby []string) (*do.GoodsDOList, error)
	List(ctx context.Context, orderby []string, opts metav1.ListMeta) (*do.GoodsDOList, error)
	GetAllGoodsIDs(ctx context.Context) ([]uint64, error)
	Create(ctx context.Context, goods *do.GoodsDO) error
	CreateInTxn(ctx context.Context, txn *gorm.DB, goods *do.GoodsDO) error
	Update(ctx context.Context, goods *do.GoodsDO) error
	UpdateInTxn(ctx context.Context, txn *gorm.DB, goods *do.GoodsDO) error
	Delete(ctx context.Context, ID uint64) error
	DeleteInTxn(ctx context.Context, txn *gorm.DB, ID uint64) error
}

// GoodsSearchStore 商品搜索存储接口
type GoodsSearchStore interface {
	Create(ctx context.Context, goods *do.GoodsSearchDO) error
	Delete(ctx context.Context, ID uint64) error
	Update(ctx context.Context, goods *do.GoodsSearchDO) error
	Search(ctx context.Context, request *GoodsFilterRequest) (*do.GoodsSearchDOList, error)
}

// GoodsFilterRequest 商品过滤请求
type GoodsFilterRequest struct {
	KeyWords     string
	CategoryIDs  []interface{}
	BrandID      int32
	PriceMin     float32
	PriceMax     float32
	IsHot        bool
	IsNew        bool
	OnSale       bool
	Pages        int32
	PagePerNums  int32
}