package controller

import (
	"github.com/gin-gonic/gin"
	"emshop/pkg/log"
)

func (us *userServer) List(ctx *gin.Context) {
	log.Info("GetUserList is called")
}
