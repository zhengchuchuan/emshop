package userop

import (
	"strconv"
	"github.com/gin-gonic/gin"
	"emshop/gin-micro/server/rest-server"
	proto "emshop/api/userop/v1"
	"emshop/internal/app/emshop/api/domain/request"
	"emshop/internal/app/emshop/api/service"
	v1 "emshop/internal/app/emshop/api/service/userop/v1"
	gin2 "emshop/internal/app/pkg/translator/gin"
	"emshop/pkg/common/core"
	"emshop/gin-micro/code"
	"emshop/pkg/errors"
	"emshop/pkg/log"
)

type userOpController struct {
	trans     restserver.I18nTranslator
	srv       service.ServiceFactory
	useropsrv v1.UserOpSrv
}

func NewUserOpController(srv service.ServiceFactory, trans restserver.I18nTranslator) *userOpController {
	return &userOpController{
		srv:   srv,
		trans: trans,
	}
}

// ==================== 用户收藏管理 ====================

func (uoc *userOpController) UserFavList(ctx *gin.Context) {
	log.Info("user fav list function called ...")
	
	// 从JWT中获取用户ID
	userId, exists := ctx.Get("user_id")
	if !exists {
		core.WriteResponse(ctx, errors.WithCode(code.ErrTokenInvalid, "用户ID不存在"), nil)
		return
	}
	
	favRequest := proto.UserFavListRequest{
		UserId: int32(userId.(int)),
	}
	
	favsResponse, err := uoc.srv.UserOp().UserFavList(ctx, &favRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	
	reMap := map[string]interface{}{
		"total": favsResponse.Total,
	}
	
	favsList := make([]interface{}, 0)
	for _, fav := range favsResponse.Data {
		favsList = append(favsList, map[string]interface{}{
			"userId":  fav.UserId,
			"goodsId": fav.GoodsId,
		})
	}
	reMap["data"] = favsList
	
	core.WriteResponse(ctx, nil, reMap)
}

func (uoc *userOpController) CreateUserFav(ctx *gin.Context) {
	log.Info("create user fav function called ...")
	
	var r request.UserFav
	
	if err := ctx.ShouldBindJSON(&r); err != nil {
		gin2.HandleValidatorError(ctx, err, uoc.trans)
		return
	}
	
	// 从JWT中获取用户ID
	userId, exists := ctx.Get("user_id")
	if !exists {
		core.WriteResponse(ctx, errors.WithCode(code.ErrTokenInvalid, "用户ID不存在"), nil)
		return
	}
	
	favRequest := proto.UserFavRequest{
		UserId:  int32(userId.(int)),
		GoodsId: r.GoodsId,
	}
	
	favResponse, err := uoc.srv.UserOp().CreateUserFav(ctx, &favRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	
	response := map[string]interface{}{
		"userId":  favResponse.UserId,
		"goodsId": favResponse.GoodsId,
	}
	
	core.WriteResponse(ctx, nil, response)
}

func (uoc *userOpController) DeleteUserFav(ctx *gin.Context) {
	log.Info("delete user fav function called ...")
	
	goodsId := ctx.Param("id")
	if goodsId == "" {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "商品ID不能为空"), nil)
		return
	}
	
	i, err := strconv.ParseInt(goodsId, 10, 32)
	if err != nil {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "商品ID格式不正确"), nil)
		return
	}
	
	// 从JWT中获取用户ID
	userId, exists := ctx.Get("user_id")
	if !exists {
		core.WriteResponse(ctx, errors.WithCode(code.ErrTokenInvalid, "用户ID不存在"), nil)
		return
	}
	
	favRequest := proto.UserFavRequest{
		UserId:  int32(userId.(int)),
		GoodsId: int32(i),
	}
	
	err = uoc.srv.UserOp().DeleteUserFav(ctx, &favRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	
	core.WriteResponse(ctx, nil, map[string]interface{}{
		"msg": "取消收藏成功",
	})
}

