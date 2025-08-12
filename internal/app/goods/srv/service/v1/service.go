package v1

import (
	"context"
	proto "emshop/api/goods/v1"
	dataV1 "emshop/internal/app/goods/srv/data/v1"
	"emshop/internal/app/goods/srv/domain/dto"
	metav1 "emshop/pkg/common/meta/v1"
)

type GoodsSrv interface {
	// 商品列表
	List(ctx context.Context, opts metav1.ListMeta, req *proto.GoodsFilterRequest, orderby []string) (*dto.GoodsDTOList, error)

	// 商品详情
	Get(ctx context.Context, ID uint64) (*dto.GoodsDTO, error)

	// 创建商品
	Create(ctx context.Context, goods *dto.GoodsDTO) error

	// 更新商品
	Update(ctx context.Context, goods *dto.GoodsDTO) error

	// 删除商品
	Delete(ctx context.Context, ID uint64) error

	//批量查询商品
	BatchGet(ctx context.Context, ids []uint64) ([]*dto.GoodsDTO, error)
}

type ServiceFactory interface {
	Goods() GoodsSrv
}

type service struct {
	factoryManager *dataV1.FactoryManager
}

func NewService(factoryManager *dataV1.FactoryManager) *service {
	return &service{factoryManager: factoryManager}
}

var _ ServiceFactory = &service{}

func (s *service) Goods() GoodsSrv {
	return newGoods(s)
}