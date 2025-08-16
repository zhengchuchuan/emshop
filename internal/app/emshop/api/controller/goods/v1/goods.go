package goods

import (
	"encoding/json"
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
		"is_hot":      goodsDTO.IsHot,
		"is_new":      goodsDTO.IsNew,
		"on_sale":     goodsDTO.OnSale,
	}

	// 添加分类信息（如果存在）
	if goodsDTO.Category != nil {
		response["category"] = map[string]interface{}{
			"id":   goodsDTO.Category.Id,
			"name": goodsDTO.Category.Name,
		}
	}

	// 添加品牌信息（如果存在）
	if goodsDTO.Brand != nil {
		response["brand"] = map[string]interface{}{
			"id":   goodsDTO.Brand.Id,
			"name": goodsDTO.Brand.Name,
			"logo": goodsDTO.Brand.Logo,
		}
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

	// 调用库存服务获取库存信息
	stockInfo, err := gc.srv.Inventory().GetStocks(ctx, id)
	if err != nil {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "获取商品库存失败: %v", err), nil)
		return
	}
	
	response := map[string]interface{}{
		"stocks":  stockInfo.Num,
		"goodsId": stockInfo.GoodsId,
	}
	
	core.WriteResponse(ctx, nil, response)
}


// ==================== 分类管理 ====================

func (gc *goodsController) CategoryList(ctx *gin.Context) {
	log.Info("category list function called ...")

	categoriesResponse, err := gc.srv.Goods().CategoryList(ctx)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	// 返回JSON格式的分类树
	var data []interface{}
	if categoriesResponse.JsonData != "" {
		if err := json.Unmarshal([]byte(categoriesResponse.JsonData), &data); err != nil {
			log.Errorf("unmarshal category json data error: %v", err)
			core.WriteResponse(ctx, errors.WithCode(code.ErrEncodingJSON, "分类数据解析失败"), nil)
			return
		}
	}

	core.WriteResponse(ctx, nil, data)
}

func (gc *goodsController) CategoryDetail(ctx *gin.Context) {
	log.Info("category detail function called ...")

	id := ctx.Param("id")
	if id == "" {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "分类ID不能为空"), nil)
		return
	}

	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "分类ID格式不正确"), nil)
		return
	}

	categoryRequest := proto.CategoryListRequest{
		Id: int32(i),
	}

	subCategoriesResponse, err := gc.srv.Goods().CategoryDetail(ctx, &categoryRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	// 构建响应数据
	subCategories := make([]interface{}, 0)
	for _, subCategory := range subCategoriesResponse.SubCategorys {
		subCategories = append(subCategories, map[string]interface{}{
			"id":              subCategory.Id,
			"name":            subCategory.Name,
			"level":           subCategory.Level,
			"parent_category": subCategory.ParentCategory,
			"is_tab":          subCategory.IsTab,
		})
	}

	response := map[string]interface{}{
		"id":              subCategoriesResponse.Info.Id,
		"name":            subCategoriesResponse.Info.Name,
		"level":           subCategoriesResponse.Info.Level,
		"parent_category": subCategoriesResponse.Info.ParentCategory,
		"is_tab":          subCategoriesResponse.Info.IsTab,
		"sub_categorys":   subCategories,
	}

	core.WriteResponse(ctx, nil, response)
}

func (gc *goodsController) CreateCategory(ctx *gin.Context) {
	log.Info("create category function called ...")

	var r request.CreateCategory

	if err := ctx.ShouldBindJSON(&r); err != nil {
		gin2.HandleValidatorError(ctx, err, gc.trans)
		return
	}

	categoryRequest := proto.CategoryInfoRequest{
		Name:           r.Name,
		ParentCategory: r.ParentCategory,
		Level:          r.Level,
		IsTab:          *r.IsTab,
	}

	categoryResponse, err := gc.srv.Goods().CreateCategory(ctx, &categoryRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	response := map[string]interface{}{
		"id":     categoryResponse.Id,
		"name":   categoryResponse.Name,
		"parent": categoryResponse.ParentCategory,
		"level":  categoryResponse.Level,
		"is_tab": categoryResponse.IsTab,
	}

	core.WriteResponse(ctx, nil, response)
}

func (gc *goodsController) UpdateCategory(ctx *gin.Context) {
	log.Info("update category function called ...")

	id := ctx.Param("id")
	if id == "" {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "分类ID不能为空"), nil)
		return
	}

	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "分类ID格式不正确"), nil)
		return
	}

	var r request.UpdateCategory

	if err := ctx.ShouldBindJSON(&r); err != nil {
		gin2.HandleValidatorError(ctx, err, gc.trans)
		return
	}

	categoryRequest := proto.CategoryInfoRequest{
		Id:   int32(i),
		Name: r.Name,
	}
	if r.IsTab != nil {
		categoryRequest.IsTab = *r.IsTab
	}

	_, err = gc.srv.Goods().UpdateCategory(ctx, &categoryRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, map[string]interface{}{
		"msg": "更新成功",
	})
}

func (gc *goodsController) DeleteCategory(ctx *gin.Context) {
	log.Info("delete category function called ...")

	id := ctx.Param("id")
	if id == "" {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "分类ID不能为空"), nil)
		return
	}

	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "分类ID格式不正确"), nil)
		return
	}

	deleteRequest := proto.DeleteCategoryRequest{
		Id: int32(i),
	}

	_, err = gc.srv.Goods().DeleteCategory(ctx, &deleteRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, map[string]interface{}{
		"msg": "删除成功",
	})
}

// ==================== 品牌管理 ====================

