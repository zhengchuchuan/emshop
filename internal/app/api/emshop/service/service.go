package service

import (
	"emshop/internal/app/api/emshop/data"
	gv1 "emshop/internal/app/api/emshop/service/goods/v1"
	iv1 "emshop/internal/app/api/emshop/service/inventory/v1"
	ov1 "emshop/internal/app/api/emshop/service/order/v1"
	sv1 "emshop/internal/app/api/emshop/service/sms/v1"
	uv1 "emshop/internal/app/api/emshop/service/user/v1"
	uopv1 "emshop/internal/app/api/emshop/service/userop/v1"
	"emshop/internal/app/pkg/options"
)

type ServiceFactory interface {
	Goods() gv1.GoodsSrv
	Users() uv1.UserSrv
	Sms() sv1.SmsSrv
	Inventory() iv1.InventorySrv
	Order() ov1.OrderSrv
	UserOp() uopv1.UserOpSrv
}

type service struct {
	data data.DataFactory

	smsOpts *options.SmsOptions

	jwtOpts *options.JwtOptions
}

func (s *service) Sms() sv1.SmsSrv {
	return sv1.NewSmsService(s.smsOpts)
}

func (s *service) Goods() gv1.GoodsSrv {
	return gv1.NewGoods(s.data)
}

func (s *service) Users() uv1.UserSrv {
	return uv1.NewUserService(s.data, s.jwtOpts)
}

func (s *service) Inventory() iv1.InventorySrv {
	return iv1.NewInventory(s.data)
}

func (s *service) Order() ov1.OrderSrv {
	return ov1.NewOrder(s.data)
}

func (s *service) UserOp() uopv1.UserOpSrv {
	return uopv1.NewUserOp(s.data)
}

func NewService(store data.DataFactory, smsOpts *options.SmsOptions, jwtOpts *options.JwtOptions) *service {
	return &service{data: store,
		smsOpts: smsOpts,
		jwtOpts: jwtOpts,
	}
}

var _ ServiceFactory = &service{}
