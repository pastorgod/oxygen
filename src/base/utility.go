package base

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"time"

	proto "github.com/gogo/protobuf/proto"
	. "logger"
)

func ProtoUInt32(val uint32) *uint32 {
	return &val
}

func ProtoInt32(val int32) *int32 {
	return &val
}

func ProtoInt64(val int64) *int64 {
	return &val
}

func ProtoBool(val bool) *bool {
	return &val
}

func ProtoString(val string) *string {
	return &val
}

//////////////////////////////////////////////////////////////////////////
func Max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func MaxInt32(a, b int32) int32 {
	if a > b {
		return a
	}

	return b
}

func MinInt32(a, b int32) int32 {
	if a < b {
		return a
	}

	return b
}

func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}

	return b
}

func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}

	return b
}

func CeilInt32(value float32) int32 {
	return int32(math.Ceil(float64(value)))
}

func FloorInt32(value float32) int32 {
	return int32(math.Floor(float64(value)))
}

func CeilInt64(value float64) int64 {
	return int64(math.Ceil(value))
}

func FloorInt64(value float64) int64 {
	return int64(math.Floor(value))
}

//////////////////////////////////////////////////////////////////////////

var defaultSource = rand.NewSource(NSecond())
var defaultRander = rand.New(defaultSource)

//var newRander = NewMersenneTwister( defaultRander.Intn(10000000) )

func RandBetween(min, max int32) int32 {

	if max <= 0 {
		max = min
	}

	//	return int32(newRander.Next( int(min), int(max) ))
	return int32((defaultRander.Intn(1<<31-1) % (int(max) - int(min) + 1)) + int(min))
}

func PrintStack(str string, args ...interface{}) {

	msg := Sprintf(str, args...)

	LOG_ERROR("***************** CALLSTACK BEGIN *********************")
	LOG_ERROR("App: %s, Time: %s", AppName(), time.Now().String())
	LOG_ERROR("Message: %s", msg)
	LOG_ERROR("CallStack:\n%s", string(debug.Stack()))
	LOG_ERROR("***************** CALLSTACK END ***********************")

	ERROR("Exception: %s", msg)
}

func SubStringRange(str string, begin, length int) string {

	rs := []rune(str)
	lth := len(rs)

	if begin < 0 {
		begin = 0
	}

	if length <= 0 {
		length = lth
	}

	if begin >= lth {
		begin = lth
	}

	end := begin + length

	if end > lth {
		end = lth
	}

	return string(rs[begin:end])
}

// 切割字符串
func SubString(str string, offset int) string {
	return SubStringRange(str, offset, -1)
}

// 后台形式运行程序
func Daemon() error {

	if 1 == os.Getppid() {
		FATAL("use error. the parent process is init.")
	}

	args := []string{os.Args[0]}

	if cfile := Configuire(); cfile != "" && cfile != (AppName()+".json") {
		args = append(args, fmt.Sprintf("-f=%s", cfile))
	}

	process, err := os.StartProcess(os.Args[0], args, &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	})

	if err != nil {
		ERROR("StartProcess fail: %v", err)
		return err
	}

	DEBUG("child process started successfully, parent exited.")

	process.Release()

	os.Exit(0)
	return nil
}

// 改变工作目录
func Chdir(path string) {

	if err := os.Chdir(path); err != nil {
		LOG_ERROR("切换工作目录失败: %s", err.Error())
		return
	}

	DEBUG("切换工作目录到: %s", path)
}

// 服务器程序名字
func AppName() string {

	index := strings.LastIndex(os.Args[0], "/")

	return SubStringRange(os.Args[0], index+1, -1)
}

// 服务器程序路径
func AppPath() string {

	if path, err := filepath.Abs(os.Args[0]); nil == err {
		return filepath.Dir(path)
	}

	index := strings.LastIndex(os.Args[0], "/")
	return SubStringRange(os.Args[0], 0, index)
}

// 主机名
func Hostname() string {

	name, err := os.Hostname()

	if err != nil {
		LOG_ERROR("Hostname: %v", err)
		return "127.0.0.1"
	}

	return name
}

