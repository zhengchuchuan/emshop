package mock

import (
	"context"
	"emshop/internal/app/user/srv/data/v1/interfaces"
	"emshop/internal/app/user/srv/domain/do"
	metav1 "emshop/pkg/common/meta/v1"
)

type users struct{}

func NewUsers() interfaces.UserStore {
	return &users{}
}

var _ interfaces.UserStore = &users{}

func (u *users) Get(ctx context.Context, ID uint64) (*do.UserDO, error) {
	return &do.UserDO{
		Mobile:   "13800138000",
		Password: "password123",
		NickName: "test_user",
		Gender:   "male",
		Role:     1,
	}, nil
}

func (u *users) GetByMobile(ctx context.Context, mobile string) (*do.UserDO, error) {
	return &do.UserDO{
		Mobile:   mobile,
		Password: "password123",
		NickName: "test_user",
		Gender:   "male",
		Role:     1,
	}, nil
}

func (u *users) List(ctx context.Context, orderby []string, opts metav1.ListMeta) (*do.UserDOList, error) {
	users := []*do.UserDO{
		{
			Mobile:   "13800138001",
			Password: "password123",
			NickName: "user1",
			Gender:   "male",
			Role:     1,
		},
		{
			Mobile:   "13800138002",
			Password: "password456",
			NickName: "user2",
			Gender:   "female",
			Role:     1,
		},
	}

	return &do.UserDOList{
		TotalCount: int64(len(users)),
		Items:      users,
	}, nil
}

func (u *users) Create(ctx context.Context, user *do.UserDO) error {
	return nil
}

func (u *users) Update(ctx context.Context, user *do.UserDO) error {
	return nil
}