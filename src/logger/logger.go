package logger

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"
	"time"
)

// log level
const (
	LogLevel_Debug  = 0
	LogLevel_Notice = 1
	LogLevel_Info   = 2
	LogLevel_Warn   = 3
	LogLevel_Error  = 4
	LogLevel_Fatal  = 5
)

var (
	defaultLogLevel       = LogLevel_Debug
	errLevels             = []string{"DEBUG", "NOTICE", "INFO", "WARN", "ERROR", "FATAL"}
	errColors             = []string{"\033[;37m", "\033[;36m", "\033[;32m", "\033[;33m", "\033[;31m", "\033[41;1;33m"}
	defaultLogger         = &Logger{file: os.Stderr, name: "#", level: defaultLogLevel}
	internalLogger        = &Logger{file: openFile("internal.log"), name: "#", level: defaultLogLevel}
	ONE_HOUR_SECOND int64 = 3600
	pid                   = syscall.Getpid()
	waitchan              = make(chan int)
)

type logInfo struct {
	time  time.Time
	level int
	depth int
	file  string
	line  int
	msg   string
}

type logQueue chan func()

type Logger struct {
	file     *os.File
	cache    logQueue
	prefix   string
	name     string
	folder   string
	level    int
	lastTime int64
}

// example: NewLogger( ".", "access.log", "AS", "DEBUG" )
//
func NewLogger(folder, filePrefix, name, levelStr string) *Logger {

	levelStr = strings.TrimSpace(levelStr)
	level := defaultLogLevel

	internalLogger.name = name

	for k, lv := range errLevels {
		if lv == levelStr {
			level = k
			break
		}
	}

	if folder != "." && folder != "./" {
		if err := os.MkdirAll(filepath.Dir(folder+"/"), 0777); err != nil && !os.IsExist(err) {
			LOG_ERROR("logger: %s", err.Error())
		}
	}

	logger := &Logger{
		file:     nil,
		cache:    make(logQueue, 200000),
		prefix:   filePrefix,
		name:     name,
		folder:   folder,
		level:    level,
		lastTime: 0,
	}

	go logger.start()

	return logger
}

func InitializeLogger(folder, filePrefix, name, levelStr string) {
	DEBUG("初始化日志: %s/%s %s %s", folder, filePrefix, name, levelStr)
	defaultLogger = NewLogger(folder, filePrefix, name, levelStr)
}

func openFile(filename string) *os.File {

	if filename == "os.Stderr" {
		return os.Stderr
	}

	if filename == "os.Stdout" {
		return os.Stdout
	}

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)

	if err != nil {
		fmt.Fprintf(os.Stderr, "openFile: %s", err.Error())
		return os.Stderr
	}

	return f
}

func SetLoggerLevel(level int) {
	defaultLogger.Level(level)
}

func SetLoggerPrefix(prefix string) {
	defaultLogger.Prefix(prefix)
	internalLogger.Prefix(prefix)
}

