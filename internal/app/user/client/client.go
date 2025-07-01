package main

import (
	"context"
	v1 "emshop/api/user/v1"
	"emshop/gin-micro/registry/consul"
	rpc "emshop/gin-micro/server/rpc-server"
	_ "emshop/gin-micro/server/rpc-server/resolver/direct"
	"emshop/gin-micro/server/rpc-server/selector"
	"emshop/gin-micro/server/rpc-server/selector/random"
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
)

func main() {
	//设置全局的负载均衡策略
	selector.SetGlobalSelector(random.NewBuilder())
	rpc.InitBuilder()

	conf := api.DefaultConfig()
	conf.Address = "127.0.0.1:8500"
	conf.Scheme = "http"
	cli, err := api.NewClient(conf)
	if err != nil {
		panic(err)
	}
	r := consul.New(cli, consul.WithHealthCheck(true))

	conn, err := rpc.DialInsecure(context.Background(),
		rpc.WithBalancerName("selector"),
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
		re, err := uc.GetUserList(context.Background(), &v1.PageInfo{Pn: 1, PSize: 10})
		if err != nil {
			panic(err)
		}
		fmt.Println(re)	
		fmt.Println("success")
		time.Sleep(time.Millisecond * 2)
	}

}
