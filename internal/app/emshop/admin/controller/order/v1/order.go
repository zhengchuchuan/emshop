package order

import (
	"strconv"
	"github.com/gin-gonic/gin"
	restserver "emshop/gin-micro/server/rest-server"
	proto "emshop/api/order/v1"
	"emshop/internal/app/emshop/admin/domain/request"
	"emshop/internal/app/emshop/admin/service"
	v1 "emshop/internal/app/emshop/admin/service/order/v1"
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

// AdminOrderList 管理员订单列表（支持多维度筛选）
func (oc *orderController) AdminOrderList(ctx *gin.Context) {
	log.Info("admin order list function called ...")
	
	var r request.AdminOrderFilter
	
	if err := ctx.ShouldBindQuery(&r); err != nil {
		gin2.HandleValidatorError(ctx, err, oc.trans)
		return
	}
	
	// 设置默认值
	if r.Pages <= 0 {
		r.Pages = 1
	}
	if r.PagePerNums <= 0 {
		r.PagePerNums = 10
	}
	
	orderRequest := &proto.OrderFilterRequest{
		UserId:      r.UserId,      // 如果为0，则查询所有用户的订单
		Pages:       r.Pages,
		PagePerNums: r.PagePerNums,
	}
	
	ordersResponse, err := oc.srv.Order().AdminOrderList(ctx, orderRequest)
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
			"userId":  order.UserId,
			"orderSn": order.OrderSn,
			"status":   order.Status,
			"payType": order.PayType,
			"total":    order.Total,
			"address":  order.Address,
			"name":     order.Name,
			"mobile":   order.Mobile,
			"addTime": order.AddTime,
		})
	}
	reMap["data"] = ordersList
	
	core.WriteResponse(ctx, nil, reMap)
}

