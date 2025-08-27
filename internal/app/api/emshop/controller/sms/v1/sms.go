package v1

import (
	"fmt"
	"time"

	upbv1 "emshop/api/user/v1"
	restserver "emshop/gin-micro/server/rest-server"
	"emshop/internal/app/api/emshop/service"
	v1 "emshop/internal/app/api/emshop/service/sms/v1"
	"emshop/internal/app/pkg/code"
	gin2 "emshop/internal/app/pkg/translator/gin"
	"emshop/pkg/common/core"
	"emshop/pkg/errors"
	"emshop/pkg/storage"

	"github.com/gin-gonic/gin"
)

type SmsController struct {
	sf    service.ServiceFactory
	trans restserver.I18nTranslator
}

func NewSmsController(sf service.ServiceFactory, trans restserver.I18nTranslator) *SmsController {
	return &SmsController{sf, trans}
}

func (sc *SmsController) SendSms(c *gin.Context) {
	var smsReq upbv1.SendSmsRequest

	if err := c.ShouldBind(&smsReq); err != nil {
		gin2.HandleValidatorError(c, err, sc.trans)
		return
	}

	// 手动验证必要字段（由于proto结构体没有binding标签）
	if smsReq.Mobile == "" {
		core.WriteResponse(c, errors.WithCode(code.ErrSmsSend, "mobile is required"), nil)
		return
	}
	if smsReq.Type != 1 && smsReq.Type != 2 {
		core.WriteResponse(c, errors.WithCode(code.ErrSmsSend, "type must be 1 or 2"), nil)
		return
	}

	smsCode := v1.GenerateSmsCode(6)
	
	// 开发环境：跳过实际短信发送，但保持完整的验证码生成和存储逻辑
	// 生产环境时需要取消注释下面的真实发送代码
	/*
	err := sc.sf.Sms().SendSms(c, smsReq.Mobile, "SMS_181850725", "{\"code\":"+smsCode+"}")
	if err != nil {
		core.WriteResponse(c, errors.WithCode(code.ErrSmsSend, "%s", err.Error()), nil)
		return
	}
	*/
	
	// 开发环境：在控制台输出验证码，方便测试
	fmt.Printf("==> 开发环境短信验证码 [%s]: %s\n", smsReq.Mobile, smsCode)

	//将验证码保存起来 - redis
	rstore := storage.RedisCluster{}
	// 区分开是登录还是注册的验证码
	key := fmt.Sprintf("%s_%d", smsReq.Mobile, smsReq.Type)
	err := rstore.SetKey(c, key, smsCode, 5*time.Minute)
	if err != nil {
		core.WriteResponse(c, errors.WithCode(code.ErrSmsSend, "%s", err.Error()), nil)
		return
	}

	core.WriteResponse(c, nil, nil)
}
