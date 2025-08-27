package v1

import (
	"context"
	"time"

	couponpb "emshop/api/coupon/v1"
	"emshop/internal/app/coupon/srv/domain/dto"
	v1 "emshop/internal/app/coupon/srv/service/v1"
	"emshop/internal/app/pkg/code"
	"emshop/pkg/errors"
	"emshop/pkg/log"
	
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type couponServer struct {
	couponpb.UnimplementedCouponServer
	srv *v1.Service
}

// NewCouponServer 创建优惠券gRPC服务器
func NewCouponServer(srv *v1.Service) couponpb.CouponServer {
	return &couponServer{
		srv: srv,
	}
}

// CreateCouponTemplate 创建优惠券模板
func (cs *couponServer) CreateCouponTemplate(ctx context.Context, req *couponpb.CreateCouponTemplateRequest) (*couponpb.CouponTemplateResponse, error) {
	log.Infof("CreateCouponTemplate: %s", req.Name)

	// 转换请求
	dto := &dto.CreateCouponTemplateDTO{
		Name:              req.Name,
		Type:              req.Type,
		DiscountType:      req.DiscountType,
		DiscountValue:     req.DiscountValue,
		MinOrderAmount:    req.MinOrderAmount,
		MaxDiscountAmount: req.MaxDiscountAmount,
		TotalCount:        req.TotalCount,
		PerUserLimit:      req.PerUserLimit,
		ValidStartTime:    time.Unix(req.ValidStartTime, 0),
		ValidEndTime:      time.Unix(req.ValidEndTime, 0),
		ValidDays:         req.ValidDays,
		Description:       req.Description,
	}

	result, err := cs.srv.CouponSrv.CreateCouponTemplate(ctx, dto)
	if err != nil {
		return nil, cs.handleError(err)
	}

	return cs.convertTemplateToProto(result), nil
}

// GetCouponTemplate 获取优惠券模板
func (cs *couponServer) GetCouponTemplate(ctx context.Context, req *couponpb.GetCouponTemplateRequest) (*couponpb.CouponTemplateResponse, error) {
	result, err := cs.srv.CouponSrv.GetCouponTemplate(ctx, req.Id)
	if err != nil {
		return nil, cs.handleError(err)
	}

	return cs.convertTemplateToProto(result), nil
}

// UpdateCouponTemplate 更新优惠券模板
func (cs *couponServer) UpdateCouponTemplate(ctx context.Context, req *couponpb.UpdateCouponTemplateRequest) (*couponpb.CouponTemplateResponse, error) {
	dto := &dto.UpdateCouponTemplateDTO{
		ID: req.Id,
	}
	
	if req.Name != nil {
		dto.Name = req.Name
	}
	if req.Status != nil {
		dto.Status = req.Status
	}
	if req.Description != nil {
		dto.Description = req.Description
	}

	result, err := cs.srv.CouponSrv.UpdateCouponTemplate(ctx, dto)
	if err != nil {
		return nil, cs.handleError(err)
	}

	return cs.convertTemplateToProto(result), nil
}

// ListCouponTemplates 获取优惠券模板列表
func (cs *couponServer) ListCouponTemplates(ctx context.Context, req *couponpb.ListCouponTemplatesRequest) (*couponpb.ListCouponTemplatesResponse, error) {
	dto := &dto.ListCouponTemplatesDTO{
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	
	if req.Status != nil {
		dto.Status = req.Status
	}

	result, err := cs.srv.CouponSrv.ListCouponTemplates(ctx, dto)
	if err != nil {
		return nil, cs.handleError(err)
	}

	// 转换响应
	items := make([]*couponpb.CouponTemplateResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, cs.convertTemplateToProto(item))
	}

	return &couponpb.ListCouponTemplatesResponse{
		TotalCount: result.TotalCount,
		Items:      items,
	}, nil
}

// ReceiveCoupon 领取优惠券
func (cs *couponServer) ReceiveCoupon(ctx context.Context, req *couponpb.ReceiveCouponRequest) (*couponpb.UserCouponResponse, error) {
	log.Infof("ReceiveCoupon: userID=%d, templateID=%d", req.UserId, req.CouponTemplateId)

	dto := &dto.ReceiveCouponDTO{
		UserID:           req.UserId,
		CouponTemplateID: req.CouponTemplateId,
	}

	result, err := cs.srv.CouponSrv.ReceiveCoupon(ctx, dto)
	if err != nil {
		return nil, cs.handleError(err)
	}

	return cs.convertUserCouponToProto(result), nil
}

// GetUserCoupons 获取用户优惠券列表
func (cs *couponServer) GetUserCoupons(ctx context.Context, req *couponpb.GetUserCouponsRequest) (*couponpb.ListUserCouponsResponse, error) {
	dto := &dto.GetUserCouponsDTO{
		UserID:   req.UserId,
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	
	if req.Status != nil {
		dto.Status = req.Status
	}

	result, err := cs.srv.CouponSrv.GetUserCoupons(ctx, dto)
	if err != nil {
		return nil, cs.handleError(err)
	}

	// 转换响应
	items := make([]*couponpb.UserCouponResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, cs.convertUserCouponToProto(item))
	}

	return &couponpb.ListUserCouponsResponse{
		TotalCount: result.TotalCount,
		Items:      items,
	}, nil
}

// GetAvailableCoupons 获取用户可用优惠券
func (cs *couponServer) GetAvailableCoupons(ctx context.Context, req *couponpb.GetAvailableCouponsRequest) (*couponpb.ListUserCouponsResponse, error) {
	dto := &dto.GetAvailableCouponsDTO{
		UserID:      req.UserId,
		OrderAmount: req.OrderAmount,
	}

	result, err := cs.srv.CouponSrv.GetAvailableCoupons(ctx, dto)
	if err != nil {
		return nil, cs.handleError(err)
	}

	// 转换响应
	items := make([]*couponpb.UserCouponResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, cs.convertUserCouponToProto(item))
	}

	return &couponpb.ListUserCouponsResponse{
		TotalCount: result.TotalCount,
		Items:      items,
	}, nil
}

