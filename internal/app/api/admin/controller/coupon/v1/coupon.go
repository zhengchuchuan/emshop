package coupon

import (
    "net/http"
    "strconv"
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

// ListTemplates 模板列表（管理员）
func (cc *couponController) ListTemplates(ctx *gin.Context) {
    // 解析查询参数
    var (
        page     = int32(1)
        pageSize = int32(10)
        status   *int32
    )
    if v := ctx.Query("page"); v != "" {
        if n, err := strconv.Atoi(v); err == nil && n > 0 {
            page = int32(n)
        }
    }
    if v := ctx.Query("pageSize"); v != "" {
        if n, err := strconv.Atoi(v); err == nil && n > 0 {
            pageSize = int32(n)
        }
    }
    if v := ctx.Query("status"); v != "" {
        if n, err := strconv.Atoi(v); err == nil {
            sv := int32(n)
            status = &sv
        }
    }

    req := &cpbv1.ListCouponTemplatesRequest{Page: page, PageSize: pageSize}
    if status != nil { req.Status = status }

    resp, err := cc.srv.Coupon().ListTemplates(ctx, req)
    core.WriteResponse(ctx, err, resp)
}

// GetTemplate 模板详情（管理员）
func (cc *couponController) GetTemplate(ctx *gin.Context) {
    idStr := ctx.Param("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil || id <= 0 {
        ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
        return
    }
    resp, err := cc.srv.Coupon().GetTemplate(ctx, id)
    core.WriteResponse(ctx, err, resp)
}

// CreateTemplate 创建模板（管理员）
func (cc *couponController) CreateTemplate(ctx *gin.Context) {
    var in CreateCouponTemplateJSON
    if err := ctx.ShouldBindJSON(&in); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid request body"})
        return
    }
    req := &cpbv1.CreateCouponTemplateRequest{
        Name:              in.Name,
        Type:              in.Type,
        DiscountType:      in.DiscountType,
        DiscountValue:     in.DiscountValue,
        MinOrderAmount:    in.MinOrderAmount,
        MaxDiscountAmount: in.MaxDiscountAmount,
        TotalCount:        in.TotalCount,
        PerUserLimit:      in.PerUserLimit,
        ValidStartTime:    in.ValidStartTime.AsUnix(),
        ValidEndTime:      in.ValidEndTime.AsUnix(),
        ValidDays:         in.ValidDays,
        Description:       in.Description,
    }
    resp, err := cc.srv.Coupon().CreateTemplate(ctx, req)
    core.WriteResponse(ctx, err, resp)
}

// UpdateTemplate 更新模板（管理员）
func (cc *couponController) UpdateTemplate(ctx *gin.Context) {
    idStr := ctx.Param("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil || id <= 0 {
        ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
        return
    }
    // 仅允许修改 name/status/description
    type body struct {
        Name        *string `json:"name"`
        Status      *int32  `json:"status"`
        Description *string `json:"description"`
    }
    var b body
    if err := ctx.ShouldBindJSON(&b); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid request body"})
        return
    }
    req := &cpbv1.UpdateCouponTemplateRequest{Id: id}
    if b.Name != nil { req.Name = b.Name }
    if b.Status != nil { req.Status = b.Status }
    if b.Description != nil { req.Description = b.Description }
    resp, err := cc.srv.Coupon().UpdateTemplate(ctx, req)
    core.WriteResponse(ctx, err, resp)
}

// DeleteTemplate 删除模板（管理员，逻辑删除：置为已结束）
func (cc *couponController) DeleteTemplate(ctx *gin.Context) {
    idStr := ctx.Param("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil || id <= 0 {
        ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
        return
    }
    resp, err := cc.srv.Coupon().DeleteTemplate(ctx, id)
    core.WriteResponse(ctx, err, resp)
}
