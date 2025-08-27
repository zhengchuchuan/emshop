package rpc

import (
	"context"
	ipbv1 "emshop/api/inventory/v1"
	"emshop/internal/app/api/admin/data"
	"emshop/pkg/log"
)


type inventory struct {
	ic ipbv1.InventoryClient
}

func NewInventory(ic ipbv1.InventoryClient) *inventory {
	return &inventory{ic}
}


// GetInventory 获取商品库存信息
func (i *inventory) GetInventory(ctx context.Context, goodsId int32) (*ipbv1.GoodsInvInfo, error) {
	log.Infof("Calling InvDetail gRPC for goods ID: %d", goodsId)
	request := &ipbv1.GoodsInvInfo{GoodsId: goodsId}
	response, err := i.ic.InvDetail(ctx, request)
	if err != nil {
		log.Errorf("InvDetail gRPC call failed for goods %d: %v", goodsId, err)
		return nil, err
	}
	log.Infof("InvDetail gRPC call successful for goods %d: stocks=%d", goodsId, response.Num)
	return response, nil
}

// BatchGetInventory 批量获取商品库存信息
func (i *inventory) BatchGetInventory(ctx context.Context, goodsIds []int32) (map[int32]*ipbv1.GoodsInvInfo, error) {
	result := make(map[int32]*ipbv1.GoodsInvInfo)
	
	// 并发获取库存信息以提高性能
	for _, goodsId := range goodsIds {
		inv, err := i.GetInventory(ctx, goodsId)
		if err != nil {
			log.Errorf("Failed to get inventory for goods %d: %v", goodsId, err)
			// 库存获取失败时，设置默认值
			result[goodsId] = &ipbv1.GoodsInvInfo{GoodsId: goodsId, Num: 0}
			continue
		}
		result[goodsId] = inv
	}
	
	return result, nil
}

// SetInventory 设置商品库存
func (i *inventory) SetInventory(ctx context.Context, request *ipbv1.GoodsInvInfo) error {
	_, err := i.ic.SetInv(ctx, request)
	return err
}

// BatchSetInventory 批量设置商品库存
func (i *inventory) BatchSetInventory(ctx context.Context, inventories []*ipbv1.GoodsInvInfo) error {
	for _, inv := range inventories {
		if err := i.SetInventory(ctx, inv); err != nil {
			log.Errorf("Failed to set inventory for goods %d: %v", inv.GoodsId, err)
			return err
		}
	}
	return nil
}

var _ data.InventoryData = &inventory{}