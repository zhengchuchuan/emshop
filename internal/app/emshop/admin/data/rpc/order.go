package rpc

import (
	"context"
	opbv1 "emshop/api/order/v1"
	"emshop/internal/app/emshop/admin/data"
	"emshop/pkg/log"
)


type order struct {
	oc opbv1.OrderClient
}

func NewOrder(oc opbv1.OrderClient) *order {
	return &order{oc}
}


// 管理员查看所有订单列表（支持多维度筛选）
func (o *order) AdminOrderList(ctx context.Context, request *opbv1.OrderFilterRequest) (*opbv1.OrderListResponse, error) {
	log.Infof("Calling AdminOrderList gRPC for request: userId=%d, pages=%d, pagePerNums=%d", 
		request.UserId, request.Pages, request.PagePerNums)
	
	// 对于管理员，如果没有指定用户ID，则查看所有订单
	response, err := o.oc.OrderList(ctx, request)
	if err != nil {
		log.Errorf("AdminOrderList gRPC call failed: %v", err)
		return nil, err
	}
	
	log.Infof("AdminOrderList gRPC call successful, returned %d orders", len(response.Data))
	return response, nil
}

// 管理员查看订单详情
func (o *order) AdminOrderDetail(ctx context.Context, request *opbv1.OrderRequest) (*opbv1.OrderInfoDetailResponse, error) {
	log.Infof("Calling AdminOrderDetail gRPC for order ID: %d, OrderSn: %s", request.Id, request.OrderSn)
	
	response, err := o.oc.OrderDetail(ctx, request)
	if err != nil {
		log.Errorf("AdminOrderDetail gRPC call failed: %v", err)
		return nil, err
	}
	
	log.Infof("AdminOrderDetail gRPC call successful")
	return response, nil
}

// 管理员更新订单状态
func (o *order) UpdateOrderStatus(ctx context.Context, request *opbv1.OrderStatus) error {
	log.Infof("Calling UpdateOrderStatus gRPC for order: %s, status: %s", request.OrderSn, request.Status)
	
	_, err := o.oc.UpdateOrderStatus(ctx, request)
	if err != nil {
		log.Errorf("UpdateOrderStatus gRPC call failed: %v", err)
		return err
	}
	
	log.Infof("UpdateOrderStatus gRPC call successful")
	return nil
}

// 按订单号查询订单
func (o *order) GetOrderByOrderSn(ctx context.Context, orderSn string) (*opbv1.OrderInfoDetailResponse, error) {
	log.Infof("Calling GetOrderByOrderSn gRPC for orderSn: %s", orderSn)
	
	request := &opbv1.OrderRequest{
		OrderSn: &orderSn,
	}
	
	response, err := o.oc.OrderDetail(ctx, request)
	if err != nil {
		log.Errorf("GetOrderByOrderSn gRPC call failed: %v", err)
		return nil, err
	}
	
	log.Infof("GetOrderByOrderSn gRPC call successful")
	return response, nil
}

// 按用户ID查询订单列表
func (o *order) GetOrdersByUserId(ctx context.Context, userId int32, pages, pagePerNums int32) (*opbv1.OrderListResponse, error) {
	log.Infof("Calling GetOrdersByUserId gRPC for userId: %d, pages: %d, pagePerNums: %d", 
		userId, pages, pagePerNums)
	
	request := &opbv1.OrderFilterRequest{
		UserId:      userId,
		Pages:       &pages,
		PagePerNums: &pagePerNums,
	}
	
	response, err := o.oc.OrderList(ctx, request)
	if err != nil {
		log.Errorf("GetOrdersByUserId gRPC call failed: %v", err)
		return nil, err
	}
	
	log.Infof("GetOrdersByUserId gRPC call successful, returned %d orders", len(response.Data))
	return response, nil
}

var _ data.OrderData = &order{}