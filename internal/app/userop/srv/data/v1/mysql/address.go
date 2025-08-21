package mysql

import (
	"context"
	code2 "emshop/gin-micro/code"
	"emshop/internal/app/userop/srv/data/v1/interfaces"
	"emshop/internal/app/userop/srv/domain/do"
	"emshop/internal/app/userop/srv/domain/dto"
	"emshop/pkg/errors"
	"emshop/pkg/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type addressRepository struct {
	// 无状态结构体，不需要db字段
}

func NewAddressRepository() interfaces.AddressStore {
	return &addressRepository{}
}

// GetAddressList 获取用户地址列表
func (r *addressRepository) GetAddressList(ctx context.Context, db *gorm.DB, userID int32) ([]*dto.AddressDTO, int64, error) {
	var addresses []do.Address
	var total int64

	result := db.WithContext(ctx).Where("user = ?", userID).Find(&addresses)
	if result.Error != nil {
		log.Errorf("get address list failed: %v", result.Error)
		return nil, 0, errors.WithCode(code2.ErrDatabase, "获取地址列表失败: %v", result.Error)
	}

	total = result.RowsAffected

	// 转换为DTO
	var dtos []*dto.AddressDTO
	for _, address := range addresses {
		dtos = append(dtos, &dto.AddressDTO{
			ID:           address.ID,
			UserID:       address.User,
			Province:     address.Province,
			City:         address.City,
			District:     address.District,
			Address:      address.Address,
			SignerName:   address.SignerName,
			SignerMobile: address.SignerMobile,
			CreatedAt:    address.CreatedAt,
			UpdatedAt:    address.UpdatedAt,
		})
	}

	return dtos, total, nil
}

// CreateAddress 创建地址
func (r *addressRepository) CreateAddress(ctx context.Context, db *gorm.DB, address *do.Address) (*do.Address, error) {
	if err := db.WithContext(ctx).Create(address).Error; err != nil {
		log.Errorf("create address failed: %v", err)
		return nil, errors.WithCode(code2.ErrDatabase, "创建地址失败: %v", err)
	}

	log.Infof("created address %d for user %d successfully", address.ID, address.User)
	return address, nil
}

// UpdateAddress 更新地址
func (r *addressRepository) UpdateAddress(ctx context.Context, db *gorm.DB, address *do.Address) error {
	// 先查找现有地址
	var existingAddress do.Address
	result := db.WithContext(ctx).Where("id = ? AND user = ?", address.ID, address.User).First(&existingAddress)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return status.Errorf(codes.NotFound, "地址不存在")
		}
		log.Errorf("find address failed: %v", result.Error)
		return errors.WithCode(code2.ErrDatabase, "查找地址失败: %v", result.Error)
	}

	// 更新字段 - 只更新非空字段
	updates := make(map[string]interface{})
	if address.Province != "" {
		updates["province"] = address.Province
	}
	if address.City != "" {
		updates["city"] = address.City
	}
	if address.District != "" {
		updates["district"] = address.District
	}
	if address.Address != "" {
		updates["address"] = address.Address
	}
	if address.SignerName != "" {
		updates["signer_name"] = address.SignerName
	}
	if address.SignerMobile != "" {
		updates["signer_mobile"] = address.SignerMobile
	}

	if err := db.WithContext(ctx).Model(&existingAddress).Updates(updates).Error; err != nil {
		log.Errorf("update address failed: %v", err)
		return errors.WithCode(code2.ErrDatabase, "更新地址失败: %v", err)
	}

	log.Infof("updated address %d for user %d successfully", address.ID, address.User)
	return nil
}

// DeleteAddress 删除地址
func (r *addressRepository) DeleteAddress(ctx context.Context, db *gorm.DB, addressID int32, userID int32) error {
	result := db.WithContext(ctx).Where("id = ? AND user = ?", addressID, userID).Delete(&do.Address{})
	if result.Error != nil {
		log.Errorf("delete address failed: %v", result.Error)
		return errors.WithCode(code2.ErrDatabase, "删除地址失败: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return status.Errorf(codes.NotFound, "地址不存在")
	}

	log.Infof("deleted address %d for user %d successfully", addressID, userID)
	return nil
}

// GetAddressByID 根据ID获取地址
func (r *addressRepository) GetAddressByID(ctx context.Context, db *gorm.DB, addressID int32, userID int32) (*do.Address, error) {
	var address do.Address
	result := db.WithContext(ctx).Where("id = ? AND user = ?", addressID, userID).First(&address)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "地址不存在")
		}
		log.Errorf("get address by id failed: %v", result.Error)
		return nil, errors.WithCode(code2.ErrDatabase, "获取地址详情失败: %v", result.Error)
	}

	return &address, nil
}