// 当前运行账户名
func Whoami() string {

	userData, err := user.Current()

	if err != nil {
		LOG_ERROR("Username: %v", err)
		return "<unknown>"
	}

	return userData.Username
}

// 退出程序
func Exit(code int) {
	os.Exit(code)
}

// 从路径中获取文件名
func FileName(path string) string {

	index := strings.LastIndex(path, "/")

	return SubStringRange(path, index+1, -1)
}

// Int32 Hash
func HashInt32(x int32) int32 {
	x = ((x >> 16) ^ x) * 0x45d9f3b
	x = ((x >> 16) ^ x) * 0x45d9f3b
	x = ((x >> 16) ^ x)
	return x
}

func HashInt32Mod(x, max int32) int32 {
	x = ((x >> 16) ^ x) * 0x45d9f3b
	x = ((x >> 16) ^ x) * 0x45d9f3b
	x = ((x >> 16) ^ x)
	return x % max
}

// 计算MD5
func Md5(str string) string {

	sum := md5.Sum([]byte(str))

	return hex.EncodeToString(sum[:])
}

// 计算Base64值
func Base64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func Base64Byte(dst []byte) string {
	return base64.StdEncoding.EncodeToString(dst)
}

func Sprintf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func ToLower(str string) string {
	return strings.ToLower(str)
}

func ToUpper(str string) string {
	return strings.ToUpper(str)
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

// 转换对象为json字符串
func ToJsonData(obj interface{}) ([]byte, error) {
	return json.Marshal(obj)
}

// 转换有格式的json字符串
func ToJsonIndentData(obj interface{}) ([]byte, error) {
	return json.MarshalIndent(obj, "", "\t")
}

// 转换成对象
func ToJsonObject(data []byte, object interface{}) error {
	return json.Unmarshal(data, object)
}

// 从文件加载json对象
func LoadJsonFromFile(fname string, result interface{}) error {

	data, err := ioutil.ReadFile(fname)

	if err != nil {
		return err
	}

	return ToJsonObject(data, result)
}

// 序列化json到文件
func SaveJsonToFile(fname string, result interface{}) error {

	data, err := ToJsonIndentData(result)

	if err != nil {
		return err
	}

	return WriteFile(fname, data)
}

// 转换为proto数据
func ToProtoData(object proto.Message) ([]byte, error) {
	return proto.Marshal(object)
}

// 转换成proto对象
func ToProtoObject(data []byte, object proto.Message) error {
	return proto.Unmarshal(data, object)
}

// 判断文件是否已经存在
func IsExist(fname string) bool {
	_, err := os.Stat(fname)
	return nil == err
}

func Atoi(s string) int {

	num, err := strconv.Atoi(s)

	if err != nil {
		ERROR("strconv.Atoi Fail, %s", err.Error())
	}

	return num
}

func Atof(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)

	if err != nil {
		ERROR("strconv.ParseFloat: %s", err.Error())
	}

	return f
}

func AtoFloat32(s string) float32 {
	return float32(Atof(s))
}

func AtoInt32(s string) int32 {
	return int32(Atoi(s))
}

func AtoInt64(s string) int64 {

	value, err := strconv.ParseInt(s, 10, 64)

	if err != nil {
		if val, ferr := strconv.ParseFloat(s, 64); nil == ferr {
			return int64(val)
		}
		ERROR("strconv.ParseInt Fail, %s", err.Error())
	}

	return value
}

func AtoUint64(s string) uint64 {

	value, err := strconv.ParseUint(s, 10, 64)

	if err != nil {
		if val, ferr := strconv.ParseFloat(s, 64); nil == ferr {
			return uint64(val)
		}
		ERROR("strconv.ParseInt Fail, %s", err.Error())
	}

	return value
}

func Itoa(i int) string {
	return strconv.Itoa(i)
}

func Split(s, sep string) []string {
	return strings.Split(s, sep)
}

const Empty string = ""

func StringIsEmpty(str string) bool {
	return Empty == str
}

