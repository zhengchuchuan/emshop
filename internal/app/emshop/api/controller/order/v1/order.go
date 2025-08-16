package order

import (
	"strconv"
	"github.com/gin-gonic/gin"
	"emshop/gin-micro/server/rest-server"
	"emshop/gin-micro/server/rest-server/middlewares"
	proto "emshop/api/order/v1"
	"emshop/internal/app/emshop/api/domain/request"
	"emshop/internal/app/emshop/api/service"
	v1 "emshop/internal/app/emshop/api/service/order/v1"
	gin2 "emshop/internal/app/pkg/translator/gin"
	"emshop/pkg/common/core"
	"emshop/gin-micro/code"
	"emshop/pkg/errors"
	"emshop/pkg/log"
)

type orderController struct {
	trans    restserver.I18nTranslator
	srv      service.ServiceFactory
	ordersrv v1.OrderSrv
}

func NewOrderController(srv service.ServiceFactory, trans restserver.I18nTranslator) *orderController {
	return &orderController{
		srv:   srv,
		trans: trans,
	}
}

// ==================== 订单管理 ====================

func (oc *orderController) OrderList(ctx *gin.Context) {
	log.Info("order list function called ...")
	
	var r request.OrderFilter
	
	if err := ctx.ShouldBindQuery(&r); err != nil {
		gin2.HandleValidatorError(ctx, err, oc.trans)
		return
	}
	
	// 从JWT中获取用户ID
	userId, exists := ctx.Get(middlewares.KeyUserID)
	if !exists {
		core.WriteResponse(ctx, errors.WithCode(code.ErrTokenInvalid, "用户ID不存在"), nil)
		return
	}
	
	orderRequest := proto.OrderFilterRequest{
		UserId:      int32(userId.(int)),
		Pages:       r.Pages,
		PagePerNums: r.PagePerNums,
	}
	
	ordersResponse, err := oc.srv.Order().OrderList(ctx, &orderRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	
	reMap := map[string]interface{}{
		"total": ordersResponse.Total,
	}
	
	ordersList := make([]interface{}, 0)
	for _, order := range ordersResponse.Data {
		ordersList = append(ordersList, map[string]interface{}{
			"id":       order.Id,
			"order_sn": order.OrderSn,
			"status":   order.Status,
			"pay_type": order.PayType,
			"total":    order.Total,
			"address":  order.Address,
			"name":     order.Name,
			"mobile":   order.Mobile,
			"add_time": order.AddTime,
		})
	}
	reMap["data"] = ordersList
	
	core.WriteResponse(ctx, nil, reMap)
}

func (oc *orderController) CreateOrder(ctx *gin.Context) {
	log.Info("create order function called ...")
	
	var r request.CreateOrder
	
	if err := ctx.ShouldBindJSON(&r); err != nil {
		gin2.HandleValidatorError(ctx, err, oc.trans)
		return
	}
	
	// 从JWT中获取用户ID
	userId, exists := ctx.Get(middlewares.KeyUserID)
	if !exists {
		core.WriteResponse(ctx, errors.WithCode(code.ErrTokenInvalid, "用户ID不存在"), nil)
		return
	}
	
	orderRequest := proto.OrderRequest{
		UserId:  int32(userId.(int)),
		Address: r.Address,
		Name:    r.Name,
		Mobile:  r.Mobile,
		Post:    r.Post,
	}
	
	_, err := oc.srv.Order().CreateOrder(ctx, &orderRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	
	core.WriteResponse(ctx, nil, map[string]interface{}{
		"msg": "订单创建成功",
	})
}

func (oc *orderController) OrderDetail(ctx *gin.Context) {
	log.Info("order detail function called ...")
	
	id := ctx.Param("id")
	if id == "" {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "订单ID不能为空"), nil)
		return
	}
	
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "订单ID格式不正确"), nil)
		return
	}
	
	// 从JWT中获取用户ID
	userId, exists := ctx.Get(middlewares.KeyUserID)
	if !exists {
		core.WriteResponse(ctx, errors.WithCode(code.ErrTokenInvalid, "用户ID不存在"), nil)
		return
	}
	
	orderRequest := proto.OrderRequest{
		Id:     int32(i),
		UserId: int32(userId.(int)),
	}
	
	orderDetailResponse, err := oc.srv.Order().OrderDetail(ctx, &orderRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	
	// 构建订单商品列表
	goodsList := make([]interface{}, 0)
	for _, item := range orderDetailResponse.Goods {
		goodsList = append(goodsList, map[string]interface{}{
			"id":          item.Id,
			"goods_id":    item.GoodsId,
			"goods_name":  item.GoodsName,
			"goods_image": item.GoodsImage,
			"goods_price": item.GoodsPrice,
			"nums":        item.Nums,
		})
	}
	
	response := map[string]interface{}{
		"id":       orderDetailResponse.OrderInfo.Id,
		"order_sn": orderDetailResponse.OrderInfo.OrderSn,
		"status":   orderDetailResponse.OrderInfo.Status,
		"pay_type": orderDetailResponse.OrderInfo.PayType,
		"total":    orderDetailResponse.OrderInfo.Total,
		"address":  orderDetailResponse.OrderInfo.Address,
		"name":     orderDetailResponse.OrderInfo.Name,
		"mobile":   orderDetailResponse.OrderInfo.Mobile,
		"add_time": orderDetailResponse.OrderInfo.AddTime,
		"goods":    goodsList,
	}
	
	core.WriteResponse(ctx, nil, response)
}