func (gc *goodsController) BrandList(ctx *gin.Context) {
	log.Info("brand list function called ...")

	var r request.BrandFilter

	if err := ctx.ShouldBindQuery(&r); err != nil {
		gin2.HandleValidatorError(ctx, err, gc.trans)
		return
	}

	brandRequest := proto.BrandFilterRequest{
		Pages:       r.Pages,
		PagePerNums: r.PagePerNums,
	}

	brandsResponse, err := gc.srv.Goods().BrandList(ctx, &brandRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	reMap := map[string]interface{}{
		"total": brandsResponse.Total,
	}

	brandsList := make([]interface{}, 0)
	for _, brand := range brandsResponse.Data {
		brandsList = append(brandsList, map[string]interface{}{
			"id":   brand.Id,
			"name": brand.Name,
			"logo": brand.Logo,
		})
	}
	reMap["data"] = brandsList

	core.WriteResponse(ctx, nil, reMap)
}

func (gc *goodsController) CreateBrand(ctx *gin.Context) {
	log.Info("create brand function called ...")

	var r request.CreateBrand

	if err := ctx.ShouldBindJSON(&r); err != nil {
		gin2.HandleValidatorError(ctx, err, gc.trans)
		return
	}

	brandRequest := proto.BrandRequest{
		Name: r.Name,
		Logo: r.Logo,
	}

	brandResponse, err := gc.srv.Goods().CreateBrand(ctx, &brandRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	response := map[string]interface{}{
		"id":   brandResponse.Id,
		"name": brandResponse.Name,
		"logo": brandResponse.Logo,
	}

	core.WriteResponse(ctx, nil, response)
}

func (gc *goodsController) UpdateBrand(ctx *gin.Context) {
	log.Info("update brand function called ...")

	id := ctx.Param("id")
	if id == "" {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "品牌ID不能为空"), nil)
		return
	}

	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "品牌ID格式不正确"), nil)
		return
	}

	var r request.UpdateBrand

	if err := ctx.ShouldBindJSON(&r); err != nil {
		gin2.HandleValidatorError(ctx, err, gc.trans)
		return
	}

	brandRequest := proto.BrandRequest{
		Id:   int32(i),
		Name: r.Name,
		Logo: r.Logo,
	}

	_, err = gc.srv.Goods().UpdateBrand(ctx, &brandRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, map[string]interface{}{
		"msg": "更新成功",
	})
}

func (gc *goodsController) DeleteBrand(ctx *gin.Context) {
	log.Info("delete brand function called ...")

	id := ctx.Param("id")
	if id == "" {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "品牌ID不能为空"), nil)
		return
	}

	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "品牌ID格式不正确"), nil)
		return
	}

	brandRequest := proto.BrandRequest{
		Id: int32(i),
	}

	_, err = gc.srv.Goods().DeleteBrand(ctx, &brandRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, map[string]interface{}{
		"msg": "删除成功",
	})
}

// ==================== 轮播图管理 ====================

func (gc *goodsController) BannerList(ctx *gin.Context) {
	log.Info("banner list function called ...")

	bannersResponse, err := gc.srv.Goods().BannerList(ctx)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	reMap := map[string]interface{}{
		"total": bannersResponse.Total,
	}

	bannersList := make([]interface{}, 0)
	for _, banner := range bannersResponse.Data {
		bannersList = append(bannersList, map[string]interface{}{
			"id":    banner.Id,
			"index": banner.Index,
			"image": banner.Image,
			"url":   banner.Url,
		})
	}
	reMap["data"] = bannersList

	core.WriteResponse(ctx, nil, reMap)
}

func (gc *goodsController) CreateBanner(ctx *gin.Context) {
	log.Info("create banner function called ...")

	var r request.CreateBanner

	if err := ctx.ShouldBindJSON(&r); err != nil {
		gin2.HandleValidatorError(ctx, err, gc.trans)
		return
	}

	bannerRequest := proto.BannerRequest{
		Index: r.Index,
		Image: r.Image,
		Url:   r.Url,
	}

	bannerResponse, err := gc.srv.Goods().CreateBanner(ctx, &bannerRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	response := map[string]interface{}{
		"id":    bannerResponse.Id,
		"index": bannerResponse.Index,
		"image": bannerResponse.Image,
		"url":   bannerResponse.Url,
	}

	core.WriteResponse(ctx, nil, response)
}

func (gc *goodsController) UpdateBanner(ctx *gin.Context) {
	log.Info("update banner function called ...")

	id := ctx.Param("id")
	if id == "" {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "轮播图ID不能为空"), nil)
		return
	}

	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "轮播图ID格式不正确"), nil)
		return
	}

	var r request.UpdateBanner

	if err := ctx.ShouldBindJSON(&r); err != nil {
		gin2.HandleValidatorError(ctx, err, gc.trans)
		return
	}

	bannerRequest := proto.BannerRequest{
		Id:    int32(i),
		Index: r.Index,
		Image: r.Image,
		Url:   r.Url,
	}

	_, err = gc.srv.Goods().UpdateBanner(ctx, &bannerRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, map[string]interface{}{
		"msg": "更新成功",
	})
}

func (gc *goodsController) DeleteBanner(ctx *gin.Context) {
	log.Info("delete banner function called ...")

	id := ctx.Param("id")
	if id == "" {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "轮播图ID不能为空"), nil)
		return
	}

	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		core.WriteResponse(ctx, errors.WithCode(code.ErrBind, "轮播图ID格式不正确"), nil)
		return
	}

	bannerRequest := proto.BannerRequest{
		Id: int32(i),
	}

	_, err = gc.srv.Goods().DeleteBanner(ctx, &bannerRequest)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, map[string]interface{}{
		"msg": "删除成功",
	})
}
