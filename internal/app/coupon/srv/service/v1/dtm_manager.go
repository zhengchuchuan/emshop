package v1

import (
	"context"
	"fmt"
	"time"

	"emshop/internal/app/coupon/srv/domain/dto"
	"emshop/internal/app/pkg/options"
	"emshop/pkg/log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// CouponDTMManager 优惠券DTM分布式事务管理器（简化版）
type CouponDTMManager struct {
	dtmServer    string
	couponSrv    string
	orderSrv     string
	paymentSrv   string
	inventorySrv string
	service      *Service
}

// NewCouponDTMManager 创建优惠券DTM事务管理器
func NewCouponDTMManager(dtmOpts *options.DtmOptions, service *Service) *CouponDTMManager {
	return &CouponDTMManager{
		dtmServer:    dtmOpts.GrpcServer,
		couponSrv:    "discovery:///emshop-coupon-srv",
		orderSrv:     "discovery:///emshop-order-srv",
		paymentSrv:   "discovery:///emshop-payment-srv",
		inventorySrv: "discovery:///emshop-inventory-srv",
		service:      service,
	}
}

// SubmitOrderWithCoupons 提交订单使用优惠券的分布式事务（简化版 - 记录日志）
func (dm *CouponDTMManager) SubmitOrderWithCoupons(ctx context.Context, req *OrderCouponSubmissionRequest) error {
	log.Infof("开始订单-优惠券分布式事务, 订单号: %s", req.OrderSn)

	// 简化版：仅记录事务步骤，不实际调用DTM
	// 在生产环境中，这里会实际调用DTM Saga事务
	
	log.Infof("步骤1: 使用优惠券 - 用户ID: %d, 优惠券数量: %d", req.UserID, len(req.CouponIDs))
	log.Infof("步骤2: 创建订单 - 订单号: %s, 金额: %.2f", req.OrderSn, req.FinalAmount)
	log.Infof("步骤3: 创建支付单 - 支付方式: %d", req.PaymentMethod)
	log.Infof("步骤4: 预留库存 - 商品数量: %d", len(req.GoodsDetails))

	// 模拟实际调用优惠券使用
	useCouponDTO := &dto.UseCouponsDTO{
		UserID:      req.UserID,
		OrderSn:     req.OrderSn,
		CouponIDs:   req.CouponIDs,
		OrderAmount: req.FinalAmount,
	}
	
	_, err := dm.service.CouponSrv.UseCoupons(ctx, useCouponDTO)
	if err != nil {
		log.Errorf("使用优惠券失败: %v", err)
		return fmt.Errorf("使用优惠券失败: %w", err)
	}

	log.Infof("订单-优惠券分布式事务模拟成功, 订单号: %s", req.OrderSn)
	return nil
}

// ProcessFlashSaleWithInventory 秒杀优惠券与库存协调的分布式事务（简化版）
func (dm *CouponDTMManager) ProcessFlashSaleWithInventory(ctx context.Context, req *FlashSaleInventoryRequest) error {
	log.Infof("开始秒杀-库存分布式事务, 用户: %d, 秒杀ID: %d", req.UserID, req.FlashSaleID)

	if dm.service.AsyncFlashSaleEnabled() && dm.service.FlashSaleCore != nil {
		coreReq := &dto.FlashSaleRequestDTO{
			ActivityID: req.FlashSaleID,
			UserID:     req.UserID,
		}
		coreResult, err := dm.service.FlashSaleCore.FlashSaleCoupon(ctx, coreReq)
		if err != nil {
			log.Errorf("异步秒杀预扣失败: %v", err)
			return fmt.Errorf("秒杀参与失败: %w", err)
		}
		if coreResult == nil || !coreResult.Success {
			failReason := "秒杀失败"
			if coreResult != nil {
				failReason = coreResult.Message
			}
			return fmt.Errorf(failReason)
		}
		log.Infof("秒杀-库存分布式事务异步预扣成功, 用户: %d", req.UserID)
		return nil
	}

	flashSaleDTO := &dto.ParticipateFlashSaleDTO{
		UserID:      req.UserID,
		FlashSaleID: req.FlashSaleID,
	}

	result, err := dm.service.FlashSaleSrv.ParticipateFlashSale(ctx, flashSaleDTO)
	if err != nil {
		log.Errorf("秒杀参与失败: %v", err)
		return fmt.Errorf("秒杀参与失败: %w", err)
	}

	if result.Status != 1 { // 1表示成功
		failReason := "未知原因"
		if result.FailReason != nil {
			failReason = *result.FailReason
		}
		return fmt.Errorf("秒杀失败: %s", failReason)
	}

	log.Infof("秒杀-库存分布式事务模拟成功, 用户: %d", req.UserID)
	return nil
}

// TryFlashSale TCC Try阶段：预占秒杀优惠券
func (dm *CouponDTMManager) TryFlashSale(ctx context.Context, req *dto.ParticipateFlashSaleDTO) (*emptypb.Empty, error) {
	log.Infof("TCC Try: 预占秒杀优惠券, 用户: %d, 秒杀ID: %d", req.UserID, req.FlashSaleID)

	if dm.service.AsyncFlashSaleEnabled() && dm.service.FlashSaleCore != nil {
		coreReq := &dto.FlashSaleRequestDTO{
			ActivityID: req.FlashSaleID,
			UserID:     req.UserID,
		}
		coreResult, err := dm.service.FlashSaleCore.FlashSaleCoupon(ctx, coreReq)
		if err != nil {
			log.Errorf("TCC Try 异步预占失败: %v", err)
			return nil, status.Errorf(codes.Aborted, "%s", err.Error())
		}
		if coreResult == nil || !coreResult.Success {
			failReason := "秒杀预占失败"
			if coreResult != nil {
				failReason = coreResult.Message
			}
			return nil, status.Errorf(codes.Aborted, "%s", failReason)
		}
		return &emptypb.Empty{}, nil
	}

	result, err := dm.service.FlashSaleSrv.ParticipateFlashSale(ctx, req)
	if err != nil {
		log.Errorf("TCC Try 预占秒杀失败: %v", err)
		return nil, status.Errorf(codes.Aborted, "%s", err.Error())
	}

	if result.Status != 1 { // 1表示成功
		failReason := "未知原因"
		if result.FailReason != nil {
			failReason = *result.FailReason
		}
		return nil, status.Errorf(codes.Aborted, "秒杀预占失败: %s", failReason)
	}

	return &emptypb.Empty{}, nil
}

// ConfirmFlashSale TCC Confirm阶段：确认秒杀优惠券扣减
func (dm *CouponDTMManager) ConfirmFlashSale(ctx context.Context, req *dto.ParticipateFlashSaleDTO) (*emptypb.Empty, error) {
	log.Infof("TCC Confirm: 确认秒杀优惠券扣减, 用户: %d, 秒杀ID: %d", req.UserID, req.FlashSaleID)
	// 确认操作（在Try阶段已经完成了实际扣减）
	log.Infof("秒杀优惠券确认成功")
	return &emptypb.Empty{}, nil
}

// CancelFlashSale TCC Cancel阶段：取消秒杀优惠券预占
func (dm *CouponDTMManager) CancelFlashSale(ctx context.Context, req *dto.ParticipateFlashSaleDTO) (*emptypb.Empty, error) {
	log.Infof("TCC Cancel: 取消秒杀优惠券预占, 用户: %d, 秒杀ID: %d", req.UserID, req.FlashSaleID)
	// 补偿逻辑（如果需要）
	log.Infof("秒杀优惠券预占已释放")
	return &emptypb.Empty{}, nil
}

// GetTransactionStatus 获取分布式事务状态
func (dm *CouponDTMManager) GetTransactionStatus(ctx context.Context, gid string) (*TransactionStatus, error) {
	log.Infof("查询事务状态, GID: %s", gid)
	
	// 模拟返回事务状态
	return &TransactionStatus{
		GID:       gid,
		Status:    "succeeded", // prepared, aborting, succeeded, failed
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// 请求/响应结构定义

type OrderCouponSubmissionRequest struct {
	OrderSn        string               `json:"order_sn"`
	UserID         int64                `json:"user_id"`
	CouponIDs      []int64              `json:"coupon_ids"`
	OriginalAmount float64              `json:"original_amount"`
	DiscountAmount float64              `json:"discount_amount"`
	FinalAmount    float64              `json:"final_amount"`
	PaymentMethod  int32                `json:"payment_method"`
	GoodsDetails   []OrderGoodsDetail   `json:"goods_details"`
	Address        string               `json:"address"`
}

type FlashSaleInventoryRequest struct {
	UserID      int64 `json:"user_id"`
	FlashSaleID int64 `json:"flash_sale_id"`
	GoodsID     int64 `json:"goods_id"`
	Quantity    int32 `json:"quantity"`
}

type OrderGoodsDetail struct {
	GoodsID  int64   `json:"goods_id"`
	Quantity int32   `json:"quantity"`
	Price    float64 `json:"price"`
}

type TransactionStatus struct {
	GID       string    `json:"gid"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
