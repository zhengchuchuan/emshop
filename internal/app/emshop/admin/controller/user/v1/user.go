package user

import (
	"net/http"
	"strconv"
	"time"

	restserver "emshop/gin-micro/server/rest-server"
	"emshop/internal/app/emshop/admin/service"
	"emshop/pkg/common/core"

	"github.com/gin-gonic/gin"
)

type userController struct {
	trans restserver.I18nTranslator
	sf    service.ServiceFactory
}

func NewUserController(trans restserver.I18nTranslator, sf service.ServiceFactory) *userController {
	return &userController{trans, sf}
}

// GetUserList 获取用户列表（管理员专用）
func (uc *userController) GetUserList(ctx *gin.Context) {
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

	userListDTO, err := uc.sf.Users().GetUserList(ctx, uint32(pn), uint32(pSize))
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	var users []gin.H
	for _, userDTO := range userListDTO.Data {
		users = append(users, gin.H{
			"id":       userDTO.Id,
			"mobile":   userDTO.Mobile,
			"name":     userDTO.NickName,
			"birthday": time.Unix(int64(userDTO.BirthDay), 0).Format("2006-01-02"),
			"gender":   userDTO.Gender,
			"role":     userDTO.Role,
		})
	}

	core.WriteResponse(ctx, nil, gin.H{
		"total": userListDTO.Total,
		"users": users,
	})
}

// GetUserByMobile 通过手机号查询用户（管理员专用）
func (uc *userController) GetUserByMobile(ctx *gin.Context) {
	mobile := ctx.Query("mobile")
	if mobile == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": "mobile parameter is required",
		})
		return
	}

	userDTO, err := uc.sf.Users().GetUserByMobile(ctx, mobile)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, gin.H{
		"id":       userDTO.Id,
		"mobile":   userDTO.Mobile,
		"name":     userDTO.NickName,
		"birthday": time.Unix(int64(userDTO.BirthDay), 0).Format("2006-01-02"),
		"gender":   userDTO.Gender,
		"role":     userDTO.Role,
	})
}

// GetUserById 通过ID查询用户（管理员专用）
func (uc *userController) GetUserById(ctx *gin.Context) {
	idStr := ctx.Param("id")
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

	userDTO, err := uc.sf.Users().GetUserById(ctx, id)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, gin.H{
		"id":       userDTO.Id,
		"mobile":   userDTO.Mobile,
		"name":     userDTO.NickName,
		"birthday": time.Unix(int64(userDTO.BirthDay), 0).Format("2006-01-02"),
		"gender":   userDTO.Gender,
		"role":     userDTO.Role,
	})
}

// UpdateUserStatus 更新用户状态（管理员专用）
func (uc *userController) UpdateUserStatus(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": "invalid id parameter",
		})
		return
	}

	var req struct {
		Status int32 `json:"status" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": "invalid request body",
		})
		return
	}

	err = uc.sf.Users().UpdateUserStatus(ctx, id, req.Status)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, gin.H{
		"msg": "User status updated successfully",
	})
}