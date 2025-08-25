package payment

import (
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
	pb "emshop/api/payment/v1"
	"emshop/internal/app/payment/srv/domain/do"
	"emshop/internal/app/payment/srv/domain/dto"
	"emshop/internal/app/payment/srv/service/v1"
	"emshop/pkg/log"
)

type paymentServer struct {
	pb.UnimplementedPaymentServer

	srv *v1.Service
}

// NewPaymentServer 创建支付服务gRPC服务器
func NewPaymentServer(srv *v1.Service) *paymentServer {
	return &paymentServer{srv: srv}
}

// CreatePayment 创建支付订单
func (ps *paymentServer) CreatePayment(ctx context.Context, req *pb.CreatePaymentRequest) (*pb.CreatePaymentResponse, error) {
	log.Infof("收到创建支付订单请求: 订单号=%s, 用户ID=%d, 金额=%f", req.OrderSn, req.UserId, req.Amount)

	// 请求参数转换
	createReq := &dto.CreatePaymentDTO{
		OrderSn:        req.OrderSn,
		UserID:         req.UserId,
		Amount:         req.Amount,
		PaymentMethod:  do.PaymentMethod(req.PaymentMethod),
		ExpiredMinutes: 15, // 默认15分钟
	}

	if req.ExpiredMinutes != nil {
		createReq.ExpiredMinutes = *req.ExpiredMinutes
	}

	// 调用业务服务
	result, err := ps.srv.PaymentSrv.CreatePayment(ctx, createReq)
	if err != nil {
		log.Errorf("创建支付订单失败: %v", err)
		return nil, err
	}

	// 响应转换
	response := &pb.CreatePaymentResponse{
		PaymentSn: result.PaymentSn,
		ExpiredAt: result.ExpiredAt.Unix(),
	}

	// 模拟生成支付链接
	if req.PaymentMethod == int32(do.PaymentMethodWechat) {
		paymentUrl := "https://pay.weixin.qq.com/mock/" + result.PaymentSn
		response.PaymentUrl = &paymentUrl
	} else if req.PaymentMethod == int32(do.PaymentMethodAlipay) {
		paymentUrl := "https://openapi.alipay.com/mock/" + result.PaymentSn
		response.PaymentUrl = &paymentUrl
	}

	log.Infof("创建支付订单成功: 支付单号=%s", result.PaymentSn)
	return response, nil
}

// CancelPayment 取消支付订单（补偿）
func (ps *paymentServer) CancelPayment(ctx context.Context, req *pb.CancelPaymentRequest) (*emptypb.Empty, error) {
	log.Infof("收到取消支付订单请求: 支付单号=%s", req.PaymentSn)

	err := ps.srv.PaymentSrv.CancelPayment(ctx, req.PaymentSn)
	if err != nil {
		log.Errorf("取消支付订单失败: %v", err)
		return nil, err
	}

	log.Infof("取消支付订单成功: 支付单号=%s", req.PaymentSn)
	return &emptypb.Empty{}, nil
}

// ConfirmPayment 确认支付成功
func (ps *paymentServer) ConfirmPayment(ctx context.Context, req *pb.ConfirmPaymentRequest) (*emptypb.Empty, error) {
	log.Infof("收到确认支付成功请求: 支付单号=%s", req.PaymentSn)

	confirmReq := &dto.ConfirmPaymentDTO{
		PaymentSn:    req.PaymentSn,
		ThirdPartySn: req.ThirdPartySn,
	}

	err := ps.srv.PaymentSrv.ConfirmPayment(ctx, confirmReq)
	if err != nil {
		log.Errorf("确认支付成功失败: %v", err)
		return nil, err
	}

	log.Infof("确认支付成功成功: 支付单号=%s", req.PaymentSn)
	return &emptypb.Empty{}, nil
}

// RefundPayment 退款（补偿）
func (ps *paymentServer) RefundPayment(ctx context.Context, req *pb.RefundPaymentRequest) (*emptypb.Empty, error) {
	log.Infof("收到退款请求: 支付单号=%s, 退款金额=%f", req.PaymentSn, req.RefundAmount)

	refundReq := &dto.RefundPaymentDTO{
		PaymentSn:    req.PaymentSn,
		RefundAmount: req.RefundAmount,
		Reason:       req.Reason,
	}

	err := ps.srv.PaymentSrv.RefundPayment(ctx, refundReq)
	if err != nil {
		log.Errorf("退款失败: %v", err)
		return nil, err
	}

	log.Infof("退款成功: 支付单号=%s", req.PaymentSn)
	return &emptypb.Empty{}, nil
}

