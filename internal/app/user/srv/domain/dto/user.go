package dto

import (
	"emshop/internal/app/user/srv/domain/do"
)

type UserDTO struct {
	do.UserDO
}

type UserDTOList struct {
	TotalCount int64      `json:"totalCount,omitempty"` //总数
	Items      []*UserDTO `json:"data"`                 //数据
}