package v1

import (
	"context"
	"fmt"
	"time"

	"emshop/gin-micro/server/rest-server/middlewares"
	"emshop/internal/app/emshop/api/data"
	"emshop/internal/app/pkg/code"
	"emshop/internal/app/pkg/options"
	"emshop/pkg/errors"
	"emshop/pkg/log"
	"emshop/pkg/storage"

	"github.com/golang-jwt/jwt/v4"
)

type UserDTO struct {
	data.User

	Token     string `json:"token"`			// JWT Token
	ExpiresAt int64  `json:"expires_at"`	// token过期时间
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


func (us *userService) MobileLogin(ctx context.Context, mobile, password string) (*UserDTO, error) {
	user, err := us.data.Users().GetByMobile(ctx, mobile)
	if err != nil {
		return nil, err
	}


	//检查密码是否正确
	err = us.data.Users().CheckPassWord(ctx, password, user.PassWord)
	if err != nil {
		// 如果是密码错误，返回特定的错误码
		if errors.IsCode(err, code.ErrUserPasswordIncorrect) {
			return nil, errors.WithCode(code.ErrUserPasswordIncorrect, "密码错误")
		}
		return nil, err
	}

	//生成token
	j := middlewares.NewJWT(us.jwtOpts.Key)
	claims := middlewares.CustomClaims{
		ID:          uint(user.ID),
		NickName:    user.NickName,
		AuthorityId: uint(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			NotBefore: jwt.NewNumericDate(time.Now()),                                   //签名的生效时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Local().Add(us.jwtOpts.Timeout)), //30天过期
			Issuer:    us.jwtOpts.Realm,
		},
	}
	token, err := j.CreateToken(claims)
	
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

	var user = &data.User{
		Mobile:   mobile,
		PassWord: password,
	}
	err = us.data.Users().Create(ctx, user)
	if err != nil {
		log.Errorf("user register failed: %v", err)
		return nil, err
	}

	// 直接生成token
	j := middlewares.NewJWT(us.jwtOpts.Key)
	claims := middlewares.CustomClaims{
		ID:          uint(user.ID),
		NickName:    user.NickName,
		AuthorityId: uint(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			NotBefore: jwt.NewNumericDate(time.Now()),                                   //签名的生效时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Local().Add(us.jwtOpts.Timeout)), //30天过期
			Issuer:    us.jwtOpts.Realm,
		},
	}
	token, err := j.CreateToken(claims)
	if err != nil {
		return nil, err
	}

	return &UserDTO{
		User:      *user,
		Token:     token,	// 向上传递token
		ExpiresAt: (time.Now().Local().Add(us.jwtOpts.Timeout)).Unix(),
	}, nil
}

func (u *userService) Update(ctx context.Context, userDTO *UserDTO) error {
	user := &data.User{
		ID:       userDTO.ID,
		Mobile:   userDTO.Mobile,
		NickName: userDTO.NickName,
		Birthday: userDTO.Birthday,
		Gender:   userDTO.Gender,
		Role:     userDTO.Role,
		PassWord: userDTO.PassWord,
	}
	return u.data.Users().Update(ctx, user)
}

func (us *userService) Get(ctx context.Context, userID uint64) (*UserDTO, error) {
	userDO, err := us.data.Users().Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &UserDTO{User: userDO}, nil
}

func (u *userService) GetByMobile(ctx context.Context, mobile string) (*UserDTO, error) {
	user, err := u.data.Users().GetByMobile(ctx, mobile)
	if err != nil {
		return nil, err
	}
	return &UserDTO{User: user}, nil
}

func (u *userService) CheckPassWord(ctx context.Context, password, EncryptedPassword string) (bool, error) {
	err := u.data.Users().CheckPassWord(ctx, password, EncryptedPassword)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (u *userService) GetUserList(ctx context.Context, pn, pSize uint32) (*UserListDTO, error) {
	userList, err := u.data.Users().List(ctx, pn, pSize)
	if err != nil {
		return nil, err
	}

	var userDTOs []*UserDTO
	for _, user := range userList.Items {
		userDTOs = append(userDTOs, &UserDTO{User: *user})
	}

	return &UserListDTO{
		TotalCount: userList.TotalCount,
		Items:      userDTOs,
	}, nil
}

var _ UserSrv = &userService{}
