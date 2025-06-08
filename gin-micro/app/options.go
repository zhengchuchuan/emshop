package app

import (
	"net/url"
	"os"
)

type options struct {
	id 	  string    // 服务的唯一标识
	name string    // 服务的名称
	endpoints []url.URL // 服务的地址列表

	sigs []os.Signal // 监听的信号
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