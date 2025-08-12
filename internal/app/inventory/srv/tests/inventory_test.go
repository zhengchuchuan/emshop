package tests

import (
	"context"
	"emshop/internal/app/inventory/srv/domain/do"
	"emshop/internal/app/inventory/srv/domain/dto"
	"emshop/internal/app/inventory/srv/global"
	"emshop/internal/app/inventory/srv/initialize"
	"emshop/internal/app/inventory/srv/config"
	v1 "emshop/internal/app/inventory/srv/service/v1"
	"testing"
	"time"
)

func init() {
	// 初始化测试配置
	global.Config = config.New()
	// 这里应该设置测试数据库配置
	global.Config.MySQLOptions.Host = "localhost"
	global.Config.MySQLOptions.Port = "3306"
	global.Config.MySQLOptions.Username = "test"
	global.Config.MySQLOptions.Password = "test"
	global.Config.MySQLOptions.Database = "inventory_test"

	global.Config.RedisOptions.Host = "localhost"
	global.Config.RedisOptions.Port = 6379

	// 初始化
	initialize.InitDB()
	initialize.InitRedis()
	initialize.InitFactory()
}

func TestInventoryService_Create(t *testing.T) {
	service := v1.NewService(global.FactoryManager.GetDataFactory(), global.Config.RedisOptions)

	inv := &dto.InventoryDTO{}
	inv.Goods = 1001
	inv.Stocks = 100

	err := service.Inventorys().Create(context.Background(), inv)
	if err != nil {
		t.Fatalf("创建库存失败: %v", err)
	}

	t.Log("创建库存成功")
}

func TestInventoryService_Get(t *testing.T) {
	service := v1.NewService(global.FactoryManager.GetDataFactory(), global.Config.RedisOptions)

	inv, err := service.Inventorys().Get(context.Background(), 1001)
	if err != nil {
		t.Fatalf("获取库存失败: %v", err)
	}

	t.Logf("库存信息: 商品ID=%d, 库存=%d", inv.Goods, inv.Stocks)
}

func TestInventoryService_Sell(t *testing.T) {
	service := v1.NewService(global.FactoryManager.GetDataFactory(), global.Config.RedisOptions)

	detail := []do.GoodsDetail{
		{Goods: 1001, Num: 10},
	}

	orderSn := "test_order_" + time.Now().Format("20060102150405")
	err := service.Inventorys().Sell(context.Background(), orderSn, detail)
	if err != nil {
		t.Fatalf("库存扣减失败: %v", err)
	}

	t.Log("库存扣减成功")
}

func TestInventoryService_TCC(t *testing.T) {
	service := v1.NewService(global.FactoryManager.GetDataFactory(), global.Config.RedisOptions)

	detail := []do.GoodsDetail{
		{Goods: 1001, Num: 5},
	}

	orderSn := "tcc_test_" + time.Now().Format("20060102150405")

	// Try阶段 - 冻结库存
	t.Log("开始Try阶段")
	err := service.Inventorys().TrySell(context.Background(), orderSn, detail)
	if err != nil {
		t.Fatalf("TCC Try阶段失败: %v", err)
	}
	t.Log("Try阶段完成")

	// Confirm阶段 - 确认扣减
	t.Log("开始Confirm阶段")
	err = service.Inventorys().ConfirmSell(context.Background(), orderSn, detail)
	if err != nil {
		t.Fatalf("TCC Confirm阶段失败: %v", err)
	}
	t.Log("Confirm阶段完成")

	t.Log("TCC事务测试成功")
}

func TestInventoryService_TCCCancel(t *testing.T) {
	service := v1.NewService(global.FactoryManager.GetDataFactory(), global.Config.RedisOptions)

	detail := []do.GoodsDetail{
		{Goods: 1001, Num: 3},
	}

	orderSn := "tcc_cancel_" + time.Now().Format("20060102150405")

	// Try阶段 - 冻结库存
	t.Log("开始Try阶段")
	err := service.Inventorys().TrySell(context.Background(), orderSn, detail)
	if err != nil {
		t.Fatalf("TCC Try阶段失败: %v", err)
	}
	t.Log("Try阶段完成")

	// Cancel阶段 - 取消冻结
	t.Log("开始Cancel阶段")
	err = service.Inventorys().CancelSell(context.Background(), orderSn, detail)
	if err != nil {
		t.Fatalf("TCC Cancel阶段失败: %v", err)
	}
	t.Log("Cancel阶段完成")

	t.Log("TCC取消测试成功")
}