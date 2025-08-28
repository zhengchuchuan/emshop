package v1

import (
	"context"

	ppbv1 "emshop/api/payment/v1"
	"emshop/internal/app/api/emshop/data"
	"emshop/internal/app/api/emshop/domain/dto/request"
	"emshop/internal/app/api/emshop/domain/dto/response"
)

// PaymentSrv 支付服务接口
type PaymentSrv interface {
	// CreatePayment 创建支付订单
	CreatePayment(ctx context.Context, userID int32, req *request.CreatePaymentRequest) (*response.CreatePaymentResponse, error)

	// GetPaymentStatus 获取支付状态
	GetPaymentStatus(ctx context.Context, paymentSN string) (*response.PaymentStatusResponse, error)

	// SimulatePayment 模拟支付
	SimulatePayment(ctx context.Context, paymentSN string, req *request.SimulatePaymentRequest) (*response.SimulatePaymentResponse, error)
}

// paymentService 支付服务实现
type paymentService struct {
	data data.DataFactory
}

// NewPaymentService 创建支付服务实例
func NewPaymentService(data data.DataFactory) PaymentSrv {
	return &paymentService{
		data: data,
	}
}

// CreatePayment 创建支付订单
func (s *paymentService) CreatePayment(ctx context.Context, userID int32, req *request.CreatePaymentRequest) (*response.CreatePaymentResponse, error) {
	// 调用RPC服务
	rpcReq := req.ToProto(userID)
	rpcResp, err := s.data.Payment().CreatePayment(ctx, rpcReq)
	if err != nil {
		return nil, err
	}

	// 转换响应
	resp := &response.CreatePaymentResponse{}
	resp.FromProto(rpcResp)

	return resp, nil
}

// GetPaymentStatus 获取支付状态
func (s *paymentService) GetPaymentStatus(ctx context.Context, paymentSN string) (*response.PaymentStatusResponse, error) {
	// 构建RPC请求
	rpcReq := &ppbv1.GetPaymentStatusRequest{
		PaymentSn: paymentSN,
	}

	// 调用RPC服务
	rpcResp, err := s.data.Payment().GetPaymentStatus(ctx, rpcReq)
	if err != nil {
		return nil, err
	}

	// 转换响应
	resp := &response.PaymentStatusResponse{}
	resp.FromProto(rpcResp)

	return resp, nil
}

// SimulatePayment 模拟支付
func (s *paymentService) SimulatePayment(ctx context.Context, paymentSN string, req *request.SimulatePaymentRequest) (*response.SimulatePaymentResponse, error) {
	// 调用RPC服务
	rpcReq := req.ToProto(paymentSN)
	err := s.data.Payment().SimulatePaymentSuccess(ctx, rpcReq)
	if err != nil {
		return &response.SimulatePaymentResponse{
			Success: false,
			Message: "模拟支付失败",
		}, nil
	}

	return &response.SimulatePaymentResponse{
		Success: true,
		Message: "模拟支付成功",
	}, nil
}

// 编译时检查接口实现
var _ PaymentSrv = (*paymentService)(nil)
