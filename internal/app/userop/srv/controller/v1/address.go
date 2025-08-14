package v1

import (
	"context"
	pb "emshop/api/userop/v1"
	servicev1 "emshop/internal/app/userop/srv/service/v1"
	"emshop/pkg/log"
)

// AddressController 地址控制器
type AddressController struct {
	pb.UnimplementedUserOpServer
	service servicev1.Service
}

// NewAddressController 创建地址控制器
func NewAddressController(service servicev1.Service) *AddressController {
	return &AddressController{
		service: service,
	}
}

// GetAddressList 获取地址列表
func (c *AddressController) GetAddressList(ctx context.Context, req *pb.AddressRequest) (*pb.AddressListResponse, error) {
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
func (c *AddressController) CreateAddress(ctx context.Context, req *pb.AddressRequest) (*pb.AddressResponse, error) {
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
func (c *AddressController) UpdateAddress(ctx context.Context, req *pb.AddressRequest) (*pb.AddressResponse, error) {
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

	return &pb.AddressResponse{}, nil
}

// DeleteAddress 删除地址
func (c *AddressController) DeleteAddress(ctx context.Context, req *pb.DeleteAddressRequest) (*pb.AddressResponse, error) {
	log.Infof("DeleteAddress request: id=%d", req.Id)

	err := c.service.AddressService().DeleteAddress(ctx, req.Id, req.UserId)
	if err != nil {
		log.Errorf("delete address failed: %v", err)
		return nil, err
	}

	return &pb.AddressResponse{}, nil
}