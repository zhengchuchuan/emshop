package v1

import (
	"context"
	proto "emshop/api/goods/v1"

	"emshop/internal/app/goods/srv/domain/do"
)

type GoodsFilterRequest struct {
	*proto.GoodsFilterRequest
	CategoryIDs []interface{}
}

type GoodsStore interface {
	Create(ctx context.Context, goods *do.GoodsSearchDO) error
	Delete(ctx context.Context, ID uint64) error
	Update(ctx context.Context, goods *do.GoodsSearchDO) error
	Search(ctx context.Context, request *GoodsFilterRequest) (*do.GoodsSearchDOList, error)
}
