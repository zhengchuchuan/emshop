package v1

import (
	"emshop/internal/app/userop/srv/data/v1/interfaces"
	"emshop/internal/app/userop/srv/data/v1/mysql"
	"gorm.io/gorm"
)

// DataFactory 数据访问工厂接口
type DataFactory interface {
	UserFav() interfaces.UserFavRepository
	Address() interfaces.AddressRepository
	Message() interfaces.MessageRepository
}

// GetDataFactory 获取数据访问工厂实例
func GetDataFactory(db *gorm.DB) DataFactory {
	return mysql.NewDataFactory(db)
}