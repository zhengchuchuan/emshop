package admin

import (
	restserver "emshop/gin-micro/server/rest-server"
	"emshop/internal/app/api/admin/config"
)

func NewAdminHTTPServer(cfg *config.Config) (*restserver.Server, error) {
	urestServer := restserver.NewServer(restserver.WithPort(cfg.Server.HttpPort),
		restserver.WithMiddlewares(cfg.Server.Middlewares),
		restserver.WithMetrics(cfg.Server.EnableMetrics),
	)

	//配置好路由
	initRouter(urestServer, cfg)

	return urestServer, nil
}
