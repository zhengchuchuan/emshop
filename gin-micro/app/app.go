package app

import (
	"emshop-admin/gin-micro/register"
	"os"
	"os/signal"
)

type App struct {
	opts options


}

func New(opts ...Option) *App {
	o := options{}
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
	// 注册服务


	// 监听退出信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, a.opts.sigs...)
	<-c

	return nil
}


// 停止服务
func (a *App) Stop() error {
	return nil
}

// 创建服务注册的结构体
func (a *App) buildInstance() (*register.ServiceInstance, error) {
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
	
	return &register.ServiceInstance{
		ID: 	   	a.opts.id,
		Name:     	a.opts.name,
		Endpoints: 	endpoints,
	}, nil
}