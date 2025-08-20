package service

import (
	"context"
	proto2 "emshop/api/goods/v1"
	proto "emshop/api/inventory/v1"
	proto3 "emshop/api/order/v1"
	"emshop/internal/app/order/srv/data/v1/mysql"
	"emshop/internal/app/order/srv/domain/do"
	"emshop/internal/app/order/srv/domain/dto"
	"emshop/internal/app/pkg/code"
	"emshop/internal/app/pkg/options"
	v1 "emshop/pkg/common/meta/v1"
	"emshop/pkg/errors"
	"emshop/pkg/log"

	"github.com/dtm-labs/client/dtmgrpc"
)

type OrderSrv interface {
	Get(ctx context.Context, orderSn string) (*dto.OrderDTO, error)
	List(ctx context.Context, userID uint64, meta v1.ListMeta, orderby []string) (*dto.OrderDTOList, error)
	Submit(ctx context.Context, order *dto.OrderDTO) error
	Create(ctx context.Context, order *dto.OrderDTO) error
	CreateCom(ctx context.Context, order *dto.OrderDTO) error //这是create的补偿
	Update(ctx context.Context, order *dto.OrderDTO) error
	
	// Cart operations
	CartItemList(ctx context.Context, userID uint64, meta v1.ListMeta) (*dto.ShopCartDTOList, error)
	CreateCartItem(ctx context.Context, cartItem *dto.ShopCartDTO) error
	UpdateCartItem(ctx context.Context, cartItem *dto.ShopCartDTO) error
	DeleteCartItem(ctx context.Context, userID, goodsID uint64) error
}

type orderService struct {
	data    mysql.DataFactory
	dtmOpts *options.DtmOptions
}

// CreateCom 是Create的补偿方法， 主要是回滚订单的创建
func (os *orderService) CreateCom(ctx context.Context, order *dto.OrderDTO) error {
	/*
		1. 删除orderinfo表
		2. 删除ordergoods表
		3. 删除order找到对应的购物车条目，删除购物车条目
	*/
	//其实不用回滚
	//你应该先查询订单是否已经存在，如果已经存在删除相关记录即可， 同时删除购物车记录
	return nil
}

// Create 创建订单
func (os *orderService) Create(ctx context.Context, order *dto.OrderDTO) error {
	/*
		1. 生成orderinfo表
		2. 生成ordergoods表
		3. 根据order找到对应的购物车条目，删除购物车条目
	*/

	var goodsids []int32
	for _, value := range order.OrderGoods {
		goodsids = append(goodsids, value.Goods)
	}

	goods, err := os.data.Goods().BatchGetGoods(context.Background(), &proto2.BatchGoodsIdInfo{Id: goodsids})
	if err != nil {
		log.Errorf("批量获取商品信息失败，goodids: %v, err:%v", goodsids, err)
		return err
	}
	if len(goods.Data) != len(goodsids) {
		log.Errorf("批量获取商品信息失败，goodids: %v, 返回值：%v, err:%v", goodsids, goods.Data, err)
		return errors.WithCode(code.ErrGoodsNotFound, "商品不存在或者部分不存在")
	}
	var goodsMap = make(map[int32]*proto2.GoodsInfoResponse)
	for _, value := range goods.Data {
		goodsMap[value.Id] = value
	}

	var orderAmount float32
	for _, value := range order.OrderGoods {
		orderAmount += goodsMap[value.Goods].ShopPrice * float32(value.Nums)
		value.GoodsName = goodsMap[value.Goods].Name
		value.GoodsPrice = goodsMap[value.Goods].ShopPrice
		value.GoodsImage = goodsMap[value.Goods].GoodsFrontImage
	}

	txn := os.data.Begin()
	defer func() {
		if err := recover(); err != nil {
			_ = txn.Rollback()
			log.Error("新建订单事务进行中出现异常，回滚")
			return
		}
	}()

	err = os.data.Orders().Create(ctx, txn, &order.OrderInfoDO)
	if err != nil {
		txn.Rollback()
		log.Errorf("创建订单失败，err:%v", err)
		return err //这个不是abort 也就是说会不停的重试
	}

	err = os.data.ShoppingCarts().DeleteByGoodsIDs(ctx, txn, uint64(order.User), goodsids)
	if err != nil {
		txn.Rollback()
		log.Errorf("删除购物车失败，goodids:%v, err:%v", goodsids, err)
		return err
	}

	txn.Commit()
	//这里有逻辑
	return nil
}

func (os *orderService) Get(ctx context.Context, orderSn string) (*dto.OrderDTO, error) {
	order, err := os.data.Orders().Get(ctx, orderSn)
	if err != nil {
		return nil, err
	}
	return &dto.OrderDTO{*order}, nil
}

func (os *orderService) List(ctx context.Context, userID uint64, meta v1.ListMeta, orderby []string) (*dto.OrderDTOList, error) {
	orders, err := os.data.Orders().List(ctx, userID, meta, orderby)
	if err != nil {
		return nil, err
	}
	var ret dto.OrderDTOList
	ret.TotalCount = orders.TotalCount
	for _, value := range orders.Items {
		ret.Items = append(ret.Items, &dto.OrderDTO{
			*value,
		})
	}
	return &ret, nil
}

