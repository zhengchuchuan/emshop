package v1

import (
	"context"
	metav1 "emshop/pkg/common/meta/v1"
)

type UserDO struct {
	// dv1.UserDO
	Name string `json:"name"` //用户名
}

type UserDOList struct {
	TotalCount int64      `json:"totalCount,omitempty"` //总数
	Items      []*UserDO `json:"data"`                 //数据
}

func List(ctx context.Context, opts metav1.ListMeta) (*UserDOList, error) {


	return &UserDOList{}, nil
}


type UserStore interface {
	/*
		有数据访问的方法，一定要有error
		参数中最好有ctx
	*/
	//用户列表
	List(ctx context.Context, opts metav1.ListMeta) (*UserDOList, error)

	// //通过手机号码查询用户
	// GetByMobile(ctx context.Context, mobile string) (*UserDO, error)

	// //通过用户ID查询用户
	// GetByID(ctx context.Context, id uint64) (*UserDO, error)

	// //创建用户
	// Create(ctx context.Context, user *UserDO) error

	// //更新用户
	// Update(ctx context.Context, user *UserDO) error
}