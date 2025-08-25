package mysql

import (
	"emshop/internal/app/logistics/srv/data/v1/interfaces"
	"gorm.io/gorm"
)

// dataFactory MySQL数据访问工厂实现
type dataFactory struct {
	db *gorm.DB
}

// NewDataFactory 创建数据访问工厂实例
func NewDataFactory(db *gorm.DB) interfaces.DataFactory {
	return &dataFactory{
		db: db,
	}
}

// DB 获取数据库连接
func (f *dataFactory) DB() *gorm.DB {
	return f.db
}

// Begin 开始事务
func (f *dataFactory) Begin() *gorm.DB {
	return f.db.Begin()
}

// LogisticsOrders 获取物流订单仓储接口
func (f *dataFactory) LogisticsOrders() interfaces.LogisticsOrdersRepo {
	return NewLogisticsOrdersRepo()
}

// LogisticsTracks 获取物流轨迹仓储接口
func (f *dataFactory) LogisticsTracks() interfaces.LogisticsTracksRepo {
	return NewLogisticsTracksRepo()
}

// LogisticsCouriers 获取配送员仓储接口
func (f *dataFactory) LogisticsCouriers() interfaces.LogisticsCouriersRepo {
	return NewLogisticsCouriersRepo()
}