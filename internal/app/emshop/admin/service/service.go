package service

import (
	"emshop/internal/app/emshop/admin/data"
	"emshop/internal/app/emshop/admin/service/user/v1"
	"emshop/internal/app/emshop/admin/service/goods/v1"
	"emshop/internal/app/emshop/admin/service/order/v1"
	"emshop/internal/app/pkg/options"
)

// ServiceFactory 服务工厂接口
type ServiceFactory interface {
	Users() user.UserSrv
	Goods() goods.GoodsSrv
	Order() order.OrderSrv
}

type serviceFactory struct {
	data data.DataFactory
	jwt  *options.JwtOptions
}

func NewService(data data.DataFactory, jwt *options.JwtOptions) ServiceFactory {
	return &serviceFactory{
		data: data,
		jwt:  jwt,
	}
}

func (s *serviceFactory) Users() user.UserSrv {
	return user.NewUserService(s.data, s.jwt)
}

func (s *serviceFactory) Goods() goods.GoodsSrv {
	return goods.NewGoodsService(s.data)
}

func (s *serviceFactory) Order() order.OrderSrv {
	return order.NewOrderService(s.data)
}