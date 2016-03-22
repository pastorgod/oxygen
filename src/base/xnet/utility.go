package xnet

import (
	"fmt"
	. "logger"
	"os"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

func PrintStack(str string, args ...interface{}) {

	msg := Sprintf(str, args...)

	LOG_ERROR("***************** CALLSTACK BEGIN *********************")
	LOG_ERROR("App: %s, Time: %s", os.Args[0], time.Now().String())
	LOG_ERROR("Message: %s", msg)
	LOG_ERROR("CallStack:\n%s", string(debug.Stack()))
	LOG_ERROR("***************** CALLSTACK END ***********************")

	ERROR("Exception: %s", msg)
}

// 定时执行方法
func AfterFunc(timeout time.Duration, fn func()) *time.Timer {

	return time.AfterFunc(timeout, func() {

		// 等到逻辑那一块一起执行
		//PushTask(func() {

		defer func() {
			if err := recover(); err != nil {
				PrintStack("AfterFunc Exception: %v", err)
			}
		}()

		fn()
		//})
	})
}

// 计算字符串hash, BKDRHash
func Hash(str string) uint32 {

	// 31 131 1313 13131 131313 etc..
	seed, hash := uint32(131), uint32(0)

	for _, c := range str {
		hash = hash*seed + uint32(c)
	}

	return (hash & 0x7FFFFFFF)
}

func Sprintf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func ToString(obj interface{}) string {
	return fmt.Sprintf("%v", obj)
}

func ToName(obj interface{}) string {
	return reflect.TypeOf(obj).Elem().Name()
}

func ToError(str string, args ...interface{}) error {
	return fmt.Errorf(str, args...)
}

func Split(s, sep string) []string {
	return strings.Split(s, sep)
}

func Assert(exp bool, msgs ...interface{}) {

	if exp {
		return
	}

	var msg = "ASSERT ERROR"

	if 0 != len(msgs) {
		msg += Sprintf(": %v", msgs)
	}

	panic(msg)
}

func Atoi(s string) int {

	num, err := strconv.Atoi(s)

	if err != nil {
		LOG_ERROR("strconv.Atoi Fail, %s", err.Error())
	}

	return num
}
