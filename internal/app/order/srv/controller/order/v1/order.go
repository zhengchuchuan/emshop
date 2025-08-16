package order

import (
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
	pb "emshop/api/order/v1"
	"emshop/internal/app/order/srv/domain/do"
	"emshop/internal/app/order/srv/domain/dto"
	"emshop/internal/app/order/srv/service/v1"
	v1 "emshop/pkg/common/meta/v1"
	"emshop/pkg/log"
)

type orderServer struct {
	pb.UnimplementedOrderServer

	srv service.ServiceFactory
}

func NewOrderServer(srv service.ServiceFactory) *orderServer {
	return &orderServer{srv: srv}
}

func (os *orderServer) CartItemList(ctx context.Context, info *pb.UserInfo) (*pb.CartItemListResponse, error) {
	cartList, err := os.srv.Orders().CartItemList(ctx, uint64(info.Id), v1.ListMeta{})
	if err != nil {
		return nil, err
	}
	
	response := &pb.CartItemListResponse{
		Total: int32(cartList.TotalCount),
		Data:  make([]*pb.ShopCartInfoResponse, len(cartList.Items)),
	}
	
	for i, item := range cartList.Items {
		response.Data[i] = &pb.ShopCartInfoResponse{
			Id:      int32(item.ID),
			UserId:  item.User,
			GoodsId: item.Goods,
			Nums:    item.Nums,
			Checked: item.Checked,
		}
	}
	
	return response, nil
}

func (os *orderServer) CreateCartItem(ctx context.Context, request *pb.CartItemRequest) (*pb.ShopCartInfoResponse, error) {
	cartItem := &dto.ShopCartDTO{
		ShoppingCartDO: do.ShoppingCartDO{
			User:    request.UserId,
			Goods:   request.GoodsId,
			Nums:    request.Nums,
			Checked: request.Checked,
		},
	}
	
	err := os.srv.Orders().CreateCartItem(ctx, cartItem)
	if err != nil {
		return nil, err
	}
	
	return &pb.ShopCartInfoResponse{
		Id:      int32(cartItem.ID),
		UserId:  cartItem.User,
		GoodsId: cartItem.Goods,
		Nums:    cartItem.Nums,
		Checked: cartItem.Checked,
	}, nil
}

func (os *orderServer) UpdateCartItem(ctx context.Context, request *pb.CartItemRequest) (*emptypb.Empty, error) {
	cartItem := &dto.ShopCartDTO{
		ShoppingCartDO: do.ShoppingCartDO{
			User:    request.UserId,
			Goods:   request.GoodsId,
			Nums:    request.Nums,
			Checked: request.Checked,
		},
	}
	
	err := os.srv.Orders().UpdateCartItem(ctx, cartItem)
	if err != nil {
		return nil, err
	}
	
	return &emptypb.Empty{}, nil
}

func (os *orderServer) DeleteCartItem(ctx context.Context, request *pb.CartItemRequest) (*emptypb.Empty, error) {
	err := os.srv.Orders().DeleteCartItem(ctx, uint64(request.UserId), uint64(request.GoodsId))
	if err != nil {
		return nil, err
	}
	
	return &emptypb.Empty{}, nil
}

