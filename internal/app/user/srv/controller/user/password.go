package user

import (
	"context"

	"emshop/internal/app/user/srv/pkg/password"
	upbv1 "emshop/api/user/v1"
)

func (us *userServer) CheckPassWord(ctx context.Context, info *upbv1.PasswordCheckInfo) (*upbv1.CheckResponse, error) {
	//校验密码
	check := password.VerifyPassword(info.Password, info.EncryptedPassword)
	return &upbv1.CheckResponse{Success: check}, nil
}
