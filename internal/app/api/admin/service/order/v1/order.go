package order

import (
	"context"
	proto "emshop/api/order/v1"
	"emshop/internal/app/api/admin/data"
	"emshop/pkg/log"
	"fmt"
)

// OrderSrv 订单服务接口
type OrderSrv interface {
	// 管理员订单列表（支持多维度查询）
	AdminOrderList(ctx context.Context, request *proto.OrderFilterRequest) (*proto.OrderListResponse, error)
	// 管理员订单详情
	AdminOrderDetail(ctx context.Context, request *proto.OrderRequest) (*proto.OrderInfoDetailResponse, error)
	// 更新订单状态
	UpdateOrderStatus(ctx context.Context, request *proto.OrderStatus) error
	// 按订单号查询
	GetOrderByOrderSn(ctx context.Context, orderSn string) (*proto.OrderInfoDetailResponse, error)
	// 按用户ID查询订单
	GetOrdersByUserId(ctx context.Context, userId int32, pages, pagePerNums int32) (*proto.OrderListResponse, error)
}

type orderService struct {
	data data.DataFactory
}

// NewOrderService 创建订单服务实例
func NewOrderService(data data.DataFactory) OrderSrv {
	return &orderService{
		data: data,
	}
}

// AdminOrderList 管理员查看所有订单列表
func (os *orderService) AdminOrderList(ctx context.Context, request *proto.OrderFilterRequest) (*proto.OrderListResponse, error) {
	log.Infof("Admin order service: AdminOrderList called with userId=%d, pages=%d, pagePerNums=%d", 
		request.UserId, request.Pages, request.PagePerNums)
	
	// 验证分页参数
	if request.Pages == nil || *request.Pages <= 0 {
		pages := int32(1)
		request.Pages = &pages
	}
	if request.PagePerNums == nil || *request.PagePerNums <= 0 {
		pagePerNums := int32(10)
		request.PagePerNums = &pagePerNums
	}
	if *request.PagePerNums > 100 {
		pagePerNums := int32(100)
		request.PagePerNums = &pagePerNums
	}
	
	response, err := os.data.Order().AdminOrderList(ctx, request)
	if err != nil {
		log.Errorf("Admin order service: AdminOrderList failed: %v", err)
		return nil, err
	}
	
	log.Infof("Admin order service: AdminOrderList successful, returned %d orders", len(response.Data))
	return response, nil
}

// AdminOrderDetail 管理员查看订单详情
func (os *orderService) AdminOrderDetail(ctx context.Context, request *proto.OrderRequest) (*proto.OrderInfoDetailResponse, error) {
	log.Infof("Admin order service: AdminOrderDetail called with ID=%d, OrderSn=%s", 
		request.Id, request.OrderSn)
	
	// 参数验证
	if request.Id <= 0 && (request.OrderSn == nil || *request.OrderSn == "") {
		return nil, fmt.Errorf("订单ID或订单号不能为空")
	}
	
	response, err := os.data.Order().AdminOrderDetail(ctx, request)
	if err != nil {
		log.Errorf("Admin order service: AdminOrderDetail failed: %v", err)
		return nil, err
	}
	
	log.Info("Admin order service: AdminOrderDetail successful")
	return response, nil
}

// UpdateOrderStatus 更新订单状态
func (os *orderService) UpdateOrderStatus(ctx context.Context, request *proto.OrderStatus) error {
	log.Infof("Admin order service: UpdateOrderStatus called for order %s to status %s", 
		request.OrderSn, request.Status)
	
	// 参数验证
	if request.OrderSn == "" {
		return fmt.Errorf("订单号不能为空")
	}
	if request.Status == "" {
		return fmt.Errorf("订单状态不能为空")
	}
	
	// 验证状态值是否有效
	validStatuses := map[string]bool{
		"PAYING":    true,  // 待支付
		"TRADE_SUCCESS": true,  // 支付成功
		"TRADE_CLOSED": true,   // 交易关闭
		"WAIT_BUYER_CONFIRM_GOODS": true, // 待发货
		"TRADE_FINISHED": true, // 交易完成
		"PAYING_ERROR": true,   // 支付失败
	}
	
	if !validStatuses[request.Status] {
		return fmt.Errorf("无效的订单状态: %s", request.Status)
	}
	
	err := os.data.Order().UpdateOrderStatus(ctx, request)
	if err != nil {
		log.Errorf("Admin order service: UpdateOrderStatus failed: %v", err)
		return err
	}
	
	log.Info("Admin order service: UpdateOrderStatus successful")
	return nil
}

// GetOrderByOrderSn 按订单号查询订单
func (os *orderService) GetOrderByOrderSn(ctx context.Context, orderSn string) (*proto.OrderInfoDetailResponse, error) {
	log.Infof("Admin order service: GetOrderByOrderSn called with orderSn=%s", orderSn)
	
	if orderSn == "" {
		return nil, fmt.Errorf("订单号不能为空")
	}
	
	response, err := os.data.Order().GetOrderByOrderSn(ctx, orderSn)
	if err != nil {
		log.Errorf("Admin order service: GetOrderByOrderSn failed: %v", err)
		return nil, err
	}
	
	log.Info("Admin order service: GetOrderByOrderSn successful")
	return response, nil
}

// GetOrdersByUserId 按用户ID查询订单列表
func (os *orderService) GetOrdersByUserId(ctx context.Context, userId int32, pages, pagePerNums int32) (*proto.OrderListResponse, error) {
	log.Infof("Admin order service: GetOrdersByUserId called with userId=%d, pages=%d, pagePerNums=%d", 
		userId, pages, pagePerNums)
	
	if userId <= 0 {
		return nil, fmt.Errorf("用户ID必须大于0")
	}
	
	// 验证分页参数
	if pages <= 0 {
		pages = 1
	}
	if pagePerNums <= 0 {
		pagePerNums = 10
	}
	if pagePerNums > 100 {
		pagePerNums = 100
	}
	
	response, err := os.data.Order().GetOrdersByUserId(ctx, userId, pages, pagePerNums)
	if err != nil {
		log.Errorf("Admin order service: GetOrdersByUserId failed: %v", err)
		return nil, err
	}
	
	log.Infof("Admin order service: GetOrdersByUserId successful, returned %d orders", len(response.Data))
	return response, nil
}