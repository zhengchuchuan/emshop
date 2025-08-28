package rpc

import (
	"context"
	ppbv1 "emshop/api/payment/v1"
	"emshop/internal/app/api/emshop/data"
	"emshop/pkg/log"
)

type payment struct {
	pc ppbv1.PaymentClient
}

func NewPayment(pc ppbv1.PaymentClient) *payment {
	return &payment{pc}
}

// CreatePayment 创建支付订单
func (p *payment) CreatePayment(ctx context.Context, request *ppbv1.CreatePaymentRequest) (*ppbv1.CreatePaymentResponse, error) {
	log.Infof("Calling CreatePayment gRPC for order: %s, amount: %.2f, method: %d",
		request.OrderSn, request.Amount, request.PaymentMethod)
	response, err := p.pc.CreatePayment(ctx, request)
	if err != nil {
		log.Errorf("CreatePayment gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("CreatePayment gRPC call successful, paymentSn: %s, expiredAt: %d",
		response.PaymentSn, response.ExpiredAt)
	return response, nil
}

// GetPaymentStatus 查询支付状态
func (p *payment) GetPaymentStatus(ctx context.Context, request *ppbv1.GetPaymentStatusRequest) (*ppbv1.PaymentStatusResponse, error) {
	log.Infof("Calling GetPaymentStatus gRPC for paymentSn: %s", request.PaymentSn)
	response, err := p.pc.GetPaymentStatus(ctx, request)
	if err != nil {
		log.Errorf("GetPaymentStatus gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("GetPaymentStatus gRPC call successful, status: %d, amount: %.2f",
		response.PaymentStatus, response.Amount)
	return response, nil
}

// SimulatePaymentSuccess 模拟支付成功
func (p *payment) SimulatePaymentSuccess(ctx context.Context, request *ppbv1.SimulatePaymentRequest) error {
	log.Infof("Calling SimulatePaymentSuccess gRPC for paymentSn: %s", request.PaymentSn)
	_, err := p.pc.SimulatePaymentSuccess(ctx, request)
	if err != nil {
		log.Errorf("SimulatePaymentSuccess gRPC call failed: %v", err)
		return err
	}
	log.Infof("SimulatePaymentSuccess gRPC call successful")
	return nil
}

// SimulatePaymentFailure 模拟支付失败
func (p *payment) SimulatePaymentFailure(ctx context.Context, request *ppbv1.SimulatePaymentRequest) error {
	log.Infof("Calling SimulatePaymentFailure gRPC for paymentSn: %s", request.PaymentSn)
	_, err := p.pc.SimulatePaymentFailure(ctx, request)
	if err != nil {
		log.Errorf("SimulatePaymentFailure gRPC call failed: %v", err)
		return err
	}
	log.Infof("SimulatePaymentFailure gRPC call successful")
	return nil
}

var _ data.PaymentData = &payment{}
