package v1

import (
	"emshop/internal/app/payment/srv/data/v1/interfaces"
	"emshop/internal/app/pkg/options"
)

// Service 支付服务工厂
type Service struct {
	PaymentSrv PaymentSrv
}

// NewService 创建支付服务工厂
func NewService(data interfaces.DataFactory, dtmOpts *options.DtmOptions, redisOpts *options.RedisOptions) *Service {
	return &Service{
		PaymentSrv: NewPaymentService(data, dtmOpts, redisOpts),
	}
}