func SetMaxProcessor(n int) {

	if n > 0 {
		runtime.GOMAXPROCS(n)
		DEBUG("setup max processor: %d", n)
	} else {
		runtime.GOMAXPROCS(runtime.NumCPU())
		DEBUG("setup max processor: %d", runtime.NumCPU())
	}
}

var PID = syscall.Getpid()

/*
func FillServerState(info *command.ServerStateResponse) {

	state := &runtime.MemStats{}
	runtime.ReadMemStats(state)

	info.Server = ProtoString(ServerName())
	info.Version = ProtoString(VERSION())
	info.ServerTime = Second()
	info.Pid = int32(PID)
	info.ThreadNum = int32(runtime.GOMAXPROCS(-1))
	info.GoNum = int32(runtime.NumGoroutine())
	info.StartupTime = ServerStartupTime()
	info.ElapsedTime = ElapsedSinceStartupTime()
	info.Configuire = ProtoString(Configuire())
	info.TaskSpeed = xnet.TaskSpeed()
	info.TaskCap = xnet.TaskCap()
	info.TaskLen = xnet.TaskLen()
	info.GcPause = int64(state.PauseNs[(state.NumGC+255)%256])
	info.GcNum = int64(state.NumGC)
	info.ConsumeMem = int64(state.Alloc)
}
*/

// 写入文件
func WriteFile(fname string, data []byte) error {

	if err := os.MkdirAll(filepath.Dir(fname), 0777); err != nil && !os.IsExist(err) {
		LOG_ERROR("创建文件夹失败! %s @ %s", err.Error(), fname)
		return err
	}

	if err := ioutil.WriteFile(fname, data, 0644); err != nil {
		LOG_ERROR("写入文件失败: %s %s", fname, err.Error())
		return err
	}

	return nil
}

func WriteStream(fname string, r io.Reader) error {

	if err := os.MkdirAll(filepath.Dir(fname), 0777); err != nil && !os.IsExist(err) {
		LOG_ERROR("创建文件夹失败! %s @ %s", err.Error(), fname)
		return err
	}

	// 创建文件
	file, err := os.Create(fname)

	if err != nil {
		LOG_ERROR("os.Create: %v", err)
		return err
	}

	defer file.Close()

	_, err = io.Copy(file, r)

	if err != nil {
		LOG_ERROR("io.Copy: %v", err)
		return err
	}

	return nil
}

func IsSlice(r interface{}) bool {

	val := reflect.ValueOf(r)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	return val.Kind() == reflect.Slice
}

func ReadLine(data []byte) []byte {

	for index, d := range data {
		if '\r' == d || '\n' == d {
			return data[:index]
		}
	}

	return data
}

func ReadLineString(data []byte) string {
	return string(ReadLine(data))
}

func VERSION() string {
	return serverVersion
}

func Version() (major, minor, revision int) {
	return major, minor, revision
}

////////////////////////////////////////////////////////////////////////////////////////////////

type IntegerMap map[int32]int32

func (this *IntegerMap) MarshalJSON() (data []byte, err error) {

	trans := make(map[string]int32, len(*this))

	for k, v := range *this {
		trans[ToString(k)] = v
	}

	return ToJsonData(trans)
}

func (this *IntegerMap) UnmarshalJSON(data []byte) error {

	trans := make(map[string]int32)

	if err := ToJsonObject(data, &trans); err != nil {
		return err
	}

	for k, v := range trans {
		(*this)[AtoInt32(k)] = v
	}

	return nil
}

type IntegerStringMap map[int32]string

func (this *IntegerStringMap) MarshalJSON() (data []byte, err error) {

	trans := make(map[string]string, len(*this))

	for k, v := range *this {
		trans[ToString(k)] = v
	}

	return ToJsonData(trans)
}

