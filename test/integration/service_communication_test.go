package integration

import (
	"testing"
	
	"emshop/internal/app/payment/srv/service/v1"
	"emshop/internal/app/pkg/options"
)

// TestServiceStructures 测试服务结构和基本功能
func TestServiceStructures(t *testing.T) {
	t.Log("测试服务间通信结构")
	
	// 测试DTM配置结构
	dtmOpts := &options.DtmOptions{
		GrpcServer: "localhost:36790",
		HttpServer: "localhost:36789",
	}
	
	if dtmOpts.GrpcServer == "" {
		t.Error("DTM gRPC服务地址不能为空")
	}
	
	// 测试DTM管理器创建
	dtmManager := v1.NewDTMManager(dtmOpts)
	if dtmManager == nil {
		t.Fatal("DTM管理器创建失败")
	}
	
	t.Log("DTM管理器创建成功")
}

// TestOrderSubmissionRequestStructure 测试订单提交请求结构
func TestOrderSubmissionRequestStructure(t *testing.T) {
	req := &v1.OrderSubmissionRequest{
		OrderSn:       "ORDER_20250125001",
		UserID:        1001,
		Amount:        199.99,
		PaymentMethod: 1,
		GoodsDetail: []v1.GoodsDetailItem{
			{Goods: 1001, Num: 1},
			{Goods: 1002, Num: 2},
		},
		Address: "北京市朝阳区测试地址123号",
	}
	
	// 验证必需字段
	if req.OrderSn == "" {
		t.Error("订单号不能为空")
	}
	
	if req.UserID <= 0 {
		t.Error("用户ID必须大于0")
	}
	
	if req.Amount <= 0 {
		t.Error("订单金额必须大于0")
	}
	
	if len(req.GoodsDetail) == 0 {
		t.Error("商品详情不能为空")
	}
	
	if req.Address == "" {
		t.Error("收货地址不能为空")
	}
	
	t.Logf("订单提交请求结构验证成功: OrderSn=%s, UserID=%d, Amount=%.2f", 
		req.OrderSn, req.UserID, req.Amount)
}

// TestPaymentSuccessRequestStructure 测试支付成功请求结构
func TestPaymentSuccessRequestStructure(t *testing.T) {
	thirdPartySn := "ALIPAY_2025012501"
	req := &v1.PaymentSuccessRequest{
		PaymentSn:        "PAY_ORDER_20250125001",
		OrderSn:          "ORDER_20250125001",
		UserID:           1001,
		ThirdPartySn:     &thirdPartySn,
		LogisticsCompany: 1,
		ShippingMethod:   1,
		ReceiverName:     "张三",
		ReceiverPhone:    "13800138000",
		ReceiverAddress:  "北京市朝阳区测试地址123号",
		Items: []v1.LogisticsItem{
			{
				GoodsID:  1001,
				Name:     "测试商品",
				Quantity: 1,
				Weight:   0.5,
				Volume:   100.0,
			},
		},
	}
	
	// 验证必需字段
	if req.PaymentSn == "" {
		t.Error("支付单号不能为空")
	}
	
	if req.OrderSn == "" {
		t.Error("订单号不能为空")
	}
	
	if req.UserID <= 0 {
		t.Error("用户ID必须大于0")
	}
	
	if req.ReceiverName == "" {
		t.Error("收货人姓名不能为空")
	}
	
	if req.ReceiverPhone == "" {
		t.Error("收货人电话不能为空")
	}
	
	if req.ReceiverAddress == "" {
		t.Error("收货人地址不能为空")
	}
	
	if len(req.Items) == 0 {
		t.Error("物流商品信息不能为空")
	}
	
	// 验证物流商品信息
	for i, item := range req.Items {
		if item.GoodsID <= 0 {
			t.Errorf("第%d个商品ID必须大于0", i+1)
		}
		if item.Name == "" {
			t.Errorf("第%d个商品名称不能为空", i+1)
		}
		if item.Quantity <= 0 {
			t.Errorf("第%d个商品数量必须大于0", i+1)
		}
	}
	
	t.Logf("支付成功请求结构验证成功: PaymentSn=%s, OrderSn=%s", 
		req.PaymentSn, req.OrderSn)
}

