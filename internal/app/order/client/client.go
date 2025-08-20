package main

import (
	"context"
	"fmt"
	"github.com/hashicorp/consul/api"
	"math/rand"
	v1 "emshop/api/order/v1"
	"emshop/gin-micro/registry/consul"
	rpc "emshop/gin-micro/server/rpc-server"
	_ "emshop/gin-micro/server/rpc-server/resolver/direct"
	"emshop/gin-micro/server/rpc-server/selector"
	"emshop/gin-micro/server/rpc-server/selector/random"
	"time"
)

func generateOrderSn(userId int32) string {
	//订单号的生成规则
	/*
		年月日时分秒+用户id+2位随机数
	*/
	now := time.Now()
	rand.Seed(time.Now().UnixNano())
	orderSn := fmt.Sprintf("%d%d%d%d%d%d%d%d",
		now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Nanosecond(),
		userId, rand.Intn(90)+10,
	)
	return orderSn
}

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
		rpc.WithEndpoint("discovery:///emshop-order-srv"),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	uc := v1.NewOrderClient(conn)

	address := "慕课网"
	orderSn := generateOrderSn(1)
	name := "bobby"
	post := "尽快发货"
	mobile := "18787878787"
	_, err = uc.SubmitOrder(context.Background(), &v1.OrderRequest{
		UserId:  1,
		Address: &address,
		OrderSn: &orderSn,
		Name:    &name,
		Post:    &post,
		Mobile:  &mobile,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("订单创建成功")
}
