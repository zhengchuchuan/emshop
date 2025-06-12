package dto

import "emshop/internal/app/goods/srv/domain/do"

type GoodsDTO struct {
	do.GoodsDO
}

type GoodsDTOList struct {
	TotalCount int         `json:"total_count,omitempty"`
	Items      []*GoodsDTO `json:"data"`
}
