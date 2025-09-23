package user

import (
    "net/http"
    "strconv"

    restserver "emshop/gin-micro/server/rest-server"
    "emshop/internal/app/api/emshop/service"
    "emshop/pkg/common/core"

    "github.com/gin-gonic/gin"
)

type userServer struct {
	trans restserver.I18nTranslator

	sf service.ServiceFactory
}

func NewUserController(trans restserver.I18nTranslator, sf service.ServiceFactory) *userServer {
	return &userServer{trans, sf}
}

func (us *userServer) GetByMobile(ctx *gin.Context) {
	mobile := ctx.Query("mobile")
	if mobile == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": "mobile parameter is required",
		})
		return
	}

	userDTO, err := us.sf.Users().GetByMobile(ctx, mobile)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, gin.H{
		"id":       userDTO.ID,
		"mobile":   userDTO.Mobile,
		"name":     userDTO.NickName,
		"birthday": userDTO.Birthday.Format("2006-01-02"),
		"gender":   userDTO.Gender,
		"role":     userDTO.Role,
	})
}

func (us *userServer) GetById(ctx *gin.Context) {
	idStr := ctx.Query("id")
	if idStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": "id parameter is required",
		})
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": "invalid id parameter",
		})
		return
	}

	userDTO, err := us.sf.Users().Get(ctx, id)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, gin.H{
		"id":       userDTO.ID,
		"mobile":   userDTO.Mobile,
		"name":     userDTO.NickName,
		"birthday": userDTO.Birthday.Format("2006-01-02"),
		"gender":   userDTO.Gender,
		"role":     userDTO.Role,
	})
}

func (us *userServer) GetUserList(ctx *gin.Context) {
	pnStr := ctx.DefaultQuery("pn", "1")
	pSizeStr := ctx.DefaultQuery("pSize", "10")

	pn, err := strconv.ParseUint(pnStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": "invalid pn parameter",
		})
		return
	}

	pSize, err := strconv.ParseUint(pSizeStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": "invalid pSize parameter",
		})
		return
	}

	// fmt.Println("GetUserList called with pn:", pn, "pSize:", pSize)

	userListDTO, err := us.sf.Users().GetUserList(ctx, uint32(pn), uint32(pSize))
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	var users []gin.H
	for _, userDTO := range userListDTO.Items {
		users = append(users, gin.H{
			"id":       userDTO.ID,
			"mobile":   userDTO.Mobile,
			"name":     userDTO.NickName,
			"birthday": userDTO.Birthday.Format("2006-01-02"),
			"gender":   userDTO.Gender,
			"role":     userDTO.Role,
		})
	}

	core.WriteResponse(ctx, nil, gin.H{
		"total": userListDTO.TotalCount,
		"users": users,
	})
}
// 已统一通过 middleware.GetUserIDFromContext 获取用户ID
