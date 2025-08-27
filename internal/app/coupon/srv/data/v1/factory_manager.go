package v1

import (
	"emshop/internal/app/coupon/srv/data/v1/interfaces"
	"emshop/internal/app/coupon/srv/data/v1/mysql"
	"emshop/internal/app/pkg/options"
	"emshop/pkg/log"
)

// FactoryManager 工厂管理器
type FactoryManager struct {
	dataFactory interfaces.DataFactory
}

// NewFactoryManager 创建工厂管理器
func NewFactoryManager(mysqlOpts *options.MySQLOptions) (*FactoryManager, error) {
	// 创建MySQL数据工厂
	dataFactory, err := mysql.NewDataFactory(mysqlOpts)
	if err != nil {
		log.Errorf("failed to create coupon mysql factory: %v", err)
		return nil, err
	}

	return &FactoryManager{
		dataFactory: dataFactory,
	}, nil
}

// GetDataFactory 获取数据工厂
func (fm *FactoryManager) GetDataFactory() interfaces.DataFactory {
	return fm.dataFactory
}

// Close 关闭所有连接
func (fm *FactoryManager) Close() error {
	if fm.dataFactory != nil {
		return fm.dataFactory.Close()
	}
	return nil
}