package register

type ServiceInstance struct {
	// 注册到服务中心的服务id
	ID string `json:"id"`
	// 服务名称
	Name string `json:"name"`
	// 服务版本
	Version string `json:"version"`

	// 服务源数据
	Metadata map[string]string `json:"metadata"`

	// 服务地址 http://127.0.1:8080 grpc://127.0.1:8080
	Endpoints []string `json:"endpoints"`

}