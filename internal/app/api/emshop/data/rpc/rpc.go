package rpc

import (
	"emshop/gin-micro/registry"
	"emshop/gin-micro/registry/consul"
	"emshop/internal/app/api/emshop/data"
	"emshop/internal/app/pkg/code"
	"emshop/internal/app/pkg/options"
	errors2 "emshop/pkg/errors"
	"fmt"
	"sync"

	cosulAPI "github.com/hashicorp/consul/api"
)

type grpcData struct {
	ud   data.UserData
	gd   data.GoodsData
	id   data.InventoryData
	od   data.OrderData
	uopd data.UserOpData
	cd   data.CouponData
	pd   data.PaymentData
	ld   data.LogisticsData
}

func (g grpcData) Goods() data.GoodsData {
	return g.gd
}

func (g grpcData) Users() data.UserData {
	return g.ud
}

func (g grpcData) Inventory() data.InventoryData {
	return g.id
}

func (g grpcData) Order() data.OrderData {
	return g.od
}

func (g grpcData) UserOp() data.UserOpData {
	return g.uopd
}

func (g grpcData) Coupon() data.CouponData {
	return g.cd
}

func (g grpcData) Payment() data.PaymentData {
	return g.pd
}

func (g grpcData) Logistics() data.LogisticsData {
	return g.ld
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

		// 创建客户端管理器，统一管理所有gRPC客户端
		clients := newGrpcClients(discovery)

		// 创建数据层实例，使用客户端管理器
		userData := NewUsers(clients.userClient)
		goodsData := NewGoods(clients.goodsClient)
		inventoryData := NewInventory(clients.inventoryClient)
		orderData := NewOrder(clients.orderClient)
		userOpData := NewUserOp(clients.userOpClient)
		couponData := NewCoupon(clients.couponClient)
		paymentData := NewPayment(clients.paymentClient)
		logisticsData := NewLogistics(clients.logisticsClient)

		dbFactory = &grpcData{
			ud:   userData,
			gd:   goodsData,
			id:   inventoryData,
			od:   orderData,
			uopd: userOpData,
			cd:   couponData,
			pd:   paymentData,
			ld:   logisticsData,
		}
	})

	if dbFactory == nil {
		return nil, errors2.WithCode(code.ErrConnectGRPC, "failed to get grpc store factory")
	}
	return dbFactory, nil
}
