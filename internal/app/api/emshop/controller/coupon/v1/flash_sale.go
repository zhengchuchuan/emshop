package coupon

import (
    "net/http"
    "strconv"

    restserver "emshop/gin-micro/server/rest-server"
    "emshop/internal/app/api/emshop/service"
    "emshop/internal/app/pkg/middleware"
    "emshop/pkg/common/core"

    "github.com/gin-gonic/gin"
)

type flashSaleController struct {
    trans restserver.I18nTranslator
    sf    service.ServiceFactory
}

func NewFlashSaleController(trans restserver.I18nTranslator, sf service.ServiceFactory) *flashSaleController {
    return &flashSaleController{trans: trans, sf: sf}
}

// ListActive 获取进行中的秒杀活动
func (fc *flashSaleController) ListActive(ctx *gin.Context) {
    resp, err := fc.sf.Coupon().ListActiveFlashSales(ctx)
    core.WriteResponse(ctx, err, resp)
}

// GetStock 获取活动库存
func (fc *flashSaleController) GetStock(ctx *gin.Context) {
    idStr := ctx.Param("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil || id <= 0 {
        ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
        return
    }
    resp, err := fc.sf.Coupon().GetFlashSaleStock(ctx, id)
    core.WriteResponse(ctx, err, resp)
}

// Participate 参与秒杀
func (fc *flashSaleController) Participate(ctx *gin.Context) {
    idStr := ctx.Param("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil || id <= 0 {
        ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
        return
    }
    uid, ok := middleware.GetUserIDFromContext(ctx)
    if !ok || uid <= 0 {
        ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未登录"})
        return
    }
    resp, err := fc.sf.Coupon().ParticipateFlashSale(ctx, int64(uid), id)
    core.WriteResponse(ctx, err, resp)
}

// MyRecord 当前用户该活动的记录
func (fc *flashSaleController) MyRecord(ctx *gin.Context) {
    idStr := ctx.Param("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil || id <= 0 {
        ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
        return
    }
    uid, ok := middleware.GetUserIDFromContext(ctx)
    if !ok || uid <= 0 {
        ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未登录"})
        return
    }
    resp, err := fc.sf.Coupon().GetUserFlashSaleRecord(ctx, int64(uid), id)
    core.WriteResponse(ctx, err, resp)
}

