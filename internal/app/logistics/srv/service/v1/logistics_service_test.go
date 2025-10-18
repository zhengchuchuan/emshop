package v1

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "emshop/internal/app/logistics/srv/domain/dto"
)

func TestLogistics_HelperCalculations(t *testing.T) {
    ls := &logisticsService{}

    // weight
    w := ls.calculateTotalWeight([]dto.OrderItemDTO{{Weight: 0.5, Quantity: 2}, {Weight: 1.2, Quantity: 1}})
    assert.InDelta(t, 2.2, w, 1e-6)

    // volume
    v := ls.calculateTotalVolume([]dto.OrderItemDTO{{Volume: 0.3, Quantity: 2}, {Volume: 0.2, Quantity: 1}})
    assert.InDelta(t, 0.8, v, 1e-6)

    // method multiplier
    assert.Equal(t, 2.0, ls.getMethodMultiplier(2))
    assert.Equal(t, 0.8, ls.getMethodMultiplier(3))
    assert.Equal(t, 1.0, ls.getMethodMultiplier(1))

    // base fee
    fee := ls.calculateBaseFee(150, 2.5)
    // 8 + (2.5-1)*2 + (150-100)*0.01 = 8 + 3 + 0.5 = 11.5
    assert.InDelta(t, 11.5, fee, 1e-6)

    // estimated days
    assert.Equal(t, int32(3), ls.getEstimatedDays(2, 600)) // fast, distance>500 -> baseDays=3
    assert.Equal(t, int32(3), ls.getEstimatedDays(3, 50))  // economy, baseDays=1 -> 3

    // distance
    d1 := ls.calculateDistance("北京市海淀区", "北京市朝阳区")
    d2 := ls.calculateDistance("北京市海淀区", "上海市徐汇区")
    d3 := ls.calculateDistance("杭州市西湖区", "苏州市园区")
    assert.Equal(t, 50.0, d1)
    assert.Equal(t, 800.0, d2)
    assert.Equal(t, 200.0, d3)
}
