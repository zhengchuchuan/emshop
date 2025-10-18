package v1

import (
    "context"
    "testing"
    "time"

    "emshop/internal/app/coupon/srv/domain/do"
    "emshop/internal/app/coupon/srv/domain/dto"
    "emshop/internal/app/coupon/srv/pkg/cache"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "gorm.io/gorm"
)

func TestCreateCouponTemplate_InvalidWindow(t *testing.T) {
    svc := NewCouponService(nil, nil, nil, nil)
    req := &dto.CreateCouponTemplateDTO{
        Name:           "invalid-window",
        Type:           int32(do.CouponTypeThreshold),
        DiscountType:   int32(do.DiscountTypeFixed),
        DiscountValue:  10,
        MinOrderAmount: 50,
        TotalCount:     100,
        PerUserLimit:   1,
        ValidStartTime: time.Now().Add(2 * time.Hour),
        ValidEndTime:   time.Now(),
    }
    _, err := svc.CreateCouponTemplate(context.Background(), req)
    assert.Error(t, err)
}

func TestGetCouponTemplate_FromCache(t *testing.T) {
    cm := &MockCacheManager{}
    now := time.Now()
    cached := &cache.CouponTemplate{
        ID: 101, Name: "满减-10", Type: int32(do.CouponTypeThreshold), DiscountType: int32(do.DiscountTypeFixed),
        DiscountValue: 10, MinAmount: 50, TotalCount: 100, UsedCount: 0,
        ValidStart: now.Add(-time.Hour), ValidEnd: now.Add(time.Hour), Status: int32(do.CouponStatusActive),
    }
    cm.On("GetCouponTemplate", mock.Anything, int64(101)).Return(cached, nil)
    svc := NewCouponService(nil, nil, nil, cm)
    dtoTpl, err := svc.GetCouponTemplate(context.Background(), 101)
    assert.NoError(t, err)
    assert.NotNil(t, dtoTpl)
    assert.Equal(t, int64(101), dtoTpl.ID)
    assert.Equal(t, "满减-10", dtoTpl.Name)
}

func TestGetCouponTemplate_DBOnMiss(t *testing.T) {
    cm := &MockCacheManager{}
    cm.On("GetCouponTemplate", mock.Anything, int64(202)).Return((*cache.CouponTemplate)(nil), nil)
    mockData := &MockDataFactory{}
    mockTplRepo := &MockCouponTemplateData{}
    mockData.On("CouponTemplates").Return(mockTplRepo)
    mockData.On("DB").Return(&gorm.DB{})

    now := time.Now()
    tplDO := &do.CouponTemplateDO{ID: 202, Name: "折扣-8折", Type: do.CouponTypeDiscount, DiscountType: do.DiscountTypePercent,
        DiscountValue: 20, MinOrderAmount: 0, MaxDiscountAmount: 0, TotalCount: 100, UsedCount: 0, PerUserLimit: 1,
        ValidStartTime: now.Add(-time.Hour), ValidEndTime: now.Add(2 * time.Hour), Status: do.CouponStatusActive,
    }
    mockTplRepo.On("Get", mock.Anything, mock.Anything, int64(202)).Return(tplDO, nil)
    svc := NewCouponService(mockData, nil, nil, cm)
    got, err := svc.GetCouponTemplate(context.Background(), 202)
    assert.NoError(t, err)
    assert.NotNil(t, got)
    assert.Equal(t, int64(202), got.ID)
    assert.Equal(t, "折扣-8折", got.Name)
}

func TestCalculateCouponDiscount_EnginePath(t *testing.T) {
    cm := &MockCacheManager{}
    now := time.Now()
    cm.On("GetCouponTemplate", mock.Anything, int64(301)).Return(&cache.CouponTemplate{
        ID: 301, Name: "满50-10", Type: int32(do.CouponTypeThreshold), DiscountType: int32(do.DiscountTypeFixed),
        DiscountValue: 10, MinAmount: 50, ValidStart: now.Add(-time.Hour), ValidEnd: now.Add(time.Hour), Status: int32(do.CouponStatusActive),
    }, nil)
    cm.On("GetUserCoupon", mock.Anything, int64(401)).Return(&cache.UserCoupon{
        ID: 401, CouponID: 301, UserID: 1001, CouponSn: "UC401", Status: int32(do.UserCouponStatusUnused),
        ObtainTime: now.Add(-2 * time.Hour), ValidStartTime: now.Add(-time.Hour), ValidEndTime: now.Add(time.Hour),
    }, nil)
    svc := NewCouponService(nil, nil, nil, cm)
    res, err := svc.CalculateCouponDiscount(context.Background(), &dto.CalculateCouponDiscountDTO{UserID: 1001, CouponIDs: []int64{401}, OrderAmount: 120, OrderItems: []*dto.OrderItemDTO{}})
    assert.NoError(t, err)
    assert.NotNil(t, res)
    assert.Contains(t, res.AppliedCoupons, int64(401))
    assert.InDelta(t, 10.0, res.DiscountAmount, 1e-6)
    assert.InDelta(t, 110.0, res.FinalAmount, 1e-6)
}

