package v1

import (
	"context"

	"emshop/internal/app/api/emshop/data"
	"emshop/internal/app/api/emshop/domain/dto/request"
	"emshop/internal/app/api/emshop/domain/dto/response"
)

// LogisticsSrv 物流服务接口
type LogisticsSrv interface {
	// GetLogisticsInfo 获取物流信息
	GetLogisticsInfo(ctx context.Context, req *request.GetLogisticsInfoRequest) (*response.LogisticsInfoResponse, error)

	// GetLogisticsTracks 获取物流轨迹
	GetLogisticsTracks(ctx context.Context, req *request.GetLogisticsTracksRequest) (*response.LogisticsTracksResponse, error)

	// CalculateShippingFee 计算运费
	CalculateShippingFee(ctx context.Context, req *request.CalculateShippingFeeRequest) (*response.ShippingFeeResponse, error)

	// GetLogisticsCompanies 获取物流公司列表
	GetLogisticsCompanies(ctx context.Context) (*response.LogisticsCompaniesResponse, error)
}

// logisticsService 物流服务实现
type logisticsService struct {
	data data.DataFactory
}

// NewLogisticsService 创建物流服务实例
func NewLogisticsService(data data.DataFactory) LogisticsSrv {
	return &logisticsService{
		data: data,
	}
}

// GetLogisticsInfo 获取物流信息
func (s *logisticsService) GetLogisticsInfo(ctx context.Context, req *request.GetLogisticsInfoRequest) (*response.LogisticsInfoResponse, error) {
	// 参数验证
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 调用RPC服务
	rpcReq := req.ToProto()
	rpcResp, err := s.data.Logistics().GetLogisticsInfo(ctx, rpcReq)
	if err != nil {
		return nil, err
	}

	// 转换响应
	resp := &response.LogisticsInfoResponse{}
	resp.FromProto(rpcResp)

	return resp, nil
}

// GetLogisticsTracks 获取物流轨迹
func (s *logisticsService) GetLogisticsTracks(ctx context.Context, req *request.GetLogisticsTracksRequest) (*response.LogisticsTracksResponse, error) {
	// 参数验证
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 调用RPC服务
	rpcReq := req.ToProto()
	rpcResp, err := s.data.Logistics().GetLogisticsTracks(ctx, rpcReq)
	if err != nil {
		return nil, err
	}

	// 转换响应
	resp := &response.LogisticsTracksResponse{}
	resp.FromProto(rpcResp)

	return resp, nil
}

// CalculateShippingFee 计算运费
func (s *logisticsService) CalculateShippingFee(ctx context.Context, req *request.CalculateShippingFeeRequest) (*response.ShippingFeeResponse, error) {
	// 调用RPC服务
	rpcReq := req.ToProto()
	rpcResp, err := s.data.Logistics().CalculateShippingFee(ctx, rpcReq)
	if err != nil {
		return nil, err
	}

	// 转换响应
	resp := &response.ShippingFeeResponse{}
	resp.FromProto(rpcResp)

	// 设置物流公司和配送方式名称
	resp.CompanyName = getLogisticsCompanyName(req.LogisticsCompany)
	resp.ShippingMethodName = getShippingMethodName(req.ShippingMethod)

	return resp, nil
}

// GetLogisticsCompanies 获取物流公司列表
func (s *logisticsService) GetLogisticsCompanies(ctx context.Context) (*response.LogisticsCompaniesResponse, error) {
	// 调用RPC服务
	rpcResp, err := s.data.Logistics().GetLogisticsCompanies(ctx)
	if err != nil {
		return nil, err
	}

	// 转换响应
	resp := &response.LogisticsCompaniesResponse{}
	resp.FromProto(rpcResp)

	return resp, nil
}

// getLogisticsCompanyName 获取物流公司名称（简单映射，实际应该从配置或数据库获取）
func getLogisticsCompanyName(company int32) string {
	switch company {
	case 1:
		return "顺丰快递"
	case 2:
		return "圆通快递"
	case 3:
		return "中通快递"
	case 4:
		return "申通快递"
	case 5:
		return "韵达快递"
	case 6:
		return "百世汇通"
	case 7:
		return "德邦快递"
	case 8:
		return "京东物流"
	default:
		return "未知快递公司"
	}
}

// getShippingMethodName 获取配送方式名称（简单映射，实际应该从配置或数据库获取）
func getShippingMethodName(method int32) string {
	switch method {
	case 1:
		return "标准快递"
	case 2:
		return "次日达"
	case 3:
		return "当日达"
	case 4:
		return "上门自提"
	default:
		return "未知配送方式"
	}
}

// 编译时检查接口实现
var _ LogisticsSrv = (*logisticsService)(nil)
