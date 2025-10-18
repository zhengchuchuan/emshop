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

func (c *coupon) CreateFlashSaleActivity(ctx context.Context, req *cpbv1.CreateFlashSaleActivityRequest) (*cpbv1.FlashSaleActivityResponse, error) {
    log.Infof("[admin] CreateFlashSaleActivity: name=%s tpl=%d", req.Name, req.CouponTemplateId)
    resp, err := c.cc.CreateFlashSaleActivity(ctx, req)
    if err != nil {
        log.Errorf("[admin] CreateFlashSaleActivity failed: %v", err)
        return nil, err
    }
    log.Infof("[admin] CreateFlashSaleActivity success, id=%d", resp.Id)
    return resp, nil
}

func (c *coupon) GetFlashSaleActivity(ctx context.Context, req *cpbv1.GetFlashSaleActivityRequest) (*cpbv1.FlashSaleActivityResponse, error) {
    log.Infof("[admin] GetFlashSaleActivity: id=%d", req.Id)
    resp, err := c.cc.GetFlashSaleActivity(ctx, req)
    if err != nil {
        log.Errorf("[admin] GetFlashSaleActivity failed: %v", err)
        return nil, err
    }
    return resp, nil
}

func (c *coupon) ListFlashSaleActivities(ctx context.Context, req *cpbv1.ListFlashSaleActivitiesRequest) (*cpbv1.ListFlashSaleActivitiesResponse, error) {
    log.Infof("[admin] ListFlashSaleActivities: status=%v page=%d", req.Status, req.Page)
    resp, err := c.cc.ListFlashSaleActivities(ctx, req)
    if err != nil {
        log.Errorf("[admin] ListFlashSaleActivities failed: %v", err)
        return nil, err
    }
    return resp, nil
}

func (c *coupon) GetCouponTemplate(ctx context.Context, req *cpbv1.GetCouponTemplateRequest) (*cpbv1.CouponTemplateResponse, error) {
    log.Infof("[admin] GetCouponTemplate: id=%d", req.Id)
    resp, err := c.cc.GetCouponTemplate(ctx, req)
    if err != nil {
        log.Errorf("[admin] GetCouponTemplate failed: %v", err)
        return nil, err
    }
    log.Infof("[admin] GetCouponTemplate success, id=%d", resp.Id)
    return resp, nil
}

func (c *coupon) UpdateCouponTemplate(ctx context.Context, req *cpbv1.UpdateCouponTemplateRequest) (*cpbv1.CouponTemplateResponse, error) {
    log.Infof("[admin] UpdateCouponTemplate: id=%d", req.Id)
    resp, err := c.cc.UpdateCouponTemplate(ctx, req)
    if err != nil {
        log.Errorf("[admin] UpdateCouponTemplate failed: %v", err)
        return nil, err
    }
    log.Infof("[admin] UpdateCouponTemplate success, id=%d", resp.Id)
    return resp, nil
}
