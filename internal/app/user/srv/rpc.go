package srv

import (
	"emshop/gin-micro/core/trace"
	rpcserver "emshop/gin-micro/server/rpc-server"
	"emshop/internal/app/user/srv/config"
)

func NewUserRPCServer(cfg *config.Config) (*rpcserver.Server, error) {
	//初始化open-telemetry的exporter
	trace.InitAgent(trace.Options{
		cfg.Telemetry.Name,
		cfg.Telemetry.Endpoint,
		cfg.Telemetry.Sampler,
		cfg.Telemetry.Batcher,
	})

	// data := mock.NewUsers()
	// srv := srv1.NewUserService(data)
	// userver := user.NewUserServer(srv)

	// rpcAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	// urpcServer := rpcserver.NewServer(
	// 	rpcserver.WithAddress(rpcAddr))
	// upb.RegisterUserServer(urpcServer.Server, userver)
	rpcserver.NewServer()
	return nil, nil
}