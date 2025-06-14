//go:build wireinject
// +build wireinject

package srv

import (
	"github.com/google/wire"
	"emshop/internal/app/pkg/options"
	"emshop/internal/app/user/srv/controller/user"
	"emshop/internal/app/user/srv/data/v1/db"
	v1 "emshop/internal/app/user/srv/service/v1"
	gapp "emshop/gin-micro/app"
	"emshop/pkg/log"
)

func initApp(*options.NacosOptions, *log.Options, *options.ServerOptions, *options.RegistryOptions, *options.TelemetryOptions, *options.MySQLOptions) (*gapp.App, error) {
	wire.Build(ProviderSet, v1.ProviderSet, db.ProviderSet, user.ProviderSet)
	return &gapp.App{}, nil
}
