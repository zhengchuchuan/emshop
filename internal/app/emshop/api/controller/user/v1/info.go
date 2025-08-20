package user

import (
	"github.com/gin-gonic/gin"
	"emshop/internal/app/pkg/jwt"
	"emshop/pkg/common/core"
)

func (us *userServer) GetUserDetail(ctx *gin.Context) {
	userID, _ := ctx.Get(jwt.KeyUserID)
	userDTO, err := us.sf.Users().Get(ctx, uint64(userID.(int)))
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	core.WriteResponse(ctx, nil, gin.H{
		"name":     userDTO.NickName,
		"birthday": userDTO.Birthday.Format("2006-01-02"),
		"gender":   userDTO.Gender,
		"mobile":   userDTO.Mobile,
	})
}
