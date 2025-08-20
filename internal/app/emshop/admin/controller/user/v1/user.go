package user

import (
	"net/http"
	"strconv"
	"time"

	restserver "emshop/gin-micro/server/rest-server"
	"emshop/internal/app/emshop/admin/service"
	"emshop/pkg/common/core"
	gin2 "emshop/internal/app/pkg/translator/gin"
	"emshop/pkg/errors"
	"emshop/gin-micro/code"
	appcode "emshop/internal/app/pkg/code"
	"emshop/pkg/log"
	upbv1 "emshop/api/user/v1"

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
	var pageInfo upbv1.PageInfo
	if err := ctx.ShouldBindQuery(&pageInfo); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid query parameters"})
		return
	}

	userListDTO, err := uc.sf.Users().GetUserList(ctx, &pageInfo)
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

type UpdateUserForm struct {
	Name     string `form:"name" json:"name" binding:"required,min=3,max=10"`
	Gender   string `form:"gender" json:"gender" binding:"required,oneof=female male"`
	Birthday string `form:"birthday" json:"birthday" binding:"required,datetime=2006-01-02"`
}

// UpdateUser 更新用户信息（管理员专用）
func (uc *userController) UpdateUser(ctx *gin.Context) {
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

	updateForm := UpdateUserForm{}
	if err := ctx.ShouldBind(&updateForm); err != nil {
		gin2.HandleValidatorError(ctx, err, uc.trans)
		return
	}

	// 将前端传递过来的日期格式转换成时间戳
	loc, _ := time.LoadLocation("Local")
	birthDay, err := time.ParseInLocation("2006-01-02", updateForm.Birthday, loc)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": "invalid birthday format",
		})
		return
	}

	// 只传递需要更新的字段
	err = uc.sf.Users().UpdateUserInfo(ctx, id, updateForm.Name, updateForm.Gender, uint64(birthDay.Unix()))
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, gin.H{
		"msg": "User updated successfully",
	})
}

// AdminLogin 管理员登录（管理员专用）
func (uc *userController) AdminLogin(ctx *gin.Context) {
	log.Info("admin login function called...")

	type AdminLoginForm struct {
		Mobile    string `json:"mobile" binding:"required,mobile"`
		Password  string `json:"password" binding:"required,min=3,max=20"`
		Captcha   string `json:"captcha" binding:"required,min=5,max=5"`
		CaptchaId string `json:"captcha_id" binding:"required"`
	}

	var loginForm AdminLoginForm
	if err := ctx.ShouldBindJSON(&loginForm); err != nil {
		gin2.HandleValidatorError(ctx, err, uc.trans)
		return
	}

	// 验证码验证
	if !store.Verify(loginForm.CaptchaId, loginForm.Captcha, true) {
		core.WriteResponse(ctx, errors.WithCode(code.ErrValidation, "验证码错误"), nil)
		return
	}

	// 直接通过登录服务验证用户和生成token
	loginResult, err := uc.sf.Users().MobileLogin(ctx, loginForm.Mobile, loginForm.Password)
	if err != nil {
		log.Errorf("Admin login failed: %v", err)
		core.WriteResponse(ctx, errors.WithCode(appcode.ErrUserNotFound, "管理员登录失败"), nil)
		return
	}

	// 验证用户角色：必须是管理员
	if loginResult.Role < 2 { // RoleAdmin = 2
		log.Warnf("Admin login denied: insufficient privileges - userID: %d, role: %d", loginResult.ID, loginResult.Role)
		core.WriteResponse(ctx, errors.WithCode(code.ErrPermissionDenied, "权限不足：仅限管理员登录"), nil)
		return
	}

	log.Infof("Admin login successful - userID: %d, role: %d", loginResult.ID, loginResult.Role)

	// 返回管理员登录结果
	core.WriteResponse(ctx, nil, gin.H{
		"id":         loginResult.ID,
		"nickName":  loginResult.NickName,
		"mobile":     loginResult.Mobile,
		"role":       loginResult.Role,
		"token":      loginResult.Token,
		"expiresAt": loginResult.ExpiresAt,
		"message":    "管理员登录成功",
	})
}