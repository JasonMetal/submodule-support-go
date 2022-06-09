package formats

import (
	"regexp"
)

// CheckMobile 验证手机号格式
func CheckMobile(str string) bool {
	regRule := "^1[345789]{1}\\d{9}$"
	reg := regexp.MustCompile(regRule)
	return reg.MatchString(str)
}

// EncodePhone 手机号脱敏
func EncodePhone(phone string) string {
	strLen := len(phone)
	if strLen < 3 {
		return "****"
	}

	if strLen <= 7 {
		return phone[0:3] + "****"
	} else {
		return phone[0:3] + "****" + phone[strLen-4:strLen]
	}
}
