package v1

import (
	proto "emshop/api/goods/v1"
	proto2 "emshop/api/inventory/v1"

	"gorm.io/gorm"
)

type DataFactory interface {
	Orders() OrderStore
	ShoppingCarts() ShopCartStore
	Goods() proto.GoodsClient
	Inventorys() proto2.InventoryClient

	Begin() *gorm.DB
}
