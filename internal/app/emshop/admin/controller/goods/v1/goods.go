package goods

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	restserver "emshop/gin-micro/server/rest-server"
	proto "emshop/api/goods/v1"
	ipbv1 "emshop/api/inventory/v1"
	"emshop/internal/app/emshop/admin/service"
	adminRequest "emshop/internal/app/emshop/admin/domain/dto/request"
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
	var r adminRequest.AdminGoodsFilter
	
	if err := ctx.ShouldBindQuery(&r); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid query parameters"})
		return
	}
	
	// 转换为protobuf请求，与API层保持一致的逻辑
	request := &proto.GoodsFilterRequest{}
	
	// 条件参数 - 只有非空时才设置
	if r.IsNew != nil {
		request.IsNew = *r.IsNew
	}
	if r.IsHot != nil {
		request.IsHot = *r.IsHot
	}
	if r.IsTab != nil {
		request.IsTab = *r.IsTab
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
	
	// 分页参数 - 设置默认值
	if r.Pages != nil {
		request.Pages = *r.Pages
	} else {
		request.Pages = 1 // 默认第1页
	}
	
	if r.PagePerNums != nil {
		request.PagePerNums = *r.PagePerNums
	} else {
		request.PagePerNums = 10 // 默认每页10条
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

// ==================== 库存管理 ====================

// GetInventory 获取商品库存（管理员专用）
func (gc *goodsController) GetInventory(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id parameter"})
		return
	}

	inventory, err := gc.sf.Goods().GetGoodsInventory(ctx, int32(id))
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, inventory)
}

// SetInventory 设置商品库存（管理员专用）
func (gc *goodsController) SetInventory(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id parameter"})
		return
	}

	var req struct {
		Num int32 `json:"num" binding:"required,min=0"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid request body"})
		return
	}

	request := &ipbv1.GoodsInvInfo{
		GoodsId: int32(id),
		Num:     req.Num,
	}

	err = gc.sf.Goods().SetGoodsInventory(ctx, request)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, gin.H{"msg": "inventory updated successfully"})
}

// BatchSetInventory 批量设置商品库存（管理员专用）
func (gc *goodsController) BatchSetInventory(ctx *gin.Context) {
	var req struct {
		Inventories []struct {
			GoodsId int32 `json:"goodsId" binding:"required"`
			Num     int32 `json:"num" binding:"required,min=0"`
		} `json:"inventories" binding:"required,dive"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid request body"})
		return
	}

	inventories := make([]*ipbv1.GoodsInvInfo, 0, len(req.Inventories))
	for _, inv := range req.Inventories {
		inventories = append(inventories, &ipbv1.GoodsInvInfo{
			GoodsId: inv.GoodsId,
			Num:     inv.Num,
		})
	}

	err := gc.sf.Goods().BatchSetGoodsInventory(ctx, inventories)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, gin.H{"msg": "batch inventory updated successfully"})
}


// ==================== 批量操作 ====================

// BatchDeleteGoods 批量删除商品（管理员专用）
func (gc *goodsController) BatchDeleteGoods(ctx *gin.Context) {
	var req struct {
		Ids []int32 `json:"ids" binding:"required,dive,min=1"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid request body"})
		return
	}

	request := &proto.BatchDeleteGoodsRequest{
		Ids: req.Ids,
	}

	response, err := gc.sf.Goods().BatchDeleteGoods(ctx, request)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, response)
}

// BatchUpdateGoodsStatus 批量更新商品状态（管理员专用）
func (gc *goodsController) BatchUpdateGoodsStatus(ctx *gin.Context) {
	var req struct {
		Ids           []int32 `json:"ids" binding:"required,dive,min=1"`
		OnSale        *bool   `json:"onSale"`
		IsHot         *bool   `json:"isHot"`
		IsNew         *bool   `json:"isNew"`
		UpdateOnSale  bool    `json:"updateOnSale"`
		UpdateIsHot   bool    `json:"updateIsHot"`
		UpdateIsNew   bool    `json:"updateIsNew"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid request body"})
		return
	}

	// 验证至少有一个更新标志为true
	if !req.UpdateOnSale && !req.UpdateIsHot && !req.UpdateIsNew {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "at least one update flag must be true"})
		return
	}

	request := &proto.BatchUpdateGoodsStatusRequest{
		Ids:           req.Ids,
		UpdateOnSale:  req.UpdateOnSale,
		UpdateIsHot:   req.UpdateIsHot,
		UpdateIsNew:   req.UpdateIsNew,
	}

	if req.OnSale != nil {
		request.OnSale = *req.OnSale
	}
	if req.IsHot != nil {
		request.IsHot = *req.IsHot
	}
	if req.IsNew != nil {
		request.IsNew = *req.IsNew
	}

	response, err := gc.sf.Goods().BatchUpdateGoodsStatus(ctx, request)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, response)
}