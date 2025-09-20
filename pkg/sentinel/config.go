package sentinel

import (
	"encoding/json"
	"fmt"

	"emshop/pkg/log"

	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/hotspot"
	"github.com/alibaba/sentinel-golang/core/system"
	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/pkg/datasource/nacos"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
)

// Config Sentinel配置结构
type Config struct {
	// 是否启用Sentinel
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Nacos配置
	Nacos NacosConfig `yaml:"nacos" json:"nacos"`

	// 应用配置
	App AppConfig `yaml:"app" json:"app"`

	// 规则配置
	Rules RulesConfig `yaml:"rules" json:"rules"`
}

// NacosConfig Nacos配置
type NacosConfig struct {
	Host      string `yaml:"host" json:"host"`
	Port      uint64 `yaml:"port" json:"port"`
	Namespace string `yaml:"namespace" json:"namespace"`
	Group     string `yaml:"group" json:"group"`
	DataId    string `yaml:"dataId" json:"dataId"`
	Username  string `yaml:"username" json:"username"`
	Password  string `yaml:"password" json:"password"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name     string `yaml:"name" json:"name"`
	LogLevel string `yaml:"logLevel" json:"logLevel"`
	LogDir   string `yaml:"logDir" json:"logDir"`
}

// RulesConfig 规则配置
type RulesConfig struct {
	// 流控规则配置文件
	FlowRulesDataId string `yaml:"flowRulesDataId" json:"flowRulesDataId"`
	// 熔断规则配置文件
	CircuitBreakerRulesDataId string `yaml:"circuitBreakerRulesDataId" json:"circuitBreakerRulesDataId"`
	// 热点参数规则配置文件
	HotspotRulesDataId string `yaml:"hotspotRulesDataId" json:"hotspotRulesDataId"`
	// 系统规则配置文件
	SystemRulesDataId string `yaml:"systemRulesDataId" json:"systemRulesDataId"`
}

// Manager Sentinel管理器
type Manager struct {
	config      *Config
	dataSources []*nacos.NacosDataSource
}

// NewManager 创建Sentinel管理器
func NewManager(config *Config) *Manager {
	return &Manager{
		config:      config,
		dataSources: make([]*nacos.NacosDataSource, 0),
	}
}

// Initialize 初始化Sentinel
func (m *Manager) Initialize() error {
	if !m.config.Enabled {
		log.Info("Sentinel is disabled, skipping initialization")
		return nil
	}

	// 初始化Sentinel
	err := api.InitDefault()
	if err != nil {
		return fmt.Errorf("初始化Sentinel失败: %w", err)
	}

	// 创建Nacos客户端
	nacosClient, err := m.createNacosClient()
	if err != nil {
		return fmt.Errorf("创建Nacos客户端失败: %w", err)
	}

	// 初始化数据源
	if err := m.initDataSources(nacosClient); err != nil {
		return fmt.Errorf("初始化数据源失败: %w", err)
	}

	log.Info("Sentinel初始化成功")
	return nil
}

// createNacosClient 创建Nacos客户端
func (m *Manager) createNacosClient() (config_client.IConfigClient, error) {
	serverConfig := []constant.ServerConfig{
		{
			IpAddr:      m.config.Nacos.Host,
			Port:        m.config.Nacos.Port,
			ContextPath: "/nacos",
		},
	}

	clientConfig := constant.ClientConfig{
		NamespaceId:         m.config.Nacos.Namespace,
		Username:            m.config.Nacos.Username,
		Password:            m.config.Nacos.Password,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              m.config.App.LogDir,
		CacheDir:            "./cache",
		LogLevel:            m.config.App.LogLevel,
	}

	return clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfig,
		"clientConfig":  clientConfig,
	})
}

// initDataSources 初始化所有数据源
func (m *Manager) initDataSources(client config_client.IConfigClient) error {
	// 初始化流控规则数据源
	if m.config.Rules.FlowRulesDataId != "" {
		if err := m.initFlowRulesDataSource(client); err != nil {
			return err
		}
	}

	// 初始化熔断规则数据源
	if m.config.Rules.CircuitBreakerRulesDataId != "" {
		if err := m.initCircuitBreakerRulesDataSource(client); err != nil {
			return err
		}
	}

	// 初始化热点参数规则数据源
	if m.config.Rules.HotspotRulesDataId != "" {
		if err := m.initHotspotRulesDataSource(client); err != nil {
			return err
		}
	}

	// 初始化系统规则数据源
	if m.config.Rules.SystemRulesDataId != "" {
		if err := m.initSystemRulesDataSource(client); err != nil {
			return err
		}
	}

	return nil
}

// initFlowRulesDataSource 初始化流控规则数据源
func (m *Manager) initFlowRulesDataSource(client config_client.IConfigClient) error {
	handler := datasource.NewFlowRulesHandler(flowRulesParser)
	dataSource, err := nacos.NewNacosDataSource(
		client,
		m.config.Nacos.Group,
		m.config.Rules.FlowRulesDataId,
		handler,
	)
	if err != nil {
		return err
	}

	if err := dataSource.Initialize(); err != nil {
		return err
	}

	m.dataSources = append(m.dataSources, dataSource)
	log.Infof("流控规则数据源初始化成功: dataId=%s", m.config.Rules.FlowRulesDataId)
	return nil
}

// initCircuitBreakerRulesDataSource 初始化熔断规则数据源
func (m *Manager) initCircuitBreakerRulesDataSource(client config_client.IConfigClient) error {
	handler := datasource.NewCircuitBreakerRulesHandler(circuitBreakerRulesParser)
	dataSource, err := nacos.NewNacosDataSource(
		client,
		m.config.Nacos.Group,
		m.config.Rules.CircuitBreakerRulesDataId,
		handler,
	)
	if err != nil {
		return err
	}

	if err := dataSource.Initialize(); err != nil {
		return err
	}

	m.dataSources = append(m.dataSources, dataSource)
	log.Infof("熔断规则数据源初始化成功: dataId=%s", m.config.Rules.CircuitBreakerRulesDataId)
	return nil
}

// initHotspotRulesDataSource 初始化热点参数规则数据源
func (m *Manager) initHotspotRulesDataSource(client config_client.IConfigClient) error {
	handler := datasource.NewHotSpotParamRulesHandler(hotspotRulesParser)
	dataSource, err := nacos.NewNacosDataSource(
		client,
		m.config.Nacos.Group,
		m.config.Rules.HotspotRulesDataId,
		handler,
	)
	if err != nil {
		return err
	}

	if err := dataSource.Initialize(); err != nil {
		return err
	}

	m.dataSources = append(m.dataSources, dataSource)
	log.Infof("热点参数规则数据源初始化成功: dataId=%s", m.config.Rules.HotspotRulesDataId)
	return nil
}

// initSystemRulesDataSource 初始化系统规则数据源
func (m *Manager) initSystemRulesDataSource(client config_client.IConfigClient) error {
	handler := datasource.NewSystemRulesHandler(systemRulesParser)
	dataSource, err := nacos.NewNacosDataSource(
		client,
		m.config.Nacos.Group,
		m.config.Rules.SystemRulesDataId,
		handler,
	)
	if err != nil {
		return err
	}

	if err := dataSource.Initialize(); err != nil {
		return err
	}

	m.dataSources = append(m.dataSources, dataSource)
	log.Infof("系统规则数据源初始化成功: dataId=%s", m.config.Rules.SystemRulesDataId)
	return nil
}

// Shutdown 关闭Sentinel管理器
func (m *Manager) Shutdown() error {
	for _, ds := range m.dataSources {
		if err := ds.Close(); err != nil {
			log.Errorf("关闭数据源失败: %v", err)
		}
	}
	log.Info("Sentinel管理器已关闭")
	return nil
}

// 规则解析器
func flowRulesParser(src []byte) (interface{}, error) {
	var rules []*flow.Rule
	if err := json.Unmarshal(src, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}

func circuitBreakerRulesParser(src []byte) (interface{}, error) {
	var rules []*circuitbreaker.Rule
	if err := json.Unmarshal(src, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}

func hotspotRulesParser(src []byte) (interface{}, error) {
	var rules []*hotspot.Rule
	if err := json.Unmarshal(src, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}

func systemRulesParser(src []byte) (interface{}, error) {
	var rules []*system.Rule
	if err := json.Unmarshal(src, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}

// DefaultConfig 返回默认配置
func DefaultConfig(serviceName string) *Config {
	return &Config{
		Enabled: true,
		Nacos: NacosConfig{
			Host:      "127.0.0.1",
			Port:      8848,
			Namespace: "",
			Group:     "sentinel-go",
			DataId:    fmt.Sprintf("%s-sentinel", serviceName),
			Username:  "",
			Password:  "",
		},
		App: AppConfig{
			Name:     serviceName,
			LogLevel: "info",
			LogDir:   "./logs",
		},
		Rules: RulesConfig{
			FlowRulesDataId:           fmt.Sprintf("%s-flow-rules", serviceName),
			CircuitBreakerRulesDataId: fmt.Sprintf("%s-circuit-breaker-rules", serviceName),
			HotspotRulesDataId:        fmt.Sprintf("%s-hotspot-rules", serviceName),
			SystemRulesDataId:         fmt.Sprintf("%s-system-rules", serviceName),
		},
	}
}
