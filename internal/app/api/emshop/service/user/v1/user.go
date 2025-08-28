package v1

import (
	"context"
	"fmt"
	"time"

	upb "emshop/api/user/v1"
	"emshop/internal/app/api/emshop/data"
	"emshop/internal/app/pkg/code"
	jwtpkg "emshop/internal/app/pkg/jwt"
	"emshop/internal/app/pkg/options"
	itime "emshop/pkg/common/time"
	"emshop/pkg/errors"
	"emshop/pkg/log"
	"emshop/pkg/storage"
)

type User struct {
	ID       uint64     `json:"id"`
	Mobile   string     `json:"mobile"`
	NickName string     `json:"nick_name"`
	Birthday itime.Time `json:"birthday"`
	Gender   string     `json:"gender"`
	Role     int32      `json:"role"`
	PassWord string     `json:"password"`
}

type UserDTO struct {
	User

	Token     string `json:"token"`      // JWT Token
	ExpiresAt int64  `json:"expires_at"` // token过期时间
}

type UserListDTO struct {
	TotalCount int64      `json:"totalCount,omitempty"`
	Items      []*UserDTO `json:"items"`
}

type UserSrv interface {
	MobileLogin(ctx context.Context, mobile, password string) (*UserDTO, error)
	Register(ctx context.Context, mobile, password, code string) (*UserDTO, error)
	Update(ctx context.Context, userDTO *UserDTO) error
	Get(ctx context.Context, userID uint64) (*UserDTO, error)
	GetByMobile(ctx context.Context, mobile string) (*UserDTO, error)
	GetUserList(ctx context.Context, pn, pSize uint32) (*UserListDTO, error)
	CheckPassWord(ctx context.Context, password, EncryptedPassword string) (bool, error)
}

type userService struct {
	//ud data.UserData
	data data.DataFactory

	jwtOpts *options.JwtOptions
}

func NewUserService(data data.DataFactory, jwtOpts *options.JwtOptions) UserSrv {
	return &userService{data: data, jwtOpts: jwtOpts}
}

// 辅助函数：将protobuf用户信息转换为本地User结构体
func protoToUser(pb *upb.UserInfoResponse) User {
	return User{
		ID:       uint64(pb.Id),
		Mobile:   pb.Mobile,
		NickName: pb.NickName,
		Birthday: itime.Time{Time: time.Unix(int64(pb.BirthDay), 0)},
		Gender:   pb.Gender,
		Role:     pb.Role,
		PassWord: pb.PassWord,
	}
}

func (us *userService) MobileLogin(ctx context.Context, mobile, password string) (*UserDTO, error) {
	userResp, err := us.data.Users().GetUserByMobile(ctx, &upb.MobileRequest{Mobile: mobile})
	if err != nil {
		return nil, err
	}

	user := protoToUser(userResp)

	//检查密码是否正确
	checkResp, err := us.data.Users().CheckPassWord(ctx, &upb.PasswordCheckInfo{
		Password:          password,
		EncryptedPassword: user.PassWord,
	})
	if err != nil {
		return nil, err
	}
	if !checkResp.Success {
		return nil, errors.WithCode(code.ErrUserPasswordIncorrect, "密码错误")
	}

	//生成token
	j := jwtpkg.NewEmshopJWT(us.jwtOpts.Key)
	token, err := j.CreateToken(
		uint(user.ID),
		uint(user.Role),
		jwtpkg.IssuerEmshopAPI,
		us.jwtOpts.Timeout,
	)

	if err != nil {
		return nil, err
	}

	return &UserDTO{
		User:      user,
		Token:     token,
		ExpiresAt: (time.Now().Local().Add(us.jwtOpts.Timeout)).Unix(),
	}, nil
}

func (us *userService) Register(ctx context.Context, mobile, password, codes string) (*UserDTO, error) {
	rstore := storage.RedisCluster{}

	// 生成验证码的key
	value, err := rstore.GetKey(ctx, fmt.Sprintf("%s_%d", mobile, 1))
	if err != nil {
		return nil, errors.WithCode(code.ErrCodeNotExist, "验证码不存在")
	}

	if value != codes {
		return nil, errors.WithCode(code.ErrCodeInCorrect, "验证码错误")
	}

	userResp, err := us.data.Users().CreateUser(ctx, &upb.CreateUserInfo{
		Mobile:   mobile,
		PassWord: password,
	})
	if err != nil {
		log.Errorf("user register failed: %v", err)
		return nil, err
	}

	user := protoToUser(userResp)

	// 直接生成token
	j := jwtpkg.NewEmshopJWT(us.jwtOpts.Key)
	token, err := j.CreateToken(
		uint(user.ID),
		uint(user.Role),
		jwtpkg.IssuerEmshopAPI,
		us.jwtOpts.Timeout,
	)
	if err != nil {
		return nil, err
	}

	return &UserDTO{
		User:      user,
		Token:     token, // 向上传递token
		ExpiresAt: (time.Now().Local().Add(us.jwtOpts.Timeout)).Unix(),
	}, nil
}

func (us *userService) Update(ctx context.Context, userDTO *UserDTO) error {
	birthDay := uint64(userDTO.Birthday.Unix())
	_, err := us.data.Users().UpdateUser(ctx, &upb.UpdateUserInfo{
		Id:       int32(userDTO.ID),
		NickName: &userDTO.NickName,
		Gender:   &userDTO.Gender,
		BirthDay: &birthDay,
	})
	return err
}

func (us *userService) Get(ctx context.Context, userID uint64) (*UserDTO, error) {
	userResp, err := us.data.Users().GetUserById(ctx, &upb.IdRequest{Id: int32(userID)})
	if err != nil {
		return nil, err
	}
	user := protoToUser(userResp)
	return &UserDTO{User: user}, nil
}

func (us *userService) GetByMobile(ctx context.Context, mobile string) (*UserDTO, error) {
	userResp, err := us.data.Users().GetUserByMobile(ctx, &upb.MobileRequest{Mobile: mobile})
	if err != nil {
		return nil, err
	}
	user := protoToUser(userResp)
	return &UserDTO{User: user}, nil
}

func (us *userService) CheckPassWord(ctx context.Context, password, EncryptedPassword string) (bool, error) {
	checkResp, err := us.data.Users().CheckPassWord(ctx, &upb.PasswordCheckInfo{
		Password:          password,
		EncryptedPassword: EncryptedPassword,
	})
	if err != nil {
		return false, err
	}
	return checkResp.Success, nil
}

func (us *userService) GetUserList(ctx context.Context, pn, pSize uint32) (*UserListDTO, error) {
	userListResp, err := us.data.Users().GetUserList(ctx, &upb.PageInfo{
		Pn:    &pn,
		PSize: &pSize,
	})
	if err != nil {
		return nil, err
	}

	var userDTOs []*UserDTO
	for _, userResp := range userListResp.Data {
		user := protoToUser(userResp)
		userDTOs = append(userDTOs, &UserDTO{User: user})
	}

	return &UserListDTO{
		TotalCount: int64(userListResp.Total),
		Items:      userDTOs,
	}, nil
}

var _ UserSrv = &userService{}
