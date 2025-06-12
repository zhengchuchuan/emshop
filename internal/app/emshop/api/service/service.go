package service

import (
	"emshop/internal/app/emshop/api/data"
	gv1 "emshop/internal/app/emshop/api/service/goods/v1"
	sv1 "emshop/internal/app/emshop/api/service/sms/v1"
	uv1 "emshop/internal/app/emshop/api/service/user/v1"
	"emshop/internal/app/pkg/options"
)

type ServiceFactory interface {
	Goods() gv1.GoodsSrv
	Users() uv1.UserSrv
	Sms() sv1.SmsSrv
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

func NewService(store data.DataFactory, smsOpts *options.SmsOptions, jwtOpts *options.JwtOptions) *service {
	return &service{data: store,
		smsOpts: smsOpts,
		jwtOpts: jwtOpts,
	}
}

var _ ServiceFactory = &service{}
