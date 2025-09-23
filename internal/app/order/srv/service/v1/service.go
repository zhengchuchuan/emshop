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
	regopts *options.RegistryOptions
}

func (s *service) Orders() OrderSrv { return newOrderService(s) }

var _ ServiceFactory = &service{}

func NewService(data mysql.DataFactory, dtmopts *options.DtmOptions, regopts *options.RegistryOptions) *service {
    return &service{data: data, dtmopts: dtmopts, regopts: regopts}
}
