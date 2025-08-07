package user

import (
	gin2 "emshop/internal/app/pkg/translator/gin"
	"emshop/pkg/log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PassWordLoginForm struct {
	Mobile    string `form:"mobile" json:"mobile" binding:"required,mobile"` //手机号码格式有规范可寻， 自定义validator
	PassWord  string `form:"password" json:"password" binding:"required,min=3,max=20"`
	Captcha   string `form:"captcha" json:"captcha" binding:"required,min=5,max=5"`
	CaptchaId string `form:"captcha_id" json:"captcha_id" binding:"required"`
}

func (us *userServer) Login(ctx *gin.Context) {
	log.Info("login is called")

	
	passwordLoginForm := PassWordLoginForm{}
	//表单验证
	if err := ctx.ShouldBind(&passwordLoginForm); err != nil {
		gin2.HandleValidatorError(ctx, err, us.trans)
		return
	}

	//验证码验证
	if !store.Verify(passwordLoginForm.CaptchaId, passwordLoginForm.Captcha, true) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"captcha": us.trans.T("business.captcha_error"),
		})
		return
	}

	userDTO, err := us.sf.Users().MobileLogin(ctx, passwordLoginForm.Mobile, passwordLoginForm.PassWord)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": us.trans.T("business.login_failed"),
		})
		return
	}
	// 返回
	ctx.JSON(http.StatusOK, gin.H{
		"id":         userDTO.ID,
		"nick_name":  userDTO.NickName,
		"token":      userDTO.Token,
		"expired_at": userDTO.ExpiresAt,
	})
}
