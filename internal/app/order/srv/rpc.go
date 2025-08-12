package srv

import (
	gpb "emshop/api/order/v1"
	"emshop/gin-micro/core/trace"
	"emshop/gin-micro/server/rpc-server"
	"emshop/internal/app/order/srv/config"
	"emshop/internal/app/order/srv/controller/order/v1"
	v1 "emshop/internal/app/order/srv/data/v1"
	v13 "emshop/internal/app/order/srv/service/v1"
	"fmt"

	"emshop/pkg/log"
)

func NewOrderRPCServer(cfg *config.Config) (*rpcserver.Server, error) {
	//初始化open-telemetry的exporter
	trace.InitAgent(trace.Options{
		cfg.Telemetry.Name,
		cfg.Telemetry.Endpoint,
		cfg.Telemetry.Sampler,
		cfg.Telemetry.Batcher,
	})

	factoryManager, err := v1.NewFactoryManager(cfg.MySQLOptions, cfg.Registry)
	if err != nil {
		log.Fatal(err.Error())
	}

	orderSrvFactory := v13.NewService(factoryManager.GetDataFactory(), cfg.Dtm)
	orderServer := order.NewOrderServer(orderSrvFactory)
	rpcAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	grpcServer := rpcserver.NewServer(rpcserver.WithAddress(rpcAddr))
	gpb.RegisterOrderServer(grpcServer.Server, orderServer)
	return grpcServer, nil
}
