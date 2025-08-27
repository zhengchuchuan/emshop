package v1

import (
	"context"

	couponpb "emshop/api/coupon/v1"
	"emshop/internal/app/coupon/srv/domain/dto"
	v1 "emshop/internal/app/coupon/srv/service/v1"
	"emshop/pkg/log"
	
	"google.golang.org/protobuf/types/known/emptypb"
)

// SubmitOrderWithCoupons 提交订单使用优惠券的分布式事务
func (cs *couponServer) SubmitOrderWithCoupons(ctx context.Context, req *couponpb.SubmitOrderWithCouponsRequest) (*couponpb.SubmitOrderWithCouponsResponse, error) {
	log.Infof("SubmitOrderWithCoupons: 订单=%s, 用户=%d, 优惠券数量=%d", req.OrderSn, req.UserId, len(req.CouponIds))

	// 转换请求
	dtmReq := &v1.OrderCouponSubmissionRequest{
		OrderSn:        req.OrderSn,
		UserID:         req.UserId,
		CouponIDs:      req.CouponIds,
		OriginalAmount: req.OriginalAmount,
		DiscountAmount: req.DiscountAmount,
		FinalAmount:    req.FinalAmount,
		PaymentMethod:  req.PaymentMethod,
		Address:        req.Address,
	}

	// 转换商品详情
	for _, item := range req.GoodsDetails {
		dtmReq.GoodsDetails = append(dtmReq.GoodsDetails, v1.OrderGoodsDetail{
			GoodsID:  item.GoodsId,
			Quantity: item.Quantity,
			Price:    item.Price,
		})
	}

	// 调用DTM管理器
	err := cs.srv.DTMManager.SubmitOrderWithCoupons(ctx, dtmReq)
	if err != nil {
		log.Errorf("分布式事务失败: %v", err)
		return nil, cs.handleError(err)
	}

	return &couponpb.SubmitOrderWithCouponsResponse{
		Success: true,
		Message: "分布式事务提交成功",
	}, nil
}

// ProcessFlashSaleWithInventory 秒杀优惠券与库存协调的分布式事务
func (cs *couponServer) ProcessFlashSaleWithInventory(ctx context.Context, req *couponpb.ProcessFlashSaleWithInventoryRequest) (*couponpb.ProcessFlashSaleWithInventoryResponse, error) {
	log.Infof("ProcessFlashSaleWithInventory: 用户=%d, 秒杀ID=%d, 商品ID=%d", req.UserId, req.FlashSaleId, req.GoodsId)

	// 转换请求
	dtmReq := &v1.FlashSaleInventoryRequest{
		UserID:      req.UserId,
		FlashSaleID: req.FlashSaleId,
		GoodsID:     req.GoodsId,
		Quantity:    req.Quantity,
	}

	// 调用DTM管理器
	err := cs.srv.DTMManager.ProcessFlashSaleWithInventory(ctx, dtmReq)
	if err != nil {
		log.Errorf("秒杀分布式事务失败: %v", err)
		return nil, cs.handleError(err)
	}

	return &couponpb.ProcessFlashSaleWithInventoryResponse{
		Success: true,
		Message: "秒杀分布式事务成功",
	}, nil
}

// TryFlashSale TCC Try阶段：预占秒杀优惠券 (DTM回调)
func (cs *couponServer) TryFlashSale(ctx context.Context, req *couponpb.ParticipateFlashSaleRequest) (*emptypb.Empty, error) {
	log.Infof("DTM TryFlashSale: 用户=%d, 秒杀ID=%d", req.UserId, req.FlashSaleId)

	dto := &dto.ParticipateFlashSaleDTO{
		UserID:      req.UserId,
		FlashSaleID: req.FlashSaleId,
	}

	return cs.srv.DTMManager.TryFlashSale(ctx, dto)
}

// ConfirmFlashSale TCC Confirm阶段：确认秒杀优惠券扣减 (DTM回调)
func (cs *couponServer) ConfirmFlashSale(ctx context.Context, req *couponpb.ParticipateFlashSaleRequest) (*emptypb.Empty, error) {
	log.Infof("DTM ConfirmFlashSale: 用户=%d, 秒杀ID=%d", req.UserId, req.FlashSaleId)

	dto := &dto.ParticipateFlashSaleDTO{
		UserID:      req.UserId,
		FlashSaleID: req.FlashSaleId,
	}

	return cs.srv.DTMManager.ConfirmFlashSale(ctx, dto)
}

// CancelFlashSale TCC Cancel阶段：取消秒杀优惠券预占 (DTM回调)
func (cs *couponServer) CancelFlashSale(ctx context.Context, req *couponpb.ParticipateFlashSaleRequest) (*emptypb.Empty, error) {
	log.Infof("DTM CancelFlashSale: 用户=%d, 秒杀ID=%d", req.UserId, req.FlashSaleId)

	dto := &dto.ParticipateFlashSaleDTO{
		UserID:      req.UserId,
		FlashSaleID: req.FlashSaleId,
	}

	return cs.srv.DTMManager.CancelFlashSale(ctx, dto)
}

// GetTransactionStatus 获取分布式事务状态
func (cs *couponServer) GetTransactionStatus(ctx context.Context, req *couponpb.GetTransactionStatusRequest) (*couponpb.GetTransactionStatusResponse, error) {
	log.Infof("GetTransactionStatus: GID=%s", req.Gid)

	status, err := cs.srv.DTMManager.GetTransactionStatus(ctx, req.Gid)
	if err != nil {
		return nil, cs.handleError(err)
	}

	return &couponpb.GetTransactionStatusResponse{
		Gid:       status.GID,
		Status:    status.Status,
		CreatedAt: status.CreatedAt.Unix(),
		UpdatedAt: status.UpdatedAt.Unix(),
	}, nil
}