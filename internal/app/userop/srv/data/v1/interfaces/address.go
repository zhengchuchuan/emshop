package interfaces

import (
	"context"
	"emshop/internal/app/userop/srv/domain/do"
	"emshop/internal/app/userop/srv/domain/dto"
	"gorm.io/gorm"
)

// AddressStore 地址数据访问接口
type AddressStore interface {
	// GetAddressList 获取用户地址列表
	GetAddressList(ctx context.Context, db *gorm.DB, userID int32) ([]*dto.AddressDTO, int64, error)
	
	// CreateAddress 创建地址
	CreateAddress(ctx context.Context, db *gorm.DB, address *do.Address) (*do.Address, error)
	
	// UpdateAddress 更新地址
	UpdateAddress(ctx context.Context, db *gorm.DB, address *do.Address) error
	
	// DeleteAddress 删除地址
	DeleteAddress(ctx context.Context, db *gorm.DB, addressID int32, userID int32) error
	
	// GetAddressByID 根据ID获取地址
	GetAddressByID(ctx context.Context, db *gorm.DB, addressID int32, userID int32) (*do.Address, error)
}