// CalculateCouponDiscount 计算优惠券折扣
func (cs *couponServer) CalculateCouponDiscount(ctx context.Context, req *couponpb.CalculateCouponDiscountRequest) (*couponpb.CalculateCouponDiscountResponse, error) {
	// 转换订单项
	orderItems := make([]*dto.OrderItemDTO, 0, len(req.OrderItems))
	for _, item := range req.OrderItems {
		orderItems = append(orderItems, &dto.OrderItemDTO{
			GoodsID:  item.GoodsId,
			Quantity: item.Quantity,
			Price:    item.Price,
		})
	}

	dto := &dto.CalculateCouponDiscountDTO{
		UserID:      req.UserId,
		CouponIDs:   req.CouponIds,
		OrderAmount: req.OrderAmount,
		OrderItems:  orderItems,
	}

	result, err := cs.srv.CouponSrv.CalculateCouponDiscount(ctx, dto)
	if err != nil {
		return nil, cs.handleError(err)
	}

	// 转换拒绝原因
	rejectedCoupons := make([]*couponpb.CouponRejection, 0, len(result.RejectedCoupons))
	for _, rejection := range result.RejectedCoupons {
		rejectedCoupons = append(rejectedCoupons, &couponpb.CouponRejection{
			CouponId: rejection.CouponID,
			Reason:   rejection.Reason,
		})
	}

	return &couponpb.CalculateCouponDiscountResponse{
		OriginalAmount:  result.OriginalAmount,
		DiscountAmount:  result.DiscountAmount,
		FinalAmount:     result.FinalAmount,
		AppliedCoupons:  result.AppliedCoupons,
		RejectedCoupons: rejectedCoupons,
	}, nil
}

// UseCoupons 使用优惠券
func (cs *couponServer) UseCoupons(ctx context.Context, req *couponpb.UseCouponsRequest) (*couponpb.UseCouponsResponse, error) {
	log.Infof("UseCoupons: orderSn=%s, userID=%d", req.OrderSn, req.UserId)

	dto := &dto.UseCouponsDTO{
		UserID:      req.UserId,
		OrderSn:     req.OrderSn,
		CouponIDs:   req.CouponIds,
		OrderAmount: req.OrderAmount,
	}

	result, err := cs.srv.CouponSrv.UseCoupons(ctx, dto)
	if err != nil {
		return nil, cs.handleError(err)
	}

	return &couponpb.UseCouponsResponse{
		DiscountAmount: result.DiscountAmount,
		UsedCoupons:    result.UsedCoupons,
	}, nil
}

