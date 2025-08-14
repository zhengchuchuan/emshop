package mysql

import (
	"emshop/internal/app/userop/srv/data/v1/interfaces"
	"gorm.io/gorm"
)

// DataFactory MySQL数据访问工厂
type DataFactory struct {
	db *gorm.DB
}

// NewDataFactory 创建MySQL数据访问工厂
func NewDataFactory(db *gorm.DB) *DataFactory {
	return &DataFactory{db: db}
}

// UserFav 获取用户收藏仓储
func (f *DataFactory) UserFav() interfaces.UserFavRepository {
	return NewUserFavRepository(f.db)
}

// Address 获取地址仓储
func (f *DataFactory) Address() interfaces.AddressRepository {
	return NewAddressRepository(f.db)
}

// Message 获取留言仓储
func (f *DataFactory) Message() interfaces.MessageRepository {
	return NewMessageRepository(f.db)
}