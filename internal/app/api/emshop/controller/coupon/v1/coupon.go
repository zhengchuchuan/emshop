package coupon

import (
	"net/http"

	restserver "emshop/gin-micro/server/rest-server"
	"emshop/internal/app/api/emshop/domain/dto/request"
	"emshop/internal/app/api/emshop/service"
	"emshop/internal/app/pkg/jwt"
	"emshop/pkg/common/core"

	"github.com/gin-gonic/gin"
)

type couponController struct {
	trans restserver.I18nTranslator
	sf    service.ServiceFactory
}

// NewCouponController 创建优惠券控制器
func NewCouponController(trans restserver.I18nTranslator, sf service.ServiceFactory) *couponController {
	return &couponController{trans, sf}
}

// ListTemplates 获取优惠券模板列表
func (cc *couponController) ListTemplates(ctx *gin.Context) {
	var req request.ListCouponTemplatesRequest

	// 绑定查询参数
	if err := ctx.ShouldBindQuery(&req); err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	// 调用服务层
	resp, err := cc.sf.Coupon().ListTemplates(ctx, &req)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, resp)
}

// ReceiveCoupon 用户领取优惠券
func (cc *couponController) ReceiveCoupon(ctx *gin.Context) {
	var req request.ReceiveCouponRequest

	// 绑定请求参数
	if err := ctx.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	// 获取用户ID
	userID := cc.getUserIDFromContext(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未登录"})
		return
	}

	// 调用服务层
	resp, err := cc.sf.Coupon().ReceiveCoupon(ctx, userID, &req)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, resp)
}

// GetUserCoupons 获取用户优惠券列表
func (cc *couponController) GetUserCoupons(ctx *gin.Context) {
	var req request.GetUserCouponsRequest

	// 绑定查询参数
	if err := ctx.ShouldBindQuery(&req); err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	// 获取用户ID
	userID := cc.getUserIDFromContext(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未登录"})
		return
	}

	// 调用服务层
	resp, err := cc.sf.Coupon().GetUserCoupons(ctx, userID, &req)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, resp)
}

// GetAvailableCoupons 获取用户可用优惠券
func (cc *couponController) GetAvailableCoupons(ctx *gin.Context) {
	var req request.GetAvailableCouponsRequest

	// 绑定查询参数
	if err := ctx.ShouldBindQuery(&req); err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	// 获取用户ID
	userID := cc.getUserIDFromContext(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未登录"})
		return
	}

	// 调用服务层
	resp, err := cc.sf.Coupon().GetAvailableCoupons(ctx, userID, &req)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, resp)
}

// CalculateDiscount 计算优惠券折扣
func (cc *couponController) CalculateDiscount(ctx *gin.Context) {
	var req request.CalculateCouponDiscountRequest

	// 绑定请求参数
	if err := ctx.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	// 获取用户ID
	userID := cc.getUserIDFromContext(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未登录"})
		return
	}

	// 调用服务层
	resp, err := cc.sf.Coupon().CalculateDiscount(ctx, userID, &req)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, resp)
}

// getUserIDFromContext 从上下文获取用户ID
func (cc *couponController) getUserIDFromContext(ctx *gin.Context) int64 {
	// 从JWT token中获取用户ID
	if claims, ok := ctx.Get("claims"); ok {
		if jwtClaims, ok := claims.(*jwt.EmshopClaims); ok {
			return int64(jwtClaims.ID)
		}
	}
	return 0
}
