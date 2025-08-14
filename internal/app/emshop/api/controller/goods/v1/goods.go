package goods

import (
	"strconv"
	"github.com/gin-gonic/gin"
	"emshop/gin-micro/server/rest-server"
	proto "emshop/api/goods/v1"
	"emshop/internal/app/emshop/api/domain/request"
	"emshop/internal/app/emshop/api/service"
	v1 "emshop/internal/app/emshop/api/service/goods/v1"
	gin2 "emshop/internal/app/pkg/translator/gin"
	"emshop/pkg/common/core"
	"emshop/gin-micro/code"
	"emshop/pkg/errors"
	"emshop/pkg/log"
)

type goodsController struct {
	trans    restserver.I18nTranslator
	srv      service.ServiceFactory
	goodssrv v1.GoodsSrv
}

func NewGoodsController(srv service.ServiceFactory, trans restserver.I18nTranslator) *goodsController {
	return &goodsController{
		srv:   srv,
		trans: trans,
	}
}

func (gc *goodsController) List(ctx *gin.Context) {
	log.Info("goods list function called ...")

	var r request.GoodsFilter

	if err := ctx.ShouldBindQuery(&r); err != nil {
		gin2.HandleValidatorError(ctx, err, gc.trans)
		return
	}

	gfr := proto.GoodsFilterRequest{
		IsNew:       r.IsNew,
		IsHot:       r.IsHot,
		PriceMax:    r.PriceMax,
		PriceMin:    r.PriceMin,
		TopCategory: r.TopCategory,
		Brand:       r.Brand,
		KeyWords:    r.KeyWords,
		Pages:       r.Pages,
		PagePerNums: r.PagePerNums,
	}

	goodsDTOList, err := gc.srv.Goods().List(ctx, &gfr)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	reMap := map[string]interface{}{
		"total": goodsDTOList.Total,
	}
	goodsList := make([]interface{}, 0)
	for _, value := range goodsDTOList.Data {
		goodsList = append(goodsList, map[string]interface{}{
			"id":          value.Id,
			"name":        value.Name,
			"goods_brief": value.GoodsBrief,
			"desc":        value.GoodsDesc,
			"ship_free":   value.ShipFree,
			"images":      value.Images,
			"desc_images": value.DescImages,
			"front_image": value.GoodsFrontImage,
			"shop_price":  value.ShopPrice,
			"category": map[string]interface{}{
				"id":   value.Category.Id,
				"name": value.Category.Name,
			},
			"brand": map[string]interface{}{
				"id":   value.Brand.Id,
				"name": value.Brand.Name,
				"logo": value.Brand.Logo,
			},
			"is_hot":  value.IsHot,
			"is_new":  value.IsNew,
			"on_sale": value.OnSale,
		})
	}
	reMap["data"] = goodsList

	core.WriteResponse(ctx, nil, reMap)
}

func (gc *goodsController) New(ctx *gin.Context) {
	log.Info("goods new function called ...")

	var r request.CreateGoods

	if err := ctx.ShouldBindJSON(&r); err != nil {
		gin2.HandleValidatorError(ctx, err, gc.trans)
		return
	}

	createGoodsInfo := proto.CreateGoodsInfo{
		Name:            r.Name,
		GoodsSn:         r.GoodsSn,
		Stocks:          r.Stocks,
		MarketPrice:     r.MarketPrice,
		ShopPrice:       r.ShopPrice,
		GoodsBrief:      r.GoodsBrief,
		GoodsDesc:       r.GoodsDesc,
		ShipFree:        r.ShipFree,
		Images:          r.Images,
		DescImages:      r.DescImages,
		GoodsFrontImage: r.GoodsFrontImage,
		IsNew:           r.IsNew,
		IsHot:           r.IsHot,
		OnSale:          r.OnSale,
		CategoryId:      r.CategoryId,
		BrandId:         r.BrandId,
	}

	goodsDTO, err := gc.srv.Goods().Create(ctx, &createGoodsInfo)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	response := map[string]interface{}{
		"id":          goodsDTO.Id,
		"name":        goodsDTO.Name,
		"goods_brief": goodsDTO.GoodsBrief,
		"desc":        goodsDTO.GoodsDesc,
		"ship_free":   goodsDTO.ShipFree,
		"images":      goodsDTO.Images,
		"desc_images": goodsDTO.DescImages,
		"front_image": goodsDTO.GoodsFrontImage,
		"shop_price":  goodsDTO.ShopPrice,
		"category": map[string]interface{}{
			"id":   goodsDTO.Category.Id,
			"name": goodsDTO.Category.Name,
		},
		"brand": map[string]interface{}{
			"id":   goodsDTO.Brand.Id,
			"name": goodsDTO.Brand.Name,
			"logo": goodsDTO.Brand.Logo,
		},
		"is_hot":  goodsDTO.IsHot,
		"is_new":  goodsDTO.IsNew,
		"on_sale": goodsDTO.OnSale,
	}

	core.WriteResponse(ctx, nil, response)
}

