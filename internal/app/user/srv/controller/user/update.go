package user

import (
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
	upbv1 "emshop/api/user/v1"
	v1 "emshop/internal/app/user/srv/data/v1"
	v12 "emshop/internal/app/user/srv/service/v1"
	"emshop/pkg/log"
	"time"
)

func (u *userServer) UpdateUser(ctx context.Context, request *upbv1.UpdateUserInfo) (*emptypb.Empty, error) {
	log.Infof("update user function called.")

	birthDay := time.Unix(int64(request.BirthDay), 0)
	userDO := v1.UserDO{
		BaseModel: v1.BaseModel{
			ID: request.Id,
		},
		NickName: request.NickName,
		Gender:   request.Gender,
		Birthday: &birthDay,
	}
	userDTO := v12.UserDTO{userDO}

	err := u.srv.Update(ctx, &userDTO)
	if err != nil {
		log.Errorf("update user: %v, error: %v", userDTO, err)
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
