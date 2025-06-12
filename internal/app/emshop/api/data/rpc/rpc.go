package rpc

import (
	"fmt"
	cosulAPI "github.com/hashicorp/consul/api"
	// gpb "emshop/api/goods/v1"
	upb "emshop/api/user/v1"
	"emshop/internal/app/emshop/api/data"
	"emshop/internal/app/pkg/code"
	"emshop/internal/app/pkg/options"
	"emshop/gin-micro/registry"
	"emshop/gin-micro/registry/consul"
	errors2 "emshop/pkg/errors"
	"sync"
)

type grpcData struct {
	gc gpb.GoodsClient
	uc upb.UserClient
}

func (g grpcData) Goods() gpb.GoodsClient {
	//TODO implement me
	panic("implement me")
}

func (g grpcData) Users() data.UserData {
	//TODO implement me
	panic("implement me")
}

func NewDiscovery(opts *options.RegistryOptions) registry.Discovery {
	c := cosulAPI.DefaultConfig()
	c.Address = opts.Address
	c.Scheme = opts.Scheme
	cli, err := cosulAPI.NewClient(c)
	if err != nil {
		panic(err)
	}
	r := consul.New(cli, consul.WithHealthCheck(true))
	return r
}

var (
	dbFactory data.DataFactory
	once      sync.Once
)

// rpc的连接， 基于服务发现
func GetDataFactoryOr(options *options.RegistryOptions) (data.DataFactory, error) {
	if options == nil && dbFactory == nil {
		return nil, fmt.Errorf("failed to get grpc store fatory")
	}

	//这里负责依赖的所有的rpc连接
	once.Do(func() {
		discovery := NewDiscovery(options)
		userClient := NewUserServiceClient(discovery)
		// goodsClient := NewGoodsServiceClient(discovery)
		dbFactory = &grpcData{
			// gc: goodsClient,
			uc: userClient,
		}
	})

	if dbFactory == nil {
		return nil, errors2.WithCode(code.ErrConnectGRPC, "failed to get grpc store factory")
	}
	return dbFactory, nil
}
