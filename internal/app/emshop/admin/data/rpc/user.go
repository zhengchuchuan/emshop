package rpc

import (
	"context"
	upbv1 "emshop/api/user/v1"
	"emshop/internal/app/emshop/admin/data"
)


type users struct {
	uc upbv1.UserClient		// 直接使用 UserClient 接口
}

func NewUsers(uc upbv1.UserClient) *users {
	return &users{uc}
}


func (u *users) CheckPassWord(ctx context.Context, request *upbv1.PasswordCheckInfo) (*upbv1.CheckResponse, error) {
	return u.uc.CheckPassWord(ctx, request)
}

func (u *users) CreateUser(ctx context.Context, request *upbv1.CreateUserInfo) (*upbv1.UserInfoResponse, error) {
	return u.uc.CreateUser(ctx, request)
}

func (u *users) UpdateUser(ctx context.Context, request *upbv1.UpdateUserInfo) (*upbv1.UserInfoResponse, error) {
	_, err := u.uc.UpdateUser(ctx, request)
	if err != nil {
		return nil, err
	}
	// UpdateUser 返回 Empty，所以我们需要重新获取用户信息
	return u.uc.GetUserById(ctx, &upbv1.IdRequest{Id: request.Id})
}

func (u *users) GetUserById(ctx context.Context, request *upbv1.IdRequest) (*upbv1.UserInfoResponse, error) {
	return u.uc.GetUserById(ctx, request)
}

func (u *users) GetUserByMobile(ctx context.Context, request *upbv1.MobileRequest) (*upbv1.UserInfoResponse, error) {
	return u.uc.GetUserByMobile(ctx, request)
}

func (u *users) GetUserList(ctx context.Context, request *upbv1.PageInfo) (*upbv1.UserListResponse, error) {
	return u.uc.GetUserList(ctx, request)
}

var _ data.UserData = &users{}