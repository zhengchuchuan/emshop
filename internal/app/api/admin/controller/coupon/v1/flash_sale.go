package coupon

import (
    "net/http"
    "strconv"

    restserver "emshop/gin-micro/server/rest-server"
    cpbv1 "emshop/api/coupon/v1"
    "emshop/internal/app/api/admin/service"
    "emshop/pkg/common/core"

    "github.com/gin-gonic/gin"
)

type flashSaleAdminController struct {
    trans restserver.I18nTranslator
    srv   service.ServiceFactory
}

func NewFlashSaleAdminController(trans restserver.I18nTranslator, srv service.ServiceFactory) *flashSaleAdminController {
    return &flashSaleAdminController{trans: trans, srv: srv}
}

func (ac *flashSaleAdminController) Create(ctx *gin.Context) {
    var in CreateFlashSaleJSON
    if err := ctx.ShouldBindJSON(&in); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid request body"})
        return
    }
    req := &cpbv1.CreateFlashSaleActivityRequest{
        CouponTemplateId: in.CouponTemplateID,
        Name:             in.Name,
        StartTime:        in.StartTime.AsUnix(),
        EndTime:          in.EndTime.AsUnix(),
        FlashSaleCount:   in.FlashSaleCount,
        PerUserLimit:     in.PerUserLimit,
    }
    resp, err := ac.srv.Coupon().CreateFlashSaleActivity(ctx, req)
    core.WriteResponse(ctx, err, resp)
}

func (ac *flashSaleAdminController) Get(ctx *gin.Context) {
    idStr := ctx.Param("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil || id <= 0 {
        ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
        return
    }
    resp, err := ac.srv.Coupon().GetFlashSaleActivity(ctx, id)
    core.WriteResponse(ctx, err, resp)
}

func (ac *flashSaleAdminController) List(ctx *gin.Context) {
    var (
        page     = int32(1)
        pageSize = int32(10)
        status   *int32
    )
    if v := ctx.Query("page"); v != "" {
        if n, err := strconv.Atoi(v); err == nil && n > 0 { page = int32(n) }
    }
    if v := ctx.Query("pageSize"); v != "" {
        if n, err := strconv.Atoi(v); err == nil && n > 0 { pageSize = int32(n) }
    }
    if v := ctx.Query("status"); v != "" {
        if n, err := strconv.Atoi(v); err == nil { s := int32(n); status = &s }
    }
    req := &cpbv1.ListFlashSaleActivitiesRequest{Page: page, PageSize: pageSize}
    if status != nil { req.Status = status }
    resp, err := ac.srv.Coupon().ListFlashSaleActivities(ctx, req)
    core.WriteResponse(ctx, err, resp)
}

// GetConfig 管理配置读取（转发到 coupon 管理接口）
func (ac *flashSaleAdminController) GetConfig(ctx *gin.Context) {
    key := ctx.Param("key")
    if key == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"msg": "empty key"})
        return
    }
    resp, err := ac.srv.Coupon().GetManageConfig(ctx, key)
    if err != nil {
        ctx.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, resp)
}

type setConfigReq struct {
    Key         string `json:"key"`
    Value       string `json:"value"`
    Description string `json:"description"`
}

// SetConfig 管理配置设置（转发到 coupon 管理接口）
func (ac *flashSaleAdminController) SetConfig(ctx *gin.Context) {
    var in setConfigReq
    if err := ctx.ShouldBindJSON(&in); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid body"})
        return
    }
    if in.Key == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"msg": "empty key"})
        return
    }
    resp, err := ac.srv.Coupon().SetManageConfig(ctx, in.Key, in.Value, in.Description)
    if err != nil {
        ctx.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, resp)
}

// Status 聚合活动状态（DB + Redis 实时库存）
func (ac *flashSaleAdminController) Status(ctx *gin.Context) {
    idStr := ctx.Param("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil || id <= 0 {
        ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
        return
    }
    // 1) 读活动详情（coupon服务会在进行中时补充 Redis 实时remain/sold）
    act, err := ac.srv.Coupon().GetFlashSaleActivity(ctx, id)
    if err != nil {
        ctx.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
        return
    }
    // 2) 再读一次库存视图（更轻量、专注库存）
    stock, err := ac.srv.Coupon().GetFlashSaleStock(ctx, id)
    if err != nil {
        // 库存接口失败不致命，返回活动信息 + 错误提示
        ctx.JSON(http.StatusOK, gin.H{"activity": act, "stock_error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, gin.H{"activity": act, "stock": stock})
}

// Start 启动/预热秒杀活动（转发到 coupon 管理接口）
func (ac *flashSaleAdminController) Start(ctx *gin.Context) {
    idStr := ctx.Param("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil || id <= 0 {
        ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
        return
    }
    resp, err := ac.srv.Coupon().StartFlashSale(ctx, id)
    if err != nil { ctx.JSON(http.StatusBadGateway, gin.H{"error": err.Error()}); return }
    ctx.JSON(http.StatusOK, resp)
}

// Stop 停止秒杀活动（转发到 coupon 管理接口）
func (ac *flashSaleAdminController) Stop(ctx *gin.Context) {
    idStr := ctx.Param("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil || id <= 0 {
        ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
        return
    }
    resp, err := ac.srv.Coupon().StopFlashSale(ctx, id)
    if err != nil { ctx.JSON(http.StatusBadGateway, gin.H{"error": err.Error()}); return }
    ctx.JSON(http.StatusOK, resp)
}
