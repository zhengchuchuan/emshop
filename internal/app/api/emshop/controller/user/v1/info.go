package user

import (
    "net/http"
    "emshop/internal/app/pkg/middleware"
    "emshop/pkg/common/core"
    "github.com/gin-gonic/gin"
)

func (us *userServer) GetUserDetail(ctx *gin.Context) {
    uid, ok := middleware.GetUserIDFromContext(ctx)
    if !ok || uid <= 0 {
        ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未登录"})
        return
    }
    userDTO, err := us.sf.Users().Get(ctx, uint64(uid))
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
