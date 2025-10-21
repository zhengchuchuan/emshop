package rpc

import (
    "context"
    cpbv1 "emshop/api/coupon/v1"
)

// gRPC 管理接口实现，调用 coupon.Coupon 服务新增的管理RPC

func (c *coupon) GetManageConfig(ctx context.Context, key string) (map[string]interface{}, error) {
    resp, err := c.cc.GetManageConfig(ctx, &cpbv1.GetManageConfigRequest{Key: key})
    if err != nil { return nil, err }
    return map[string]interface{}{"key": resp.Key, "value": resp.Value, "source": resp.Source}, nil
}

func (c *coupon) SetManageConfig(ctx context.Context, key, value, desc string) (map[string]interface{}, error) {
    _, err := c.cc.SetManageConfig(ctx, &cpbv1.SetManageConfigRequest{Key: key, Value: value, Description: desc})
    if err != nil { return nil, err }
    return map[string]interface{}{"ok": true}, nil
}

func (c *coupon) StartFlashSale(ctx context.Context, activityID int64) (map[string]interface{}, error) {
    _, err := c.cc.StartFlashSaleActivity(ctx, &cpbv1.StartFlashSaleRequest{Id: activityID})
    if err != nil { return nil, err }
    return map[string]interface{}{"ok": true, "activity_id": activityID}, nil
}

func (c *coupon) StopFlashSale(ctx context.Context, activityID int64) (map[string]interface{}, error) {
    _, err := c.cc.StopFlashSaleActivity(ctx, &cpbv1.StopFlashSaleRequest{Id: activityID})
    if err != nil { return nil, err }
    return map[string]interface{}{"ok": true, "activity_id": activityID}, nil
}

