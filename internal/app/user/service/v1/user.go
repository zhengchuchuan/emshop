package v1

import (
	"context"
	dv1 "emshop-admin/internal/app/user/data/v1"
	metav1 "emshop-admin/pkg/common/meta/v1"
)

type UserDTO struct {
	dv1.UserDO
	// Name string `json:"name"` //用户名
	// userStore dv1.UserStore
}

type UserDTOList struct {
	TotalCount int64      `json:"totalCount,omitempty"` //总数
	Items      []*UserDTO `json:"data"`                 //数据
}


type userService struct {
	userStrore dv1.UserStore
}

func NewUserService(us dv1.UserStore) *userService {
	return &userService{
		userStrore: us,
	}
}

func  (u *userService)List(ctx context.Context, opts metav1.ListMeta) (*UserDTOList, error) {
	
	/*
		1. data层的接口必须先写好
		2. 我期望测试的时候每次测试底层的data层的数据按照我期望的返回
			1. 先手动去插入一些数据
			2. 去删除一些数据
		3. 如果data层的方法有bug， 坑爹， 我们的代码想要具备好的可测试性
	*/
	doList, err := u.userStrore.List(ctx, opts)
	if err != nil {
		return nil, err
	}
	// 代码不方便写单元测试用例
	var userDTOList UserDTOList
	for _, value := range doList.Items {
		projectDTO := UserDTO{*value}
		userDTOList.Items = append(userDTOList.Items, &projectDTO)
	}
	return &userDTOList, nil
}