func (uoc *userOpController) GetUserFavDetail(ctx *gin.Context) {
	log.Info("get user fav detail function called ...")
	
	goodsId := ctx.Param("id")
	if goodsId == "" {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "商品ID不能为空"), nil)
		return
	}
	
	i, err := strconv.ParseInt(goodsId, 10, 32)
	if err != nil {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "商品ID格式不正确"), nil)
		return
	}
	
	// 从JWT中获取用户ID
	userId, exists := ctx.Get("user_id")
	if !exists {
		core.WriteResponse(ctx, errors.WithCode(code.ErrTokenInvalid, "用户ID不存在"), nil)
		return
	}
	
	favRequest := proto.UserFavRequest{
		UserId:  int32(userId.(int)),
		GoodsId: int32(i),
	}
	
	favResponse, err := uoc.srv.UserOp().GetUserFavDetail(ctx, &favRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	
	response := map[string]interface{}{
		"userId":  favResponse.UserId,
		"goodsId": favResponse.GoodsId,
	}
	
	core.WriteResponse(ctx, nil, response)
}

// ==================== 用户地址管理 ====================

func (uoc *userOpController) GetAddressList(ctx *gin.Context) {
	log.Info("get address list function called ...")
	
	// 从JWT中获取用户ID
	userId, exists := ctx.Get("user_id")
	if !exists {
		core.WriteResponse(ctx, errors.WithCode(code.ErrTokenInvalid, "用户ID不存在"), nil)
		return
	}
	
	addressRequest := proto.AddressRequest{
		UserId: int32(userId.(int)),
	}
	
	addressResponse, err := uoc.srv.UserOp().GetAddressList(ctx, &addressRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	
	reMap := map[string]interface{}{
		"total": addressResponse.Total,
	}
	
	addressList := make([]interface{}, 0)
	for _, addr := range addressResponse.Data {
		addressList = append(addressList, map[string]interface{}{
			"id":           addr.Id,
			"province":     addr.Province,
			"city":         addr.City,
			"district":     addr.District,
			"address":      addr.Address,
			"signerName":  addr.SignerName,
			"signerMobile": addr.SignerMobile,
		})
	}
	reMap["data"] = addressList
	
	core.WriteResponse(ctx, nil, reMap)
}

func (uoc *userOpController) CreateAddress(ctx *gin.Context) {
	log.Info("create address function called ...")
	
	var r request.CreateAddress
	
	if err := ctx.ShouldBindJSON(&r); err != nil {
		gin2.HandleValidatorError(ctx, err, uoc.trans)
		return
	}
	
	// 从JWT中获取用户ID
	userId, exists := ctx.Get("user_id")
	if !exists {
		core.WriteResponse(ctx, errors.WithCode(code.ErrTokenInvalid, "用户ID不存在"), nil)
		return
	}
	
	addressRequest := proto.AddressRequest{
		UserId:       int32(userId.(int)),
		Province:     r.Province,
		City:         r.City,
		District:     r.District,
		Address:      r.Address,
		SignerName:   r.SignerName,
		SignerMobile: r.SignerMobile,
	}
	
	addressResponse, err := uoc.srv.UserOp().CreateAddress(ctx, &addressRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	
	response := map[string]interface{}{
		"id":           addressResponse.Id,
		"province":     addressResponse.Province,
		"city":         addressResponse.City,
		"district":     addressResponse.District,
		"address":      addressResponse.Address,
		"signerName":  addressResponse.SignerName,
		"signerMobile": addressResponse.SignerMobile,
	}
	
	core.WriteResponse(ctx, nil, response)
}

func (uoc *userOpController) UpdateAddress(ctx *gin.Context) {
	log.Info("update address function called ...")
	
	id := ctx.Param("id")
	if id == "" {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "地址ID不能为空"), nil)
		return
	}
	
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "地址ID格式不正确"), nil)
		return
	}
	
	var r request.UpdateAddress
	
	if err := ctx.ShouldBindJSON(&r); err != nil {
		gin2.HandleValidatorError(ctx, err, uoc.trans)
		return
	}
	
	// 从JWT中获取用户ID
	userId, exists := ctx.Get("user_id")
	if !exists {
		core.WriteResponse(ctx, errors.WithCode(code.ErrTokenInvalid, "用户ID不存在"), nil)
		return
	}
	
	addressRequest := proto.AddressRequest{
		Id:           int32(i),
		UserId:       int32(userId.(int)),
		Province:     r.Province,
		City:         r.City,
		District:     r.District,
		Address:      r.Address,
		SignerName:   r.SignerName,
		SignerMobile: r.SignerMobile,
	}
	
	err = uoc.srv.UserOp().UpdateAddress(ctx, &addressRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	
	core.WriteResponse(ctx, nil, map[string]interface{}{
		"msg": "更新成功",
	})
}

