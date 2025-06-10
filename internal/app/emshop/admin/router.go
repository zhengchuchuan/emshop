package admin

import (
	"emshop/internal/app/emshop/admin/controller"
	"emshop/gin-micro/server/rest-server"
)

func initRouter(g *restserver.Server) {
	v1 := g.Group("/v1")
	ugroup := v1.Group("/user")
	ucontroller := controller.NewUserController()
	ugroup.GET("list", ucontroller.List)
}
