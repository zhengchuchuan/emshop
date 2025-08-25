package main

import (
	"context"
	"emshop/internal/app/logistics/srv/app"
	apputil "emshop/pkg/app"
	"emshop/pkg/log"
	"math/rand"
	"time"
)

// Name 定义服务名称
const Name = "emshop-logistics-srv"

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	logisticsApp := apputil.NewApp(Name,
		apputil.WithDescription("E-commerce logistics service"),
		apputil.WithOptions(app.NewLogisticsOptions()),
		apputil.WithRunFunc(run),
	).BuildCommand()

	if err := logisticsApp.Execute(); err != nil {
		log.Fatalw(err.Error())
	}
}

func run(ctx context.Context, basename string, opts interface{}) error {
	return app.Run(ctx, opts)
}