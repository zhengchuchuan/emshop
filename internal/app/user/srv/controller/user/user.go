package user

import (
	upbv1 "emshop/api/user/v1"
	srv1 "emshop/internal/app/user/srv/service/v1"
)

type userServer struct {
	srv srv1.UserSrv
	upbv1.UnimplementedUserServer // 添加这一行嵌入结构体
}

func NewUserServer(srv srv1.UserSrv) *userServer {
	return &userServer{
		srv: srv,
	}
}