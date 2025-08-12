package mysql

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	code2 "emshop/gin-micro/code"
	"emshop/internal/app/pkg/code"
	"emshop/internal/app/user/srv/data/v1/interfaces"
	"emshop/internal/app/user/srv/domain/do"
	metav1 "emshop/pkg/common/meta/v1"
	"emshop/pkg/errors"
)

type users struct {
	factory *mysqlFactory
}

func newUsers(factory *mysqlFactory) interfaces.UserStore {
	return &users{factory: factory}
}

var _ interfaces.UserStore = &users{}

func (u *users) Get(ctx context.Context, ID uint64) (*do.UserDO, error) {
	user := do.UserDO{}
	err := u.factory.db.First(&user, ID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrUserNotFound, "%s", err.Error())
		}
		return nil, errors.WithCode(code2.ErrDatabase, "%s", err.Error())
	}
	return &user, nil
}

func (u *users) GetByMobile(ctx context.Context, mobile string) (*do.UserDO, error) {
	user := do.UserDO{}

	err := u.factory.db.Where("mobile=?", mobile).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrUserNotFound, "%s", err.Error())
		}
		return nil, errors.WithCode(code2.ErrDatabase, "%s", err.Error())
	}
	return &user, nil
}

func (u *users) Create(ctx context.Context, user *do.UserDO) error {
	tx := u.factory.db.Create(user)
	if tx.Error != nil {
		return errors.WithCode(code2.ErrDatabase, "%s", tx.Error.Error())
	}
	return nil
}

func (u *users) Update(ctx context.Context, user *do.UserDO) error {
	tx := u.factory.db.Save(user)
	if tx.Error != nil {
		return errors.WithCode(code2.ErrDatabase, "%s", tx.Error.Error())
	}
	return nil
}

func (u *users) List(ctx context.Context, orderby []string, opts metav1.ListMeta) (*do.UserDOList, error) {
	ret := &do.UserDOList{}

	var limit, offset int
	if opts.PageSize == 0 {
		limit = 10
	} else {
		limit = opts.PageSize
	}

	if opts.Page > 0 {
		offset = (opts.Page - 1) * limit
	}

	query := u.factory.db
	for _, value := range orderby {
		query = query.Order(value)
	}

	if err := query.Model(&do.UserDO{}).Count(&ret.TotalCount).Error; err != nil {
		return nil, errors.WithCode(code2.ErrDatabase, "%s", err.Error())
	}
	fmt.Println("TotalCount:", ret.TotalCount)
	
	if err := query.Offset(offset).Limit(limit).Find(&ret.Items).Error; err != nil {
		return nil, errors.WithCode(code2.ErrDatabase, "%s", err.Error())
	}
	return ret, nil
}