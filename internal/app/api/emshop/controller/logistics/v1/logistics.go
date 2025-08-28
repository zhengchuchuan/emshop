package logistics

import (
	restserver "emshop/gin-micro/server/rest-server"
	"emshop/internal/app/api/emshop/domain/dto/request"
	"emshop/internal/app/api/emshop/service"
	"emshop/pkg/common/core"

	"github.com/gin-gonic/gin"
)

type logisticsController struct {
	trans restserver.I18nTranslator
	sf    service.ServiceFactory
}

// NewLogisticsController 创建物流控制器
func NewLogisticsController(trans restserver.I18nTranslator, sf service.ServiceFactory) *logisticsController {
	return &logisticsController{trans, sf}
}

// GetLogisticsInfo 获取物流信息
func (lc *logisticsController) GetLogisticsInfo(ctx *gin.Context) {
	var req request.GetLogisticsInfoRequest

	// 绑定查询参数
	if err := ctx.ShouldBindQuery(&req); err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	// 调用服务层
	resp, err := lc.sf.Logistics().GetLogisticsInfo(ctx, &req)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, resp)
}

// GetLogisticsTracks 获取物流轨迹
func (lc *logisticsController) GetLogisticsTracks(ctx *gin.Context) {
	var req request.GetLogisticsTracksRequest

	// 绑定查询参数
	if err := ctx.ShouldBindQuery(&req); err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	// 调用服务层
	resp, err := lc.sf.Logistics().GetLogisticsTracks(ctx, &req)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, resp)
}

// CalculateShippingFee 计算运费
func (lc *logisticsController) CalculateShippingFee(ctx *gin.Context) {
	var req request.CalculateShippingFeeRequest

	// 绑定请求参数
	if err := ctx.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	// 调用服务层
	resp, err := lc.sf.Logistics().CalculateShippingFee(ctx, &req)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, resp)
}

// GetLogisticsCompanies 获取物流公司列表
func (lc *logisticsController) GetLogisticsCompanies(ctx *gin.Context) {
	// 调用服务层
	resp, err := lc.sf.Logistics().GetLogisticsCompanies(ctx)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, resp)
}
