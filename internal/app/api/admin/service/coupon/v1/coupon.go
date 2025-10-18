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
    // ListTemplates 分页查询模板（可按状态筛选）
    ListTemplates(ctx context.Context, req *cpbv1.ListCouponTemplatesRequest) (*cpbv1.ListCouponTemplatesResponse, error)
    // GetTemplate 获取模板详情
    GetTemplate(ctx context.Context, id int64) (*cpbv1.CouponTemplateResponse, error)
    // CreateTemplate 创建模板
    CreateTemplate(ctx context.Context, req *cpbv1.CreateCouponTemplateRequest) (*cpbv1.CouponTemplateResponse, error)
    // UpdateTemplate 更新模板（名称/状态/描述）
    UpdateTemplate(ctx context.Context, req *cpbv1.UpdateCouponTemplateRequest) (*cpbv1.CouponTemplateResponse, error)
    // DeleteTemplate 逻辑删除模板（置为已结束）
    DeleteTemplate(ctx context.Context, id int64) (*cpbv1.CouponTemplateResponse, error)
    // Flash sale
    CreateFlashSaleActivity(ctx context.Context, req *cpbv1.CreateFlashSaleActivityRequest) (*cpbv1.FlashSaleActivityResponse, error)
    GetFlashSaleActivity(ctx context.Context, id int64) (*cpbv1.FlashSaleActivityResponse, error)
    ListFlashSaleActivities(ctx context.Context, req *cpbv1.ListFlashSaleActivitiesRequest) (*cpbv1.ListFlashSaleActivitiesResponse, error)
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

func (s *couponService) ListTemplates(ctx context.Context, req *cpbv1.ListCouponTemplatesRequest) (*cpbv1.ListCouponTemplatesResponse, error) {
    return s.data.Coupon().ListCouponTemplates(ctx, req)
}

func (s *couponService) GetTemplate(ctx context.Context, id int64) (*cpbv1.CouponTemplateResponse, error) {
    return s.data.Coupon().GetCouponTemplate(ctx, &cpbv1.GetCouponTemplateRequest{Id: id})
}

func (s *couponService) CreateTemplate(ctx context.Context, req *cpbv1.CreateCouponTemplateRequest) (*cpbv1.CouponTemplateResponse, error) {
    return s.data.Coupon().CreateCouponTemplate(ctx, req)
}

func (s *couponService) UpdateTemplate(ctx context.Context, req *cpbv1.UpdateCouponTemplateRequest) (*cpbv1.CouponTemplateResponse, error) {
    return s.data.Coupon().UpdateCouponTemplate(ctx, req)
}

func (s *couponService) DeleteTemplate(ctx context.Context, id int64) (*cpbv1.CouponTemplateResponse, error) {
    // 逻辑删除：将状态置为已结束(3)
    status := int32(3)
    return s.data.Coupon().UpdateCouponTemplate(ctx, &cpbv1.UpdateCouponTemplateRequest{Id: id, Status: &status})
}

// ===== Flash sale (admin) =====
func (s *couponService) CreateFlashSaleActivity(ctx context.Context, req *cpbv1.CreateFlashSaleActivityRequest) (*cpbv1.FlashSaleActivityResponse, error) {
    return s.data.Coupon().CreateFlashSaleActivity(ctx, req)
}

func (s *couponService) GetFlashSaleActivity(ctx context.Context, id int64) (*cpbv1.FlashSaleActivityResponse, error) {
    return s.data.Coupon().GetFlashSaleActivity(ctx, &cpbv1.GetFlashSaleActivityRequest{Id: id})
}

func (s *couponService) ListFlashSaleActivities(ctx context.Context, req *cpbv1.ListFlashSaleActivitiesRequest) (*cpbv1.ListFlashSaleActivitiesResponse, error) {
    return s.data.Coupon().ListFlashSaleActivities(ctx, req)
}
