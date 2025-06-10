package app

import (
	"emshop/gin-micro/registry"
	gs "emshop/gin-micro/server"
	"emshop/pkg/log"
	"net/url"
	"syscall"
	"time"

	"os"
	"os/signal"
	"sync"

	"context"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type App struct {
	opts options

	lk	sync.Mutex // 锁, 用于保护instance的并发访问
	instance *registry.ServiceInstance // 服务注册的实例

	cancel func()
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


	//现在启动了两个server，一个是restserver，一个是rpcserver
	/*
		这两个server是否必须同时启动成功？
		如果有一个启动失败，那么我们就要停止另外一个server
		如果启动了多个， 如果其中一个启动失败，其他的应该被取消
			如果剩余的server的状态：
				1. 还没有开始调用start
					stop
				2. start进行中
					调用进行中的cancel
				3. start已经完成
					调用stop
		如果我们的服务启动了然后这个时候用户立马进行了访问
	*/

	// servers 可以添加种server
	var servers []gs.Server
	if a.opts.restServer != nil {
		servers = append(servers, a.opts.restServer)
	}
	if a.opts.rpcServer != nil {
		servers = append(servers, a.opts.rpcServer)
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel	// 保存取消函数,后续的 Stop() 方法可以调用它来取消所有的操作
	eg, ctx := errgroup.WithContext(ctx)
	wg := sync.WaitGroup{}
	for _, srv := range servers {
		//启动server
		//在启动一个goroutine 去监听是否有err产生
		srv := srv
		eg.Go(func() error {
			<-ctx.Done() //wait for stop signal
			//不可能无休止的等待stop
			sctx, cancel := context.WithTimeout(context.Background(), a.opts.stopTimeout)
			defer cancel()
			return srv.Stop(sctx)
		})

		wg.Add(1)
		eg.Go(func() error {
			wg.Done()
			log.Info("start rest server")
			return srv.Start(ctx)
		})
	}

	wg.Wait()

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


	//监听退出信息
	c := make(chan os.Signal, 1)
	signal.Notify(c, a.opts.sigs...)
	eg.Go(func() error {
		// 等待上下文取消或接收到退出信号
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c:
			return a.Stop()
		}
	})
	if err := eg.Wait(); err != nil {
		return err
	}
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