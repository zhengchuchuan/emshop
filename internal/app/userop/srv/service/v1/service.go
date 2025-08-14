package v1

import (
	datav1 "emshop/internal/app/userop/srv/data/v1"
)

// Service 服务接口
type Service interface {
	UserFavService() UserFavService
	AddressService() AddressService
	MessageService() MessageService
}

type service struct {
	dataFactory datav1.DataFactory
}

// NewService 创建服务实例
func NewService(dataFactory datav1.DataFactory) Service {
	return &service{
		dataFactory: dataFactory,
	}
}

func (s *service) UserFavService() UserFavService {
	return NewUserFavService(s.dataFactory)
}

func (s *service) AddressService() AddressService {
	return NewAddressService(s.dataFactory)
}

func (s *service) MessageService() MessageService {
	return NewMessageService(s.dataFactory)
}