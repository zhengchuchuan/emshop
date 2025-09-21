package user

import (
    "emshop/pkg/common/core"
    "github.com/gin-gonic/gin"
)

func (us *userServer) GetUserDetail(ctx *gin.Context) {
    userID := us.getUserIDFromContext(ctx)
    userDTO, err := us.sf.Users().Get(ctx, uint64(userID))
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
