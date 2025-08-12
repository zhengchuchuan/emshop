// Package srv 这是一个inventory服务启动示例
// 注意：这只是一个示例文件，展示如何使用补全的功能
// 如果要作为独立程序运行，请将package改为main
package srv

import (
	"context"
	"emshop/internal/app/inventory/srv/config"
	"emshop/internal/app/inventory/srv/domain/do"
	"emshop/internal/app/inventory/srv/domain/dto"
	"emshop/internal/app/inventory/srv/global"
	"emshop/internal/app/inventory/srv/initialize"
	v1 "emshop/internal/app/inventory/srv/service/v1"
	"emshop/pkg/log"
	"fmt"
)

func MainExample() {
	// 1. 初始化配置
	global.Config = config.New()
	
	// 这里应该从配置文件或环境变量加载实际配置
	// 示例配置
	global.Config.MySQLOptions.Host = "localhost"
	global.Config.MySQLOptions.Port = "3306"
	global.Config.MySQLOptions.Username = "emshop"
	global.Config.MySQLOptions.Password = "password"
	global.Config.MySQLOptions.Database = "emshop_inventory"
	
	global.Config.RedisOptions.Host = "localhost"
	global.Config.RedisOptions.Port = 6379

	// 2. 初始化各种组件
	initialize.InitDB()
	initialize.InitRedis()
	initialize.InitFactory()

	// 3. 创建服务实例
	service := v1.NewService(global.FactoryManager.GetDataFactory(), global.Config.RedisOptions)

	// 4. 演示库存功能
	demonstrateInventoryFeatures(service)

	log.Info("Inventory service initialized successfully!")
}

func demonstrateInventoryFeatures(service v1.ServiceFactory) {
	ctx := context.Background()

	// 创建库存
	fmt.Println("=== 创建库存 ===")
	inv := &dto.InventoryDTO{}
	inv.Goods = 1001
	inv.Stocks = 100

	if err := service.Inventorys().Create(ctx, inv); err != nil {
		log.Errorf("创建库存失败: %v", err)
	} else {
		log.Info("库存创建成功")
	}

	// 查询库存
	fmt.Println("=== 查询库存 ===")
	result, err := service.Inventorys().Get(ctx, 1001)
	if err != nil {
		log.Errorf("查询库存失败: %v", err)
	} else {
		log.Infof("库存信息: 商品ID=%d, 库存=%d", result.Goods, result.Stocks)
	}

	// 库存扣减
	fmt.Println("=== 库存扣减 ===")
	detail := []do.GoodsDetail{
		{Goods: 1001, Num: 10},
	}
	
	if err := service.Inventorys().Sell(ctx, "demo_order_001", detail); err != nil {
		log.Errorf("库存扣减失败: %v", err)
	} else {
		log.Info("库存扣减成功")
	}

	// TCC分布式事务演示
	fmt.Println("=== TCC分布式事务演示 ===")
	tccDetail := []do.GoodsDetail{
		{Goods: 1001, Num: 5},
	}

	// Try阶段
	if err := service.Inventorys().TrySell(ctx, "tcc_demo_001", tccDetail); err != nil {
		log.Errorf("TCC Try失败: %v", err)
	} else {
		log.Info("TCC Try成功")
	}

	// Confirm阶段
	if err := service.Inventorys().ConfirmSell(ctx, "tcc_demo_001", tccDetail); err != nil {
		log.Errorf("TCC Confirm失败: %v", err)
	} else {
		log.Info("TCC Confirm成功")
	}

	fmt.Println("=== 演示完成 ===")
}