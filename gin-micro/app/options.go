package app

import (
	// "emshop-admin/gin-micro/registry" // Replace with the actual path to the registry package
	"emshop-admin/gin-micro/registry"
	"net/url"
	"os"
	"time"
)

type options struct {
	id 	  string    // 服务的唯一标识
	name string    // 服务的名称
	endpoints []url.URL // 服务的地址列表

	sigs []os.Signal // 监听的信号

	registrar registry.Registrar // 服务注册中心的注册器
	registrarTimeout time.Duration

	// //stop超时时间
	// stopTimeout time.Duration

	// restServer *restserver.Server
	// rpcServer  *rpcserver.Server
}


// 函数选项模式
type Option func(o *options)




func WithID(id string) Option {
	return func(o *options) {
		o.id = id
	}
}

func WithName(name string) Option {
	return func(o *options) {
		o.name = name
	}
}

func WithOptions(endpotints []url.URL) Option {
	return func(o *options) {
		o.endpoints = endpotints
	}
}

func WithSigs(sigs []os.Signal) Option {
	return func(o *options) {
		o.sigs = sigs
	}
}