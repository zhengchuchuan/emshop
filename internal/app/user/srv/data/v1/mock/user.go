package mock

import (
	"context"

	dv1 "emshop/internal/app/user/srv/data/v1"
	metav1 "emshop/pkg/common/meta/v1"
)

type users struct{
	users []*dv1.UserDO
}

func NewUsers() *users {
	return &users{}
}

func (u *users) List(ctx context.Context, orderby []string, opts metav1.ListMeta) (*dv1.UserDOList, error) {
 
	// 模拟返回数据
	return &dv1.UserDOList{
		TotalCount: 1,
		Items: u.users,
	}, nil
}

// 添加 Create 方法
func (u *users) Create(ctx context.Context, user *dv1.UserDO) error {
    // 模拟实现
    u.users = append(u.users, user)
    return nil
}

// 添加 GetByMobile 方法
func (u *users) GetByMobile(ctx context.Context, mobile string) (*dv1.UserDO, error) {
    // 模拟实现
    return nil, nil
}

// 添加 GetByID 方法
func (u *users) GetByID(ctx context.Context, id uint64) (*dv1.UserDO, error) {
    // 模拟实现
    return nil, nil
}

// 添加 Update 方法
func (u *users) Update(ctx context.Context, user *dv1.UserDO) error {
    // 模拟实现
    return nil
}