package user

import (
	srv1 "emshop-admin/internal/app/user/srv/service/v1"
)

type userServer struct {
	srv srv1.UserSrv
}

func NewUserServer(srv srv1.UserSrv) *userServer {
	return &userServer{
		srv: srv,
	}
}