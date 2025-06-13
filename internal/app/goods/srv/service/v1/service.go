package v1

import (
	v1 "emshop/internal/app/goods/srv/data/v1"
	v12 "emshop/internal/app/goods/srv/data_search/v1"
)

type ServiceFactory interface {
	Goods() GoodsSrv
}

type service struct {
	data       v1.DataFactory
	dataSearch v12.SearchFactory
}

func NewService(store v1.DataFactory, dataSearch v12.SearchFactory) *service {
	return &service{data: store, dataSearch: dataSearch}
}

var _ ServiceFactory = &service{}

func (s *service) Goods() GoodsSrv {
	return newGoods(s)
}
