package v1

import (
	"context"
	"emshop/internal/app/payment/srv/data/v1/interfaces"
	"emshop/internal/app/payment/srv/domain/do"
	"emshop/internal/app/payment/srv/domain/dto"
	"emshop/internal/app/pkg/code"
	"emshop/internal/app/pkg/options"
	"emshop/pkg/errors"
	"emshop/pkg/log"
	"fmt"
	"time"
	"crypto/rand"
	"math/big"
)

// PaymentSrv 支付服务接口
type PaymentSrv interface {
	// Saga正向操作 - 订单提交事务
	CreatePayment(ctx context.Context, req *dto.CreatePaymentDTO) (*dto.PaymentStatusDTO, error)
	CancelPayment(ctx context.Context, paymentSn string) error

	// Saga正向操作 - 支付成功事务  
	ConfirmPayment(ctx context.Context, req *dto.ConfirmPaymentDTO) error
	RefundPayment(ctx context.Context, req *dto.RefundPaymentDTO) error

	// 查询操作
	GetPaymentStatus(ctx context.Context, paymentSn string) (*dto.PaymentStatusDTO, error)

	// 模拟操作（用于演示和测试）
	SimulatePaymentSuccess(ctx context.Context, paymentSn string, thirdPartySn *string) error
	SimulatePaymentFailure(ctx context.Context, paymentSn string) error

	// 库存预留接口（用于分布式事务）
	ReserveStock(ctx context.Context, req *dto.ReserveStockDTO) error
	ReleaseReserved(ctx context.Context, req *dto.ReleaseReservedDTO) error
}

type paymentService struct {
	data         interfaces.DataFactory
	dtmOpts      *options.DtmOptions
	redisOptions *options.RedisOptions
}

// NewPaymentService 创建支付服务实例
func NewPaymentService(data interfaces.DataFactory, dtmOpts *options.DtmOptions, redisOpts *options.RedisOptions) PaymentSrv {
	return &paymentService{
		data:         data,
		dtmOpts:      dtmOpts,
		redisOptions: redisOpts,
	}
}

// generatePaymentSn 生成支付单号
func (ps *paymentService) generatePaymentSn() string {
	timestamp := time.Now().Format("20060102150405")
	n, _ := rand.Int(rand.Reader, big.NewInt(9999))
	return fmt.Sprintf("PAY%s%04d", timestamp, n.Int64())
}

// CreatePayment 创建支付订单
func (ps *paymentService) CreatePayment(ctx context.Context, req *dto.CreatePaymentDTO) (*dto.PaymentStatusDTO, error) {
	log.Infof("创建支付订单: 订单号=%s, 用户ID=%d, 金额=%f", req.OrderSn, req.UserID, req.Amount)

	// 参数验证
	if req.Amount <= 0 {
		return nil, errors.WithCode(code.ErrPaymentAmountInvalid, "支付金额必须大于0")
	}

	// 检查该订单是否已经创建支付订单
	existingPayment, err := ps.data.PaymentOrders().GetByOrderSn(ctx, ps.data.DB(), req.OrderSn)
	if err != nil && !errors.IsCode(err, code.ErrPaymentNotFound) {
		return nil, err
	}
	if existingPayment != nil {
		return nil, errors.WithCode(code.ErrPaymentExists, "该订单已存在支付订单")
	}

	// 生成支付单号
	paymentSn := ps.generatePaymentSn()

	// 设置过期时间
	expiredMinutes := req.ExpiredMinutes
	if expiredMinutes <= 0 {
		expiredMinutes = 15 // 默认15分钟
	}
	expiredAt := time.Now().Add(time.Duration(expiredMinutes) * time.Minute)

	// 创建支付订单
	payment := &do.PaymentOrderDO{
		PaymentSn:     paymentSn,
		OrderSn:       req.OrderSn,
		UserID:        req.UserID,
		Amount:        req.Amount,
		PaymentMethod: req.PaymentMethod,
		PaymentStatus: do.PaymentStatusPending,
		ExpiredAt:     expiredAt,
	}

	// 开启事务
	tx := ps.data.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 创建支付订单
	if err := ps.data.PaymentOrders().Create(ctx, tx, payment); err != nil {
		tx.Rollback()
		return nil, errors.WithCode(code.ErrCreatePaymentFailed, "创建支付订单失败")
	}

	// 记录日志
	paymentLog := &do.PaymentLogDO{
		PaymentSn:    paymentSn,
		Action:       "create",
		StatusFrom:   nil,
		StatusTo:     func() *int32 { s := int32(do.PaymentStatusPending); return &s }(),
		Remark:       "创建支付订单",
		OperatorType: "system",
	}
	if err := ps.data.PaymentLogs().Create(ctx, tx, paymentLog); err != nil {
		log.Warnf("创建支付日志失败: %v", err)
		// 日志失败不影响主流程
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, errors.WithCode(code.ErrConnectDB, "提交事务失败")
	}

	log.Infof("支付订单创建成功: 支付单号=%s", paymentSn)

	// 返回支付状态
	return &dto.PaymentStatusDTO{
		PaymentSn:     paymentSn,
		OrderSn:       req.OrderSn,
		PaymentStatus: do.PaymentStatusPending,
		Amount:        req.Amount,
		PaymentMethod: req.PaymentMethod,
		ExpiredAt:     expiredAt,
	}, nil
}