// Submit 提交订单， 这里是基于可靠消息最终一致性的思想， saga事务来解决订单生成的问题
func (os *orderService) Submit(ctx context.Context, order *dto.OrderDTO) error {
	//先从购物车中获取商品信息
	list, err := os.data.ShoppingCarts().List(ctx, uint64(order.User), true, v1.ListMeta{}, []string{})
	if err != nil {
		log.Errorf("获取购物车信息失败，err:%v", err)
		return err
	}

	if len(list.Items) == 0 {
		log.Errorf("购物车中没有商品，无法下单")
		return errors.WithCode(code.ErrNoGoodsSelect, "没有选择商品")
	}

	var orderGoods []*do.OrderGoods
	var orderItems []*proto3.OrderItemResponse
	for _, value := range list.Items {
		orderGoods = append(orderGoods, &do.OrderGoods{
			Goods: value.Goods,
			Nums:  value.Nums,
		})

		orderItems = append(orderItems, &proto3.OrderItemResponse{
			GoodsId: value.Goods,
			Nums:    value.Nums,
		})
	}
	order.OrderGoods = orderGoods

	//基于可靠消息最终一致性的思想， saga事务来解决订单生成的问题
	var goodsInfo []*proto.GoodsInvInfo
	for _, value := range order.OrderGoods {
		goodsInfo = append(goodsInfo, &proto.GoodsInvInfo{
			GoodsId: value.Goods,
			Num:     value.Nums,
		})
	}
	req := &proto.SellInfo{
		GoodsInfo: goodsInfo,
		OrderSn:   order.OrderSn,
	}
	oReq := &proto3.OrderRequest{
		OrderSn:    &order.OrderSn,
		UserId:     order.User,
		Address:    &order.Address,
		Name:       &order.SignerName,
		Mobile:     &order.SingerMobile,
		Post:       &order.Post,
		OrderItems: orderItems,
	}

	// 注意：这里的qsBusi和gBusi是服务的地址， 需要根据实际情况修改,此处直接写死了consul的地址
	qsBusi := "discovery:///emshop-inventory-srv"
	gBusi := "discovery:///emshop-order-srv"
	// saga事务分为正向和补偿两个阶段
	// 正向阶段： Sell -> CreateOrder
	// 补偿阶段： Reback -> CreateOrderCom
	// 通过 DTM 的 Saga 模式，串联库存扣减和订单创建两个服务，保证跨服务的数据一致性。
	saga := dtmgrpc.NewSagaGrpc(os.dtmOpts.GrpcServer, order.OrderSn).
		Add(qsBusi+"/Inventory/Sell", qsBusi+"/Inventory/Reback", req).
		Add(gBusi+"/Order/CreateOrder", gBusi+"/Order/CreateOrderCom", oReq)
	saga.WaitResult = true
	err = saga.Submit()
	//通过OrderSn查询一下， 当前的状态如何状态一直值Submitted那么就你一直不要给前端返回， 如果是failed那么你提示给前端说下单失败，重新下单
	return err
}

func (os *orderService) Update(ctx context.Context, order *dto.OrderDTO) error {
	return os.data.Orders().Update(ctx, nil, &order.OrderInfoDO)
}

// Cart operations implementation
func (os *orderService) CartItemList(ctx context.Context, userID uint64, meta v1.ListMeta) (*dto.ShopCartDTOList, error) {
	shopCartDOList, err := os.data.ShoppingCarts().List(ctx, userID, false, meta, []string{})
	if err != nil {
		return nil, err
	}
	
	result := &dto.ShopCartDTOList{
		TotalCount: shopCartDOList.TotalCount,
		Items:      make([]*dto.ShopCartDTO, len(shopCartDOList.Items)),
	}
	
	for i, item := range shopCartDOList.Items {
		result.Items[i] = &dto.ShopCartDTO{
			ShoppingCartDO: *item,
		}
	}
	
	return result, nil
}

func (os *orderService) CreateCartItem(ctx context.Context, cartItem *dto.ShopCartDTO) error {
	// Check if the cart item already exists
	existingItem, err := os.data.ShoppingCarts().Get(ctx, uint64(cartItem.User), uint64(cartItem.Goods))
	if err == nil {
		// Item exists, update the quantity
		existingItem.Nums += cartItem.Nums
		return os.data.ShoppingCarts().UpdateNum(ctx, existingItem)
	}
	
	// Item doesn't exist, create new
	return os.data.ShoppingCarts().Create(ctx, &cartItem.ShoppingCartDO)
}

func (os *orderService) UpdateCartItem(ctx context.Context, cartItem *dto.ShopCartDTO) error {
	return os.data.ShoppingCarts().UpdateNum(ctx, &cartItem.ShoppingCartDO)
}

func (os *orderService) DeleteCartItem(ctx context.Context, userID, goodsID uint64) error {
	cartItem, err := os.data.ShoppingCarts().Get(ctx, userID, goodsID)
	if err != nil {
		return err
	}
	return os.data.ShoppingCarts().Delete(ctx, uint64(cartItem.ID))
}

func newOrderService(sv *service) *orderService {
	return &orderService{
		data:    sv.data,
		dtmOpts: sv.dtmopts,
	}
}

var _ OrderSrv = &orderService{}
