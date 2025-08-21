package v1

import (
	"context"
	datav1 "emshop/internal/app/userop/srv/data/v1"
	"emshop/internal/app/userop/srv/domain/do"
	"emshop/internal/app/userop/srv/domain/dto"
)

// AddressService 地址服务接口
type AddressService interface {
	GetAddressList(ctx context.Context, userID int32) ([]*dto.AddressDTO, int64, error)
	CreateAddress(ctx context.Context, req *AddressCreateRequest) (*do.Address, error)
	UpdateAddress(ctx context.Context, req *AddressUpdateRequest) error
	DeleteAddress(ctx context.Context, addressID int32, userID int32) error
}

// AddressCreateRequest 创建地址请求
type AddressCreateRequest struct {
	UserID       int32
	Province     string
	City         string
	District     string
	Address      string
	SignerName   string
	SignerMobile string
}

// AddressUpdateRequest 更新地址请求
type AddressUpdateRequest struct {
	ID           int32
	UserID       int32
	Province     string
	City         string
	District     string
	Address      string
	SignerName   string
	SignerMobile string
}

type addressService struct {
	dataFactory datav1.DataFactory
}

// NewAddressService 创建地址服务
func NewAddressService(dataFactory datav1.DataFactory) AddressService {
	return &addressService{
		dataFactory: dataFactory,
	}
}

func (s *addressService) GetAddressList(ctx context.Context, userID int32) ([]*dto.AddressDTO, int64, error) {
	return s.dataFactory.Address().GetAddressList(ctx, s.dataFactory.DB(), userID)
}

func (s *addressService) CreateAddress(ctx context.Context, req *AddressCreateRequest) (*do.Address, error) {
	address := &do.Address{
		User:         req.UserID,
		Province:     req.Province,
		City:         req.City,
		District:     req.District,
		Address:      req.Address,
		SignerName:   req.SignerName,
		SignerMobile: req.SignerMobile,
	}
	return s.dataFactory.Address().CreateAddress(ctx, s.dataFactory.DB(), address)
}

func (s *addressService) UpdateAddress(ctx context.Context, req *AddressUpdateRequest) error {
	address := &do.Address{
		BaseModel: do.BaseModel{
			ID: req.ID,
		},
		User:         req.UserID,
		Province:     req.Province,
		City:         req.City,
		District:     req.District,
		Address:      req.Address,
		SignerName:   req.SignerName,
		SignerMobile: req.SignerMobile,
	}
	return s.dataFactory.Address().UpdateAddress(ctx, s.dataFactory.DB(), address)
}

func (s *addressService) DeleteAddress(ctx context.Context, addressID int32, userID int32) error {
	return s.dataFactory.Address().DeleteAddress(ctx, s.dataFactory.DB(), addressID, userID)
}