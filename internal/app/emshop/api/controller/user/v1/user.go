package user

import (
	"emshop/internal/app/emshop/api/service"
	"emshop/gin-micro/server/rest-server"
)

type userServer struct {
	trans restserver.I18nTranslator

	sf service.ServiceFactory
}

func NewUserController(trans restserver.I18nTranslator, sf service.ServiceFactory) *userServer {
	return &userServer{trans, sf}
}
