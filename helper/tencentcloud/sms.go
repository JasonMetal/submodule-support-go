package tencentcloud

import (
	"encoding/json"
	"errors"
	"gitee.com/DXTeam/idea-go.git/bootstrap"
	"gitee.com/DXTeam/idea-go.git/helper/slices"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"strings"
)

type SmsParams struct {
	TemplateId, SignName   string
	TemplateParams, Mobile []string
}

// Send 发送短信
func (s *SmsParams) Send() (b []byte, err error) {
	c := bootstrap.TencentSmsClient

	// 使用新签名
	if s.SignName != "" {
		c.Request.SignName = common.StringPtr(s.SignName)
	}
	if s.TemplateId == "" {
		err := errors.New("模板ID不能为空!")
		return nil, err
	}
	if len(s.Mobile) == 0 {
		err := errors.New("手机号不能为空!")
		return nil, err
	}

	// 处理手机号 自动兼容 带+号的手机
	for key, value := range s.Mobile {
		//如果存在+号则跳过
		if strings.Index(value, "+") >= 0 {
			continue
		}
		s.Mobile[key] = c.MobileCode + value
	}
	//去重手机号
	s.Mobile = slices.Unique(s.Mobile)
	// 处理模板参数
	if len(s.TemplateParams) > 0 {
		c.Request.TemplateParamSet = common.StringPtrs(s.TemplateParams)
	}
	// 处理模板ID
	c.Request.TemplateId = common.StringPtr(s.TemplateId)
	c.Request.PhoneNumberSet = common.StringPtrs(s.Mobile)
	response, err := c.Client.SendSms(c.Request)

	b, err = json.Marshal(response.Response)
	return b, err
}
