//go:build wireinject
// +build wireinject

package srv

import (
	"github.com/google/wire"
	"emshop/internal/app/pkg/options"
	"emshop/internal/app/user/srv/controller/user"
	v1data "emshop/internal/app/user/srv/data/v1"
	v1 "emshop/internal/app/user/srv/service/v1"
	gapp "emshop/gin-micro/app"
	"emshop/pkg/log"
)

// wire Injector
func initApp(*options.NacosOptions, *log.Options, *options.ServerOptions, *options.RegistryOptions, *options.TelemetryOptions, *options.MySQLOptions) (*gapp.App, error) {
	// 会在wire_gen文件中生成注入器的具体实现
	wire.Build(ProviderSet, v1.ProviderSet, v1data.ProviderSet, user.ProviderSet)
	return &gapp.App{}, nil
}
