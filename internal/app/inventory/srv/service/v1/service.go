package v1

import (
	v1 "emshop/internal/app/inventory/srv/data/v1"
	"emshop/internal/app/pkg/options"
	"fmt"

	goredislib "github.com/go-redis/redis/v8"

	redsyncredis "github.com/go-redsync/redsync/v4/redis"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
)

type ServiceFactory interface {
	Inventorys() InventorySrv
}

type service struct {
	data v1.DataFactory

	redisOptions *options.RedisOptions
	pool         redsyncredis.Pool
}

func (s *service) Inventorys() InventorySrv {
	//TODO implement me
	return newInventoryService(s)
}

func NewService(store v1.DataFactory, redisOptions *options.RedisOptions) *service {
	client := goredislib.NewClient(&goredislib.Options{
		Addr: fmt.Sprintf("%s:%d", redisOptions.Host, redisOptions.Port),
	})
	pool := goredis.NewPool(client) // or, pool := redigo.NewPool(...)

	return &service{data: store, redisOptions: redisOptions, pool: pool}
}

var _ ServiceFactory = &service{}
