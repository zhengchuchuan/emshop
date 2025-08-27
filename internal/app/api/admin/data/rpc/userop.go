package rpc

import (
	uoppbv1 "emshop/api/userop/v1"
	"emshop/internal/app/api/admin/data"
)


type userop struct {
	uopc uoppbv1.UserOpClient
}

func NewUserOp(uopc uoppbv1.UserOpClient) *userop {
	return &userop{uopc}
}


// 用户操作相关方法可以根据需要添加

var _ data.UserOpData = &userop{}