package v1

import (
	"context"
	pb "emshop/api/userop/v1"
	servicev1 "emshop/internal/app/userop/srv/service/v1"
	"emshop/pkg/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

// UserOpController 用户操作统一控制器
type UserOpController struct {
	pb.UnimplementedUserOpServer
	service servicev1.Service
}

// NewUserOpController 创建用户操作控制器
func NewUserOpController(service servicev1.Service) *UserOpController {
	return &UserOpController{
		service: service,
	}
}

// ==================== 用户收藏管理 ====================

// UserFavList 获取用户收藏列表
func (c *UserOpController) UserFavList(ctx context.Context, req *pb.UserFavListRequest) (*pb.UserFavListResponse, error) {
	log.Infof("UserFavList request: user_id=%d", req.UserId)

	favs, total, err := c.service.UserFavService().GetUserFavList(ctx, req.UserId, 0)
	if err != nil {
		log.Errorf("get user fav list failed: %v", err)
		return nil, err
	}

	var data []*pb.UserFavResponse
	for _, fav := range favs {
		data = append(data, &pb.UserFavResponse{
			UserId:  fav.UserID,
			GoodsId: fav.GoodsID,
		})
	}

	return &pb.UserFavListResponse{
		Total: int32(total),
		Data:  data,
	}, nil
}

// CreateUserFav 创建用户收藏
func (c *UserOpController) CreateUserFav(ctx context.Context, req *pb.UserFavRequest) (*pb.UserFavResponse, error) {
	log.Infof("CreateUserFav request: user_id=%d, goods_id=%d", req.UserId, req.GoodsId)

	userFav, err := c.service.UserFavService().CreateUserFav(ctx, req.UserId, req.GoodsId)
	if err != nil {
		log.Errorf("create user fav failed: %v", err)
		return nil, err
	}

	return &pb.UserFavResponse{
		UserId:  userFav.User,
		GoodsId: userFav.Goods,
	}, nil
}

// DeleteUserFav 删除用户收藏
func (c *UserOpController) DeleteUserFav(ctx context.Context, req *pb.UserFavRequest) (*emptypb.Empty, error) {
	log.Infof("DeleteUserFav request: user_id=%d, goods_id=%d", req.UserId, req.GoodsId)

	err := c.service.UserFavService().DeleteUserFav(ctx, req.UserId, req.GoodsId)
	if err != nil {
		log.Errorf("delete user fav failed: %v", err)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// GetUserFavDetail 获取用户收藏详情（检查是否收藏）
func (c *UserOpController) GetUserFavDetail(ctx context.Context, req *pb.UserFavRequest) (*pb.UserFavResponse, error) {
	log.Infof("GetUserFavDetail request: user_id=%d, goods_id=%d", req.UserId, req.GoodsId)

	userFav, err := c.service.UserFavService().GetUserFavDetail(ctx, req.UserId, req.GoodsId)
	if err != nil {
		log.Errorf("get user fav detail failed: %v", err)
		return nil, err
	}

	return &pb.UserFavResponse{
		UserId:  userFav.User,
		GoodsId: userFav.Goods,
	}, nil
}

// ==================== 用户地址管理 ====================

// GetAddressList 获取地址列表
func (c *UserOpController) GetAddressList(ctx context.Context, req *pb.AddressRequest) (*pb.AddressListResponse, error) {
	log.Infof("GetAddressList request: user_id=%d", req.UserId)

	addresses, total, err := c.service.AddressService().GetAddressList(ctx, req.UserId)
	if err != nil {
		log.Errorf("get address list failed: %v", err)
		return nil, err
	}

	var data []*pb.AddressResponse
	for _, addr := range addresses {
		data = append(data, &pb.AddressResponse{
			Id:           addr.ID,
			UserId:       addr.UserID,
			Province:     addr.Province,
			City:         addr.City,
			District:     addr.District,
			Address:      addr.Address,
			SignerName:   addr.SignerName,
			SignerMobile: addr.SignerMobile,
		})
	}

	return &pb.AddressListResponse{
		Total: int32(total),
		Data:  data,
	}, nil
}

// CreateAddress 创建地址
func (c *UserOpController) CreateAddress(ctx context.Context, req *pb.AddressRequest) (*pb.AddressResponse, error) {
	log.Infof("CreateAddress request: user_id=%d", req.UserId)

	createReq := &servicev1.AddressCreateRequest{
		UserID:       req.UserId,
		Province:     req.Province,
		City:         req.City,
		District:     req.District,
		Address:      req.Address,
		SignerName:   req.SignerName,
		SignerMobile: req.SignerMobile,
	}

	address, err := c.service.AddressService().CreateAddress(ctx, createReq)
	if err != nil {
		log.Errorf("create address failed: %v", err)
		return nil, err
	}

	return &pb.AddressResponse{
		Id:           address.ID,
		UserId:       address.User,
		Province:     address.Province,
		City:         address.City,
		District:     address.District,
		Address:      address.Address,
		SignerName:   address.SignerName,
		SignerMobile: address.SignerMobile,
	}, nil
}

// UpdateAddress 更新地址
func (c *UserOpController) UpdateAddress(ctx context.Context, req *pb.AddressRequest) (*emptypb.Empty, error) {
	log.Infof("UpdateAddress request: id=%d, user_id=%d", req.Id, req.UserId)

	updateReq := &servicev1.AddressUpdateRequest{
		ID:           req.Id,
		UserID:       req.UserId,
		Province:     req.Province,
		City:         req.City,
		District:     req.District,
		Address:      req.Address,
		SignerName:   req.SignerName,
		SignerMobile: req.SignerMobile,
	}

	err := c.service.AddressService().UpdateAddress(ctx, updateReq)
	if err != nil {
		log.Errorf("update address failed: %v", err)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// DeleteAddress 删除地址
func (c *UserOpController) DeleteAddress(ctx context.Context, req *pb.DeleteAddressRequest) (*emptypb.Empty, error) {
	log.Infof("DeleteAddress request: id=%d", req.Id)

	err := c.service.AddressService().DeleteAddress(ctx, req.Id, req.UserId)
	if err != nil {
		log.Errorf("delete address failed: %v", err)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// ==================== 用户留言管理 ====================

// MessageList 获取留言列表
func (c *UserOpController) MessageList(ctx context.Context, req *pb.MessageRequest) (*pb.MessageListResponse, error) {
	log.Infof("MessageList request: user_id=%d", req.UserId)

	messages, total, err := c.service.MessageService().GetMessageList(ctx, req.UserId)
	if err != nil {
		log.Errorf("get message list failed: %v", err)
		return nil, err
	}

	var data []*pb.MessageResponse
	for _, msg := range messages {
		data = append(data, &pb.MessageResponse{
			Id:          msg.ID,
			UserId:      msg.UserID,
			MessageType: msg.MessageType,
			Subject:     msg.Subject,
			Message:     msg.Message,
			File:        msg.File,
		})
	}

	return &pb.MessageListResponse{
		Total: int32(total),
		Data:  data,
	}, nil
}

// CreateMessage 创建留言
func (c *UserOpController) CreateMessage(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {
	log.Infof("CreateMessage request: user_id=%d, message_type=%d", req.UserId, req.MessageType)

	createReq := &servicev1.MessageCreateRequest{
		UserID:      req.UserId,
		MessageType: req.MessageType,
		Subject:     req.Subject,
		Message:     req.Message,
		File:        req.File,
	}

	message, err := c.service.MessageService().CreateMessage(ctx, createReq)
	if err != nil {
		log.Errorf("create message failed: %v", err)
		return nil, err
	}

	return &pb.MessageResponse{
		Id:          message.ID,
		UserId:      message.User,
		MessageType: message.MessageType,
		Subject:     message.Subject,
		Message:     message.Message,
		File:        message.File,
	}, nil
}