// CancelPayment 取消支付订单（补偿操作）
func (ps *paymentService) CancelPayment(ctx context.Context, paymentSn string) error {
	log.Infof("取消支付订单: 支付单号=%s", paymentSn)

	// 查询支付订单
	payment, err := ps.data.PaymentOrders().Get(ctx, ps.data.DB(), paymentSn)
	if err != nil {
		if errors.IsCode(err, code.ErrPaymentNotFound) {
			// 支付订单不存在，认为取消成功（幂等性）
			log.Infof("支付订单不存在，取消操作成功: %s", paymentSn)
			return nil
		}
		return err
	}

	// 检查状态
	if payment.PaymentStatus == do.PaymentStatusCancelled {
		// 已经取消，幂等性
		return nil
	}

	if payment.PaymentStatus == do.PaymentStatusPaid {
		return errors.WithCode(code.ErrPaymentCannotCancel, "支付订单已支付，无法取消")
	}

	// 开启事务
	tx := ps.data.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 更新状态为取消
	oldStatus := payment.PaymentStatus
	if err := ps.data.PaymentOrders().UpdateStatus(ctx, tx, paymentSn, do.PaymentStatusCancelled); err != nil {
		tx.Rollback()
		return err
	}

	// 记录日志
	paymentLog := &do.PaymentLogDO{
		PaymentSn:    paymentSn,
		Action:       "cancel",
		StatusFrom:   func() *int32 { s := int32(oldStatus); return &s }(),
		StatusTo:     func() *int32 { s := int32(do.PaymentStatusCancelled); return &s }(),
		Remark:       "DTM补偿操作：取消支付订单",
		OperatorType: "system",
	}
	if err := ps.data.PaymentLogs().Create(ctx, tx, paymentLog); err != nil {
		log.Warnf("创建支付日志失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return errors.WithCode(code.ErrConnectDB, "提交事务失败")
	}

	log.Infof("支付订单取消成功: 支付单号=%s", paymentSn)
	return nil
}

// ConfirmPayment 确认支付成功
func (ps *paymentService) ConfirmPayment(ctx context.Context, req *dto.ConfirmPaymentDTO) error {
	log.Infof("确认支付成功: 支付单号=%s", req.PaymentSn)

	// 查询支付订单
	payment, err := ps.data.PaymentOrders().Get(ctx, ps.data.DB(), req.PaymentSn)
	if err != nil {
		return err
	}

	// 检查状态
	if payment.PaymentStatus == do.PaymentStatusPaid {
		// 已经支付成功，幂等性
		return nil
	}

	if payment.PaymentStatus != do.PaymentStatusPending {
		return errors.WithCode(code.ErrPaymentStatusInvalid, "支付状态无效，无法确认支付")
	}

	// 检查是否过期
	if time.Now().After(payment.ExpiredAt) {
		return errors.WithCode(code.ErrPaymentExpired, "支付已过期")
	}

	// 开启事务
	tx := ps.data.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 更新支付状态和信息
	now := time.Now()
	if err := ps.data.PaymentOrders().UpdateStatus(ctx, tx, req.PaymentSn, do.PaymentStatusPaid); err != nil {
		tx.Rollback()
		return err
	}

	if req.ThirdPartySn != nil {
		if err := ps.data.PaymentOrders().UpdatePaidInfo(ctx, tx, req.PaymentSn, req.ThirdPartySn, &now); err != nil {
			tx.Rollback()
			return err
		}
	}

	// 记录日志
	paymentLog := &do.PaymentLogDO{
		PaymentSn:    req.PaymentSn,
		Action:       "confirm_pay",
		StatusFrom:   func() *int32 { s := int32(do.PaymentStatusPending); return &s }(),
		StatusTo:     func() *int32 { s := int32(do.PaymentStatusPaid); return &s }(),
		Remark:       "确认支付成功",
		OperatorType: "system",
	}
	if err := ps.data.PaymentLogs().Create(ctx, tx, paymentLog); err != nil {
		log.Warnf("创建支付日志失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return errors.WithCode(code.ErrConnectDB, "提交事务失败")
	}

	log.Infof("确认支付成功完成: 支付单号=%s", req.PaymentSn)
	return nil
}

// RefundPayment 退款（补偿操作）
func (ps *paymentService) RefundPayment(ctx context.Context, req *dto.RefundPaymentDTO) error {
	log.Infof("执行退款: 支付单号=%s, 退款金额=%f", req.PaymentSn, req.RefundAmount)

	// 查询支付订单
	payment, err := ps.data.PaymentOrders().Get(ctx, ps.data.DB(), req.PaymentSn)
	if err != nil {
		return err
	}

	// 检查状态
	if payment.PaymentStatus == do.PaymentStatusRefunded {
		// 已经退款，幂等性
		return nil
	}

	if payment.PaymentStatus != do.PaymentStatusPaid {
		return errors.WithCode(code.ErrPaymentStatusInvalid, "只有已支付的订单才能退款")
	}

	// 检查退款金额
	if req.RefundAmount <= 0 || req.RefundAmount > payment.Amount {
		return errors.WithCode(code.ErrPaymentAmountInvalid, "退款金额无效")
	}

	// 开启事务
	tx := ps.data.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 更新状态为退款中
	oldStatus := payment.PaymentStatus
	if err := ps.data.PaymentOrders().UpdateStatus(ctx, tx, req.PaymentSn, do.PaymentStatusRefunding); err != nil {
		tx.Rollback()
		return err
	}

	// 模拟退款处理（实际环境中会调用第三方支付接口）
	// 这里直接标记为退款成功
	if err := ps.data.PaymentOrders().UpdateStatus(ctx, tx, req.PaymentSn, do.PaymentStatusRefunded); err != nil {
		tx.Rollback()
		return err
	}

	// 记录日志
	remark := "DTM补偿操作：退款"
	if req.Reason != nil {
		remark += fmt.Sprintf("，原因：%s", *req.Reason)
	}
	
	paymentLog := &do.PaymentLogDO{
		PaymentSn:    req.PaymentSn,
		Action:       "refund",
		StatusFrom:   func() *int32 { s := int32(oldStatus); return &s }(),
		StatusTo:     func() *int32 { s := int32(do.PaymentStatusRefunded); return &s }(),
		Remark:       remark,
		OperatorType: "system",
	}
	if err := ps.data.PaymentLogs().Create(ctx, tx, paymentLog); err != nil {
		log.Warnf("创建支付日志失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return errors.WithCode(code.ErrConnectDB, "提交事务失败")
	}

	log.Infof("退款成功: 支付单号=%s", req.PaymentSn)
	return nil
}

// GetPaymentStatus 查询支付状态
func (ps *paymentService) GetPaymentStatus(ctx context.Context, paymentSn string) (*dto.PaymentStatusDTO, error) {
	payment, err := ps.data.PaymentOrders().Get(ctx, ps.data.DB(), paymentSn)
	if err != nil {
		return nil, err
	}

	return &dto.PaymentStatusDTO{
		PaymentSn:     payment.PaymentSn,
		OrderSn:       payment.OrderSn,
		PaymentStatus: payment.PaymentStatus,
		Amount:        payment.Amount,
		PaymentMethod: payment.PaymentMethod,
		PaidAt:        payment.PaidAt,
		ExpiredAt:     payment.ExpiredAt,
	}, nil
}

// SimulatePaymentSuccess 模拟支付成功
func (ps *paymentService) SimulatePaymentSuccess(ctx context.Context, paymentSn string, thirdPartySn *string) error {
	log.Infof("模拟支付成功: 支付单号=%s", paymentSn)
	
	return ps.ConfirmPayment(ctx, &dto.ConfirmPaymentDTO{
		PaymentSn:    paymentSn,
		ThirdPartySn: thirdPartySn,
	})
}

// SimulatePaymentFailure 模拟支付失败
func (ps *paymentService) SimulatePaymentFailure(ctx context.Context, paymentSn string) error {
	log.Infof("模拟支付失败: 支付单号=%s", paymentSn)

	// 查询支付订单
	payment, err := ps.data.PaymentOrders().Get(ctx, ps.data.DB(), paymentSn)
	if err != nil {
		return err
	}

	// 检查状态
	if payment.PaymentStatus == do.PaymentStatusFailed {
		// 已经失败，幂等性
		return nil
	}

	if payment.PaymentStatus != do.PaymentStatusPending {
		return errors.WithCode(code.ErrPaymentStatusInvalid, "只有待支付的订单才能模拟支付失败")
	}

	// 开启事务
	tx := ps.data.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 更新状态为失败
	oldStatus := payment.PaymentStatus
	if err := ps.data.PaymentOrders().UpdateStatus(ctx, tx, paymentSn, do.PaymentStatusFailed); err != nil {
		tx.Rollback()
		return err
	}

	// 记录日志
	paymentLog := &do.PaymentLogDO{
		PaymentSn:    paymentSn,
		Action:       "pay_fail",
		StatusFrom:   func() *int32 { s := int32(oldStatus); return &s }(),
		StatusTo:     func() *int32 { s := int32(do.PaymentStatusFailed); return &s }(),
		Remark:       "模拟支付失败",
		OperatorType: "system",
	}
	if err := ps.data.PaymentLogs().Create(ctx, tx, paymentLog); err != nil {
		log.Warnf("创建支付日志失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return errors.WithCode(code.ErrConnectDB, "提交事务失败")
	}

	log.Infof("模拟支付失败完成: 支付单号=%s", paymentSn)
	return nil
}

// ReserveStock 预留库存（用于订单提交事务）
func (ps *paymentService) ReserveStock(ctx context.Context, req *dto.ReserveStockDTO) error {
	log.Infof("预留库存: 订单号=%s, 商品数量=%d", req.OrderSn, len(req.GoodsInfo))

	if len(req.GoodsInfo) == 0 {
		return nil
	}

	// 检查是否已经预留过（幂等性）
	existingReservations, err := ps.data.StockReservations().GetByOrderSn(ctx, ps.data.DB(), req.OrderSn)
	if err != nil && !errors.IsCode(err, code.ErrStockReservationNotFound) {
		return err
	}

	if len(existingReservations) > 0 {
		// 已经预留过，检查状态
		for _, reservation := range existingReservations {
			if reservation.Status == do.StockReservationStatusReserved {
				log.Infof("库存已预留，操作成功: 订单号=%s", req.OrderSn)
				return nil // 幂等性
			}
		}
	}

	// 开启事务
	tx := ps.data.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 创建库存预留记录
	var reservations []*do.StockReservationDO
	for _, goodsInfo := range req.GoodsInfo {
		reservation := &do.StockReservationDO{
			OrderSn:     req.OrderSn,
			GoodsID:     goodsInfo.Goods,
			ReservedNum: goodsInfo.Num,
			Status:      do.StockReservationStatusReserved,
			ReservedAt:  time.Now(),
		}
		reservations = append(reservations, reservation)
	}

	// 批量创建预留记录
	if err := ps.data.StockReservations().BatchCreate(ctx, tx, reservations); err != nil {
		tx.Rollback()
		return errors.WithCode(code.ErrStockReservationFailed, "创建库存预留记录失败")
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return errors.WithCode(code.ErrConnectDB, "提交事务失败")
	}

	log.Infof("库存预留成功: 订单号=%s", req.OrderSn)
	return nil
}

// ReleaseReserved 释放预留库存（补偿操作）
func (ps *paymentService) ReleaseReserved(ctx context.Context, req *dto.ReleaseReservedDTO) error {
	log.Infof("释放预留库存: 订单号=%s, 商品数量=%d", req.OrderSn, len(req.GoodsInfo))

	if len(req.GoodsInfo) == 0 {
		return nil
	}

	// 查询预留记录
	reservations, err := ps.data.StockReservations().GetByOrderSn(ctx, ps.data.DB(), req.OrderSn)
	if err != nil {
		if errors.IsCode(err, code.ErrStockReservationNotFound) {
			// 没有预留记录，认为释放成功（幂等性）
			log.Infof("没有预留记录，释放操作成功: 订单号=%s", req.OrderSn)
			return nil
		}
		return err
	}

	// 检查是否已经释放（幂等性）
	allReleased := true
	for _, reservation := range reservations {
		if reservation.Status != do.StockReservationStatusReleased {
			allReleased = false
			break
		}
	}
	if allReleased {
		log.Infof("库存已释放，操作成功: 订单号=%s", req.OrderSn)
		return nil
	}

	// 开启事务
	tx := ps.data.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 批量更新状态为已释放
	if err := ps.data.StockReservations().BatchUpdateStatus(ctx, tx, req.OrderSn, do.StockReservationStatusReleased); err != nil {
		tx.Rollback()
		return errors.WithCode(code.ErrStockReleaseFailed, "释放库存预留失败")
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return errors.WithCode(code.ErrConnectDB, "提交事务失败")
	}

	log.Infof("释放预留库存成功: 订单号=%s", req.OrderSn)
	return nil
}