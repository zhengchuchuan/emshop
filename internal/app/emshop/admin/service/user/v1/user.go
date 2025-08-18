package user

import (
	"context"
	upbv1 "emshop/api/user/v1"
	"emshop/internal/app/emshop/admin/data"
	"emshop/pkg/log"
)

// UserSrv 管理员用户服务接口
type UserSrv interface {
	GetUserList(ctx context.Context, page, pageSize uint32) (*upbv1.UserListResponse, error)
	GetUserById(ctx context.Context, id uint64) (*upbv1.UserInfoResponse, error)
	GetUserByMobile(ctx context.Context, mobile string) (*upbv1.UserInfoResponse, error)
	UpdateUserStatus(ctx context.Context, id uint64, status int32) error
}

type userService struct {
	data data.DataFactory
}

func NewUserService(data data.DataFactory) UserSrv {
	return &userService{data: data}
}

func (u *userService) GetUserList(ctx context.Context, page, pageSize uint32) (*upbv1.UserListResponse, error) {
	log.Infof("Admin GetUserList called with page: %d, pageSize: %d", page, pageSize)
	
	request := &upbv1.PageInfo{
		Pn:    page,
		PSize: pageSize,
	}
	
	return u.data.Users().GetUserList(ctx, request)
}

func (u *userService) GetUserById(ctx context.Context, id uint64) (*upbv1.UserInfoResponse, error) {
	log.Infof("Admin GetUserById called with id: %d", id)
	
	request := &upbv1.IdRequest{
		Id: int32(id),
	}
	
	return u.data.Users().GetUserById(ctx, request)
}

func (u *userService) GetUserByMobile(ctx context.Context, mobile string) (*upbv1.UserInfoResponse, error) {
	log.Infof("Admin GetUserByMobile called with mobile: %s", mobile)
	
	request := &upbv1.MobileRequest{
		Mobile: mobile,
	}
	
	return u.data.Users().GetUserByMobile(ctx, request)
}

func (u *userService) UpdateUserStatus(ctx context.Context, id uint64, status int32) error {
	log.Infof("Admin UpdateUserStatus called with id: %d, status: %d", id, status)
	
	// 这里可以添加更多管理员特有的业务逻辑，比如权限检查、审计日志等
	request := &upbv1.UpdateUserInfo{
		Id: int32(id),
		// 根据实际需要设置状态字段
	}
	
	_, err := u.data.Users().UpdateUser(ctx, request)
	return err
}