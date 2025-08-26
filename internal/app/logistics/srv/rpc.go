package srv

import (
	"fmt"
	logisticspb "emshop/api/logistics/v1"
	"emshop/internal/app/logistics/srv/config"
	v1 "emshop/internal/app/logistics/srv/controller/logistics/v1"
	datav1 "emshop/internal/app/logistics/srv/data/v1"
	service "emshop/internal/app/logistics/srv/service/v1"
	"emshop/gin-micro/core/trace"
	"emshop/gin-micro/server/rpc-server"
	"emshop/pkg/log"
)

func NewLogisticsRPCServer(cfg *config.Config) (*rpcserver.Server, error) {
	//初始化open-telemetry的exporter
	trace.InitAgent(trace.Options{
		Name:     cfg.Telemetry.Name,
		Endpoint: cfg.Telemetry.Endpoint,
		Sampler:  cfg.Telemetry.Sampler,
		Batcher:  cfg.Telemetry.Batcher,
	})

	// 创建数据工厂管理器
	factoryManager, err := datav1.NewFactoryManager(cfg.MySQLOptions)
	if err != nil {
		log.Fatal(err.Error())
		return nil, err
	}

	// 创建业务服务层
	logisticsSrv := service.NewLogisticsService(factoryManager.GetDataFactory(), cfg.Redis)

	// 创建控制器
	logisticsServer := v1.NewLogisticsController(logisticsSrv)
	
	rpcAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	grpcServer := rpcserver.NewServer(rpcserver.WithAddress(rpcAddr))

	logisticspb.RegisterLogisticsServer(grpcServer.Server, logisticsServer)

	log.Infof("物流gRPC服务注册成功，监听地址: %s", rpcAddr)

	return grpcServer, nil
}