// ReleaseCoupons 释放优惠券
func (cs *couponServer) ReleaseCoupons(ctx context.Context, req *couponpb.ReleaseCouponsRequest) (*emptypb.Empty, error) {
	log.Infof("ReleaseCoupons: orderSn=%s", req.OrderSn)

	dto := &dto.ReleaseCouponsDTO{
		OrderSn: req.OrderSn,
	}

	err := cs.srv.CouponSrv.ReleaseCoupons(ctx, dto)
	if err != nil {
		return nil, cs.handleError(err)
	}

	return &emptypb.Empty{}, nil
}

// convertTemplateToProto 转换模板DTO为Protobuf
func (cs *couponServer) convertTemplateToProto(dto *dto.CouponTemplateDTO) *couponpb.CouponTemplateResponse {
	return &couponpb.CouponTemplateResponse{
		Id:                dto.ID,
		Name:              dto.Name,
		Type:              dto.Type,
		DiscountType:      dto.DiscountType,
		DiscountValue:     dto.DiscountValue,
		MinOrderAmount:    dto.MinOrderAmount,
		MaxDiscountAmount: dto.MaxDiscountAmount,
		TotalCount:        dto.TotalCount,
		UsedCount:         dto.UsedCount,
		PerUserLimit:      dto.PerUserLimit,
		ValidStartTime:    dto.ValidStartTime.Unix(),
		ValidEndTime:      dto.ValidEndTime.Unix(),
		ValidDays:         dto.ValidDays,
		Status:            dto.Status,
		Description:       dto.Description,
		CreatedAt:         dto.CreatedAt.Unix(),
	}
}

// convertUserCouponToProto 转换用户优惠券DTO为Protobuf
func (cs *couponServer) convertUserCouponToProto(dto *dto.UserCouponDTO) *couponpb.UserCouponResponse {
	resp := &couponpb.UserCouponResponse{
		Id:               dto.ID,
		CouponTemplateId: dto.CouponTemplateID,
		UserId:           dto.UserID,
		CouponCode:       dto.CouponCode,
		Status:           dto.Status,
		ReceivedAt:       dto.ReceivedAt.Unix(),
		ExpiredAt:        dto.ExpiredAt.Unix(),
	}

	if dto.OrderSn != nil {
		resp.OrderSn = dto.OrderSn
	}
	if dto.UsedAt != nil {
		usedAt := dto.UsedAt.Unix()
		resp.UsedAt = &usedAt
	}
	if dto.Template != nil {
		resp.Template = cs.convertTemplateToProto(dto.Template)
	}

	return resp
}

// handleError 处理错误并转换为gRPC错误
func (cs *couponServer) handleError(err error) error {
	if err == nil {
		return nil
	}

	log.Errorf("gRPC error: %v", err)

	// 根据业务错误码转换为gRPC状态码
	if errors.IsCode(err, code.ErrResourceNotFound) {
		return status.Errorf(codes.NotFound, err.Error())
	}
	if errors.IsCode(err, code.ErrInvalidRequest) {
		return status.Errorf(codes.InvalidArgument, err.Error())
	}
	if errors.IsCode(err, code.ErrResourceNotAvailable) {
		return status.Errorf(codes.Unavailable, err.Error())
	}
	if errors.IsCode(err, code.ErrResourceLimitExceeded) {
		return status.Errorf(codes.ResourceExhausted, err.Error())
	}
	if errors.IsCode(err, code.ErrDatabase) {
		return status.Errorf(codes.Internal, "数据库错误")
	}
	if errors.IsCode(err, code.ErrRedis) {
		return status.Errorf(codes.Internal, "缓存错误")
	}

	// 默认返回内部错误
	return status.Errorf(codes.Internal, "内部服务错误")
}