func (uoc *userOpController) DeleteAddress(ctx *gin.Context) {
	log.Info("delete address function called ...")
	
	id := ctx.Param("id")
	if id == "" {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "地址ID不能为空"), nil)
		return
	}
	
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "地址ID格式不正确"), nil)
		return
	}
	
	// 从JWT中获取用户ID
	userId, exists := ctx.Get("user_id")
	if !exists {
		core.WriteResponse(ctx, errors.WithCode(code.ErrTokenInvalid, "用户ID不存在"), nil)
		return
	}
	
	deleteRequest := proto.DeleteAddressRequest{
		Id:     int32(i),
		UserId: int32(userId.(int)),
	}
	
	err = uoc.srv.UserOp().DeleteAddress(ctx, &deleteRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	
	core.WriteResponse(ctx, nil, map[string]interface{}{
		"msg": "删除成功",
	})
}

// ==================== 用户留言管理 ====================

func (uoc *userOpController) MessageList(ctx *gin.Context) {
	log.Info("message list function called ...")
	
	// 从JWT中获取用户ID
	userId, exists := ctx.Get("user_id")
	if !exists {
		core.WriteResponse(ctx, errors.WithCode(code.ErrTokenInvalid, "用户ID不存在"), nil)
		return
	}
	
	messageRequest := proto.MessageRequest{
		UserId: int32(userId.(int)),
	}
	
	messageResponse, err := uoc.srv.UserOp().MessageList(ctx, &messageRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	
	reMap := map[string]interface{}{
		"total": messageResponse.Total,
	}
	
	messageList := make([]interface{}, 0)
	for _, msg := range messageResponse.Data {
		messageList = append(messageList, map[string]interface{}{
			"id":           msg.Id,
			"messageType": msg.MessageType,
			"subject":      msg.Subject,
			"message":      msg.Message,
			"file":         msg.File,
		})
	}
	reMap["data"] = messageList
	
	core.WriteResponse(ctx, nil, reMap)
}

func (uoc *userOpController) CreateMessage(ctx *gin.Context) {
	log.Info("create message function called ...")
	
	var r request.CreateMessage
	
	if err := ctx.ShouldBindJSON(&r); err != nil {
		gin2.HandleValidatorError(ctx, err, uoc.trans)
		return
	}
	
	// 从JWT中获取用户ID
	userId, exists := ctx.Get("user_id")
	if !exists {
		core.WriteResponse(ctx, errors.WithCode(code.ErrTokenInvalid, "用户ID不存在"), nil)
		return
	}
	
	messageRequest := proto.MessageRequest{
		UserId:      int32(userId.(int)),
		MessageType: r.MessageType,
		Subject:     r.Subject,
		Message:     r.Message,
		File:        r.File,
	}
	
	messageResponse, err := uoc.srv.UserOp().CreateMessage(ctx, &messageRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	
	response := map[string]interface{}{
		"id":           messageResponse.Id,
		"messageType": messageResponse.MessageType,
		"subject":      messageResponse.Subject,
		"message":      messageResponse.Message,
		"file":         messageResponse.File,
	}
	
	core.WriteResponse(ctx, nil, response)
}