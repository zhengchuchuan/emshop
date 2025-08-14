package rpc

import (
	"context"
	"time"
	opbv1 "emshop/api/order/v1"
	"emshop/gin-micro/server/rpc-server"
	"emshop/gin-micro/server/rpc-server/client-interceptors"
	"emshop/internal/app/emshop/api/data"
	"emshop/gin-micro/registry"
	"emshop/pkg/log"
	"google.golang.org/grpc"
)

const orderserviceName = "discovery:///emshop-order-srv"

type order struct {
	oc opbv1.OrderClient
}

func NewOrder(oc opbv1.OrderClient) *order {
	return &order{oc}
}

func NewOrderServiceClient(r registry.Discovery) opbv1.OrderClient {
	log.Infof("Initializing gRPC connection to service: %s", orderserviceName)
	conn, err := rpcserver.DialInsecure(
		context.Background(),
		rpcserver.WithEndpoint(orderserviceName),
		rpcserver.WithDiscovery(r),
		rpcserver.WithClientTimeout(10*time.Second),
		rpcserver.WithClientOptions(grpc.WithNoProxy()),
		rpcserver.WithClientUnaryInterceptor(clientinterceptors.UnaryTracingInterceptor),
	)
	if err != nil {
		log.Errorf("Failed to create gRPC connection: %v", err)
		panic(err)
	}
	log.Info("gRPC connection established successfully")
	c := opbv1.NewOrderClient(conn)
	return c
}

// ==================== 订单管理 ====================

func (o *order) OrderList(ctx context.Context, request *opbv1.OrderFilterRequest) (*opbv1.OrderListResponse, error) {
	log.Infof("Calling OrderList gRPC for user: %d", request.UserId)
	response, err := o.oc.OrderList(ctx, request)
	if err != nil {
		log.Errorf("OrderList gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("OrderList gRPC call successful, total: %d", response.Total)
	return response, nil
}

func (o *order) CreateOrder(ctx context.Context, request *opbv1.OrderRequest) (*opbv1.OrderInfoResponse, error) {
	log.Infof("Calling CreateOrder gRPC for user: %d", request.UserId)
	_, err := o.oc.CreateOrder(ctx, request)
	if err != nil {
		log.Errorf("CreateOrder gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("CreateOrder gRPC call successful")
	return &opbv1.OrderInfoResponse{}, nil
}

func (o *order) OrderDetail(ctx context.Context, request *opbv1.OrderRequest) (*opbv1.OrderInfoDetailResponse, error) {
	log.Infof("Calling OrderDetail gRPC for order: %d", request.Id)
	response, err := o.oc.OrderDetail(ctx, request)
	if err != nil {
		log.Errorf("OrderDetail gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("OrderDetail gRPC call successful")
	return response, nil
}

func (o *order) UpdateOrderStatus(ctx context.Context, request *opbv1.OrderStatus) (*opbv1.OrderInfoResponse, error) {
	log.Infof("Calling UpdateOrderStatus gRPC for order: %s", request.OrderSn)
	_, err := o.oc.UpdateOrderStatus(ctx, request)
	if err != nil {
		log.Errorf("UpdateOrderStatus gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("UpdateOrderStatus gRPC call successful")
	return &opbv1.OrderInfoResponse{}, nil
}

// ==================== 购物车管理 ====================

func (o *order) CartItemList(ctx context.Context, request *opbv1.UserInfo) (*opbv1.CartItemListResponse, error) {
	log.Infof("Calling CartItemList gRPC for user: %d", request.Id)
	response, err := o.oc.CartItemList(ctx, request)
	if err != nil {
		log.Errorf("CartItemList gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("CartItemList gRPC call successful, total: %d", response.Total)
	return response, nil
}

func (o *order) CreateCartItem(ctx context.Context, request *opbv1.CartItemRequest) (*opbv1.ShopCartInfoResponse, error) {
	log.Infof("Calling CreateCartItem gRPC for user: %d, goods: %d", request.UserId, request.GoodsId)
	response, err := o.oc.CreateCartItem(ctx, request)
	if err != nil {
		log.Errorf("CreateCartItem gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("CreateCartItem gRPC call successful, cart item ID: %d", response.Id)
	return response, nil
}

func (o *order) UpdateCartItem(ctx context.Context, request *opbv1.CartItemRequest) (*opbv1.ShopCartInfoResponse, error) {
	log.Infof("Calling UpdateCartItem gRPC for cart item: %d", request.Id)
	_, err := o.oc.UpdateCartItem(ctx, request)
	if err != nil {
		log.Errorf("UpdateCartItem gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("UpdateCartItem gRPC call successful")
	return &opbv1.ShopCartInfoResponse{}, nil
}

func (o *order) DeleteCartItem(ctx context.Context, request *opbv1.CartItemRequest) (*opbv1.ShopCartInfoResponse, error) {
	log.Infof("Calling DeleteCartItem gRPC for cart item: %d", request.Id)
	_, err := o.oc.DeleteCartItem(ctx, request)
	if err != nil {
		log.Errorf("DeleteCartItem gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("DeleteCartItem gRPC call successful")
	return &opbv1.ShopCartInfoResponse{}, nil
}

var _ data.OrderData = &order{}