func WaitLogger(timeout time.Duration) {

	LOG_INFO("WaitLogger: %d", len(defaultLogger.cache))

	go func() {
		if nil == defaultLogger.cache || 0 == cap(defaultLogger.cache) {
			close(waitchan)
		} else {
			defaultLogger.cache <- func() { close(waitchan) }
		}
	}()

	select {
	case <-time.After(timeout):
		return
	case <-waitchan:
		return
	}
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////
func (this *Logger) Level(level int) {
	this.level = level
}

func (this *Logger) Prefix(prefix string) {
	this.name = prefix
}

func (this *Logger) start() {

	ticker := time.Tick(time.Second)

	for {
		select {
		case fn, ok := <-this.cache:
			if !ok {
				return
			}
			fn()
		case <-ticker:
			if len(this.cache) > 10 {
				LOG_DEBUG("logger_task: %d/%d", len(this.cache), cap(this.cache))
			}
		case <-waitchan:
			return
		}
	}
}

func (this *Logger) format(t time.Time, level int, file string, line int, msg string, for_console bool) string {

	// 2014/04/29-09:17:46 PL DEBUG client_task.cpp:169 Platfrom Message.
	lvStr := errLevels[level]

	year, month, day := t.Date()
	hour, min, sec := t.Clock()

	prefix, suffix := "", ""

	if for_console {
		prefix = errColors[level]
		suffix = "\033[0m"
	}

	return fmt.Sprintf("%d/%02d/%02d-%02d:%02d:%02d %s[%d] %s%s %s:%d %s%s\n",
		year, month, day, hour, min, sec, this.name, pid, prefix, lvStr, file, line, msg, suffix)
}

func (this *Logger) prepareLogFile(now time.Time) {

	if this.file != nil {
		this.file.Sync()
		this.file.Close()
		this.file = nil
	}

	// access.20140507-10

	year, month, day := now.Date()
	hour, _, _ := now.Clock()

	filename := fmt.Sprintf("%s/%s.%d%02d%02d-%02d", this.folder, this.prefix, year, month, day, hour)

	filename, _ = filepath.Abs(filename)

	this.file = openFile(filename)
	this.lastTime = now.Unix()

	argv := []string{"-s", "-f", filename, this.folder + "/" + this.prefix}

	// ln --force -s old new
	// ln 必须使用绝对路径
	excuteCommand("ln", argv)
}

func (this *Logger) writeLog(now time.Time, level int, file string, line int, msg string) {

	sec := now.Unix()

	if this.prefix != "" && (nil == this.file || sec/ONE_HOUR_SECOND != this.lastTime/ONE_HOUR_SECOND) {
		this.prepareLogFile(now)
	}

	// 渲染颜色
	renderColor := false

	if os.Stderr == this.file || os.Stdout == this.file {
		renderColor = true
	}

	str := this.format(now, level, file, line, msg, renderColor)

	if this.file != nil {
		if _, err := this.file.WriteString(str); nil != err {
			fmt.Fprintf(os.Stderr, "logger: %s, %s", err.Error(), str)
			this.prepareLogFile(now)
		}
		//this.file.Sync()
	}
}

func (this *Logger) logmsg(level, depth int, msg string) {

	_, file, line, ok := runtime.Caller(depth)

	if !ok {
		file = "???"
		line = 0
	}

	if rets := strings.Split(file, "/"); rets != nil {
		file = rets[len(rets)-1]
	}

	if nil == this.cache || 0 == cap(this.cache) {
		this.writeLog(time.Now(), level, file, line, msg)
	} else {
		info := logInfo{time: time.Now(), level: level, file: file, line: line, msg: msg}
		this.cache <- func() { this.writeLog(info.time, info.level, info.file, info.line, info.msg) }
	}
}

func excuteCommand(cmdStr string, args []string) error {

	cmd := exec.Command(cmdStr, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		DEBUG("ExcuteCommand Fail. ", err.Error())
		return err
	}

	if err := cmd.Wait(); err != nil {
		DEBUG("command to exit: %s", err.Error())
		return err
	}

	cmd.Process.Kill()
	return nil
}

const depth = 3

func (this *Logger) Debug(format string, args ...interface{}) {

	if LogLevel_Debug >= this.level {
		this.logmsg(LogLevel_Debug, depth, fmt.Sprintf(format, args...))
	}
}

func (this *Logger) Notice(format string, args ...interface{}) {

	if LogLevel_Notice >= this.level {
		this.logmsg(LogLevel_Notice, depth, fmt.Sprintf(format, args...))
	}
}

func (this *Logger) Info(format string, args ...interface{}) {

	if LogLevel_Info >= this.level {
		this.logmsg(LogLevel_Info, depth, fmt.Sprintf(format, args...))
	}
}

func (this *Logger) Warn(format string, args ...interface{}) {

	if LogLevel_Warn >= this.level {
		this.logmsg(LogLevel_Warn, depth, fmt.Sprintf(format, args...))
	}
}

func (this *Logger) Error(format string, args ...interface{}) {

	if LogLevel_Error >= this.level {
		this.logmsg(LogLevel_Error, depth, fmt.Sprintf(format, args...))
	}
}

func (this *Logger) Fatal(format string, args ...interface{}) {

	if LogLevel_Fatal >= this.level {
		this.logmsg(LogLevel_Fatal, depth, fmt.Sprintf(format, args...))

		debug.PrintStack()
		os.Exit(1)
	}
}

///////////////////////////////////////////////////////////////////////////

func DEBUG(format string, args ...interface{}) {

	if LogLevel_Debug >= defaultLogger.level {
		defaultLogger.logmsg(LogLevel_Debug, 2, fmt.Sprintf(format, args...))
	}
}

func NOTICE(format string, args ...interface{}) {

	if LogLevel_Notice >= defaultLogger.level {
		defaultLogger.logmsg(LogLevel_Notice, 2, fmt.Sprintf(format, args...))
	}
}

func INFO(format string, args ...interface{}) {

	if LogLevel_Info >= defaultLogger.level {
		defaultLogger.logmsg(LogLevel_Info, 2, fmt.Sprintf(format, args...))
	}
}

func WARN(format string, args ...interface{}) {

	if LogLevel_Warn >= defaultLogger.level {
		defaultLogger.logmsg(LogLevel_Warn, 2, fmt.Sprintf(format, args...))
	}
}

func ERROR(format string, args ...interface{}) {

	if LogLevel_Error >= defaultLogger.level {
		defaultLogger.logmsg(LogLevel_Error, 2, fmt.Sprintf(format, args...))
	}
}

func FATAL(format string, args ...interface{}) {

	if LogLevel_Fatal >= defaultLogger.level {

		msg := fmt.Sprintf(format, args...)

		defaultLogger.logmsg(LogLevel_Fatal, 2, msg)

		panic(msg)
	}
}

//////////////////////////////////////////////////////////////////////////
func LOG_DEBUG(format string, args ...interface{}) {

	if LogLevel_Debug >= internalLogger.level {
		internalLogger.logmsg(LogLevel_Debug, 2, fmt.Sprintf(format, args...))
	}
}

func LOG_NOTICE(format string, args ...interface{}) {

	if LogLevel_Notice >= internalLogger.level {
		internalLogger.logmsg(LogLevel_Notice, 2, fmt.Sprintf(format, args...))
	}
}

func LOG_INFO(format string, args ...interface{}) {

	if LogLevel_Info >= internalLogger.level {
		internalLogger.logmsg(LogLevel_Info, 2, fmt.Sprintf(format, args...))
	}
}

func LOG_WARN(format string, args ...interface{}) {

	if LogLevel_Warn >= internalLogger.level {
		internalLogger.logmsg(LogLevel_Warn, 2, fmt.Sprintf(format, args...))
	}
}

func LOG_ERROR(format string, args ...interface{}) {

	if LogLevel_Error >= internalLogger.level {
		internalLogger.logmsg(LogLevel_Error, 2, fmt.Sprintf(format, args...))
	}
}

func LOG_FATAL(format string, args ...interface{}) {

	if LogLevel_Fatal >= internalLogger.level {

		msg := fmt.Sprintf(format, args...)

		internalLogger.logmsg(LogLevel_Fatal, 2, msg)

		panic(msg)
	}
}
