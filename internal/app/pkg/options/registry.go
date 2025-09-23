package options

import (
	"emshop/pkg/errors"

	"github.com/spf13/pflag"
)

type RegistryOptions struct {
    Address string `mapstructure:"address" json:"address,omitempty"`
    Scheme  string `mapstructure:"scheme" json:"scheme,omitempty"`
    // gRPC/TCP health check interval in seconds
    HealthCheckInterval int `mapstructure:"health-check-interval" json:"health-check-interval,omitempty"`
    // Consul deregister critical service after N seconds
    DeregisterCriticalAfter int `mapstructure:"deregister-critical-after" json:"deregister-critical-after,omitempty"`
    // Health check timeout in seconds (applies to both gRPC and TCP checks)
    CheckTimeout int `mapstructure:"check-timeout" json:"check-timeout,omitempty"`
}

func NewRegistryOptions() *RegistryOptions {
    return &RegistryOptions{
        Address: "127.0.0.1:8500", // 默认consul的地址
        Scheme:  "http",
        HealthCheckInterval:       10,
        DeregisterCriticalAfter:   600,
        CheckTimeout:              5,
    }
}

func (o *RegistryOptions) Validate() []error {
	errs := []error{}
	if o.Address == "" || o.Scheme == "" {
		errs = append(errs, errors.New("address an scheme is empty"))
	}
	return errs
}

func (o *RegistryOptions) AddFlags(fs *pflag.FlagSet) {
    fs.StringVar(&o.Address, "consul.address", o.Address, ""+
        "consul address, if left , default is 127.0.0.1:8500")

    fs.StringVar(&o.Scheme, "consul.scheme", o.Scheme, ""+
        "registry schema, if left , default is http")

    fs.IntVar(&o.HealthCheckInterval, "consul.health-check-interval", o.HealthCheckInterval, "health check interval seconds for consul registered services")
    fs.IntVar(&o.DeregisterCriticalAfter, "consul.deregister-critical-after", o.DeregisterCriticalAfter, "seconds after which consul deregisters critical services")
    fs.IntVar(&o.CheckTimeout, "consul.check-timeout", o.CheckTimeout, "health check timeout seconds (applies to gRPC/TCP checks)")
}
