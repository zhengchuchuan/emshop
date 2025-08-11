package user

import (
	"context"
	"emshop/internal/app/user/srv/domain/do"
	"emshop/internal/app/user/srv/domain/dto"

	upbv1 "emshop/api/user/v1"
	"emshop/pkg/log"
)

// controller层应该是很薄的一层， 参数校验，日志打印，错误处理，调用service层
func (u *userServer) CreateUser(ctx context.Context, request *upbv1.CreateUserInfo) (*upbv1.UserInfoResponse, error) {
	log.Infof("create user function called.")

	userDO := do.UserDO{
		Mobile:   request.Mobile,
		NickName: request.NickName,
		Password: request.PassWord, // 原始密码传给Service层处理
	}
	userDTO := dto.UserDTO{UserDO: userDO}

	err := u.srv.Create(ctx, &userDTO)
	if err != nil {
		log.Errorf("create user: %v, error: %v", userDTO, err)
		return nil, err
	}

	userInfoRsp := DTOToResponse(userDTO)
	return userInfoRsp, nil
}
