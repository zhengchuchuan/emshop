package db

import (
	"context"

	"gorm.io/gorm"

	dv1 "emshop/internal/app/user/srv/data/v1"
	metav1 "emshop/pkg/common/meta/v1"
)

type users struct {
	db *gorm.DB
}

func NewUsers(db *gorm.DB) *users {
	return &users{db: db}
}

func (u *users) List(ctx context.Context, orderby []string, opts metav1.ListMeta) (*dv1.UserDOList, error) {
	// 实现gorm查询逻辑
	// u.db.Where()
	return nil, nil
}

