package v1

import (
	"context"

	"emshop/internal/app/api/emshop/data"
	"emshop/internal/app/api/emshop/domain/dto/request"
	"emshop/internal/app/api/emshop/domain/dto/response"
)

// CouponSrv 优惠券服务接口
type CouponSrv interface {
	// ListTemplates 获取优惠券模板列表
	ListTemplates(ctx context.Context, req *request.ListCouponTemplatesRequest) (*response.CouponTemplateListResponse, error)

	// ReceiveCoupon 用户领取优惠券
	ReceiveCoupon(ctx context.Context, userID int64, req *request.ReceiveCouponRequest) (*response.ReceiveCouponResponse, error)

	// GetUserCoupons 获取用户优惠券列表
	GetUserCoupons(ctx context.Context, userID int64, req *request.GetUserCouponsRequest) (*response.UserCouponListResponse, error)

	// GetAvailableCoupons 获取用户可用优惠券
	GetAvailableCoupons(ctx context.Context, userID int64, req *request.GetAvailableCouponsRequest) (*response.AvailableCouponsResponse, error)

	// CalculateDiscount 计算优惠券折扣
	CalculateDiscount(ctx context.Context, userID int64, req *request.CalculateCouponDiscountRequest) (*response.CouponDiscountResponse, error)
}

// couponService 优惠券服务实现
type couponService struct {
	data data.DataFactory
}

// NewCouponService 创建优惠券服务实例
func NewCouponService(data data.DataFactory) CouponSrv {
	return &couponService{
		data: data,
	}
}

// ListTemplates 获取优惠券模板列表
func (s *couponService) ListTemplates(ctx context.Context, req *request.ListCouponTemplatesRequest) (*response.CouponTemplateListResponse, error) {
	// 参数验证
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 调用RPC服务
	rpcReq := req.ToProto()
	rpcResp, err := s.data.Coupon().ListCouponTemplates(ctx, rpcReq)
	if err != nil {
		return nil, err
	}

	// 转换响应
	resp := &response.CouponTemplateListResponse{}
	resp.FromProto(rpcResp)

	return resp, nil
}

// ReceiveCoupon 用户领取优惠券
func (s *couponService) ReceiveCoupon(ctx context.Context, userID int64, req *request.ReceiveCouponRequest) (*response.ReceiveCouponResponse, error) {
	// 参数验证
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 调用RPC服务
	rpcReq := req.ToProto(userID)
	rpcResp, err := s.data.Coupon().ReceiveCoupon(ctx, rpcReq)
	if err != nil {
		return nil, err
	}

	// 转换响应
	resp := &response.ReceiveCouponResponse{}
	resp.FromProto(rpcResp)

	return resp, nil
}

// GetUserCoupons 获取用户优惠券列表
func (s *couponService) GetUserCoupons(ctx context.Context, userID int64, req *request.GetUserCouponsRequest) (*response.UserCouponListResponse, error) {
	// 参数验证
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 调用RPC服务
	rpcReq := req.ToProto(userID)
	rpcResp, err := s.data.Coupon().GetUserCoupons(ctx, rpcReq)
	if err != nil {
		return nil, err
	}

	// 转换响应
	resp := &response.UserCouponListResponse{}
	resp.FromProto(rpcResp)

	return resp, nil
}

// GetAvailableCoupons 获取用户可用优惠券
func (s *couponService) GetAvailableCoupons(ctx context.Context, userID int64, req *request.GetAvailableCouponsRequest) (*response.AvailableCouponsResponse, error) {
	// 参数验证
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 调用RPC服务
	rpcReq := req.ToProto(userID)
	rpcResp, err := s.data.Coupon().GetAvailableCoupons(ctx, rpcReq)
	if err != nil {
		return nil, err
	}

	// 转换响应
	resp := &response.AvailableCouponsResponse{}
	resp.FromProto(rpcResp)

	return resp, nil
}

// CalculateDiscount 计算优惠券折扣
func (s *couponService) CalculateDiscount(ctx context.Context, userID int64, req *request.CalculateCouponDiscountRequest) (*response.CouponDiscountResponse, error) {
	// 参数验证
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 调用RPC服务
	rpcReq := req.ToProto(userID)
	rpcResp, err := s.data.Coupon().CalculateCouponDiscount(ctx, rpcReq)
	if err != nil {
		return nil, err
	}

	// 转换响应
	resp := &response.CouponDiscountResponse{}
	resp.FromProto(rpcResp)

	return resp, nil
}

// 编译时检查接口实现
var _ CouponSrv = (*couponService)(nil)
