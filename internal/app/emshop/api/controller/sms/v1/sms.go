package v1

import (
	"fmt"
	"time"

	
	"emshop/internal/app/emshop/api/service"
	v1 "emshop/internal/app/emshop/api/service/sms/v1"
	"emshop/internal/app/pkg/code"
	gin2 "emshop/internal/app/pkg/translator/gin"
	"emshop/pkg/common/core"
	"emshop/pkg/errors"
	"emshop/pkg/storage"
	"emshop/gin-micro/server/rest-server"

	"github.com/gin-gonic/gin"
)

type SendSmsForm struct {
	Mobile string `form:"mobile" json:"mobile" binding:"required,mobile"` //手机号码格式有规范可寻， 自定义validator
	Type   uint   `form:"type" json:"type" binding:"required,oneof=1 2"`	// 1: 注册, 2: 登录
	
}

type SmsController struct {
	sf    service.ServiceFactory
	trans restserver.I18nTranslator
}

func NewSmsController(sf service.ServiceFactory, trans restserver.I18nTranslator) *SmsController {
	return &SmsController{sf, trans}
}

func (sc *SmsController) SendSms(c *gin.Context) {
	sendSmsForm := SendSmsForm{}
	if err := c.ShouldBind(&sendSmsForm); err != nil {
		gin2.HandleValidatorError(c, err, sc.trans)
	}

	smsCode := v1.GenerateSmsCode(6)
	
	// 开发环境：跳过实际短信发送，但保持完整的验证码生成和存储逻辑
	// 生产环境时需要取消注释下面的真实发送代码
	/*
	err := sc.sf.Sms().SendSms(c, sendSmsForm.Mobile, "SMS_181850725", "{\"code\":"+smsCode+"}")
	if err != nil {
		core.WriteResponse(c, errors.WithCode(code.ErrSmsSend, "%s", err.Error()), nil)
		return
	}
	*/
	
	// 开发环境：在控制台输出验证码，方便测试
	fmt.Printf("==> 开发环境短信验证码 [%s]: %s\n", sendSmsForm.Mobile, smsCode)

	//将验证码保存起来 - redis
	rstore := storage.RedisCluster{}
	// 区分开是登录还是注册的验证码
	key := fmt.Sprintf("%s_%d", sendSmsForm.Mobile, sendSmsForm.Type)
	err := rstore.SetKey(c, key, smsCode, 5*time.Minute)
	if err != nil {
		core.WriteResponse(c, errors.WithCode(code.ErrSmsSend, "%s", err.Error()), nil)
		return
	}

	core.WriteResponse(c, nil, nil)
}
