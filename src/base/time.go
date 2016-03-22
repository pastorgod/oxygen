package base

import (
	. "logger"
	"time"
)

// 秒(sec)
func Second() int64 {
	return Now().Unix()
}

// 毫秒(ms)
func MSecond() int64 {
	return USecond() / 1000
}

// 微秒(us)
func USecond() int64 {
	return NSecond() / 1000
}

// 纳秒(ns)
func NSecond() int64 {
	return Now().UnixNano()
}

// 休眠指定秒
func Sleep(second int) {
	time.Sleep(time.Second * time.Duration(second))
}

// 休眠指定毫秒
func MSleep(ms int) {
	time.Sleep(time.Millisecond * time.Duration(ms))
}

const (
	Format_One_Minute int64 = 1 * 60
	Format_One_Hour   int64 = Format_One_Minute * 60
	Format_One_Day    int64 = Format_One_Hour * 24

	// 一天24小时的秒钟
	ONE_DAY_SECONDS int64 = 3600 * 24
)

// 将秒转换为可视化时间
func FormatElapsedTime(sec int64) (day, hour, minute, second int32) {

	if sec >= Format_One_Day {
		day = int32(sec / Format_One_Day)
		sec %= Format_One_Day
	}

	if sec >= Format_One_Hour {
		hour = int32(sec / Format_One_Hour)
		sec %= Format_One_Hour
	}

	if sec >= Format_One_Minute {
		minute = int32(sec / Format_One_Minute)
		sec %= Format_One_Minute
	}

	second = int32(sec)
	return
}

// 格式化为字符串
func FormatElapsedTimeStr(sec int64) string {

	day, hour, minute, second := FormatElapsedTime(sec)

	return Sprintf("%d天%d小时%d分钟%d秒", day, hour, minute, second)
}

func FormatUnix(t int64) string {
	now := Unix(t, 0)
	return Sprintf("%d-%02d-%02d %02d:%02d:%02d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
}

var (
	FullTimeFormat    = "2006-01-02 15:04:05"
	SectionDateFormat = "2006-01-02"
	SectionTimeFormat = "15:04:05"
)

// 策划填写的字符串时间一般都是 本地时间 所以也也应该已本地时间去解析 在取unix时间才是对的
// 从字符串解析过来的时间都是本地时间格式包含时区偏移量

// 解析 年-月-日 时:分:秒 格式的时间
func ParseTime(str string) (time.Time, error) {

	time, err := time.ParseInLocation(FullTimeFormat, str, time.Local)

	if err != nil {
		LOG_ERROR("时间格式错误: %s", err.Error())
		PrintStack("ParseTime: %s", str)
		return time, err
	}

	return time, nil
}

// 解析 年-月-日 格式的时间
func ParseSectionDate(str string) (time.Time, bool) {

	time, err := time.ParseInLocation(SectionDateFormat, str, time.Local)

	if err != nil {
		LOG_ERROR("时间格式错误: %s", err.Error())
		PrintStack("ParseSectionDate: %s", str)
		return time, false
	}

	return time, true
}

// 解析 时:分:秒 格式的时间
func ParseSectionTime(str string) (time.Time, bool) {

	time, err := time.ParseInLocation(SectionTimeFormat, str, time.Local)

	if err != nil {
		LOG_ERROR("时间格式错误: %s", err.Error())
		PrintStack("ParseSectionTime: %s", str)
		return time, false
	}

	return time, true
}

// a时间 - b时间 返回结果 秒
func TimeSub(a, b time.Time) int64 {
	return int64(a.Sub(b) / time.Second)
}

var set_time_offset time.Duration

// 返回当前时间
func Now() time.Time {
	return time.Now().Add(set_time_offset)
}

// 返回服务器时区
func ZoneOffset() int32 {
	_, offset := time.Now().Zone()

	return int32(offset)
}

// 调整时间
func Adjust(t time.Time) {

	duration := t.Sub(time.Now())

	if duration >= 0 {
		DEBUG("Adjust: %d, %s", duration, FormatElapsedTimeStr(int64(duration/time.Second)))
	} else {
		DEBUG("Adjust: %d, %s", duration, FormatElapsedTimeStr(int64(-duration/time.Second)))
	}

	set_time_offset = duration

	DEBUG("设置之后的时间: %s", ServerTimeStr())
}

// 返回指定秒 + 纳秒的时间
func Unix(sec int64, nsec int64) time.Time {
	return time.Unix(sec, nsec)
}

// 返回日期
func Date(year, month, day, hour, minute, second int32) time.Time {
	return time.Date(int(year), time.Month(month), int(day), int(hour), int(minute), int(second), 0, time.Local)
}

// 检查是否需要重置
func CheckReset(last_time int64, hour int) bool {

	if 0 == last_time {
		return false
	}

	now := Now()
	now_time := now.Unix()

	// 最后重置的时间超过一天 则需要重置
	// 当前时间 - 最后一次重置的时间大于等于24小时
	if (now_time - last_time) >= ONE_DAY_SECONDS {
		return true
	}

	y, m, d := now.Date()
	reset_time := time.Date(y, m, d, hour, 0, 0, 0, time.Local).Unix()

	// 如果当天 则判断是否是重置之前
	if last_time < reset_time {
		return true
	}

	return false
}

// 转换到当天开始
func ToDayBegin(t int64) time.Time {
	t1 := Unix(t, 0)
	return time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, time.UTC)
}

// 转换到当天结束
func ToDayEnd(t int64) time.Time {
	t1 := Unix(t, 0)
	return time.Date(t1.Year(), t1.Month(), t1.Day(), 23, 59, 59, 0, time.UTC)
}

// 检查时间是否在今天
func CheckInDay(t1 int64) bool {
	now_year, now_month, now_day := Now().Date()
	last_year, last_month, last_day := time.Unix(t1, 0).Date()

	return now_year == last_year && now_month == last_month && now_day == last_day
}

// 检查两个时间是否在同一月
func CheckInMonth(t1 int64) bool {
	now_year, now_month, _ := Now().Date()
	last_year, last_month, _ := time.Unix(t1, 0).Date()

	return now_year == last_year && now_month == last_month
}

var ms_server_start_time int64 = Second()

// 服务器启动时的时间
func ServerStartupTime() int64 {
	return ms_server_start_time
}

// 从服务器启动以来所经过的时间
func ElapsedSinceStartupTime() int64 {
	return Second() - ServerStartupTime()
}

// 当前时间的字符串格式
func ServerTimeStr() string {
	return Now().String()
}

// 获得指定时刻剩余时长
func Timeleft(hour, minute, second int) time.Duration {

	now := Now()
	y, m, d := now.Date()

	// 奖励发送的那一刻
	rewardTime := time.Date(y, m, d, hour, minute, second, 0, time.Local)

	var timeout time.Duration

	// 当前时间 > 开始时间则设置为下一天的时间
	if now.After(rewardTime) {
		// 明天的奖励时间
		rewardTime = time.Date(y, m, d+1, hour, minute, second, 0, time.Local)
		timeout = rewardTime.Sub(now)
	} else {
		// 今天的奖励时间还没有到
		timeout = rewardTime.Sub(now)
	}

	return timeout
}

var ms_startup_time = Second()

func SetupStartupTime(str string) {
	t, err := ParseTime(str)
	if nil != err {
		ERROR("******** 时间格式错误 ********* %s, %v", str, err)
		return
	}
	ms_startup_time = t.Unix()
	INFO("************ 开服时间: %s ***********", str)
}

// 服务器开服时间
func StartupTime() int64 {
	return ms_startup_time
}
