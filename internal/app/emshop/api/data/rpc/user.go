package rpc

import (
	"context"
	"emshop/gin-micro/server/rpc-server"
	"emshop/gin-micro/server/rpc-server/client-interceptors"
	"emshop/pkg/log"
	"time"

	upbv1 "emshop/api/user/v1"
	"emshop/internal/app/emshop/api/data"
	"emshop/gin-micro/registry"
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

func (u *users) CheckPassWord(ctx context.Context, request *upbv1.PasswordCheckInfo) (*upbv1.CheckResponse, error) {
	return u.uc.CheckPassWord(ctx, request)
}

func (u *users) CreateUser(ctx context.Context, request *upbv1.CreateUserInfo) (*upbv1.UserInfoResponse, error) {
	return u.uc.CreateUser(ctx, request)
}

func (u *users) UpdateUser(ctx context.Context, request *upbv1.UpdateUserInfo) (*upbv1.UserInfoResponse, error) {
	_, err := u.uc.UpdateUser(ctx, request)
	if err != nil {
		return nil, err
	}
	// UpdateUser 返回 Empty，所以我们需要重新获取用户信息
	return u.uc.GetUserById(ctx, &upbv1.IdRequest{Id: request.Id})
}

func (u *users) GetUserById(ctx context.Context, request *upbv1.IdRequest) (*upbv1.UserInfoResponse, error) {
	return u.uc.GetUserById(ctx, request)
}

func (u *users) GetUserByMobile(ctx context.Context, request *upbv1.MobileRequest) (*upbv1.UserInfoResponse, error) {
	return u.uc.GetUserByMobile(ctx, request)
}

func (u *users) GetUserList(ctx context.Context, request *upbv1.PageInfo) (*upbv1.UserListResponse, error) {
	return u.uc.GetUserList(ctx, request)
}

var _ data.UserData = &users{}
