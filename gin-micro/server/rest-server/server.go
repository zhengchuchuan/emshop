package restserver

import (
	"context"
	"fmt"
	"strings"
	"github.com/penglongli/gin-metrics/ginmetrics"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	mws "emshop/gin-micro/server/rest-server/middlewares"
	"emshop/gin-micro/server/rest-server/pprof"
	"emshop/gin-micro/server/rest-server/validation"
	"emshop/pkg/errors"
	"emshop/pkg/log"
)

// I18nTranslator 定义翻译器接口，与universal-translator兼容
type I18nTranslator interface {
	T(key string, params ...interface{}) string
}

// i18nTranslatorWrapper 实现了I18nTranslator接口，包装go-i18n/v2的Localizer
type i18nTranslatorWrapper struct {
	localizer *i18n.Localizer
}

// T 翻译方法，与universal-translator的T方法兼容
func (wrapper *i18nTranslatorWrapper) T(key string, params ...interface{}) string {
	// 添加空指针保护
	if wrapper == nil || wrapper.localizer == nil {
		log.Warnf("translator not initialized, returning original key: %s", key)
		return key
	}
	
	// 构造完整的消息ID，支持嵌套路径
	var messageID string
	if strings.Contains(key, ".") {
		messageID = key
	} else {
		messageID = fmt.Sprintf("validation.%s", key)
	}
	
	config := &i18n.LocalizeConfig{
		MessageID: messageID,
		DefaultMessage: &i18n.Message{
			ID:    messageID,
			Other: key, // 没有翻译时直接返回原key
		},
	}
	
	// 如果有参数，将第一个参数作为TemplateData
	if len(params) > 0 {
		if templateData, ok := params[0].(map[string]interface{}); ok {
			config.TemplateData = templateData
		}
	}
	
	result, err := wrapper.localizer.Localize(config)
	if err != nil {
		log.Debugf("translation not found for key: %s, returning original key", key)
		return key
	}
	
	return result
}


type JwtInfo struct {
	// defaults to "JWT"
	Realm string
	// defaults to empty
	Key string
	// defaults to 7 days
	Timeout time.Duration
	// defaults to 7 days
	MaxRefresh time.Duration
}

// wrapper for gin.Engine
type Server struct {
	*gin.Engine

	//端口号， 默认值 8080
	port int

	//开发日志模式， 默认值 debug
	mode string

	//是否开启健康检查接口， 默认开启， 如果开启会自动添加 /health 接口
	healthz bool

	//是否开启pprof接口， 默认开启， 如果开启会自动添加 /debug/pprof 接口
	enableProfiling bool

	//是否开启metrics接口， 默认开启， 如果开启会自动添加 /metrics 接口
	enableMetrics bool

	//中间件
	middlewares []string

	//jwt配置信息
	jwt *JwtInfo

	//翻译器, 默认值 zh
	transName string
	// 翻译文件目录路径，默认为空（使用内置翻译）
	localesDir string
	// go-i18n/v2 localizer用于翻译消息
	localizer *i18n.Localizer
	// 当前语言环境
	locale    string

	server *http.Server

	serviceName string

	//路由初始化回调函数和配置
	routerInitFunc func(*Server, interface{})
	routerInitConfig interface{}
}

func NewServer(opts ...ServerOption) *Server {
	// 默认的配置
	srv := &Server{
		port:            8080,
		mode:            "debug",
		healthz:         true,
		enableProfiling: true,
		jwt: &JwtInfo{
			"JWT",
			"mwGDMGtSpdwXaiihF5WnEgRajSFpdZj8",
			7 * 24 * time.Hour,
			7 * 24 * time.Hour,
		},
		Engine:      gin.Default(),
		transName:   "zh",
		serviceName: "gin-micro",
	}

	for _, o := range opts {
		o(srv)
	}

	// gin集成链路追踪
	srv.Use(mws.TracingHandler(srv.serviceName))


	for _, m := range srv.middlewares {
		mw, ok := mws.Middlewares[m]
		if !ok {
			log.Warnf("can not find middleware: %s", m)
			// 没有找到中间件，跳过
			continue
			// 如果需要严格检查，可以取消注释下面这行代码
			//panic(errors.Errorf("can not find middleware: %s", m))
		}

		log.Infof("intall middleware: %s", m)
		srv.Use(mw)
	}

	return srv
}

// Localizer 获取当前的i18n localizer
func (s *Server) Localizer() *i18n.Localizer {
	return s.localizer
}

// GetLocale 获取当前语言环境
func (s *Server) GetLocale() string {
	return s.locale
}

// Translator 返回一个兼容universal-translator接口的翻译器
// 这个方法用于保持与现有控制器的兼容性
func (s *Server) Translator() I18nTranslator {
	return &i18nTranslatorWrapper{localizer: s.localizer}
}

// start rest server
func (s *Server) Start(ctx context.Context) error {
	log.Infof("Starting REST server with transName: %s, localesDir: %s", s.transName, s.localesDir)
	
	//设置开发模式，打印路由信息
	if s.mode != gin.DebugMode && s.mode != gin.ReleaseMode && s.mode != gin.TestMode {
		return errors.New("mode must be one of debug/release/test")
	}

	//设置开发模式，打印路由信息
	gin.SetMode(s.mode)
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		log.Infof("%-6s %-s --> %s(%d handlers)", httpMethod, absolutePath, handlerName, nuHandlers)
	}

	//TODO 初始化翻译器
	err := s.initTrans(s.transName)
	if err != nil {
		log.Errorf("initTrans error %s", err.Error())
		return err
	}
	
	// 初始化路由（在翻译器初始化之后）
	if s.routerInitFunc != nil {
		log.Info("initializing routes after translator setup")
		s.routerInitFunc(s, s.routerInitConfig)
	}

	//注册mobile验证码
	validation.RegisterMobile(s.localizer)

	//根据配置初始化pprof路由
	if s.enableProfiling {
		pprof.Register(s.Engine)
	}

	// 注册prometheus监控
	if s.enableMetrics {
		// get global Monitor object
		m := ginmetrics.GetMonitor()
		// +optional set metric path, default /debug/metrics
		m.SetMetricPath("/metrics")
		// +optional set slow time, default 5s
		// +optional set request duration, default {0.1, 0.3, 1.2, 5, 10}
		// used to p95, p99
		m.SetDuration([]float64{0.1, 0.3, 1.2, 5, 10})
		m.Use(s)
	}

	log.Infof("rest server is running on port: %d", s.port)
	address := fmt.Sprintf(":%d", s.port)
	// 使用http的server 优雅退出, 自己维护
	s.server = &http.Server{
		Addr:    address,
		Handler: s.Engine,
	}
	_ = s.SetTrustedProxies(nil)
	if err = s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	log.Infof("rest server is stopping")
	if err := s.server.Shutdown(ctx); err != nil {
		log.Errorf("rest server shutdown error: %s", err.Error())
		return err
	}
	log.Info("rest server stopped")
	return nil
}
