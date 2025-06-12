package v1

import (
	"context"
	"emshop/internal/app/pkg/code"
	dv1 "emshop/internal/app/user/srv/data/v1"
	metav1 "emshop/pkg/common/meta/v1"
	"emshop/pkg/errors"
)

type UserDTO struct {
	dv1.UserDO
}
type UserDTOList struct {
	TotalCount int64      `json:"totalCount,omitempty"` //总数
	Items      []*UserDTO `json:"data"`                 //数据
}

type userService struct {
	userStrore dv1.UserStore
}

type UserSrv interface {
	List(ctx context.Context, orderby []string, opts metav1.ListMeta) (*UserDTOList, error)
	Create(ctx context.Context, user *UserDTO) error
	Update(ctx context.Context, user *UserDTO) error
	GetByID(ctx context.Context, ID uint64) (*UserDTO, error)
	GetByMobile(ctx context.Context, mobile string) (*UserDTO, error)
}



func NewUserService(us dv1.UserStore) UserSrv {
	return &userService{
		userStrore: us,
	}
}
var _ UserSrv = &userService{}




func (u *userService) Create(ctx context.Context, user *UserDTO) error {
	// 检查用户是否存在的逻辑在servic层实现
	//先判断用户是否存在
	_, err := u.userStrore.GetByMobile(ctx, user.Mobile)
	if err != nil && errors.IsCode(err, code.ErrUserNotFound) {
		return u.userStrore.Create(ctx, &user.UserDO)
	}

	//这里应该区别到底是什么错误，用户已经存在？ 数据访问错误？
	return errors.WithCode(code.ErrUserAlreadyExists, "用户已经存在")
}

func (u *userService) Update(ctx context.Context, user *UserDTO) error {
	//先查询用户是否存在
	_, err := u.userStrore.GetByID(ctx, uint64(user.ID))
	if err != nil {
		return err
	}

	return u.userStrore.Update(ctx, &user.UserDO)
}

func (u *userService) GetByID(ctx context.Context, ID uint64) (*UserDTO, error) {
	userDO, err := u.userStrore.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}

	return &UserDTO{*userDO}, nil
}

func (u *userService) GetByMobile(ctx context.Context, mobile string) (*UserDTO, error) {
	userDO, err := u.userStrore.GetByMobile(ctx, mobile)
	if err != nil {
		return nil, err
	}

	return &UserDTO{*userDO}, nil
}


func (u *userService) List(ctx context.Context, orderby []string, opts metav1.ListMeta) (*UserDTOList, error) {
	//这里是业务逻辑1
	/*
		1. data层的接口必须先写好
		2. 我期望测试的时候每次测试底层的data层的数据按照我期望的返回
			1. 先手动去插入一些数据
			2. 去删除一些数据
		3. 如果data层的方法有bug， 坑爹， 我们的代码想要具备好的可测试性
	*/

	doList, err := u.userStrore.List(ctx, orderby, opts)
	if err != nil {
		return nil, err
	}

	//业务逻辑2
	//代码不方便写单元测试用例
	var userDTOList UserDTOList
	for _, value := range doList.Items {
		projectDTO := UserDTO{*value}
		userDTOList.Items = append(userDTOList.Items, &projectDTO)
	}

	//业务逻辑3
	return &userDTOList, nil
}