func (this *IntegerStringMap) UnmarshalJSON(data []byte) error {

	trans := make(map[string]string)

	if err := ToJsonObject(data, &trans); err != nil {
		return err
	}

	for k, v := range trans {
		(*this)[AtoInt32(k)] = v
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////

func OnlineKey() string {
	return "GAMESERVERS:ONLINECOUNT"
}

func LoginKey(accid string) string {
	return Sprintf("LOGIN:%s", accid)
}

func ReloginKey(server, accid string) string {
	return Sprintf("RELOGIN:%s:%s", server, accid)
}

func UserListKey(server string) string {
	return Sprintf("%s:USERLIST", server)
}

func PayKey(server string) string {
	return Sprintf("PAY:%s", server)
}

func GiftBagKey(activity_id int32) string {
	return Sprintf("GIFTBAG:%d", activity_id)
}

func GiftCodeKey(activity_id int32) string {
	return Sprintf("GIFTCODE:%d", activity_id)
}

func CompenRechargeKey() string {
	return "COMPENSATION:RECHARGES"
}

func CompenAllKey() string {
	return "COMPENSATION:ALLPLAYERS"
}

func AccountDelKey() string {
	return "KROEA:DEL:ACCOUNT"
}

/*
func InitLogger(conf *command.LoggerConf) {

	log_path, log_file := conf.GetLogPath(), conf.GetLogFile()

	InitializeLogger(log_path, log_file, conf.GetLogName(), conf.GetLogLevel())

	// 重定向 stdout | stderr 到文件
	RedirectErrorToFile()
}
*/

// 加载配置文件
func LoadConfiguire(conf interface{}) {

	if err := LoadJsonFromFile(Configuire(), conf); err != nil {
		ERROR("加载配置文件失败: %s %s %v", Configuire(), ToName(conf), err)
		Sleep(1)
		Exit(1)
	}
}

// 加载配置文件
func LoadConfig() (*ConfigFile, error) {
	return ReadConfigFile(Configuire())
}

///////////////////////////////////////////////////////////////////////////////////////////////
var serverName string
var serverVersion string
var major, minor, revision int

var daemon *bool = flag.Bool("d", false, "run as daemon mode.")
var chelp *bool = flag.Bool("help", false, "show this message.")
var cfile *string = flag.String("f", "", "load configuire file.")
var upgrade *bool = flag.Bool("u", false, "hot upgrade(kill -HUP PID).")
var cpid *string = flag.String("p", "your.pid", "pid file(/tmp/your.pid).")
var cid *uint = flag.Uint("cid", 0, "component id")

func SetServerName(name string) {
	serverName = name
}

func ServerName() string {
	return serverName
}

func ServerNamePtr() *string {
	return ProtoString(serverName)
}

func Configuire() string {

	if nil == cfile || "" == *cfile {
		return Sprintf("%s.json", AppName())
	}

	return *cfile
}

func PidFile() string {

	if nil == cpid || "your.pid" == *cpid {
		return Sprintf("/tmp/%s.pid", AppName())
	}

	return *cpid
}

func NeedToUpgrade() bool {
	return *upgrade
}

func NeedToDaemon() bool {
	return *daemon
}

func Cid() uint32 {
	if nil != cid {
		return uint32(*cid)
	}
	return 0
}

func init() {

	// parse command args.
	flag.Parse()

	// 帮助或者没有输入配置文件参数则打印帮助并退出
	if true == *chelp {
		flag.PrintDefaults()
		os.Exit(0)
	}

	// 需要以 daemon 模式运行的
	if true == *daemon {
		if err := Daemon(); err != nil {
			FATAL("run as daemon mode to fail. %v", err)
		}
	}

	LOG_DEBUG("\n\n***************************** %s %s ******************************", AppName(), time.Now().String())

	if data, err := ioutil.ReadFile(AppPath() + "/" + "VERSION"); err != nil {
		LOG_FATAL("读取版本文件失败: %s", err.Error())
	} else {
		serverVersion = ReadLineString(data)

		vers := strings.Split(serverVersion, ".")

		if len(vers) != 3 {
			LOG_FATAL("版本号格式错误: %s", serverVersion)
		}

		major = Atoi(vers[0])
		minor = Atoi(vers[1])
		revision = Atoi(vers[2])
		/*
			INFO( "currently version: Major: %d, Minor: %d, Revision: %d",
			major,
			minor,
			revision )
		*/
	}
}
