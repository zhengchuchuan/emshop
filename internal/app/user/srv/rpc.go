package srv

import (
	rpcserver "emshop/gin-micro/server/rpc-server"
	"emshop/internal/app/user/srv/config"
)

func NewUserRPCServer(cfg *config.Config) (*rpcserver.Server, error) {
	rpcserver.NewServer()
	return nil, nil
}