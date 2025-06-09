package main

import (
	"context"
	v1 "emshop-admin/api/user/v1"
	"emshop-admin/gin-micro/registry/consul"
	rpc "emshop-admin/gin-micro/server/rpc-server"
	_ "emshop-admin/gin-micro/server/rpc-server/resolver/direct"
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
)

func main() {
	//设置全局的负载均衡策略
	// selector.SetGlobalSelector(random.NewBuilder())
	// rpc.InitBuilder()

	// 注册consul的服务发现配置
	conf := api.DefaultConfig()
	conf.Address = "127.0.0.1:8500"
	conf.Scheme = "http"
	cli, err := api.NewClient(conf)
	if err != nil {
		panic(err)
	}
	r := consul.New(cli, consul.WithHealthCheck(true))

	conn, err := rpc.DialInsecure(context.Background(),
		// rpc.WithBalancerName("selector"),
		rpc.WithDiscovery(r),
		rpc.WithClientTimeout(time.Second*5000),
		rpc.WithEndpoint("discovery:///emshop-user-srv"),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	uc := v1.NewUserClient(conn)

	for {
		_, err := uc.GetUserList(context.Background(), &v1.PageInfo{})
		if err != nil {
			panic(err)
		}

		fmt.Println("success")
		time.Sleep(time.Millisecond * 2)
	}

}
