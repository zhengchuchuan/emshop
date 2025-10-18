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
