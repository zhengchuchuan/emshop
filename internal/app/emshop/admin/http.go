package admin

import (
	"emshop/internal/app/emshop/admin/config"
	"emshop/gin-micro/server/rest-server"
)

func NewAdminHTTPServer(cfg *config.Config) (*restserver.Server, error) {
	urestServer := restserver.NewServer(restserver.WithPort(cfg.Server.HttpPort),
		restserver.WithMiddlewares(cfg.Server.Middlewares),
		restserver.WithMetrics(true),
	)

	//配置好路由
	initRouter(urestServer)

	return urestServer, nil
}
