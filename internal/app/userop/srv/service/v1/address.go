package v1

import (
	"context"
	datav1 "emshop/internal/app/userop/srv/data/v1"
	"emshop/internal/app/userop/srv/data/v1/interfaces"
	"emshop/internal/app/userop/srv/data/v1/mysql"
	"emshop/internal/app/userop/srv/domain/do"
	"emshop/internal/app/userop/srv/domain/dto"
	"emshop/pkg/log"
	"gorm.io/gorm"
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
	// 预加载的核心组件（日常CRUD操作）
	addressDAO  interfaces.AddressStore
	db          *gorm.DB
	
	// 保留工厂引用（复杂操作和扩展）
	dataFactory mysql.DataFactory
}

// NewAddressService 创建地址服务
func NewAddressService(dataFactory datav1.DataFactory) AddressService {
	// 适配器模式：将datav1.DataFactory转换为mysql.DataFactory
	mysqlFactory, ok := dataFactory.(mysql.DataFactory)
	if !ok {
		log.Errorf("dataFactory is not mysql.DataFactory type")
		return &addressService{
			dataFactory: dataFactory.(mysql.DataFactory),
		}
	}
	
	return &addressService{
		// 预加载核心组件，避免每次方法调用时重复获取
		addressDAO:  mysqlFactory.Address(),
		db:          mysqlFactory.DB(),
		
		// 保留工厂引用用于复杂操作
		dataFactory: mysqlFactory,
	}
}

func (s *addressService) GetAddressList(ctx context.Context, userID int32) ([]*dto.AddressDTO, int64, error) {
	log.Debugf("Getting address list for user: %d", userID)
	
	// 直接使用预加载的DAO
	addressList, total, err := s.addressDAO.GetAddressList(ctx, s.db, userID)
	if err != nil {
		log.Errorf("Failed to get address list for user %d: %v", userID, err)
		return nil, 0, err
	}
	
	log.Debugf("Successfully got address list for user %d, total: %d", userID, total)
	return addressList, total, nil
}

func (s *addressService) CreateAddress(ctx context.Context, req *AddressCreateRequest) (*do.Address, error) {
	log.Debugf("Creating address for user: %d, address: %s", req.UserID, req.Address)
	
	address := &do.Address{
		User:         req.UserID,
		Province:     req.Province,
		City:         req.City,
		District:     req.District,
		Address:      req.Address,
		SignerName:   req.SignerName,
		SignerMobile: req.SignerMobile,
	}
	
	// 直接使用预加载的DAO
	createdAddress, err := s.addressDAO.CreateAddress(ctx, s.db, address)
	if err != nil {
		log.Errorf("Failed to create address for user %d: %v", req.UserID, err)
		return nil, err
	}
	
	log.Infof("Successfully created address for user %d, addressID: %d", req.UserID, createdAddress.ID)
	return createdAddress, nil
}

func (s *addressService) UpdateAddress(ctx context.Context, req *AddressUpdateRequest) error {
	log.Debugf("Updating address: ID=%d, userID=%d", req.ID, req.UserID)
	
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
	
	// 直接使用预加载的DAO
	err := s.addressDAO.UpdateAddress(ctx, s.db, address)
	if err != nil {
		log.Errorf("Failed to update address: ID=%d, userID=%d, error=%v", req.ID, req.UserID, err)
		return err
	}
	
	log.Infof("Successfully updated address: ID=%d, userID=%d", req.ID, req.UserID)
	return nil
}

func (s *addressService) DeleteAddress(ctx context.Context, addressID int32, userID int32) error {
	log.Debugf("Deleting address: ID=%d, userID=%d", addressID, userID)
	
	// 直接使用预加载的DAO
	err := s.addressDAO.DeleteAddress(ctx, s.db, addressID, userID)
	if err != nil {
		log.Errorf("Failed to delete address: ID=%d, userID=%d, error=%v", addressID, userID, err)
		return err
	}
	
	log.Infof("Successfully deleted address: ID=%d, userID=%d", addressID, userID)
	return nil
}