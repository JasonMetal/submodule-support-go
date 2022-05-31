package formats

import (
	"regexp"
)

//验证手机号格式
func CheckMobile(str string) bool {
	regRule := "^1[345789]{1}\\d{9}$"
	reg := regexp.MustCompile(regRule)
	return reg.MatchString(str)
}
