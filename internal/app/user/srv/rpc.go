package srv

import (
	"fmt"
	"log"

	"emshop/gin-micro/core/trace"
	rpcserver "emshop/gin-micro/server/rpc-server"
	"emshop/internal/app/user/srv/config"
	"emshop/internal/app/user/srv/controller/user"
	"emshop/internal/app/user/srv/data/v1/db"
	upb "emshop/api/user/v1"
	srv1 "emshop/internal/app/user/srv/service/v1"
)

func NewUserRPCServer(cfg *config.Config) (*rpcserver.Server, error) {
    //初始化open-telemetry的exporter
    trace.InitAgent(trace.Options{
        Name:     cfg.Telemetry.Name,     
        Endpoint: cfg.Telemetry.Endpoint, 
        Sampler:  cfg.Telemetry.Sampler,  
        Batcher:  cfg.Telemetry.Batcher,  
    })


	gormDB, err := db.GetDBFactoryOr(cfg.MySQLOptions)
	if err != nil {
		log.Fatal(err.Error())
		
	}
    data := db.NewUsers(gormDB)
    srv := srv1.NewUserService(data)
    userver := user.NewUserServer(srv)

    rpcAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
    urpcServer := rpcserver.NewServer(rpcserver.WithAddress(rpcAddr))
    upb.RegisterUserServer(urpcServer.Server, userver)


	return urpcServer, nil
}