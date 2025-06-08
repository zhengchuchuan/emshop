package srv

import (
	rpcserver "emshop-admin/gin-micro/server/rpc-server"
	"emshop-admin/internal/app/user/srv/config"
)

func NewUserRPCServer(cfg *config.Config) (*rpcserver.Server, error) {
	rpcserver.NewServer()
	return nil, nil
}