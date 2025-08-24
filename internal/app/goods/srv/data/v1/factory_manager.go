package v1

import (
	"emshop/internal/app/goods/srv/data/v1/elasticsearch"
	"emshop/internal/app/goods/srv/data/v1/mysql"
	"emshop/internal/app/goods/srv/data/v1/sync"
	"emshop/internal/app/pkg/options"
	"emshop/pkg/log"
)

// FactoryManager 工厂管理器
type FactoryManager struct {
	dataFactory   mysql.DataFactory
	syncManager   sync.DataSyncManagerInterface
	esOptions     *options.EsOptions
}

// NewFactoryManager 创建工厂管理器
func NewFactoryManager(mysqlOpts *options.MySQLOptions, esOpts *options.EsOptions) (*FactoryManager, error) {
	// 创建搜索引擎工厂
	searchFactory, err := elasticsearch.NewElasticsearchFactory(esOpts)
	if err != nil {
		log.Errorf("failed to create elasticsearch factory: %v", err)
		return nil, err
	}

	// 创建MySQL数据工厂
	dataFactory, err := mysql.NewMySQLFactory(mysqlOpts, searchFactory)
	if err != nil {
		log.Errorf("failed to create mysql factory: %v", err)
		return nil, err
	}

	// 创建数据同步管理器
	syncManager := sync.NewDataSyncManager(dataFactory, searchFactory)

	return &FactoryManager{
		dataFactory: dataFactory,
		syncManager: syncManager,
		esOptions:   esOpts,
	}, nil
}

// GetDataFactory 获取数据工厂
func (fm *FactoryManager) GetDataFactory() mysql.DataFactory {
	return fm.dataFactory
}

// GetSyncManager 获取同步管理器
func (fm *FactoryManager) GetSyncManager() sync.DataSyncManagerInterface {
	return fm.syncManager
}

// GetEsOptions 获取ES配置选项
func (fm *FactoryManager) GetEsOptions() *options.EsOptions {
	return fm.esOptions
}

// Close 关闭所有连接
func (fm *FactoryManager) Close() error {
	if fm.dataFactory != nil {
		return fm.dataFactory.Close()
	}
	return nil
}