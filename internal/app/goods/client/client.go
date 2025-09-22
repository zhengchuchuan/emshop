package main

import (
	"context"
	"fmt"
	"github.com/hashicorp/consul/api"
	v1 "emshop/api/goods/v1"
	"emshop/gin-micro/registry/consul"
	rpc "emshop/gin-micro/server/rpc-server"
	_ "emshop/gin-micro/server/rpc-server/resolver/direct"
	"emshop/gin-micro/server/rpc-server/selector"
	"emshop/gin-micro/server/rpc-server/selector/p2c"
	"time"
)

func main() {
	//设置全局的负载均衡策略
	selector.SetGlobalSelector(p2c.NewBuilder())
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
		rpc.WithEndpoint("discovery:///emshop-goods-srv"),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	uc := v1.NewGoodsClient(conn)

	keyWords := "猕猴桃"
	re, err := uc.GoodsList(context.Background(), &v1.GoodsFilterRequest{
		KeyWords: &keyWords,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(re)

}
