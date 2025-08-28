package v1

import (
	"context"
	proto "emshop/api/order/v1"
	"emshop/internal/app/api/emshop/data"
)

type OrderSrv interface {
	// 订单管理
	OrderList(ctx context.Context, request *proto.OrderFilterRequest) (*proto.OrderListResponse, error)
	CreateOrder(ctx context.Context, request *proto.OrderRequest) (*proto.OrderInfoResponse, error)
	OrderDetail(ctx context.Context, request *proto.OrderRequest) (*proto.OrderInfoDetailResponse, error)
	UpdateOrderStatus(ctx context.Context, request *proto.OrderStatus) error

	// 购物车管理
	CartItemList(ctx context.Context, request *proto.UserInfo) (*proto.CartItemListResponse, error)
	CreateCartItem(ctx context.Context, request *proto.CartItemRequest) (*proto.ShopCartInfoResponse, error)
	UpdateCartItem(ctx context.Context, request *proto.CartItemRequest) error
	DeleteCartItem(ctx context.Context, request *proto.CartItemRequest) error
}

type orderService struct {
	data data.DataFactory
}

// ==================== 订单管理 ====================

func (os *orderService) OrderList(ctx context.Context, request *proto.OrderFilterRequest) (*proto.OrderListResponse, error) {
	return os.data.Order().OrderList(ctx, request)
}

func (os *orderService) CreateOrder(ctx context.Context, request *proto.OrderRequest) (*proto.OrderInfoResponse, error) {
	_, err := os.data.Order().CreateOrder(ctx, request)
	if err != nil {
		return nil, err
	}
	return &proto.OrderInfoResponse{}, nil
}

func (os *orderService) OrderDetail(ctx context.Context, request *proto.OrderRequest) (*proto.OrderInfoDetailResponse, error) {
	return os.data.Order().OrderDetail(ctx, request)
}

func (os *orderService) UpdateOrderStatus(ctx context.Context, request *proto.OrderStatus) error {
	_, err := os.data.Order().UpdateOrderStatus(ctx, request)
	return err
}

// ==================== 购物车管理 ====================

func (os *orderService) CartItemList(ctx context.Context, request *proto.UserInfo) (*proto.CartItemListResponse, error) {
	return os.data.Order().CartItemList(ctx, request)
}

func (os *orderService) CreateCartItem(ctx context.Context, request *proto.CartItemRequest) (*proto.ShopCartInfoResponse, error) {
	return os.data.Order().CreateCartItem(ctx, request)
}

func (os *orderService) UpdateCartItem(ctx context.Context, request *proto.CartItemRequest) error {
	_, err := os.data.Order().UpdateCartItem(ctx, request)
	return err
}

func (os *orderService) DeleteCartItem(ctx context.Context, request *proto.CartItemRequest) error {
	_, err := os.data.Order().DeleteCartItem(ctx, request)
	return err
}

func NewOrder(data data.DataFactory) *orderService {
	return &orderService{data: data}
}

var _ OrderSrv = &orderService{}
