package user

import (
	"context"
	upbv1 "emshop/api/user/v1"
	"emshop/internal/app/api/admin/data"
	jwtpkg "emshop/internal/app/pkg/jwt"
	"emshop/internal/app/pkg/options"
	"emshop/pkg/log"
	"time"
)

// AdminUserDTO 管理员用户数据传输对象
type AdminUserDTO struct {
	ID       uint64 `json:"id"`
	Mobile   string `json:"mobile"`
	NickName string `json:"nick_name"`
	Role     int32  `json:"role"`
	PassWord string `json:"password"`
	Token    string `json:"token"`
	ExpiresAt int64 `json:"expires_at"`
}

// UserSrv 管理员用户服务接口
type UserSrv interface {
	GetUserList(ctx context.Context, pageInfo *upbv1.PageInfo) (*upbv1.UserListResponse, error)
	GetUserById(ctx context.Context, id uint64) (*upbv1.UserInfoResponse, error)
	GetUserByMobile(ctx context.Context, mobile string) (*upbv1.UserInfoResponse, error)
	UpdateUserStatus(ctx context.Context, id uint64, status int32) error
	UpdateUser(ctx context.Context, user *upbv1.UserInfoResponse) error
	UpdateUserInfo(ctx context.Context, id uint64, nickName, gender string, birthday uint64) error
	// 添加登录相关方法
	MobileLogin(ctx context.Context, mobile, password string) (*AdminUserDTO, error)
	CheckPassWord(ctx context.Context, password, encryptedPassword string) (bool, error)
}

type userService struct {
	data data.DataFactory
	jwt  *options.JwtOptions
}

func NewUserService(data data.DataFactory, jwt *options.JwtOptions) UserSrv {
	return &userService{
		data: data,
		jwt:  jwt,
	}
}

func (u *userService) GetUserList(ctx context.Context, pageInfo *upbv1.PageInfo) (*upbv1.UserListResponse, error) {
	if pageInfo.Pn != nil && pageInfo.PSize != nil {
		log.Infof("Admin GetUserList called with page: %d, pageSize: %d", *pageInfo.Pn, *pageInfo.PSize)
	} else {
		log.Infof("Admin GetUserList called with no pagination (return all data)")
	}
	
	return u.data.Users().GetUserList(ctx, pageInfo)
}

func (u *userService) GetUserById(ctx context.Context, id uint64) (*upbv1.UserInfoResponse, error) {
	log.Infof("Admin GetUserById called with id: %d", id)
	
	request := &upbv1.IdRequest{
		Id: int32(id),
	}
	
	return u.data.Users().GetUserById(ctx, request)
}

func (u *userService) GetUserByMobile(ctx context.Context, mobile string) (*upbv1.UserInfoResponse, error) {
	log.Infof("Admin GetUserByMobile called with mobile: %s", mobile)
	
	request := &upbv1.MobileRequest{
		Mobile: mobile,
	}
	
	return u.data.Users().GetUserByMobile(ctx, request)
}

func (u *userService) UpdateUserStatus(ctx context.Context, id uint64, status int32) error {
	log.Infof("Admin UpdateUserStatus called with id: %d, status: %d", id, status)
	
	// 这里可以添加更多管理员特有的业务逻辑，比如权限检查、审计日志等
	request := &upbv1.UpdateUserInfo{
		Id: int32(id),
		// 根据实际需要设置状态字段
	}
	
	_, err := u.data.Users().UpdateUser(ctx, request)
	return err
}

func (u *userService) UpdateUser(ctx context.Context, user *upbv1.UserInfoResponse) error {
	log.Infof("Admin UpdateUser called with id: %d", user.Id)
	
	request := &upbv1.UpdateUserInfo{
		Id:       user.Id,
		NickName: &user.NickName,
		Gender:   &user.Gender,
		BirthDay: &user.BirthDay,
	}
	
	_, err := u.data.Users().UpdateUser(ctx, request)
	return err
}

func (u *userService) UpdateUserInfo(ctx context.Context, id uint64, nickName, gender string, birthday uint64) error {
	log.Infof("Admin UpdateUserInfo called with id: %d, nickName: %s, gender: %s, birthday: %d", id, nickName, gender, birthday)
	
	request := &upbv1.UpdateUserInfo{
		Id:       int32(id),
		NickName: &nickName,
		Gender:   &gender,
		BirthDay: &birthday,
	}
	
	_, err := u.data.Users().UpdateUser(ctx, request)
	return err
}

// MobileLogin 管理员手机号登录
func (u *userService) MobileLogin(ctx context.Context, mobile, password string) (*AdminUserDTO, error) {
	log.Infof("Admin MobileLogin called with mobile: %s", mobile)
	
	// 获取用户信息
	userResp, err := u.GetUserByMobile(ctx, mobile)
	if err != nil {
		log.Errorf("Admin login failed: user not found - %v", err)
		return nil, err
	}
	
	// 验证密码
	isValid, err := u.CheckPassWord(ctx, password, userResp.PassWord)
	if err != nil {
		log.Errorf("Admin login failed: password check error - %v", err)
		return nil, err
	}
	
	if !isValid {
		log.Warnf("Admin login failed: incorrect password for user ID: %d", userResp.Id)
		return nil, err
	}
	
	// 生成JWT Token
	j := jwtpkg.NewEmshopJWT(u.jwt.Key)
	token, err := j.CreateToken(
		uint(userResp.Id),
		uint(userResp.Role), // 这里是角色信息
		jwtpkg.IssuerEmshopAdmin,
		u.jwt.Timeout,
	)
	if err != nil {
		log.Errorf("Admin login failed: token generation error - %v", err)
		return nil, err
	}
	
	return &AdminUserDTO{
		ID:       uint64(userResp.Id),
		Mobile:   userResp.Mobile,
		NickName: userResp.NickName,
		Role:     userResp.Role,
		PassWord: userResp.PassWord,
		Token:    token,
		ExpiresAt: time.Now().Add(u.jwt.Timeout).Unix(),
	}, nil
}

// CheckPassWord 验证密码
func (u *userService) CheckPassWord(ctx context.Context, password, encryptedPassword string) (bool, error) {
	log.Infof("Admin CheckPassWord called")
	
	request := &upbv1.PasswordCheckInfo{
		Password:          password,
		EncryptedPassword: encryptedPassword,
	}
	
	response, err := u.data.Users().CheckPassWord(ctx, request)
	if err != nil {
		log.Errorf("Admin password check failed: %v", err)
		return false, err
	}
	
	return response.Success, nil
}