func (gc *goodsController) Sync(ctx *gin.Context) {
	log.Info("goods sync function called ...")

	var r request.SyncData

	if err := ctx.ShouldBindJSON(&r); err != nil {
		gin2.HandleValidatorError(ctx, err, gc.trans)
		return
	}

	syncRequest := proto.SyncDataRequest{
		ForceSync: r.ForceSync, // 是否强制全量同步
		GoodsIds:  r.GoodsIds,	// 同步的商品id列表
	}

	syncResponse, err := gc.srv.Goods().SyncData(ctx, &syncRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	response := map[string]interface{}{
		"success":      syncResponse.Success,
		"message":      syncResponse.Message,
		"synced_count": syncResponse.SyncedCount,
		"failed_count": syncResponse.FailedCount,
		"errors":       syncResponse.Errors,
	}

	core.WriteResponse(ctx, nil, response)
}

func (gc *goodsController) Detail(ctx *gin.Context) {
	log.Info("goods detail function called ...")

	id := ctx.Param("id")
	if id == "" {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "商品ID不能为空"), nil)
		return
	}

	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "商品ID格式不正确"), nil)
		return
	}

	goodsDetailRequest := proto.GoodInfoRequest{
		Id: int32(i),
	}

	goodsDTO, err := gc.srv.Goods().Detail(ctx, &goodsDetailRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	response := map[string]interface{}{
		"id":          goodsDTO.Id,
		"name":        goodsDTO.Name,
		"goods_brief": goodsDTO.GoodsBrief,
		"desc":        goodsDTO.GoodsDesc,
		"ship_free":   goodsDTO.ShipFree,
		"images":      goodsDTO.Images,
		"desc_images": goodsDTO.DescImages,
		"front_image": goodsDTO.GoodsFrontImage,
		"shop_price":  goodsDTO.ShopPrice,
		"category": map[string]interface{}{
			"id":   goodsDTO.Category.Id,
			"name": goodsDTO.Category.Name,
		},
		"brand": map[string]interface{}{
			"id":   goodsDTO.Brand.Id,
			"name": goodsDTO.Brand.Name,
			"logo": goodsDTO.Brand.Logo,
		},
		"is_hot":  goodsDTO.IsHot,
		"is_new":  goodsDTO.IsNew,
		"on_sale": goodsDTO.OnSale,
	}

	core.WriteResponse(ctx, nil, response)
}

func (gc *goodsController) Delete(ctx *gin.Context) {
	log.Info("goods delete function called ...")

	id := ctx.Param("id")
	if id == "" {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "商品ID不能为空"), nil)
		return
	}

	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "商品ID格式不正确"), nil)
		return
	}

	deleteGoodsInfo := proto.DeleteGoodsInfo{
		Id: int32(i),
	}

	_, err = gc.srv.Goods().Delete(ctx, &deleteGoodsInfo)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, map[string]interface{}{
		"msg": "删除成功",
	})
}

func (gc *goodsController) Update(ctx *gin.Context) {
	log.Info("goods update function called ...")

	id := ctx.Param("id")
	if id == "" {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "商品ID不能为空"), nil)
		return
	}

	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "商品ID格式不正确"), nil)
		return
	}

	var r request.UpdateGoods

	if err := ctx.ShouldBindJSON(&r); err != nil {
		gin2.HandleValidatorError(ctx, err, gc.trans)
		return
	}

	updateGoodsInfo := proto.CreateGoodsInfo{
		Id:              int32(i),
		Name:            r.Name,
		GoodsSn:         r.GoodsSn,
		Stocks:          r.Stocks,
		MarketPrice:     r.MarketPrice,
		ShopPrice:       r.ShopPrice,
		GoodsBrief:      r.GoodsBrief,
		GoodsDesc:       r.GoodsDesc,
		ShipFree:        r.ShipFree,
		Images:          r.Images,
		DescImages:      r.DescImages,
		GoodsFrontImage: r.GoodsFrontImage,
		IsNew:           r.IsNew,
		IsHot:           r.IsHot,
		OnSale:          r.OnSale,
		CategoryId:      r.CategoryId,
		BrandId:         r.BrandId,
	}

	_, err = gc.srv.Goods().Update(ctx, &updateGoodsInfo)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, map[string]interface{}{
		"msg": "更新成功",
	})
}

func (gc *goodsController) UpdateStatus(ctx *gin.Context) {
	log.Info("goods update status function called ...")

	id := ctx.Param("id")
	if id == "" {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "商品ID不能为空"), nil)
		return
	}

	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "商品ID格式不正确"), nil)
		return
	}

	var r request.UpdateGoodsStatus

	if err := ctx.ShouldBindJSON(&r); err != nil {
		gin2.HandleValidatorError(ctx, err, gc.trans)
		return
	}

	updateGoodsInfo := proto.CreateGoodsInfo{
		Id:     int32(i),
		IsHot:  *r.IsHot,
		IsNew:  *r.IsNew,
		OnSale: *r.OnSale,
	}

	_, err = gc.srv.Goods().Update(ctx, &updateGoodsInfo)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, map[string]interface{}{
		"msg": "修改成功",
	})
}

func (gc *goodsController) Stocks(ctx *gin.Context) {
	log.Info("goods stocks function called ...")

	id := ctx.Param("id")
	if id == "" {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "商品ID不能为空"), nil)
		return
	}

	// TODO: 实现商品库存查询逻辑
	// 这里需要调用库存服务或者商品服务获取库存信息
	
	core.WriteResponse(ctx, nil, map[string]interface{}{
		"stocks": 0,
		"msg":    "TODO: 商品库存功能待实现",
	})
}
