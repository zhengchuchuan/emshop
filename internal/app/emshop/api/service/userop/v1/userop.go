package v1

import (
	"context"
	proto "emshop/api/userop/v1"
	"emshop/internal/app/emshop/api/data"
)

type UserOpSrv interface {
	// 用户收藏管理
	UserFavList(ctx context.Context, request *proto.UserFavListRequest) (*proto.UserFavListResponse, error)
	CreateUserFav(ctx context.Context, request *proto.UserFavRequest) (*proto.UserFavResponse, error)
	DeleteUserFav(ctx context.Context, request *proto.UserFavRequest) error
	GetUserFavDetail(ctx context.Context, request *proto.UserFavRequest) (*proto.UserFavResponse, error)
	
	// 用户地址管理
	GetAddressList(ctx context.Context, request *proto.AddressRequest) (*proto.AddressListResponse, error)
	CreateAddress(ctx context.Context, request *proto.AddressRequest) (*proto.AddressResponse, error)
	UpdateAddress(ctx context.Context, request *proto.AddressRequest) error
	DeleteAddress(ctx context.Context, request *proto.DeleteAddressRequest) error
	
	// 用户留言管理
	MessageList(ctx context.Context, request *proto.MessageRequest) (*proto.MessageListResponse, error)
	CreateMessage(ctx context.Context, request *proto.MessageRequest) (*proto.MessageResponse, error)
}

type userOpService struct {
	data data.DataFactory
}

// ==================== 用户收藏管理 ====================

func (uos *userOpService) UserFavList(ctx context.Context, request *proto.UserFavListRequest) (*proto.UserFavListResponse, error) {
	return uos.data.UserOp().UserFavList(ctx, request)
}

func (uos *userOpService) CreateUserFav(ctx context.Context, request *proto.UserFavRequest) (*proto.UserFavResponse, error) {
	return uos.data.UserOp().CreateUserFav(ctx, request)
}

func (uos *userOpService) DeleteUserFav(ctx context.Context, request *proto.UserFavRequest) error {
	_, err := uos.data.UserOp().DeleteUserFav(ctx, request)
	return err
}

func (uos *userOpService) GetUserFavDetail(ctx context.Context, request *proto.UserFavRequest) (*proto.UserFavResponse, error) {
	return uos.data.UserOp().GetUserFavDetail(ctx, request)
}

// ==================== 用户地址管理 ====================

func (uos *userOpService) GetAddressList(ctx context.Context, request *proto.AddressRequest) (*proto.AddressListResponse, error) {
	return uos.data.UserOp().GetAddressList(ctx, request)
}

func (uos *userOpService) CreateAddress(ctx context.Context, request *proto.AddressRequest) (*proto.AddressResponse, error) {
	return uos.data.UserOp().CreateAddress(ctx, request)
}

func (uos *userOpService) UpdateAddress(ctx context.Context, request *proto.AddressRequest) error {
	_, err := uos.data.UserOp().UpdateAddress(ctx, request)
	return err
}

func (uos *userOpService) DeleteAddress(ctx context.Context, request *proto.DeleteAddressRequest) error {
	_, err := uos.data.UserOp().DeleteAddress(ctx, request)
	return err
}

// ==================== 用户留言管理 ====================

func (uos *userOpService) MessageList(ctx context.Context, request *proto.MessageRequest) (*proto.MessageListResponse, error) {
	return uos.data.UserOp().MessageList(ctx, request)
}

func (uos *userOpService) CreateMessage(ctx context.Context, request *proto.MessageRequest) (*proto.MessageResponse, error) {
	return uos.data.UserOp().CreateMessage(ctx, request)
}

func NewUserOp(data data.DataFactory) *userOpService {
	return &userOpService{data: data}
}

var _ UserOpSrv = &userOpService{}