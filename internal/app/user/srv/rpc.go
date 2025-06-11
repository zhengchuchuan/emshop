package srv

import (
	
	"fmt"

	rpcserver "emshop/gin-micro/server/rpc-server"
	"emshop/internal/app/user/srv/config"
	"emshop/internal/app/user/srv/controller/user"
	"emshop/internal/app/user/srv/data/v1/mock"
	"emshop/gin-micro/core/trace"
	upb "emshop/api/user/v1"
)

func NewUserRPCServer(cfg *config.Config) (*rpcserver.Server, error) {
	//初始化open-telemetry的exporter
	trace.InitAgent(trace.Options{
		cfg.Telemetry.Name,
		cfg.Telemetry.Endpoint,
		cfg.Telemetry.Sampler,
		cfg.Telemetry.Batcher,
	})

	data := mock.NewUsers()
	srv := srv1.NewUserService(data)
	userver := user.NewUserServer(srv)

	rpcAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	urpcServer := rpcserver.NewServer(
		rpcserver.WithAddress(rpcAddr))
	upb.RegisterUserServer(urpcServer.Server, userver)
	rpcserver.NewServer()
	return urpcServer, nil
}