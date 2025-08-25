package srv

import (
	gpb "emshop/api/payment/v1"
	"emshop/gin-micro/core/trace"
	"emshop/gin-micro/server/rpc-server"
	"emshop/internal/app/payment/srv/config"
	"emshop/internal/app/payment/srv/controller/payment/v1"
	v1data "emshop/internal/app/payment/srv/data/v1/mysql"
	v1service "emshop/internal/app/payment/srv/service/v1"
	"fmt"

	"emshop/pkg/log"
)

func NewPaymentRPCServer(cfg *config.Config) (*rpcserver.Server, error) {
	//初始化open-telemetry的exporter
	trace.InitAgent(trace.Options{
		cfg.Telemetry.Name,
		cfg.Telemetry.Endpoint,
		cfg.Telemetry.Sampler,
		cfg.Telemetry.Batcher,
	})

	// 初始化数据工厂
	dataFactory, err := v1data.NewDataFactory(cfg.MySQLOptions)
	if err != nil {
		log.Fatal(err.Error())
		return nil, err
	}

	// 初始化服务工厂
	paymentSrvFactory := v1service.NewService(dataFactory, cfg.Dtm, cfg.Redis)
	
	// 创建gRPC服务器
	paymentServer := payment.NewPaymentServer(paymentSrvFactory)
	rpcAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	grpcServer := rpcserver.NewServer(rpcserver.WithAddress(rpcAddr))
	
	// 注册服务
	gpb.RegisterPaymentServer(grpcServer.Server, paymentServer)
	
	return grpcServer, nil
}