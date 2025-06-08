package srv

import (
	"fmt"

	gapp "emshop-admin/gin-micro/app"
	"emshop-admin/internal/app/user/srv/config"
	"emshop-admin/pkg/app"
	"emshop-admin/pkg/log"
)

func NewApp(basename string) *app.App {
	cfg := config.New()
	appl := app.NewApp("user",
		"emshop-admin",
		app.WithOptions(cfg),
		app.WithRunFunc(run(cfg)),
		//app.WithNoConfig(), //设置不读取配置文件
	)
	return appl
}

func NewUserApp(cfg *config.Config) (*gapp.App, error) {
	// 初始化log
	log.Init(cfg.Log)
	defer log.Flush()

	// 服务注册
	rpcServer, err := NewUserRPCServer(cfg)

	return nil, nil
}

func run(cfg *config.Config) app.RunFunc {
	return func(baseName string) error {
		fmt.Println(cfg.Log.Level)
		return nil
	}
}

