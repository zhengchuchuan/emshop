package service

import (
	"emshop/internal/app/order/srv/data/v1/mysql"
	"emshop/internal/app/pkg/options"
)

type ServiceFactory interface {
	Orders() OrderSrv
}

type service struct {
	data    mysql.DataFactory
	dtmopts *options.DtmOptions
}

func (s *service) Orders() OrderSrv {
	return newOrderService(s)
}

var _ ServiceFactory = &service{}

func NewService(data mysql.DataFactory, dtmopts *options.DtmOptions) *service {
	return &service{data: data, dtmopts: dtmopts}
}
