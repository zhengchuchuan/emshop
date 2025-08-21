package v1

import (
	"emshop/internal/app/userop/srv/data/v1/interfaces"
	"emshop/internal/app/userop/srv/data/v1/mysql"
	"gorm.io/gorm"
)

// DataFactory 数据访问工厂接口
type DataFactory interface {
	UserFav() interfaces.UserFavStore
	Address() interfaces.AddressStore
	Message() interfaces.MessageStore
	
	// 事务支持
	Begin() *gorm.DB
	
	// DB连接访问
	DB() *gorm.DB
	
	// 关闭连接
	Close() error
}

// GetDataFactory 获取数据访问工厂实例
func GetDataFactory(db *gorm.DB) DataFactory {
	return mysql.NewDataFactory(db)
}