// AdminOrderDetail 管理员查看订单详情
func (oc *orderController) AdminOrderDetail(ctx *gin.Context) {
	log.Info("admin order detail function called ...")
	
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
	
	orderRequest := &proto.OrderRequest{
		Id: int32(i),
	}
	
	orderDetailResponse, err := oc.srv.Order().AdminOrderDetail(ctx, orderRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	
	// 构建订单商品列表
	goodsList := make([]interface{}, 0)
	for _, item := range orderDetailResponse.Goods {
		goodsList = append(goodsList, map[string]interface{}{
			"id":          item.Id,
			"goodsId":    item.GoodsId,
			"goodsName":  item.GoodsName,
			"goodsImage": item.GoodsImage,
			"goodsPrice": item.GoodsPrice,
			"nums":        item.Nums,
		})
	}
	
	response := map[string]interface{}{
		"id":       orderDetailResponse.OrderInfo.Id,
		"userId":  orderDetailResponse.OrderInfo.UserId,
		"orderSn": orderDetailResponse.OrderInfo.OrderSn,
		"status":   orderDetailResponse.OrderInfo.Status,
		"payType": orderDetailResponse.OrderInfo.PayType,
		"total":    orderDetailResponse.OrderInfo.Total,
		"address":  orderDetailResponse.OrderInfo.Address,
		"name":     orderDetailResponse.OrderInfo.Name,
		"mobile":   orderDetailResponse.OrderInfo.Mobile,
		"addTime": orderDetailResponse.OrderInfo.AddTime,
		"goods":    goodsList,
	}
	
	core.WriteResponse(ctx, nil, response)
}

// UpdateOrderStatus 更新订单状态
func (oc *orderController) UpdateOrderStatus(ctx *gin.Context) {
	log.Info("admin update order status function called ...")
	
	id := ctx.Param("id")
	if id == "" {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "订单ID不能为空"), nil)
		return
	}
	
	var r request.UpdateOrderStatusRequest
	
	if err := ctx.ShouldBindJSON(&r); err != nil {
		gin2.HandleValidatorError(ctx, err, oc.trans)
		return
	}
	
	// 先根据ID获取订单详情以获得订单号
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "订单ID格式不正确"), nil)
		return
	}
	
	// 获取订单详情
	orderDetailRequest := &proto.OrderRequest{
		Id: int32(i),
	}
	
	orderDetail, err := oc.srv.Order().AdminOrderDetail(ctx, orderDetailRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	
	// 更新订单状态
	statusRequest := &proto.OrderStatus{
		Id:      int32(i),
		OrderSn: orderDetail.OrderInfo.OrderSn,
		Status:  r.Status,
	}
	
	err = oc.srv.Order().UpdateOrderStatus(ctx, statusRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	
	core.WriteResponse(ctx, nil, map[string]interface{}{
		"msg": "订单状态更新成功",
	})
}

// GetOrderByOrderSn 按订单号查询订单
func (oc *orderController) GetOrderByOrderSn(ctx *gin.Context) {
	log.Info("admin get order by order sn function called ...")
	
	orderSn := ctx.Param("order_sn")
	if orderSn == "" {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "订单号不能为空"), nil)
		return
	}
	
	orderDetailResponse, err := oc.srv.Order().GetOrderByOrderSn(ctx, orderSn)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	
	// 构建订单商品列表
	goodsList := make([]interface{}, 0)
	for _, item := range orderDetailResponse.Goods {
		goodsList = append(goodsList, map[string]interface{}{
			"id":          item.Id,
			"goodsId":    item.GoodsId,
			"goodsName":  item.GoodsName,
			"goodsImage": item.GoodsImage,
			"goodsPrice": item.GoodsPrice,
			"nums":        item.Nums,
		})
	}
	
	response := map[string]interface{}{
		"id":       orderDetailResponse.OrderInfo.Id,
		"userId":  orderDetailResponse.OrderInfo.UserId,
		"orderSn": orderDetailResponse.OrderInfo.OrderSn,
		"status":   orderDetailResponse.OrderInfo.Status,
		"payType": orderDetailResponse.OrderInfo.PayType,
		"total":    orderDetailResponse.OrderInfo.Total,
		"address":  orderDetailResponse.OrderInfo.Address,
		"name":     orderDetailResponse.OrderInfo.Name,
		"mobile":   orderDetailResponse.OrderInfo.Mobile,
		"addTime": orderDetailResponse.OrderInfo.AddTime,
		"goods":    goodsList,
	}
	
	core.WriteResponse(ctx, nil, response)
}

// GetOrdersByUserId 按用户ID查询订单列表
func (oc *orderController) GetOrdersByUserId(ctx *gin.Context) {
	log.Info("admin get orders by user id function called ...")
	
	userIdStr := ctx.Param("user_id")
	if userIdStr == "" {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "用户ID不能为空"), nil)
		return
	}
	
	userId, err := strconv.ParseInt(userIdStr, 10, 32)
	if err != nil {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "用户ID格式不正确"), nil)
		return
	}
	
	var r request.OrderSearchByUserIdRequest
	if err := ctx.ShouldBindQuery(&r); err != nil {
		gin2.HandleValidatorError(ctx, err, oc.trans)
		return
	}
	
	// 设置默认值
	if r.Pages <= 0 {
		r.Pages = 1
	}
	if r.PagePerNums <= 0 {
		r.PagePerNums = 10
	}
	
	ordersResponse, err := oc.srv.Order().GetOrdersByUserId(ctx, int32(userId), r.Pages, r.PagePerNums)
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
			"userId":  order.UserId,
			"orderSn": order.OrderSn,
			"status":   order.Status,
			"payType": order.PayType,
			"total":    order.Total,
			"address":  order.Address,
			"name":     order.Name,
			"mobile":   order.Mobile,
			"addTime": order.AddTime,
		})
	}
	reMap["data"] = ordersList
	
	core.WriteResponse(ctx, nil, reMap)
}