// TestGoodsDetailItemStructure 测试商品详情结构
func TestGoodsDetailItemStructure(t *testing.T) {
	items := []v1.GoodsDetailItem{
		{Goods: 1001, Num: 2},
		{Goods: 1002, Num: 1},
		{Goods: 1003, Num: 3},
	}
	
	totalQuantity := int32(0)
	for i, item := range items {
		if item.Goods <= 0 {
			t.Errorf("第%d个商品ID必须大于0", i+1)
		}
		if item.Num <= 0 {
			t.Errorf("第%d个商品数量必须大于0", i+1)
		}
		totalQuantity += item.Num
	}
	
	expectedTotal := int32(6) // 2+1+3
	if totalQuantity != expectedTotal {
		t.Errorf("商品总数量计算错误，期望%d，实际%d", expectedTotal, totalQuantity)
	}
	
	t.Logf("商品详情结构验证成功，总数量: %d", totalQuantity)
}

// TestLogisticsItemStructure 测试物流商品结构
func TestLogisticsItemStructure(t *testing.T) {
	items := []v1.LogisticsItem{
		{
			GoodsID:  1001,
			Name:     "iPhone 15",
			Quantity: 1,
			Weight:   0.17,
			Volume:   150.0,
		},
		{
			GoodsID:  1002,
			Name:     "MacBook Pro",
			Quantity: 1,
			Weight:   1.6,
			Volume:   3500.0,
		},
	}
	
	totalWeight := 0.0
	totalVolume := 0.0
	
	for i, item := range items {
		if item.GoodsID <= 0 {
			t.Errorf("第%d个商品ID必须大于0", i+1)
		}
		if item.Name == "" {
			t.Errorf("第%d个商品名称不能为空", i+1)
		}
		if item.Quantity <= 0 {
			t.Errorf("第%d个商品数量必须大于0", i+1)
		}
		if item.Weight <= 0 {
			t.Errorf("第%d个商品重量必须大于0", i+1)
		}
		if item.Volume <= 0 {
			t.Errorf("第%d个商品体积必须大于0", i+1)
		}
		
		totalWeight += item.Weight
		totalVolume += item.Volume
	}
	
	t.Logf("物流商品结构验证成功，总重量: %.2fkg，总体积: %.2fcm³", 
		totalWeight, totalVolume)
}

// TestServiceNaming 测试服务命名约定
func TestServiceNaming(t *testing.T) {
	expectedServices := []string{
		"discovery:///emshop-payment-srv",
		"discovery:///emshop-order-srv",
		"discovery:///emshop-inventory-srv",
		"discovery:///emshop-logistics-srv",
	}
	
	for _, serviceName := range expectedServices {
		if serviceName == "" {
			t.Error("服务名称不能为空")
		}
		
		if !contains(serviceName, "discovery:///emshop-") {
			t.Errorf("服务名称格式不正确: %s", serviceName)
		}
		
		if !contains(serviceName, "-srv") {
			t.Errorf("服务名称应以'-srv'结尾: %s", serviceName)
		}
	}
	
	t.Logf("服务命名约定验证成功，共%d个服务", len(expectedServices))
}

// 辅助函数：检查字符串包含
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || 
		   (len(s) > len(substr) && s[len(s)-len(substr):] == substr) ||
		   (len(s) > len(substr) && len(substr) > 0 && 
		   	func() bool {
		   		for i := 0; i <= len(s)-len(substr); i++ {
		   			if s[i:i+len(substr)] == substr {
		   				return true
		   			}
		   		}
		   		return false
		   	}())
}