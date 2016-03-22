package base

import (
	"regexp"
)

// 邮箱验证
var msMailRegexp = regexp.MustCompile("^([a-zA-Z0-9_-])+@([a-zA-Z0-9_-])+(.[a-zA-Z0-9_-])+")

// 手机号验证
var msPhoneRegexp = regexp.MustCompile("^(13[0-9]|14[57]|15[0-35-9]|18[07-9])\\d{8}$")

// IDFA 验证
var msidfaRegexp = regexp.MustCompile("^[A-Z0-9]{8}-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{12}$")

func IsEmail(str string) bool {
	return msMailRegexp.MatchString(str)
}

func IsPhoneNumber(str string) bool {
	return msPhoneRegexp.MatchString(str)
}

func IsIDFA(str string) bool {
	return msidfaRegexp.MatchString(str)
}
