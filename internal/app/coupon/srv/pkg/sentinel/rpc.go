package sentinel

import (
	"fmt"

	cpb "emshop/api/coupon/v1"
	"emshop/gin-micro/core/trace"
	rpcserver "emshop/gin-micro/server/rpc-server"
	"emshop/internal/app/pkg/options"
	"emshop/pkg/sentinel"

	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/pkg/datasource/nacos"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
)

// NewCouponNacosDataSource 创建优惠券服务的Nacos数据源
func NewCouponNacosDataSource(opts *options.NacosOptions) (*nacos.NacosDataSource, error) {
	// Nacos服务器地址
	sc := []constant.ServerConfig{
		{
			ContextPath: "/nacos",
			Port:        opts.Port,
			IpAddr:      opts.Host,
		},
	}

	// Nacos客户端配置
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

	// 注册流控规则Handler
	h := datasource.NewFlowRulesHandler(datasource.FlowRuleJsonArrayParser)
	
	// 创建NacosDataSource数据源
	nds, err := nacos.NewNacosDataSource(client, opts.Group, opts.DataId, h)
	if err != nil {
		return nil, err
	}
	return nds, nil
}

// NewCouponRPCServer 创建优惠券服务的RPC服务器
func NewCouponRPCServer(
	telemetry *options.TelemetryOptions,
	serverOpts *options.ServerOptions,
	couponServer cpb.CouponServer,
	dataNacos *nacos.NacosDataSource,
) (*rpcserver.Server, error) {
	
	// 初始化open-telemetry的exporter
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
		// 创建优惠券服务专用的降级处理器
		fallbackHandler := sentinel.NewBusinessFallbackHandler("emshop-coupon-srv", true)
		
		// 配置Sentinel拦截器
		interceptorConfig := sentinel.DefaultServerInterceptorConfig("coupon-srv")
		interceptorConfig.FallbackFunc = fallbackHandler.Handle
		
		// 创建Sentinel拦截器
		sentinelInterceptor := sentinel.NewUnaryServerInterceptor(interceptorConfig)
		opts = append(opts, rpcserver.WithUnaryInterceptor(sentinelInterceptor))
		
		// 初始化Nacos数据源
		err := dataNacos.Initialize()
		if err != nil {
			return nil, err
		}
	}
	
	// 创建RPC服务器
	couponRPCServer := rpcserver.NewServer(opts...)

	// 注册优惠券服务
	cpb.RegisterCouponServer(couponRPCServer.Server, couponServer)

	return couponRPCServer, nil
}

// NewCouponSentinelManager 创建优惠券服务的Sentinel管理器
func NewCouponSentinelManager() *sentinel.Manager {
	config := sentinel.DefaultConfig("emshop-coupon-srv")
	
	// 自定义优惠券服务的配置
	config.Rules.FlowRulesDataId = "coupon-flow-rules"
	config.Rules.CircuitBreakerRulesDataId = "coupon-circuit-breaker-rules"
	config.Rules.HotspotRulesDataId = "coupon-hotspot-rules"
	config.Rules.SystemRulesDataId = "coupon-system-rules"
	
	return sentinel.NewManager(config)
}

// GenerateCouponRules 生成优惠券服务的业务规则
func GenerateCouponRules() *sentinel.BusinessRules {
	rules := sentinel.DefaultBusinessRules()
	
	// 针对优惠券服务的特殊配置
	if rules.Coupon != nil {
		// 秒杀场景下的优惠券发放限流更严格
		rules.Coupon.IssueQPS = 200  // 降低发放QPS
		rules.Coupon.UseQPS = 1000   // 提高使用QPS
		
		// 热点参数配置：针对特定用户或优惠券
		rules.Coupon.UserHotspot.Count = 5        // 每个用户每秒最多5次请求
		rules.Coupon.UserHotspot.DurationInSec = 1
		rules.Coupon.UserHotspot.SpecificItems = []sentinel.HotspotItem{
			{Value: "vip_user", Threshold: 20},    // VIP用户更高的阈值
			{Value: "normal_user", Threshold: 3},  // 普通用户更低的阈值
		}
		
		// 熔断器配置：更敏感的错误检测
		rules.Coupon.CircuitBreaker.ErrorRatio = 0.2           // 20%错误率触发熔断
		rules.Coupon.CircuitBreaker.SlowRatio = 0.4            // 40%慢调用触发熔断
		rules.Coupon.CircuitBreaker.SlowTimeMs = 500           // 500ms为慢调用
		rules.Coupon.CircuitBreaker.MinRequestAmount = 5       // 最少5个请求
		rules.Coupon.CircuitBreaker.RecoveryTimeoutSec = 5     // 5秒恢复时间
	}
	
	return rules
}