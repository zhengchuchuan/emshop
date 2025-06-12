package v1

import (
	"time"

	
	"emshop/internal/app/emshop/api/service"
	v1 "emshop/internal/app/emshop/api/service/sms/v1"
	"emshop/internal/app/pkg/code"
	gin2 "emshop/internal/app/pkg/translator/gin"
	"emshop/pkg/common/core"
	"emshop/pkg/errors"
	"emshop/pkg/storage"

	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
)

type SendSmsForm struct {
	Mobile string `form:"mobile" json:"mobile" binding:"required,mobile"` //手机号码格式有规范可寻， 自定义validator
	Type   uint   `form:"type" json:"type" binding:"required,oneof=1 2"`
	//1. 注册发送短信验证码和动态验证码登录发送验证码
}

type SmsController struct {
	sf    service.ServiceFactory
	trans ut.Translator
}

func NewSmsController(sf service.ServiceFactory, trans ut.Translator) *SmsController {
	return &SmsController{sf, trans}
}

func (sc *SmsController) SendSms(c *gin.Context) {
	sendSmsForm := SendSmsForm{}
	if err := c.ShouldBind(&sendSmsForm); err != nil {
		gin2.HandleValidatorError(c, err, sc.trans)
	}

	smsCode := v1.GenerateSmsCode(6)
	err := sc.sf.Sms().SendSms(c, sendSmsForm.Mobile, "SMS_181850725", "{\"code\":"+smsCode+"}")
	if err != nil {
		core.WriteResponse(c, errors.WithCode(code.ErrSmsSend, err.Error()), nil)
		return
	}

	//将验证码保存起来 - redis
	rstore := storage.RedisCluster{}
	err = rstore.SetKey(c, sendSmsForm.Mobile, smsCode, 5*time.Minute)
	if err != nil {
		core.WriteResponse(c, errors.WithCode(code.ErrSmsSend, err.Error()), nil)
		return
	}

	core.WriteResponse(c, nil, nil)
}
