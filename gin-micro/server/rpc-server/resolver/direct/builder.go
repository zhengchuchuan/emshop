package direct

import (
	"strings"

	"google.golang.org/grpc/resolver"
)


func init() {
	resolver.Register(NewBuilder())
}

type directBuilder struct{}

//	direct://<authority>/127.0.0.1:9000,
func NewBuilder() *directBuilder {
	return &directBuilder{}
}


func (d *directBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	addrs := make([]resolver.Address, 0)
	// 1. 提取路径部分：/127.0.0.1:9000,127.0.0.1:9001  
    // 2. 去掉前缀"/"：127.0.0.1:9000,127.0.0.1:9001
    // 3. 按逗号分割：["127.0.0.1:9000", "127.0.0.1:9001"]
	for _, addr := range strings.Split(strings.TrimPrefix(target.URL.Path, "/"), ",") {
		addrs = append(addrs, resolver.Address{Addr: addr})
	}

	// 直接更新gRPC连接状态，无需监听变化
	err := cc.UpdateState(resolver.State{Addresses: addrs})
	if err != nil {
		return nil, err
	}
	return newDirectResolver(), nil
}

func (d *directBuilder) Scheme() string {
	return "direct"
}

var _ resolver.Builder = &directBuilder{}
