package rpc

import (
    "context"
    cpbv1 "emshop/api/coupon/v1"
    "emshop/internal/app/api/admin/data"
    "emshop/pkg/log"
)

type coupon struct {
    cc cpbv1.CouponClient
}

func NewCoupon(cc cpbv1.CouponClient) data.CouponData {
    return &coupon{cc: cc}
}

func (c *coupon) ListCouponTemplates(ctx context.Context, req *cpbv1.ListCouponTemplatesRequest) (*cpbv1.ListCouponTemplatesResponse, error) {
    log.Infof("[admin] ListCouponTemplates with status=%v page=%d pageSize=%d", req.Status, req.Page, req.PageSize)
    resp, err := c.cc.ListCouponTemplates(ctx, req)
    if err != nil {
        log.Errorf("[admin] ListCouponTemplates failed: %v", err)
        return nil, err
    }
    log.Infof("[admin] ListCouponTemplates success, total=%d", resp.TotalCount)
    return resp, nil
}

func (c *coupon) CreateCouponTemplate(ctx context.Context, req *cpbv1.CreateCouponTemplateRequest) (*cpbv1.CouponTemplateResponse, error) {
    log.Infof("[admin] CreateCouponTemplate: name=%s type=%d discountType=%d", req.Name, req.Type, req.DiscountType)
    resp, err := c.cc.CreateCouponTemplate(ctx, req)
    if err != nil {
        log.Errorf("[admin] CreateCouponTemplate failed: %v", err)
        return nil, err
    }
    log.Infof("[admin] CreateCouponTemplate success, id=%d", resp.Id)
    return resp, nil
}

