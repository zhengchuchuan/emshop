package coupon

import (
    "net/http"
    "time"

    restserver "emshop/gin-micro/server/rest-server"
    cpbv1 "emshop/api/coupon/v1"
    "emshop/internal/app/api/admin/service"
    "emshop/pkg/common/core"

    "github.com/gin-gonic/gin"
)

type couponController struct {
    trans restserver.I18nTranslator
    srv   service.ServiceFactory
}

func NewCouponController(trans restserver.I18nTranslator, srv service.ServiceFactory) *couponController {
    return &couponController{trans: trans, srv: srv}
}

// EnsureDefaultTemplate 如果没有可用的优惠券模板，则创建一个默认模板
func (cc *couponController) EnsureDefaultTemplate(ctx *gin.Context) {
    tpl, err := cc.srv.Coupon().EnsureDefaultTemplate(ctx)
    if err != nil {
        core.WriteResponse(ctx, err, nil)
        return
    }
    // 返回统一响应结构
    type ensureResp struct {
        Id               int64   `json:"id"`
        Name             string  `json:"name"`
        Type             int32   `json:"type"`
        DiscountType     int32   `json:"discount_type"`
        DiscountValue    float64 `json:"discount_value"`
        MinOrderAmount   float64 `json:"min_order_amount"`
        MaxDiscountAmount float64 `json:"max_discount_amount"`
        TotalCount       int32   `json:"total_count"`
        UsedCount        int32   `json:"used_count"`
        PerUserLimit     int32   `json:"per_user_limit"`
        ValidStartTime   int64   `json:"valid_start_time"`
        ValidEndTime     int64   `json:"valid_end_time"`
        ValidDays        int32   `json:"valid_days"`
        Status           int32   `json:"status"`
        Description      string  `json:"description"`
        CreatedAt        int64   `json:"created_at"`
    }
    toResp := func(in *cpbv1.CouponTemplateResponse) *ensureResp {
        if in == nil { return nil }
        return &ensureResp{
            Id: in.Id,
            Name: in.Name,
            Type: in.Type,
            DiscountType: in.DiscountType,
            DiscountValue: in.DiscountValue,
            MinOrderAmount: in.MinOrderAmount,
            MaxDiscountAmount: in.MaxDiscountAmount,
            TotalCount: in.TotalCount,
            UsedCount: in.UsedCount,
            PerUserLimit: in.PerUserLimit,
            ValidStartTime: in.ValidStartTime,
            ValidEndTime: in.ValidEndTime,
            ValidDays: in.ValidDays,
            Status: in.Status,
            Description: in.Description,
            CreatedAt: in.CreatedAt,
        }
    }
    ctx.JSON(http.StatusOK, gin.H{
        "code": 0,
        "message": "ok",
        "data": toResp(tpl),
        "server_time": time.Now().Unix(),
    })
}

