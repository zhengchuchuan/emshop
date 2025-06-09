package app

import (
	"emshop-admin/gin-micro/registry"
	"emshop-admin/pkg/log"
	"net/url"
	"syscall"
	"time"

	"os"
	"os/signal"
	"sync"

	"context"

	"github.com/google/uuid"
)

type App struct {
	opts options

	lk	sync.Mutex // 锁, 用于保护instance的并发访问
	instance *registry.ServiceInstance // 服务注册的实例
}

func New(opts ...Option) *App {
	// 通过函数选项模式设置默认值

	// 默认选项
	o := options{
		sigs:             []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT},
		registrarTimeout: 10 * time.Second,
		stopTimeout:      10 * time.Second,
	}

	if id, err := uuid.NewUUID(); err == nil {
		o.id = id.String()
	} else {
		log.Errorf("generate uuid error: %s", err)
	}

	for _, opt := range opts {
		opt(&o)
	}

	return &App{
		opts: o,
	}
}


// 启动整个服务
func (a *App) Run() error {
	// 注册的信息
	instance, err := a.buildInstance()
	if err != nil {
		return err
	}

	// 可能被其他goroutine修改,需要保护此变量
	a.lk.Lock()
	a.instance = instance
	a.lk.Unlock()
	// 启动RPC服务
	// if a.opts.rpcServer != nil {
	// 	err := a.opts.rpcServer.Start(context.Background())
	// 	if err != nil {
	// 		log.Errorf("start rpc server error: %s", err)
	// 		return err
	// 	}
	// }
	// go func() {
	// 	err := a.opts.rpcServer.Start(context.Background())
	// 	if err != nil {
	// 		log.Errorf("start rpc server error: %s", err)
	// 		panic(err)
	// 	}
	// }()


	// 注册服务
	if a.opts.registrar != nil {
		rctx, rcancel := context.WithTimeout(context.Background(), a.opts.registrarTimeout)
		a.opts.registrar.Register(rctx, a.instance)

		defer rcancel()

		if err != nil {
			log.Errorf("register service error: %s", err)
			return err
		}
	}


	// 监听退出信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, a.opts.sigs...)
	<-c

	return nil
}


// 停止服务
func (a *App) Stop() error {
	a.lk.Lock()
	instance := a.instance
	a.lk.Unlock()

	log.Info("start deregister service")

	if a.opts.registrar != nil && instance != nil {
		rctx, rcancel := context.WithTimeout(context.Background(), a.opts.registrarTimeout)
		defer rcancel()
		if err := a.opts.registrar.Deregister(rctx, instance); err != nil {
			log.Errorf("deregister service error: %s", err)
			return err
		}
	}



	return nil
}

// 创建服务注册的结构体
func (a *App) buildInstance() (*registry.ServiceInstance, error) {
	// 初始化一些组件
	// 1. 初始化日志
	// 2. 初始化配置
	// 3. 初始化数据库连接
	// 4. 初始化缓存连接
	// 5. 初始化服务注册中心连接
	// 6. 初始化其他组件
	endpoints := make([]string, 0)
	for _, e := range a.opts.endpoints {
		endpoints = append(endpoints, e.String())
	}

	//从rpcserver， restserver去主动获取这些信息
	if a.opts.rpcServer != nil {
		if a.opts.rpcServer.Endpoint() != nil {
			endpoints = append(endpoints, a.opts.rpcServer.Endpoint().String())
		} else {
			u := &url.URL{
				Scheme: "grpc",
				Host:   a.opts.rpcServer.Address(),
			}
			endpoints = append(endpoints, u.String())
		}
	}
	
	return &registry.ServiceInstance{
		ID: 	   	a.opts.id,
		Name:     	a.opts.name,
		Endpoints: 	endpoints,
	}, nil
}