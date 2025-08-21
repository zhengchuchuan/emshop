package mysql

import (
	"emshop/internal/app/userop/srv/data/v1/interfaces"
	"gorm.io/gorm"
)

// DataFactory MySQL数据访问工厂接口
type DataFactory interface {
	// 主存储接口
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

// mysqlFactory MySQL数据工厂实现
type mysqlFactory struct {
	db *gorm.DB

	// DAO单例
	userFavDAO interfaces.UserFavStore
	addressDAO interfaces.AddressStore
	messageDAO interfaces.MessageStore
}

func (mf *mysqlFactory) Begin() *gorm.DB {
	return mf.db.Begin()
}

func (mf *mysqlFactory) DB() *gorm.DB {
	return mf.db
}

func (mf *mysqlFactory) Close() error {
	sqlDB, err := mf.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (mf *mysqlFactory) UserFav() interfaces.UserFavStore {
	return mf.userFavDAO
}

func (mf *mysqlFactory) Address() interfaces.AddressStore {
	return mf.addressDAO
}

func (mf *mysqlFactory) Message() interfaces.MessageStore {
	return mf.messageDAO
}

var _ DataFactory = &mysqlFactory{}

// NewDataFactory 创建 MySQL数据访问工厂
func NewDataFactory(db *gorm.DB) DataFactory {
	// 创建工厂实例
	factory := &mysqlFactory{
		db: db,
	}

	// 创建DAO实例
	factory.userFavDAO = NewUserFavRepository()
	factory.addressDAO = NewAddressRepository()
	factory.messageDAO = NewMessageRepository()

	return factory
}