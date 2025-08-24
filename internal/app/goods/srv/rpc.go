package srv

import (
	"fmt"
	gpb "emshop/api/goods/v1"
	"emshop/internal/app/goods/srv/config"
	v12 "emshop/internal/app/goods/srv/controller/v1"
	dataV1 "emshop/internal/app/goods/srv/data/v1"
	v1 "emshop/internal/app/goods/srv/service/v1"
	"emshop/gin-micro/core/trace"
	"emshop/gin-micro/server/rpc-server"

	"emshop/pkg/log"
)

func NewGoodsRPCServer(cfg *config.Config) (*rpcserver.Server, *dataV1.FactoryManager, error) {
	//初始化open-telemetry的exporter
	trace.InitAgent(trace.Options{
		Name:     cfg.Telemetry.Name,
		Endpoint: cfg.Telemetry.Endpoint,
		Sampler:  cfg.Telemetry.Sampler,
		Batcher:  cfg.Telemetry.Batcher,
	})

	// 使用新的工厂管理器
	factoryManager, err := dataV1.NewFactoryManager(cfg.MySQLOptions, cfg.EsOptions)
	if err != nil {
		log.Fatal(err.Error())
		return nil, nil, err
	}

	// 创建服务层
	srvFactory := v1.NewService(factoryManager)
	goodsServer := v12.NewGoodsServer(srvFactory)
	rpcAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	grpcServer := rpcserver.NewServer(rpcserver.WithAddress(rpcAddr))

	gpb.RegisterGoodsServer(grpcServer.Server, goodsServer)

	//r := gin.Default()
	//upb.RegisterUserServerHTTPServer(userver, r)
	//r.Run(":8075")
	return grpcServer, factoryManager, nil
}
