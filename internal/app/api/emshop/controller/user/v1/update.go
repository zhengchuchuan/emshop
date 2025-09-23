package user

import (
    "net/http"
    "time"

    "emshop/internal/app/api/emshop/domain/dto/request"
    "emshop/internal/app/pkg/middleware"
    gin2 "emshop/internal/app/pkg/translator/gin"
    "emshop/pkg/common/core"
    jtime "emshop/pkg/common/time"

    "github.com/gin-gonic/gin"
)

func (us *userServer) UpdateUser(ctx *gin.Context) {
	var req request.UpdateUserRequest

	// 表单验证
	if err := ctx.ShouldBind(&req); err != nil {
		gin2.HandleValidatorError(ctx, err, us.trans)
		return
	}

    // 获取当前用户ID（统一使用中间件助手）
    uid, ok := middleware.GetUserIDFromContext(ctx)
    if !ok || uid <= 0 {
        ctx.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "message": "用户未登录"})
        return
    }
    userIDInt := uint64(uid)

	// 将请求数据转换为proto结构
	updateReq, err := req.ToProto(userIDInt)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	// 获取现有用户信息
	userDTO, err := us.sf.Users().Get(ctx, userIDInt)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	// 更新用户信息
	if updateReq.NickName != nil {
		userDTO.NickName = *updateReq.NickName
	}
	if updateReq.Gender != nil {
		userDTO.Gender = *updateReq.Gender
	}
	if updateReq.BirthDay != nil {
		userDTO.Birthday = jtime.Time{Time: time.Unix(int64(*updateReq.BirthDay), 0)}
	}

	err = us.sf.Users().Update(ctx, userDTO)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	core.WriteResponse(ctx, nil, gin.H{"Message": "用户信息更新成功"})
}