// GetPaymentStatus 查询支付状态
func (ps *paymentServer) GetPaymentStatus(ctx context.Context, req *pb.GetPaymentStatusRequest) (*pb.PaymentStatusResponse, error) {
	log.Debugf("收到查询支付状态请求: 支付单号=%s", req.PaymentSn)

	result, err := ps.srv.PaymentSrv.GetPaymentStatus(ctx, req.PaymentSn)
	if err != nil {
		log.Errorf("查询支付状态失败: %v", err)
		return nil, err
	}

	// 响应转换
	response := &pb.PaymentStatusResponse{
		PaymentSn:     result.PaymentSn,
		OrderSn:       result.OrderSn,
		PaymentStatus: int32(result.PaymentStatus),
		Amount:        result.Amount,
		PaymentMethod: int32(result.PaymentMethod),
		ExpiredAt:     result.ExpiredAt.Unix(),
	}

	if result.PaidAt != nil {
		paidAt := result.PaidAt.Unix()
		response.PaidAt = &paidAt
	}

	return response, nil
}

// SimulatePaymentSuccess 模拟支付成功
func (ps *paymentServer) SimulatePaymentSuccess(ctx context.Context, req *pb.SimulatePaymentRequest) (*emptypb.Empty, error) {
	log.Infof("收到模拟支付成功请求: 支付单号=%s", req.PaymentSn)

	err := ps.srv.PaymentSrv.SimulatePaymentSuccess(ctx, req.PaymentSn, req.ThirdPartySn)
	if err != nil {
		log.Errorf("模拟支付成功失败: %v", err)
		return nil, err
	}

	log.Infof("模拟支付成功成功: 支付单号=%s", req.PaymentSn)
	return &emptypb.Empty{}, nil
}

// SimulatePaymentFailure 模拟支付失败
func (ps *paymentServer) SimulatePaymentFailure(ctx context.Context, req *pb.SimulatePaymentRequest) (*emptypb.Empty, error) {
	log.Infof("收到模拟支付失败请求: 支付单号=%s", req.PaymentSn)

	err := ps.srv.PaymentSrv.SimulatePaymentFailure(ctx, req.PaymentSn)
	if err != nil {
		log.Errorf("模拟支付失败失败: %v", err)
		return nil, err
	}

	log.Infof("模拟支付失败成功: 支付单号=%s", req.PaymentSn)
	return &emptypb.Empty{}, nil
}

// ReserveStock 预留库存（订单提交事务用）
func (ps *paymentServer) ReserveStock(ctx context.Context, req *pb.ReserveStockRequest) (*emptypb.Empty, error) {
	log.Infof("收到预留库存请求: 订单号=%s, 商品数量=%d", req.OrderSn, len(req.GoodsInfo))

	// 转换商品信息
	goodsInfo := make([]do.GoodsDetail, len(req.GoodsInfo))
	for i, goods := range req.GoodsInfo {
		goodsInfo[i] = do.GoodsDetail{
			Goods: goods.GoodsId,
			Num:   goods.Num,
		}
	}

	reserveReq := &dto.ReserveStockDTO{
		OrderSn:   req.OrderSn,
		GoodsInfo: goodsInfo,
	}

	err := ps.srv.PaymentSrv.ReserveStock(ctx, reserveReq)
	if err != nil {
		log.Errorf("预留库存失败: %v", err)
		return nil, err
	}

	log.Infof("预留库存成功: 订单号=%s", req.OrderSn)
	return &emptypb.Empty{}, nil
}

// ReleaseReserved 释放预留库存（补偿）
func (ps *paymentServer) ReleaseReserved(ctx context.Context, req *pb.ReleaseReservedRequest) (*emptypb.Empty, error) {
	log.Infof("收到释放预留库存请求: 订单号=%s, 商品数量=%d", req.OrderSn, len(req.GoodsInfo))

	// 转换商品信息
	goodsInfo := make([]do.GoodsDetail, len(req.GoodsInfo))
	for i, goods := range req.GoodsInfo {
		goodsInfo[i] = do.GoodsDetail{
			Goods: goods.GoodsId,
			Num:   goods.Num,
		}
	}

	releaseReq := &dto.ReleaseReservedDTO{
		OrderSn:   req.OrderSn,
		GoodsInfo: goodsInfo,
	}

	err := ps.srv.PaymentSrv.ReleaseReserved(ctx, releaseReq)
	if err != nil {
		log.Errorf("释放预留库存失败: %v", err)
		return nil, err
	}

	log.Infof("释放预留库存成功: 订单号=%s", req.OrderSn)
	return &emptypb.Empty{}, nil
}