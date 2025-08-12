package v1

import (
	"emshop/internal/app/order/srv/data/v1/mysql"
	"emshop/internal/app/pkg/options"
	"emshop/pkg/log"
)

// FactoryManager 工厂管理器
type FactoryManager struct {
	dataFactory mysql.DataFactory
}

// NewFactoryManager 创建工厂管理器
func NewFactoryManager(mysqlOpts *options.MySQLOptions, registryOpts *options.RegistryOptions) (*FactoryManager, error) {
	// 创建RPC客户端
	goodsClient := mysql.GetGoodsClient(registryOpts)
	invClient := mysql.GetInventoryClient(registryOpts)

	// 创建MySQL数据工厂
	dataFactory, err := mysql.NewMySQLFactory(mysqlOpts, goodsClient, invClient)
	if err != nil {
		log.Errorf("failed to create mysql factory: %v", err)
		return nil, err
	}

	return &FactoryManager{
		dataFactory: dataFactory,
	}, nil
}

// GetDataFactory 获取数据工厂
func (fm *FactoryManager) GetDataFactory() mysql.DataFactory {
	return fm.dataFactory
}

// Close 关闭所有连接
func (fm *FactoryManager) Close() error {
	if fm.dataFactory != nil {
		return fm.dataFactory.Close()
	}
	return nil
}