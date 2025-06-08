package mock

import (
	"context"

	dv1 "emshop-admin/internal/app/user/data/v1"
	metav1 "emshop-admin/pkg/common/meta/v1"
)

type users struct{}

func NewUsers() *users {
	return &users{}
}

func (u *users) List(ctx context.Context, opts metav1.ListMeta) (*dv1.UserDOList, error) {
	// 模拟返回数据
	return &dv1.UserDOList{
		TotalCount: 1,
		Items: []*dv1.UserDO{
			{
				Name: "testuser",
			},
		},
	}, nil
}