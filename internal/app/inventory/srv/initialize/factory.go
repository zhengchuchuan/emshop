package initialize

import (
	"emshop/internal/app/inventory/srv/data/v1"
	"emshop/internal/app/inventory/srv/global"
	"emshop/pkg/log"
)

// InitFactory 初始化数据工厂管理器
func InitFactory() {
	var err error
	global.FactoryManager, err = v1.NewFactoryManager(global.Config.MySQLOptions)
	if err != nil {
		log.Fatalf("failed to create factory manager: %v", err)
	}
}