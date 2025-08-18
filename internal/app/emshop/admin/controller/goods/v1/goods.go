package goods

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	restserver "emshop/gin-micro/server/rest-server"
	proto "emshop/api/goods/v1"
	"emshop/internal/app/emshop/admin/service"
	"emshop/pkg/common/core"
)

type goodsController struct {
	trans restserver.I18nTranslator
	sf    service.ServiceFactory
}

func NewGoodsController(sf service.ServiceFactory, trans restserver.I18nTranslator) *goodsController {
	return &goodsController{
		sf:    sf,
		trans: trans,
	}
}

// ==================== 商品管理 ====================

// List 商品列表（管理员专用）
func (gc *goodsController) List(ctx *gin.Context) {
	var r struct {
		IsNew       *bool   `form:"isNew"`
		IsHot       *bool   `form:"isHot"`
		PriceMax    *int32  `form:"priceMax"`
		PriceMin    *int32  `form:"priceMin"`
		TopCategory *int32  `form:"topCategory"`
		Brand       *int32  `form:"brand"`
		KeyWords    *string `form:"keyWords"`
		Pages       *int32  `form:"pages"`
		PagePerNums *int32  `form:"pagePerNums"`
	}

	if err := ctx.ShouldBindQuery(&r); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid query parameters"})
		return
	}

	request := &proto.GoodsFilterRequest{}
	if r.IsNew != nil {
		request.IsNew = *r.IsNew
	}
	if r.IsHot != nil {
		request.IsHot = *r.IsHot
	}
	if r.PriceMax != nil {
		request.PriceMax = *r.PriceMax
	}
	if r.PriceMin != nil {
		request.PriceMin = *r.PriceMin
	}
	if r.TopCategory != nil {
		request.TopCategory = *r.TopCategory
	}
	if r.Brand != nil {
		request.Brand = *r.Brand
	}
	if r.KeyWords != nil {
		request.KeyWords = *r.KeyWords
	}
	if r.Pages != nil {
		request.Pages = *r.Pages
	}
	if r.PagePerNums != nil {
		request.PagePerNums = *r.PagePerNums
	}

	response, err := gc.sf.Goods().GetGoodsList(ctx, request)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, response)
}

// Create 创建商品（管理员专用）
func (gc *goodsController) Create(ctx *gin.Context) {
	var req proto.CreateGoodsInfo
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid request body"})
		return
	}

	response, err := gc.sf.Goods().CreateGoods(ctx, &req)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, response)
}

// Update 更新商品（管理员专用）
func (gc *goodsController) Update(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id parameter"})
		return
	}

	var req proto.CreateGoodsInfo
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid request body"})
		return
	}

	req.Id = int32(id)
	response, err := gc.sf.Goods().UpdateGoods(ctx, &req)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, response)
}

// Delete 删除商品（管理员专用）
func (gc *goodsController) Delete(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id parameter"})
		return
	}

	response, err := gc.sf.Goods().DeleteGoods(ctx, id)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, response)
}

// Detail 商品详情（管理员专用）
func (gc *goodsController) Detail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id parameter"})
		return
	}

	response, err := gc.sf.Goods().GetGoodsDetail(ctx, id)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, response)
}

// Sync 同步商品数据（管理员专用）
func (gc *goodsController) Sync(ctx *gin.Context) {
	var req proto.SyncDataRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid request body"})
		return
	}

	response, err := gc.sf.Goods().SyncGoodsData(ctx, &req)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, response)
}

// ==================== 分类管理 ====================

// CategoryList 分类列表（管理员专用）
func (gc *goodsController) CategoryList(ctx *gin.Context) {
	response, err := gc.sf.Goods().GetAllCategoriesList(ctx)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, response)
}

// CreateCategory 创建分类（管理员专用）
func (gc *goodsController) CreateCategory(ctx *gin.Context) {
	var req proto.CategoryInfoRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid request body"})
		return
	}

	response, err := gc.sf.Goods().CreateCategory(ctx, &req)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, response)
}

// UpdateCategory 更新分类（管理员专用）
func (gc *goodsController) UpdateCategory(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id parameter"})
		return
	}

	var req proto.CategoryInfoRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid request body"})
		return
	}

	req.Id = int32(id)
	response, err := gc.sf.Goods().UpdateCategory(ctx, &req)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, response)
}

// DeleteCategory 删除分类（管理员专用）
func (gc *goodsController) DeleteCategory(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id parameter"})
		return
	}

	response, err := gc.sf.Goods().DeleteCategory(ctx, id)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, response)
}

// ==================== 品牌管理 ====================

// BrandList 品牌列表（管理员专用）
func (gc *goodsController) BrandList(ctx *gin.Context) {
	var r struct {
		Pages       *int32 `form:"pages"`
		PagePerNums *int32 `form:"pagePerNums"`
	}

	if err := ctx.ShouldBindQuery(&r); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid query parameters"})
		return
	}

	request := &proto.BrandFilterRequest{}
	if r.Pages != nil {
		request.Pages = *r.Pages
	}
	if r.PagePerNums != nil {
		request.PagePerNums = *r.PagePerNums
	}

	response, err := gc.sf.Goods().GetBrandsList(ctx, request)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, response)
}

// CreateBrand 创建品牌（管理员专用）
func (gc *goodsController) CreateBrand(ctx *gin.Context) {
	var req proto.BrandRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid request body"})
		return
	}

	response, err := gc.sf.Goods().CreateBrand(ctx, &req)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, response)
}

// UpdateBrand 更新品牌（管理员专用）
func (gc *goodsController) UpdateBrand(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id parameter"})
		return
	}

	var req proto.BrandRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid request body"})
		return
	}

	req.Id = int32(id)
	response, err := gc.sf.Goods().UpdateBrand(ctx, &req)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, response)
}

// DeleteBrand 删除品牌（管理员专用）
func (gc *goodsController) DeleteBrand(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id parameter"})
		return
	}

	response, err := gc.sf.Goods().DeleteBrand(ctx, id)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, response)
}

// ==================== 轮播图管理 ====================

// BannerList 轮播图列表（管理员专用）
func (gc *goodsController) BannerList(ctx *gin.Context) {
	response, err := gc.sf.Goods().GetBannersList(ctx)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, response)
}

// CreateBanner 创建轮播图（管理员专用）
func (gc *goodsController) CreateBanner(ctx *gin.Context) {
	var req proto.BannerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid request body"})
		return
	}

	response, err := gc.sf.Goods().CreateBanner(ctx, &req)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, response)
}

// UpdateBanner 更新轮播图（管理员专用）
func (gc *goodsController) UpdateBanner(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id parameter"})
		return
	}

	var req proto.BannerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid request body"})
		return
	}

	req.Id = int32(id)
	response, err := gc.sf.Goods().UpdateBanner(ctx, &req)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, response)
}

// DeleteBanner 删除轮播图（管理员专用）
func (gc *goodsController) DeleteBanner(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id parameter"})
		return
	}

	response, err := gc.sf.Goods().DeleteBanner(ctx, id)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, response)
}