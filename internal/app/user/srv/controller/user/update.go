package user

import (
	"context"
	upbv1 "emshop/api/user/v1"
	"emshop/internal/app/user/srv/domain/do"
	"emshop/internal/app/user/srv/domain/dto"
	"emshop/pkg/db"
	"emshop/pkg/log"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (u *userServer) UpdateUser(ctx context.Context, request *upbv1.UpdateUserInfo) (*emptypb.Empty, error) {
	log.Infof("update user function called.")

	birthDay := time.Unix(int64(request.BirthDay), 0)
	userDO := do.UserDO{
		BaseModel: db.BaseModel{
			ID: request.Id,
		},
		NickName: request.NickName,
		Gender:   request.Gender,
		Birthday: &birthDay,
	}
	userDTO := dto.UserDTO{userDO}

	err := u.srv.Update(ctx, &userDTO)
	if err != nil {
		log.Errorf("update user: %v, error: %v", userDTO, err)
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
