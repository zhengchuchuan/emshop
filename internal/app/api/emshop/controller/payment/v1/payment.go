package payment

import (
    "net/http"

    restserver "emshop/gin-micro/server/rest-server"
    "emshop/internal/app/api/emshop/domain/dto/request"
    "emshop/internal/app/api/emshop/service"
    "emshop/internal/app/pkg/jwt"
    "emshop/pkg/common/core"

    "github.com/gin-gonic/gin"
)

type paymentController struct {
	trans restserver.I18nTranslator
	sf    service.ServiceFactory
}

// NewPaymentController 创建支付控制器
func NewPaymentController(trans restserver.I18nTranslator, sf service.ServiceFactory) *paymentController {
	return &paymentController{trans, sf}
}

// CreatePayment 创建支付订单
func (pc *paymentController) CreatePayment(ctx *gin.Context) {
	var req request.CreatePaymentRequest

	// 绑定请求参数
	if err := ctx.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	// 获取用户ID
	userID := pc.getUserIDFromContext(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未登录"})
		return
	}

	// 调用服务层
	resp, err := pc.sf.Payment().CreatePayment(ctx, int32(userID), &req)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, resp)
}

// GetPaymentStatus 获取支付状态
func (pc *paymentController) GetPaymentStatus(ctx *gin.Context) {
	paymentSN := ctx.Param("paymentSN")
	if paymentSN == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "支付单号不能为空"})
		return
	}

	// 调用服务层
	resp, err := pc.sf.Payment().GetPaymentStatus(ctx, paymentSN)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, resp)
}

// SimulatePayment 模拟支付
func (pc *paymentController) SimulatePayment(ctx *gin.Context) {
	paymentSN := ctx.Param("paymentSN")
	if paymentSN == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "支付单号不能为空"})
		return
	}

	var req request.SimulatePaymentRequest

	// 绑定请求参数
	if err := ctx.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	// 调用服务层
	resp, err := pc.sf.Payment().SimulatePayment(ctx, paymentSN, &req)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, resp)
}

// getUserIDFromContext 从上下文获取用户ID
func (pc *paymentController) getUserIDFromContext(ctx *gin.Context) int64 {
    // 从中间件设置的上下文键获取用户ID
    if v, ok := ctx.Get(jwt.KeyUserID); ok {
        if id, ok := v.(int); ok {
            return int64(id)
        }
        if id64, ok := v.(int64); ok {
            return id64
        }
    }
    return 0
}
