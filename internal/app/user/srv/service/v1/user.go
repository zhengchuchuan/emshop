package v1

import (
	"context"
	"emshop/internal/app/pkg/code"
	"emshop/internal/app/user/srv/domain/do"
	"emshop/internal/app/user/srv/domain/dto"
	"emshop/internal/app/user/srv/pkg/password"
	metav1 "emshop/pkg/common/meta/v1"
	"emshop/pkg/errors"
)


type userService struct {
	userStrore do.UserStore
}

type UserSrv interface {
	List(ctx context.Context, orderby []string, opts metav1.ListMeta) (*dto.UserDTOList, error)
	Create(ctx context.Context, user *dto.UserDTO) error
	Update(ctx context.Context, user *dto.UserDTO) error
	GetByID(ctx context.Context, ID uint64) (*dto.UserDTO, error)
	GetByMobile(ctx context.Context, mobile string) (*dto.UserDTO, error)
}



func NewUserService(us do.UserStore) UserSrv {
	return &userService{
		userStrore: us,
	}
}
var _ UserSrv = &userService{}




func (u *userService) Create(ctx context.Context, user *dto.UserDTO) error {
	// 检查用户是否存在的逻辑在service层实现
	//先判断用户是否存在
	_, err := u.userStrore.GetByMobile(ctx, user.Mobile)
	if err != nil && errors.IsCode(err, code.ErrUserNotFound) {
		// 密码加密逻辑应该在service层
		encryptedPassword, err := password.EncryptPassword(user.Password)
		if err != nil {
			return errors.WithCode(code.ErrEncryptionFailed, "密码加密失败")
		}
		
		// 更新用户密码为加密后的密码
		user.Password = encryptedPassword
		
		return u.userStrore.Create(ctx, &user.UserDO)
	}

	//这里应该区别到底是什么错误，用户已经存在？ 数据访问错误？
	return errors.WithCode(code.ErrUserAlreadyExists, "用户已经存在")
}

func (u *userService) Update(ctx context.Context, user *dto.UserDTO) error {
	//先查询用户是否存在
	_, err := u.userStrore.GetByID(ctx, uint64(user.ID))
	if err != nil {
		return err
	}

	return u.userStrore.Update(ctx, &user.UserDO)
}

func (u *userService) GetByID(ctx context.Context, ID uint64) (*dto.UserDTO, error) {
	userDO, err := u.userStrore.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}

	return &dto.UserDTO{UserDO: *userDO}, nil
}

func (u *userService) GetByMobile(ctx context.Context, mobile string) (*dto.UserDTO, error) {
	userDO, err := u.userStrore.GetByMobile(ctx, mobile)
	if err != nil {
		return nil, err
	}

	return &dto.UserDTO{UserDO: *userDO}, nil
}


func (u *userService) List(ctx context.Context, orderby []string, opts metav1.ListMeta) (*dto.UserDTOList, error) {
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
	var userDTOList dto.UserDTOList
	userDTOList.TotalCount = doList.TotalCount  // 设置总数
	for _, value := range doList.Items {
		projectDTO := dto.UserDTO{UserDO: *value}
		userDTOList.Items = append(userDTOList.Items, &projectDTO)
	}

	//业务逻辑3
	return &userDTOList, nil
}
