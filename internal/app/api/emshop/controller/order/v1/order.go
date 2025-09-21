package order

import (
	proto "emshop/api/order/v1"
	"emshop/gin-micro/code"
	restserver "emshop/gin-micro/server/rest-server"
	"emshop/internal/app/api/emshop/domain/dto/request"
	"emshop/internal/app/api/emshop/service"
	"emshop/internal/app/pkg/jwt"
	gin2 "emshop/internal/app/pkg/translator/gin"
	"emshop/pkg/common/core"
	"emshop/pkg/errors"
	"emshop/pkg/log"
	"strconv"

	"github.com/gin-gonic/gin"
)

type orderController struct {
	trans restserver.I18nTranslator
	srv   service.ServiceFactory
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
	userID := oc.getUserIDFromContext(ctx)
	if userID == 0 {
		core.WriteResponse(ctx, errors.WithCode(code.ErrTokenInvalid, "用户ID不存在"), nil)
		return
	}

	orderRequest := proto.OrderFilterRequest{
		UserId: int32(userID),
	}

	// 分页参数 - 设置默认值
	if r.Pages != nil {
		orderRequest.Pages = r.Pages
	} else {
		pages := int32(1)
		orderRequest.Pages = &pages
	}

	if r.PagePerNums != nil {
		orderRequest.PagePerNums = r.PagePerNums
	} else {
		pagePerNums := int32(10)
		orderRequest.PagePerNums = &pagePerNums
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
			"id":      order.Id,
			"orderSn": order.OrderSn,
			"status":  order.Status,
			"payType": order.PayType,
			"total":   order.Total,
			"address": order.Address,
			"name":    order.Name,
			"mobile":  order.Mobile,
			"addTime": order.AddTime,
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
	userID := oc.getUserIDFromContext(ctx)
	if userID == 0 {
		core.WriteResponse(ctx, errors.WithCode(code.ErrTokenInvalid, "用户ID不存在"), nil)
		return
	}

	orderRequest := proto.OrderRequest{
		UserId:  int32(userID),
		Address: &r.Address,
		Name:    &r.Name,
		Mobile:  &r.Mobile,
		Post:    &r.Post,
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
	userID := oc.getUserIDFromContext(ctx)
	if userID == 0 {
		core.WriteResponse(ctx, errors.WithCode(code.ErrTokenInvalid, "用户ID不存在"), nil)
		return
	}

	orderRequest := proto.OrderRequest{
		Id:     int32(i),
		UserId: int32(userID),
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
			"id":         item.Id,
			"goodsId":    item.GoodsId,
			"goodsName":  item.GoodsName,
			"goodsImage": item.GoodsImage,
			"goodsPrice": item.GoodsPrice,
			"nums":       item.Nums,
		})
	}

	response := map[string]interface{}{
		"id":      orderDetailResponse.OrderInfo.Id,
		"orderSn": orderDetailResponse.OrderInfo.OrderSn,
		"status":  orderDetailResponse.OrderInfo.Status,
		"payType": orderDetailResponse.OrderInfo.PayType,
		"total":   orderDetailResponse.OrderInfo.Total,
		"address": orderDetailResponse.OrderInfo.Address,
		"name":    orderDetailResponse.OrderInfo.Name,
		"mobile":  orderDetailResponse.OrderInfo.Mobile,
		"addTime": orderDetailResponse.OrderInfo.AddTime,
		"goods":   goodsList,
	}

	core.WriteResponse(ctx, nil, response)
}

// ==================== 购物车管理 ====================

func (oc *orderController) CartList(ctx *gin.Context) {
	log.Info("cart list function called ...")

	// 从JWT中获取用户ID
	userID := oc.getUserIDFromContext(ctx)
	if userID == 0 {
		core.WriteResponse(ctx, errors.WithCode(code.ErrTokenInvalid, "用户ID不存在"), nil)
		return
	}

	userRequest := proto.UserInfo{
		Id: int32(userID),
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
			"id":      item.Id,
			"goodsId": item.GoodsId,
			"nums":    item.Nums,
			"checked": item.Checked,
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
	userID := oc.getUserIDFromContext(ctx)
	if userID == 0 {
		core.WriteResponse(ctx, errors.WithCode(code.ErrTokenInvalid, "用户ID不存在"), nil)
		return
	}

	checked := true
	cartRequest := proto.CartItemRequest{
		UserId:  int32(userID),
		GoodsId: r.GoodsId,
		Nums:    &r.Nums,
		Checked: &checked,
	}

	cartResponse, err := oc.srv.Order().CreateCartItem(ctx, &cartRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	response := map[string]interface{}{
		"id":      cartResponse.Id,
		"goodsId": cartResponse.GoodsId,
		"nums":    cartResponse.Nums,
		"checked": cartResponse.Checked,
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
	userID := oc.getUserIDFromContext(ctx)
	if userID == 0 {
		core.WriteResponse(ctx, errors.WithCode(code.ErrTokenInvalid, "用户ID不存在"), nil)
		return
	}

	cartRequest := proto.CartItemRequest{
		Id:     int32(i),
		UserId: int32(userID),
		Nums:   &r.Nums,
	}
	if r.Checked != nil {
		cartRequest.Checked = r.Checked
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
	userID := oc.getUserIDFromContext(ctx)
	if userID == 0 {
		core.WriteResponse(ctx, errors.WithCode(code.ErrTokenInvalid, "用户ID不存在"), nil)
		return
	}

	// 先获取购物车列表，找到对应的商品ID
	userRequest := proto.UserInfo{
		Id: int32(userID),
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
		UserId:  int32(userID),
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

// getUserIDFromContext 从上下文获取用户ID
func (oc *orderController) getUserIDFromContext(ctx *gin.Context) int64 {
    // 从中间件设置的上下文键获取用户ID
    if v, ok := ctx.Get(jwt.KeyUserID); ok {
        if id, ok := v.(int); ok {
            return int64(id)
        }
        if id64, ok := v.(int64); ok {
            return id64
        }
    }
    return 0
}
