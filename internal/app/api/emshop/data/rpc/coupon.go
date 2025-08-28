package rpc

import (
	"context"
	cpbv1 "emshop/api/coupon/v1"
	"emshop/internal/app/api/emshop/data"
	"emshop/pkg/log"
)

type coupon struct {
	cc cpbv1.CouponClient
}

func NewCoupon(cc cpbv1.CouponClient) *coupon {
	return &coupon{cc}
}

// ListCouponTemplates 获取优惠券模板列表
func (c *coupon) ListCouponTemplates(ctx context.Context, request *cpbv1.ListCouponTemplatesRequest) (*cpbv1.ListCouponTemplatesResponse, error) {
	log.Infof("Calling ListCouponTemplates gRPC with page: %d, pageSize: %d", request.Page, request.PageSize)
	response, err := c.cc.ListCouponTemplates(ctx, request)
	if err != nil {
		log.Errorf("ListCouponTemplates gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("ListCouponTemplates gRPC call successful, total: %d", response.TotalCount)
	return response, nil
}

// GetCouponTemplate 获取优惠券模板详情
func (c *coupon) GetCouponTemplate(ctx context.Context, request *cpbv1.GetCouponTemplateRequest) (*cpbv1.CouponTemplateResponse, error) {
	log.Infof("Calling GetCouponTemplate gRPC with ID: %d", request.Id)
	response, err := c.cc.GetCouponTemplate(ctx, request)
	if err != nil {
		log.Errorf("GetCouponTemplate gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("GetCouponTemplate gRPC call successful, template name: %s", response.Name)
	return response, nil
}

// ReceiveCoupon 领取优惠券
func (c *coupon) ReceiveCoupon(ctx context.Context, request *cpbv1.ReceiveCouponRequest) (*cpbv1.UserCouponResponse, error) {
	log.Infof("Calling ReceiveCoupon gRPC for user: %d, template: %d", request.UserId, request.CouponTemplateId)
	response, err := c.cc.ReceiveCoupon(ctx, request)
	if err != nil {
		log.Errorf("ReceiveCoupon gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("ReceiveCoupon gRPC call successful, coupon ID: %d, code: %s", response.Id, response.CouponCode)
	return response, nil
}

// GetUserCoupons 获取用户优惠券列表
func (c *coupon) GetUserCoupons(ctx context.Context, request *cpbv1.GetUserCouponsRequest) (*cpbv1.ListUserCouponsResponse, error) {
	log.Infof("Calling GetUserCoupons gRPC for user: %d, page: %d, pageSize: %d", request.UserId, request.Page, request.PageSize)
	response, err := c.cc.GetUserCoupons(ctx, request)
	if err != nil {
		log.Errorf("GetUserCoupons gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("GetUserCoupons gRPC call successful, total: %d", response.TotalCount)
	return response, nil
}

// GetAvailableCoupons 获取用户可用优惠券
func (c *coupon) GetAvailableCoupons(ctx context.Context, request *cpbv1.GetAvailableCouponsRequest) (*cpbv1.ListUserCouponsResponse, error) {
	log.Infof("Calling GetAvailableCoupons gRPC for user: %d, orderAmount: %.2f", request.UserId, request.OrderAmount)
	response, err := c.cc.GetAvailableCoupons(ctx, request)
	if err != nil {
		log.Errorf("GetAvailableCoupons gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("GetAvailableCoupons gRPC call successful, available coupons: %d", response.TotalCount)
	return response, nil
}

// CalculateCouponDiscount 计算优惠券折扣
func (c *coupon) CalculateCouponDiscount(ctx context.Context, request *cpbv1.CalculateCouponDiscountRequest) (*cpbv1.CalculateCouponDiscountResponse, error) {
	log.Infof("Calling CalculateCouponDiscount gRPC for user: %d, orderAmount: %.2f, coupons: %v",
		request.UserId, request.OrderAmount, request.CouponIds)
	response, err := c.cc.CalculateCouponDiscount(ctx, request)
	if err != nil {
		log.Errorf("CalculateCouponDiscount gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("CalculateCouponDiscount gRPC call successful, discountAmount: %.2f, finalAmount: %.2f",
		response.DiscountAmount, response.FinalAmount)
	return response, nil
}

var _ data.CouponData = &coupon{}