// 这个是给分布式事务saga调用的，目前没为api提供的目的
func (os *orderServer) CreateOrder(ctx context.Context, request *pb.OrderRequest) (*emptypb.Empty, error) {
	orderGoods := make([]*do.OrderGoods, len(request.OrderItems))
	for i, item := range request.OrderItems {
		orderGoods[i] = &do.OrderGoods{
			Goods: item.GoodsId,
			Nums:  item.Nums,
		}
	}

	err := os.srv.Orders().Create(ctx, &dto.OrderDTO{
		OrderInfoDO: do.OrderInfoDO{
			User:         request.UserId,
			Address:      request.Address,
			SignerName:   request.Name,
			SingerMobile: request.Mobile,
			Post:         request.Post,
			OrderSn:      request.OrderSn,
			OrderGoods:   orderGoods,
		},
	})
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (os *orderServer) CreateOrderCom(ctx context.Context, request *pb.OrderRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

//// 订单号的生成， 订单号-雪花算法，目前的订单号生成算法有问题： 不是递增
//func generateOrderSn(userId int32) string {
//	//订单号的生成规则
//	/*
//		年月日时分秒+用户id+2位随机数
//	*/
//	now := time.Now()
//	rand.Seed(time.Now().UnixNano())
//	orderSn := fmt.Sprintf("%d%d%d%d%d%d%d%d",
//		now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Nanosecond(),
//		userId, rand.Intn(90)+10,
//	)
//	return orderSn
//}

/*
订单提交的时候应该是先生成订单号
订单号会单独做一个接口，订单查询，以及一系列的关联我们应该采用order_sn，不要再去采用id去关联
*/
func (os *orderServer) SubmitOrder(ctx context.Context, request *pb.OrderRequest) (*emptypb.Empty, error) {
	//从购物车中得到选中的商品
	orderDTO := dto.OrderDTO{
		OrderInfoDO: do.OrderInfoDO{
			User:         request.UserId,
			Address:      request.Address,
			SignerName:   request.Name,
			SingerMobile: request.Mobile,
			Post:         request.Post,
			OrderSn:      request.OrderSn,
		},
	}
	err := os.srv.Orders().Submit(ctx, &orderDTO)
	if err != nil {
		log.Errorf("新建订单失败: %v", err)
		return nil, err
	}
	//另外一款解决ioc的库，wire
	return &emptypb.Empty{}, nil
}

func (os *orderServer) OrderList(ctx context.Context, request *pb.OrderFilterRequest) (*pb.OrderListResponse, error) {
	orderList, err := os.srv.Orders().List(ctx, uint64(request.UserId), v1.ListMeta{
		Page:     int(request.Pages),
		PageSize: int(request.PagePerNums),
	}, []string{})
	if err != nil {
		return nil, err
	}
	
	response := &pb.OrderListResponse{
		Total: int32(orderList.TotalCount),
		Data:  make([]*pb.OrderInfoResponse, len(orderList.Items)),
	}
	
	for i, order := range orderList.Items {
		response.Data[i] = &pb.OrderInfoResponse{
			Id:      int32(order.ID),
			UserId:  order.User,
			OrderSn: order.OrderSn,
			PayType: order.PayType,
			Status:  order.Status,
			Post:    order.Post,
			Total:   order.OrderMount,
			Address: order.Address,
			Name:    order.SignerName,
			Mobile:  order.SingerMobile,
			AddTime: order.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	
	return response, nil
}

func (os *orderServer) OrderDetail(ctx context.Context, request *pb.OrderRequest) (*pb.OrderInfoDetailResponse, error) {
	order, err := os.srv.Orders().Get(ctx, request.OrderSn)
	if err != nil {
		return nil, err
	}
	
	orderGoods := make([]*pb.OrderItemResponse, len(order.OrderGoods))
	for i, item := range order.OrderGoods {
		orderGoods[i] = &pb.OrderItemResponse{
			GoodsId: item.Goods,
			Nums:    item.Nums,
		}
	}
	
	return &pb.OrderInfoDetailResponse{
		OrderInfo: &pb.OrderInfoResponse{
			Id:      int32(order.ID),
			UserId:  order.User,
			OrderSn: order.OrderSn,
			PayType: order.PayType,
			Status:  order.Status,
			Post:    order.Post,
			Total:   order.OrderMount,
			Address: order.Address,
			Name:    order.SignerName,
			Mobile:  order.SingerMobile,
			AddTime: order.CreatedAt.Format("2006-01-02 15:04:05"),
		},
		Goods: orderGoods,
	}, nil
}

func (os *orderServer) UpdateOrderStatus(ctx context.Context, status *pb.OrderStatus) (*emptypb.Empty, error) {
	order, err := os.srv.Orders().Get(ctx, status.OrderSn)
	if err != nil {
		return nil, err
	}
	
	order.Status = status.Status
	
	err = os.srv.Orders().Update(ctx, order)
	if err != nil {
		return nil, err
	}
	
	return &emptypb.Empty{}, nil
}

var _ pb.OrderServer = &orderServer{}
