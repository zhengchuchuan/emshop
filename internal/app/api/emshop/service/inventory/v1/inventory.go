package v1

import (
	"context"
	ipb "emshop/api/inventory/v1"
	"emshop/internal/app/api/emshop/data"
	"strconv"
)

type InventorySrv interface {
	GetStocks(ctx context.Context, goodsID string) (*ipb.GoodsInvInfo, error)
}

type inventoryService struct {
	data data.DataFactory
}

func (is *inventoryService) GetStocks(ctx context.Context, goodsID string) (*ipb.GoodsInvInfo, error) {
	goodsIDInt, err := strconv.Atoi(goodsID)
	if err != nil {
		return nil, err
	}
	
	request := &ipb.GoodsInvInfo{
		GoodsId: int32(goodsIDInt),
	}
	
	return is.data.Inventory().InvDetail(ctx, request)
}

func NewInventory(data data.DataFactory) InventorySrv {
	return &inventoryService{data: data}
}

var _ InventorySrv = &inventoryService{}