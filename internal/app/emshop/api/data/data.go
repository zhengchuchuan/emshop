package data

import (
	gpb "emshop/api/goods/v1"
)

type DataFactory interface {
	Goods() gpb.GoodsClient
	Users() UserData
}