// ==================== 购物车管理 ====================

func (oc *orderController) CartList(ctx *gin.Context) {
	log.Info("cart list function called ...")
	
	// 从JWT中获取用户ID
	userId, exists := ctx.Get(middlewares.KeyUserID)
	if !exists {
		core.WriteResponse(ctx, errors.WithCode(code.ErrTokenInvalid, "用户ID不存在"), nil)
		return
	}
	
	userRequest := proto.UserInfo{
		Id: int32(userId.(int)),
	}
	
	cartResponse, err := oc.srv.Order().CartItemList(ctx, &userRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	
	reMap := map[string]interface{}{
		"total": cartResponse.Total,
	}
	
	cartList := make([]interface{}, 0)
	for _, item := range cartResponse.Data {
		cartList = append(cartList, map[string]interface{}{
			"id":       item.Id,
			"goods_id": item.GoodsId,
			"nums":     item.Nums,
			"checked":  item.Checked,
		})
	}
	reMap["data"] = cartList
	
	core.WriteResponse(ctx, nil, reMap)
}

func (oc *orderController) AddToCart(ctx *gin.Context) {
	log.Info("add to cart function called ...")
	
	var r request.AddToCart
	
	if err := ctx.ShouldBindJSON(&r); err != nil {
		gin2.HandleValidatorError(ctx, err, oc.trans)
		return
	}
	
	// 从JWT中获取用户ID
	userId, exists := ctx.Get(middlewares.KeyUserID)
	if !exists {
		core.WriteResponse(ctx, errors.WithCode(code.ErrTokenInvalid, "用户ID不存在"), nil)
		return
	}
	
	cartRequest := proto.CartItemRequest{
		UserId:  int32(userId.(int)),
		GoodsId: r.GoodsId,
		Nums:    r.Nums,
		Checked: true,
	}
	
	cartResponse, err := oc.srv.Order().CreateCartItem(ctx, &cartRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	
	response := map[string]interface{}{
		"id":       cartResponse.Id,
		"goods_id": cartResponse.GoodsId,
		"nums":     cartResponse.Nums,
		"checked":  cartResponse.Checked,
	}
	
	core.WriteResponse(ctx, nil, response)
}

func (oc *orderController) UpdateCartItem(ctx *gin.Context) {
	log.Info("update cart item function called ...")
	
	id := ctx.Param("id")
	if id == "" {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "购物车ID不能为空"), nil)
		return
	}
	
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "购物车ID格式不正确"), nil)
		return
	}
	
	var r request.UpdateCartItem
	
	if err := ctx.ShouldBindJSON(&r); err != nil {
		gin2.HandleValidatorError(ctx, err, oc.trans)
		return
	}
	
	// 从JWT中获取用户ID
	userId, exists := ctx.Get(middlewares.KeyUserID)
	if !exists {
		core.WriteResponse(ctx, errors.WithCode(code.ErrTokenInvalid, "用户ID不存在"), nil)
		return
	}
	
	cartRequest := proto.CartItemRequest{
		Id:     int32(i),
		UserId: int32(userId.(int)),
		Nums:   r.Nums,
	}
	if r.Checked != nil {
		cartRequest.Checked = *r.Checked
	}
	
	err = oc.srv.Order().UpdateCartItem(ctx, &cartRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	
	core.WriteResponse(ctx, nil, map[string]interface{}{
		"msg": "更新成功",
	})
}

func (oc *orderController) DeleteCartItem(ctx *gin.Context) {
	log.Info("delete cart item function called ...")
	
	id := ctx.Param("id")
	if id == "" {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "购物车ID不能为空"), nil)
		return
	}
	
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "购物车ID格式不正确"), nil)
		return
	}
	
	// 从JWT中获取用户ID
	userId, exists := ctx.Get(middlewares.KeyUserID)
	if !exists {
		core.WriteResponse(ctx, errors.WithCode(code.ErrTokenInvalid, "用户ID不存在"), nil)
		return
	}
	
	// 先获取购物车列表，找到对应的商品ID
	userRequest := proto.UserInfo{
		Id: int32(userId.(int)),
	}
	
	cartResponse, err := oc.srv.Order().CartItemList(ctx, &userRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	
	// 找到对应的购物车条目
	var goodsId int32
	found := false
	for _, item := range cartResponse.Data {
		if item.Id == int32(i) {
			goodsId = item.GoodsId
			found = true
			break
		}
	}
	
	if !found {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "购物车条目不存在"), nil)
		return
	}
	
	cartRequest := proto.CartItemRequest{
		UserId:  int32(userId.(int)),
		GoodsId: goodsId,
	}
	
	err = oc.srv.Order().DeleteCartItem(ctx, &cartRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	
	core.WriteResponse(ctx, nil, map[string]interface{}{
		"msg": "删除成功",
	})
}