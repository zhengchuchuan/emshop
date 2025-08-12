package srv

import (
	gpb "emshop/api/inventory/v1"
	"emshop/internal/app/inventory/srv/config"
	"emshop/gin-micro/core/trace"
	"emshop/gin-micro/server/rpc-server"
	v12 "emshop/internal/app/inventory/srv/controller/v1"
	v1 "emshop/internal/app/inventory/srv/data/v1"
	v13 "emshop/internal/app/inventory/srv/service/v1"
	"fmt"

	"emshop/pkg/log"
)

func NewInventoryRPCServer(cfg *config.Config) (*rpcserver.Server, error) {
	//初始化open-telemetry的exporter
	trace.InitAgent(trace.Options{
		cfg.Telemetry.Name,
		cfg.Telemetry.Endpoint,
		cfg.Telemetry.Sampler,
		cfg.Telemetry.Batcher,
	})

	//有点繁琐，wire， ioc-golang
	factoryManager, err := v1.NewFactoryManager(cfg.MySQLOptions)
	if err != nil {
		log.Fatal(err.Error())
	}
	invService := v13.NewService(factoryManager.GetDataFactory(), cfg.RedisOptions)
	invServer := v12.NewInventoryServer(invService)
	rpcAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	grpcServer := rpcserver.NewServer(rpcserver.WithAddress(rpcAddr))
	gpb.RegisterInventoryServer(grpcServer.Server, invServer)
	//r := gin.Default()
	//upb.RegisterUserServerHTTPServer(userver, r)
	//r.Run(":8075")
	return grpcServer, nil
}
