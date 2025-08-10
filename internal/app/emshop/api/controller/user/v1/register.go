package user

import (
	"github.com/gin-gonic/gin"
	gin2 "emshop/internal/app/pkg/translator/gin"
	"emshop/pkg/common/core"
)

type RegisterForm struct {
	Mobile   string `form:"mobile" json:"mobile" binding:"required,mobile"` //手机号码格式有规范可寻， 自定义validator
	PassWord string `form:"password" json:"password" binding:"required,min=3,max=20"`
	Code     string `form:"code" json:"code" binding:"required,min=6,max=6"`
}

func (us *userServer) Register(ctx *gin.Context) {
	regForm := RegisterForm{}
	if err := ctx.ShouldBind(&regForm); err != nil {
		if us.trans != nil {
			gin2.HandleValidatorError(ctx, err, us.trans)
		} else {
			core.WriteResponse(ctx, err, nil)
		}
		return
	}

	userDTO, err := us.sf.Users().Register(ctx, regForm.Mobile, regForm.PassWord, regForm.Code)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	core.WriteResponse(ctx, nil, gin.H{
		"id":         userDTO.ID,
		"nick_name":  userDTO.NickName,
		"token":      userDTO.Token,
		"expired_at": userDTO.ExpiresAt,
	})

}
