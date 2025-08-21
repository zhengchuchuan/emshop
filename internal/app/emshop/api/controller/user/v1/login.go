package user

import (
	upbv1 "emshop/api/user/v1"
	"emshop/internal/app/pkg/code"
	gin2 "emshop/internal/app/pkg/translator/gin"
	"emshop/pkg/errors"
	"emshop/pkg/log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// var store = base64Captcha.DefaultMemStore

func (us *userServer) Login(ctx *gin.Context) {
	log.Info("login is called")

	var loginReq upbv1.UserLoginRequest

	//表单验证
	if err := ctx.ShouldBind(&loginReq); err != nil {
		gin2.HandleValidatorError(ctx, err, us.trans)
		return
	}

	// 手动验证必要字段（由于proto结构体没有binding标签）
	if loginReq.Mobile == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": us.trans.T("business.mobile_required"),
		})
		return
	}
	if loginReq.Password == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": us.trans.T("business.password_required"),
		})
		return
	}
	if loginReq.Captcha == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": us.trans.T("business.captcha_required"),
		})
		return
	}
	if loginReq.CaptchaId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": us.trans.T("business.captcha_id_required"),
		})
		return
	}

	//验证码验证
	if !store.Verify(loginReq.CaptchaId, loginReq.Captcha, true) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"captcha": us.trans.T("business.captcha_error"),
		})
		return
	}

	userDTO, err := us.sf.Users().MobileLogin(ctx, loginReq.Mobile, loginReq.Password)
	if err != nil {
		// 根据错误类型返回不同的状态码和错误信息
		if errors.IsCode(err, code.ErrUserNotFound) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"msg": us.trans.T("business.user_not_found"),
			})
			return
		}

		if errors.IsCode(err, code.ErrUserPasswordIncorrect) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"msg": us.trans.T("business.password_incorrect"),
			})
			return
		}

		// 其他未知错误返回内部服务器错误
		log.Errorf("login failed with unknown error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": us.trans.T("business.login_failed"),
		})
		return
	}
	// 返回
	ctx.JSON(http.StatusOK, gin.H{
		"id":        userDTO.ID,
		"nickName":  userDTO.NickName,
		"token":     userDTO.Token,
		"expiredAt": userDTO.ExpiresAt,
	})
}
