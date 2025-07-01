package user

import (
	"context"
	v12 "emshop/internal/app/user/srv/data/v1"
	v1 "emshop/internal/app/user/srv/service/v1"

	upbv1 "emshop/api/user/v1"
	"emshop/pkg/log"
)

// controller层应该是很薄的一层， 参数校验，日志打印，错误处理，调用service层
func (u *userServer) CreateUser(ctx context.Context, request *upbv1.CreateUserInfo) (*upbv1.UserInfoResponse, error) {
	log.Infof("create user function called.")

	userDO := v12.UserDO{
		Mobile:   request.Mobile,
		NickName: request.NickName,
		Password: request.PassWord,
	}
	userDTO := v1.UserDTO{userDO}

	err := u.srv.Create(ctx, &userDTO)
	if err != nil {
		log.Errorf("create user: %v, error: %v", userDTO, err)
		return nil, err
	}

	userInfoRsp := DTOToResponse(userDTO)
	return userInfoRsp, nil
}
