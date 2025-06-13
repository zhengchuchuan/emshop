package dto

import "emshop/internal/app/order/srv/domain/do"

type ShopCartDTO struct {
	do.ShoppingCartDO
}

type ShopCartDTOList struct {
	TotalCount int64          `json:"totalCount,omitempty"`
	Items      []*ShopCartDTO `json:"data"`
}
