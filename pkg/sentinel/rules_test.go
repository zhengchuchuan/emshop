package sentinel_test

import (
    "testing"
    "github.com/alibaba/sentinel-golang/core/flow"
    sentinel "emshop/pkg/sentinel"
    "github.com/stretchr/testify/assert"
)

func TestBusinessRules_Defaults(t *testing.T) {
    br := sentinel.DefaultBusinessRules()
    // Flash sale
    assert.Equal(t, 1000.0, br.FlashSale.FlashSaleQPS)
    assert.Equal(t, 2000.0, br.FlashSale.ProductDetailQPS)
    assert.Equal(t, 500.0, br.FlashSale.OrderQPS)
    // Payment
    assert.Equal(t, 200.0, br.Payment.PaymentQPS)
    assert.Equal(t, 1000.0, br.Payment.PaymentQueryQPS)
    assert.Equal(t, 100.0, br.Payment.RefundQPS)
    // Inventory
    assert.Equal(t, 300.0, br.Inventory.DeductQPS)
    assert.Equal(t, 2000.0, br.Inventory.QueryQPS)
    assert.Equal(t, 200.0, br.Inventory.RestoreQPS)
    // Coupon
    assert.Equal(t, 500.0, br.Coupon.IssueQPS)
    assert.Equal(t, 800.0, br.Coupon.UseQPS)
    assert.Equal(t, 1500.0, br.Coupon.QueryQPS)
    // User
    assert.Equal(t, 300.0, br.User.LoginQPS)
    assert.Equal(t, 100.0, br.User.RegisterQPS)
    assert.Equal(t, 1000.0, br.User.QueryQPS)
    // Goods
    assert.Equal(t, 2000.0, br.Goods.ListQPS)
    assert.Equal(t, 3000.0, br.Goods.DetailQPS)
    assert.Equal(t, 1500.0, br.Goods.SearchQPS)
    // Order
    assert.Equal(t, 500.0, br.Order.CreateQPS)
    assert.Equal(t, 1500.0, br.Order.QueryQPS)
    assert.Equal(t, 200.0, br.Order.CancelQPS)
}

func TestBusinessRules_GenerateFlowRules_Coupon(t *testing.T) {
    br := sentinel.DefaultBusinessRules()
    rules, err := br.GenerateFlowRules("emshop-coupon-srv")
    assert.NoError(t, err)
    assert.Len(t, rules, 3)
    th := resourceThresholds(rules)
    assert.Equal(t, br.Coupon.IssueQPS, th["coupon-srv:IssueCoupon"])
    assert.Equal(t, br.Coupon.UseQPS, th["coupon-srv:UseCoupon"])
    assert.Equal(t, br.Coupon.QueryQPS, th["coupon-srv:GetUserCoupons"])
}

func TestBusinessRules_GenerateFlowRules_User(t *testing.T) {
    br := sentinel.DefaultBusinessRules()
    rules, err := br.GenerateFlowRules("emshop-user-srv")
    assert.NoError(t, err)
    assert.Len(t, rules, 3)
    th := resourceThresholds(rules)
    assert.Equal(t, br.User.RegisterQPS, th["user-srv:CreateUser"])
    assert.Equal(t, br.User.LoginQPS, th["user-srv:GetUserByMobile"])
    assert.Equal(t, br.User.QueryQPS, th["user-srv:GetUserById"])
}

func TestBusinessRules_GenerateFlowRules_Inventory(t *testing.T) {
    br := sentinel.DefaultBusinessRules()
    rules, err := br.GenerateFlowRules("emshop-inventory-srv")
    assert.NoError(t, err)
    assert.Len(t, rules, 2)
    th := resourceThresholds(rules)
    assert.Equal(t, br.Inventory.DeductQPS, th["inventory-srv:Sell"])
    assert.Equal(t, br.Inventory.QueryQPS, th["inventory-srv:InvDetail"])
}

func resourceThresholds(rules []*flow.Rule) map[string]float64 {
    m := make(map[string]float64, len(rules))
    for _, r := range rules {
        m[r.Resource] = r.Threshold
    }
    return m
}

