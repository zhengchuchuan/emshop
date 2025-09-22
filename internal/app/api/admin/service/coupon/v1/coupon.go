package coupon

import (
    "context"
    "time"

    cpbv1 "emshop/api/coupon/v1"
    "emshop/internal/app/api/admin/data"
    "emshop/pkg/log"
)

// CouponSrv 管理端优惠券服务
type CouponSrv interface {
    // EnsureDefaultTemplate 检查是否存在可用模板，不存在则创建一个默认模板并返回
    EnsureDefaultTemplate(ctx context.Context) (*cpbv1.CouponTemplateResponse, error)
}

type couponService struct {
    data data.DataFactory
}

func NewCouponService(d data.DataFactory) CouponSrv {
    return &couponService{data: d}
}

func (s *couponService) EnsureDefaultTemplate(ctx context.Context) (*cpbv1.CouponTemplateResponse, error) {
    // 查询可用模板（status=1 Active），取1条看是否存在
    status := int32(1)
    listReq := &cpbv1.ListCouponTemplatesRequest{
        Status:   &status,
        Page:     1,
        PageSize: 1,
    }
    listResp, err := s.data.Coupon().ListCouponTemplates(ctx, listReq)
    if err != nil {
        return nil, err
    }
    if listResp != nil && listResp.TotalCount > 0 && len(listResp.Items) > 0 {
        // 已有可用模板，返回第一条
        log.Infof("[admin] Found existing coupon template id=%d", listResp.Items[0].Id)
        return listResp.Items[0], nil
    }

    // 创建默认模板（满100减10，长期有效30天，总量不限，每人限领1）
    now := time.Now()
    createReq := &cpbv1.CreateCouponTemplateRequest{
        Name:             "默认满减券",
        Type:             1,      // 满减券
        DiscountType:     1,      // 固定金额
        DiscountValue:    10.0,   // 减10元
        MinOrderAmount:   100.0,  // 满100可用
        MaxDiscountAmount: 0,     // 固定金额无需上限
        TotalCount:       0,      // 0表示不限量
        PerUserLimit:     1,      // 每人限领1
        ValidStartTime:   now.Add(-time.Hour).Unix(),
        ValidEndTime:     now.Add(30 * 24 * time.Hour).Unix(),
        ValidDays:        0,
        Description:      "系统默认创建的满减券模板",
    }
    return s.data.Coupon().CreateCouponTemplate(ctx, createReq)
}

