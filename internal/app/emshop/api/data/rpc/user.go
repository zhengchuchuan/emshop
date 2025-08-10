package rpc

import (
	"context"
	"emshop/internal/app/pkg/code"
	"emshop/gin-micro/server/rpc-server"
	"emshop/gin-micro/server/rpc-server/client-interceptors"
	"emshop/pkg/errors"
	"emshop/pkg/log"
	"time"

	upbv1 "emshop/api/user/v1"
	"emshop/internal/app/emshop/api/data"
	"emshop/gin-micro/registry"
	itime "emshop/pkg/common/time"
	"google.golang.org/grpc"
)

const serviceName = "discovery:///emshop-user-srv"

type users struct {
	uc upbv1.UserClient		// 直接使用 UserClient 接口
}

func NewUsers(uc upbv1.UserClient) *users {
	return &users{uc}
}

func NewUserServiceClient(r registry.Discovery) upbv1.UserClient {
	log.Infof("Initializing gRPC connection to service: %s", serviceName)
	conn, err := rpcserver.DialInsecure(
		context.Background(),
		rpcserver.WithEndpoint(serviceName),
		rpcserver.WithDiscovery(r),	// 使用服务发现
		rpcserver.WithClientTimeout(10*time.Second), // 增加连接超时时间到10秒
		rpcserver.WithClientOptions(grpc.WithNoProxy()), // 禁用代理
		rpcserver.WithClientUnaryInterceptor(clientinterceptors.UnaryTracingInterceptor), // 添加链路追踪拦截器
	)
	if err != nil {
		log.Errorf("Failed to create gRPC connection: %v", err)
		panic(err)
	}
	log.Info("gRPC connection established successfully")
	c := upbv1.NewUserClient(conn)
	return c
}

func (u *users) CheckPassWord(ctx context.Context, password, encryptedPwd string) error {
	cres, err := u.uc.CheckPassWord(ctx, &upbv1.PasswordCheckInfo{
		Password:          password,
		EncryptedPassword: encryptedPwd,
	})
	if err != nil {
		return err
	}
	if cres.Success {
		return nil
	}
	return errors.WithCode(code.ErrUserPasswordIncorrect, "密码错误")
}

func (u *users) Create(ctx context.Context, user *data.User) error {
	protoUser := &upbv1.CreateUserInfo{
		Mobile:   user.Mobile,
		NickName: user.NickName,
		PassWord: user.PassWord,
	}
	log.Infof("Calling CreateUser gRPC for mobile: %s", user.Mobile)
	userRsp, err := u.uc.CreateUser(ctx, protoUser)
	if err != nil {
		log.Errorf("CreateUser gRPC call failed: %v", err)
		return err
	}
	log.Infof("CreateUser gRPC call successful, user ID: %d", userRsp.Id)
	user.ID = uint64(userRsp.Id)
	return nil
}

func (u *users) Update(ctx context.Context, user *data.User) error {
	protoUser := &upbv1.UpdateUserInfo{
		Id:       int32(user.ID),
		NickName: user.NickName,
		Gender:   user.Gender,
		BirthDay: uint64(user.Birthday.Unix()),
	}
	_, err := u.uc.UpdateUser(ctx, protoUser)
	if err != nil {
		return err
	}
	return nil
}

func (u *users) Get(ctx context.Context, userID uint64) (data.User, error) {
	user, err := u.uc.GetUserById(ctx, &upbv1.IdRequest{
		Id: int32(userID),
	})
	if err != nil {
		return data.User{}, err
	}

	return data.User{
		ID:       uint64(user.Id),
		Mobile:   user.Mobile,
		NickName: user.NickName,
		Birthday: itime.Time{time.Unix(int64(user.BirthDay), 0)},
		Gender:   user.Gender,
		Role:     user.Role,
		PassWord: user.PassWord,
	}, nil
}

func (u *users) GetByMobile(ctx context.Context, mobile string) (data.User, error) {
	user, err := u.uc.GetUserByMobile(ctx, &upbv1.MobileRequest{
		Mobile: mobile,
	})
	if err != nil {
		return data.User{}, err
	}

	return data.User{
		ID:       uint64(user.Id),
		Mobile:   user.Mobile,
		NickName: user.NickName,
		Birthday: itime.Time{time.Unix(int64(user.BirthDay), 0)},
		Gender:   user.Gender,
		Role:     user.Role,
		PassWord: user.PassWord,
	}, nil
}

var _ data.UserData = &users{}
