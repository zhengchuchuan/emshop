package rpc

import (
	"context"
	uoppbv1 "emshop/api/userop/v1"
	"emshop/internal/app/api/emshop/data"
	"emshop/pkg/log"
)

type userop struct {
	uoc uoppbv1.UserOpClient
}

func NewUserOp(uoc uoppbv1.UserOpClient) *userop {
	return &userop{uoc}
}

// ==================== 用户收藏管理 ====================

func (uo *userop) UserFavList(ctx context.Context, request *uoppbv1.UserFavListRequest) (*uoppbv1.UserFavListResponse, error) {
	log.Infof("Calling UserFavList gRPC for user: %d", request.UserId)
	response, err := uo.uoc.UserFavList(ctx, request)
	if err != nil {
		log.Errorf("UserFavList gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("UserFavList gRPC call successful, total: %d", response.Total)
	return response, nil
}

func (uo *userop) CreateUserFav(ctx context.Context, request *uoppbv1.UserFavRequest) (*uoppbv1.UserFavResponse, error) {
	log.Infof("Calling CreateUserFav gRPC for user: %d, goods: %d", request.UserId, request.GoodsId)
	response, err := uo.uoc.CreateUserFav(ctx, request)
	if err != nil {
		log.Errorf("CreateUserFav gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("CreateUserFav gRPC call successful")
	return response, nil
}

func (uo *userop) DeleteUserFav(ctx context.Context, request *uoppbv1.UserFavRequest) (*uoppbv1.UserFavResponse, error) {
	log.Infof("Calling DeleteUserFav gRPC for user: %d, goods: %d", request.UserId, request.GoodsId)
	_, err := uo.uoc.DeleteUserFav(ctx, request)
	if err != nil {
		log.Errorf("DeleteUserFav gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("DeleteUserFav gRPC call successful")
	return &uoppbv1.UserFavResponse{}, nil
}

func (uo *userop) GetUserFavDetail(ctx context.Context, request *uoppbv1.UserFavRequest) (*uoppbv1.UserFavResponse, error) {
	log.Infof("Calling GetUserFavDetail gRPC for user: %d, goods: %d", request.UserId, request.GoodsId)
	response, err := uo.uoc.GetUserFavDetail(ctx, request)
	if err != nil {
		log.Errorf("GetUserFavDetail gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("GetUserFavDetail gRPC call successful")
	return response, nil
}

// ==================== 用户地址管理 ====================

func (uo *userop) GetAddressList(ctx context.Context, request *uoppbv1.AddressRequest) (*uoppbv1.AddressListResponse, error) {
	log.Infof("Calling GetAddressList gRPC for user: %d", request.UserId)
	response, err := uo.uoc.GetAddressList(ctx, request)
	if err != nil {
		log.Errorf("GetAddressList gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("GetAddressList gRPC call successful, total: %d", response.Total)
	return response, nil
}

func (uo *userop) CreateAddress(ctx context.Context, request *uoppbv1.AddressRequest) (*uoppbv1.AddressResponse, error) {
	log.Infof("Calling CreateAddress gRPC for user: %d", request.UserId)
	response, err := uo.uoc.CreateAddress(ctx, request)
	if err != nil {
		log.Errorf("CreateAddress gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("CreateAddress gRPC call successful, address ID: %d", response.Id)
	return response, nil
}

func (uo *userop) UpdateAddress(ctx context.Context, request *uoppbv1.AddressRequest) (*uoppbv1.AddressResponse, error) {
	log.Infof("Calling UpdateAddress gRPC for address: %d", request.Id)
	_, err := uo.uoc.UpdateAddress(ctx, request)
	if err != nil {
		log.Errorf("UpdateAddress gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("UpdateAddress gRPC call successful")
	return &uoppbv1.AddressResponse{}, nil
}

func (uo *userop) DeleteAddress(ctx context.Context, request *uoppbv1.DeleteAddressRequest) (*uoppbv1.AddressResponse, error) {
	log.Infof("Calling DeleteAddress gRPC for address: %d", request.Id)
	_, err := uo.uoc.DeleteAddress(ctx, request)
	if err != nil {
		log.Errorf("DeleteAddress gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("DeleteAddress gRPC call successful")
	return &uoppbv1.AddressResponse{}, nil
}

// ==================== 用户留言管理 ====================

func (uo *userop) MessageList(ctx context.Context, request *uoppbv1.MessageRequest) (*uoppbv1.MessageListResponse, error) {
	log.Infof("Calling MessageList gRPC for user: %d", request.UserId)
	response, err := uo.uoc.MessageList(ctx, request)
	if err != nil {
		log.Errorf("MessageList gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("MessageList gRPC call successful, total: %d", response.Total)
	return response, nil
}

func (uo *userop) CreateMessage(ctx context.Context, request *uoppbv1.MessageRequest) (*uoppbv1.MessageResponse, error) {
	log.Infof("Calling CreateMessage gRPC for user: %d, type: %d", request.UserId, request.MessageType)
	response, err := uo.uoc.CreateMessage(ctx, request)
	if err != nil {
		log.Errorf("CreateMessage gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("CreateMessage gRPC call successful, message ID: %d", response.Id)
	return response, nil
}

var _ data.UserOpData = &userop{}
