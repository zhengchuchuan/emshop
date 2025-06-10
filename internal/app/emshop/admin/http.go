package admin

import (
	"emshop/internal/app/user/srv/config"
	"emshop/gin-micro/server/rest-server"
)

func NewUserHTTPServer(cfg *config.Config) (*restserver.Server, error) {
	urestServer := restserver.NewServer(restserver.WithPort(cfg.Server.HttpPort),
		restserver.WithMiddlewares(cfg.Server.Middlewares),
		restserver.WithMetrics(true),
	)

	//配置好路由
	initRouter(urestServer)

	return urestServer, nil
}
