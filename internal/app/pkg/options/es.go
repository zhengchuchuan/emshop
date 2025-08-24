package options

import "github.com/spf13/pflag"

type EsOptions struct {
	Host                string `json:"host" mapstructure:"host"`
	Port                string `json:"port" mapstructure:"port"`
	EnableServiceSync   bool   `json:"enable_service_sync" mapstructure:"enable_service_sync"`
}

func NewEsOptions() *EsOptions {
	return &EsOptions{
		Host:              "127.0.0.1",
		Port:              "9200",
		EnableServiceSync: false, // Default to false, rely on Canal for sync
	}
}

func (e *EsOptions) Validate() []error {
	errs := []error{}
	return errs
}

func (e *EsOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&e.Host, "es.host", e.Host, ""+
		"es service host address. If left blank, the following related es options will be ignored.")

	fs.StringVar(&e.Port, "es.port", e.Port, ""+
		"es service port If left blank, the following related es options will be ignored..")

	fs.BoolVar(&e.EnableServiceSync, "es.enable-service-sync", e.EnableServiceSync, ""+
		"enable service-level Elasticsearch synchronization. When false, relies on Canal for sync.")
}
