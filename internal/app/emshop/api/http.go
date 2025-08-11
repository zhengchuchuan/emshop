package admin

import (
	"emshop/internal/app/emshop/api/config"
	"emshop/gin-micro/server/rest-server"
)

func NewAPIHTTPServer(cfg *config.Config) (*restserver.Server, error) {
	aRestServer := restserver.NewServer(
		restserver.WithPort(cfg.Server.HttpPort),
		restserver.WithMiddlewares(cfg.Server.Middlewares),
		restserver.WithMetrics(true),
		restserver.WithTransNames(cfg.I18n.Locale),
		restserver.WithLocalesDir(cfg.I18n.LocalesDir),
		restserver.WithRouterInit(func(server *restserver.Server, configInterface interface{}) {
			initRouter(server, configInterface.(*config.Config))
		}, cfg), // 延迟路由初始化，在翻译器初始化后执行
	)

	return aRestServer, nil
}
