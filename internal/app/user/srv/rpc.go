package srv

import (
	"fmt"

	upb "emshop/api/user/v1"
	"emshop/gin-micro/core/trace"
	rpcserver "emshop/gin-micro/server/rpc-server"
	"emshop/internal/app/pkg/options"
	"emshop/pkg/sentinel"

	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/pkg/datasource/nacos"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
)

func NewNacosDataSource(opts *options.NacosOptions) (*nacos.NacosDataSource, error) {
	//nacos server地址
	sc := []constant.ServerConfig{
		{
			ContextPath: "/nacos",
			Port:        opts.Port,
			IpAddr:      opts.Host,
		},
	}

	//nacos client 相关参数配置,具体配置可参考github.com/nacos-group/nacos-sdk-go
	cc := constant.ClientConfig{
		NamespaceId: opts.Namespace,
		TimeoutMs:   5000,
		LogDir:      "./logs",
	}

	client, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": sc,
		"clientConfig":  cc,
	})
	if err != nil {
		return nil, err
	}

	//注册流控规则Handler
	h := datasource.NewFlowRulesHandler(datasource.FlowRuleJsonArrayParser)
	//创建NacosDataSource数据源
	nds, err := nacos.NewNacosDataSource(client, opts.Group, opts.DataId, h)
	if err != nil {
		return nil, err
	}
	return nds, nil
}

func NewUserRPCServer(telemetry *options.TelemetryOptions, serverOpts *options.ServerOptions, userver upb.UserServer, dataNacos *nacos.NacosDataSource) (*rpcserver.Server, error) {
	//初始化open-telemetry的exporter
	trace.InitAgent(trace.Options{
		Name:     telemetry.Name,
		Endpoint: telemetry.Endpoint,
		Sampler:  telemetry.Sampler,
		Batcher:  telemetry.Batcher,
	})

	rpcAddr := fmt.Sprintf("%s:%d", serverOpts.Host, serverOpts.Port)

	var opts []rpcserver.ServerOption
	opts = append(opts, rpcserver.WithAddress(rpcAddr))

	if serverOpts.EnableLimit {
		// 使用新的统一Sentinel拦截器
		fallbackHandler := sentinel.NewBusinessFallbackHandler("emshop-user-srv", false)
		interceptorConfig := sentinel.DefaultServerInterceptorConfig("user-srv")
		interceptorConfig.FallbackFunc = fallbackHandler.Handle

		sentinelInterceptor := sentinel.NewUnaryServerInterceptor(interceptorConfig)
		opts = append(opts, rpcserver.WithUnaryInterceptor(sentinelInterceptor))

		//初始化nacos
		err := dataNacos.Initialize()
		if err != nil {
			return nil, err
		}
	}
	urpcServer := rpcserver.NewServer(opts...)

	upb.RegisterUserServer(urpcServer.Server, userver)

	//r := gin.Default()
	//upb.RegisterUserServerHTTPServer(userver, r)
	//r.Run(":8075")
	return urpcServer, nil
}
