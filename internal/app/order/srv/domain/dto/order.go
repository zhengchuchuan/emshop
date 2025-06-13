package dto

import "emshop/internal/app/order/srv/domain/do"

type OrderDTO struct {
	do.OrderInfoDO
}

type OrderDTOList struct {
	TotalCount int64       `json:"totalCount,omitempty"`
	Items      []*OrderDTO `json:"data"`
}
