package base

/*
import(
	"db"
	"os"
	"time"
	"command"
	"base/xnet"
.	"logger"
)

var ServiceUnavailable = ProtoString( "ServiceUnavailable" )

type IAppHandler interface {

	// on accepted session.
	OnAccept( xnet.ISession ) bool
}

type ServerApp struct {
	key		uint32		// uint6(ServerType) | uint10(zone_id) | uint16(number)
	gpmd	*GPMDClient	// gpmd client.
	conf	*ConfigFile	// config.
	zprefix	string		// zone prefix.
}

func NewServerApp( t command.ServerType, numbers... uint32 ) *ServerApp {

	Assert( 0 == len(numbers) || 0 != numbers[0], "server number error" )

	number := topServerNumber

	if len(numbers) > 0 {
		number = numbers[0]
	}

	return &ServerApp { key  : FetchServerKey( t, 0, number ) }
}

func(this *ServerApp) SvrKey() uint32 {
	return this.key
}

func(this *ServerApp) SvrType() command.ServerType {
	return FetchServerType( this.SvrKey() )
}

func(this *ServerApp) SvrZone() uint32 {
	return FetchServerZone( this.SvrKey() )
}

func(this *ServerApp) SvrNumber() uint32 {
	return FetchServerNumber( this.SvrKey() )
}

func(this *ServerApp) SetSvrNumber( number uint32 ) {
	Assert( nil == this.gpmd, "use error." )
	this.key = FetchServerKey( this.SvrType(), this.SvrZone(), number )
}

func(this *ServerApp) SetSvrZone( zone uint32 ) {
	Assert( nil == this.gpmd, "use error." )
	this.key = FetchServerKey( this.SvrType(), zone, this.SvrNumber() )
}

func(this *ServerApp) String() string {
	return FetchServerName( this.SvrKey() )
}

func(this *ServerApp) ZoneName() string {
	return Sprintf( "%s%d", this.zprefix, this.SvrZone() )
}

func(this *ServerApp) Register( name string, fn interface{} ) {
	this.gpmd.Register( name, fn )
}

func(this *ServerApp) RegisterService( service interface{}, alias...string ) {
	this.gpmd.RegisterService( service, alias... )
}

func(this *ServerApp) Config( reload...bool ) (c *ConfigFile, err error) {

	// 不重载配置文件则优先加载缓存
	if 0 == len(reload) || false == reload[0] {
		if this.conf != nil {
			return this.conf, nil
		}
	}

	// 加载配置文件
	if this.conf, err = LoadConfig(); err != nil {
		ERROR( "load [%s] to fail. %s", Configuire(), err.Error() )
		return nil, err
	}

	return this.conf, nil
}

func(this *ServerApp) GetHarborPort( base_port uint32 ) uint32 {
	return base_port + base_port / 10 * this.SvrZone() + base_port / 100 * uint32(this.SvrType()) + this.SvrNumber()
}

func(this *ServerApp) GetString( section, option string ) (string, error) {
	conf, err := this.Config()
	if err != nil {
		return "", err
	}
	return conf.GetString( section, option )
}

func(this *ServerApp) GetOptString( section, option, def string ) string {
	if value, err := this.GetString( section, option ); nil == err {
		return value
	}
	return def
}

func(this *ServerApp) GetInt64( section, option string ) (int64, error) {
	conf, err := this.Config()
	if err != nil {
		return 0, err
	}
	return conf.GetInt64( section, option )
}

func(this *ServerApp) GetOptInt64( section, option string, def int64 ) int64 {
	if value, err := this.GetInt64( section, option ); nil == err {
		return value
	}
	return def
}

func(this *ServerApp) GetInt32( section, option string ) (int32, error) {
	value, err := this.GetInt64( section, option )
	if err != nil {
		return 0, err
	}
	return int32(value), nil
}

func(this *ServerApp) GetOptInt32( section, option string, def int32 ) int32 {
	if value, err := this.GetInt32( section, option ); nil == err {
		return value
	}
	return def
}

func(this *ServerApp) GetUInt32( section, option string ) (uint32, error) {
	value, err := this.GetInt64( section, option )
	if err != nil {
		return 0, err
	}
	return uint32(value), nil
}

func(this *ServerApp) GetOptUInt32( section, option string, def uint32 ) uint32 {
	if value, err := this.GetUInt32( section, option ); nil == err {
		return value
	}
	return def
}

func(this *ServerApp) GetBool( section, option string )(bool, error) {
	conf, err := this.Config()
	if err != nil {
		return false, err
	}
	return conf.GetBool( section, option )
}

func(this *ServerApp) GetOptBool( section, option string, def bool ) bool {
	if value, err := this.GetBool( section, option ); nil == err {
		return value
	}
	return def
}

func(this *ServerApp) GetFloat( section, option string )(float64, error) {
	conf, err := this.Config()
	if err != nil {
		return 0, err
	}
	return conf.GetFloat( section, option )
}

func(this *ServerApp) GetOptFloat( section, option string, def float64 ) float64 {
	if value, err := this.GetFloat( section, option ); nil == err {
		return value
	}
	return def
}

func(this *ServerApp) Initialize( handler IAppHandler ) bool {

	// 多服结构需要检查检查cid
	switch this.SvrType() {
	case command.ServerType_LoginApp, command.ServerType_BaseApp, command.ServerType_CellApp:
		if Cid() <=0 {
			ERROR( "missing command param (-cid)" )
			return false
		}

		// 设置服务器编号
		this.SetSvrNumber( Cid() )
	}

	// 日志标示
	SetLoggerPrefix( ToUpper( this.SvrType().String() ) )

	// 初始化 gpmd
	if err := this.initGPMD(); err != nil {
		ERROR( "init gpmd failed: %v", err )
		return false
	}

	INFO( "******[ Server: %s, Zone: %d, Number: %d ]******", this.String(), this.SvrZone(), this.SvrNumber() )

	go this.acceptLoop( handler.OnAccept )

//	return this.InitLogger()
	return true
}

func(this *ServerApp) OnAccept(xnet.ISession) bool {
	return true
}

func(this *ServerApp) acceptLoop( handler func(xnet.ISession) bool ) {
	this.gpmd.AcceptLoop( handler )
}

func(this *ServerApp) initZone() (err error) {

	// 读取大区ID
	zoneid, err := this.GetInt64( DefaultSection, "zone" )

	if err != nil {
		return err
	}

	if zoneid <= 0 || zoneid > 1024 {
		return ToError( "zone => 1 - 1024" )
	}

	// 读取大区名
	this.zprefix, err = this.GetString( DefaultSection, "prefix" )

	if err != nil {
		return err
	}

	// 初始化大区ID
	this.SetSvrZone( uint32(zoneid) )

	return nil
}

func(this *ServerApp) initGPMD() (err error) {

	if err = this.initZone(); err != nil {
		ERROR( "初始化大区信息失败! %s", err.Error() )
		return err
	}

	// gpmd server addr
	gpmd_server_addr, err := this.GetString( "gpmd", "addr" )

	if err != nil {
		ERROR( "missing option 'gpmd/addr'" )
		return err
	}

	// 优先使用本级别配置端口
	local_port, err := this.GetUInt32( this.String(), "port" )

	// 如果没有配置则自动生成
	if nil != err {
		// 读取本地harbor端口
		base_port, err := this.GetUInt32( DefaultSection, "base_port" )

		if err != nil || base_port < 1000 {
			ERROR( "missing option 'default/base_port'" )
			return err
		}

		// 自动计算端口号
		local_port = this.GetHarborPort( base_port )
	}

	// local_port + zone * type
	local_addr := Sprintf( ":%d", local_port )

	if host, err := this.GetString( this.String(), "host" ); nil == err {
		local_addr = Sprintf( "%s:%d", host, local_port )
	}

	INFO( "gpmd: %s, local: %s", gpmd_server_addr, local_addr )

	if this.gpmd, err = NewGPMDClient( this.SvrKey(), gpmd_server_addr, local_addr ); err != nil {
		return err
	}

	return nil
}

func(this *ServerApp) InitMongodb( key string, tables...db.IRecord ) bool {

	mgodb, err := this.GetString( this.String(), key )

	if err != nil {
		ERROR( "InitMongodb: %s %s", key, err.Error() )
		return false
	}

	return db.InitializeMongodb( mgodb, tables... )
}

func(this *ServerApp) InitRedis( key string ) bool {

	redis, err := this.GetString( this.String(), key )

	if err != nil {
		ERROR( "InitRedis: %s %s", key, err.Error() )
		return false
	}

	return db.InitializeRedis( redis )
}

func(this *ServerApp) InitLogger() bool {

	// 先读取默认配置
	log_path, _ := this.GetString( DefaultSection, "log_path" )

	// 优先使用本级别配置
	if value, err := this.GetString( this.String(), "log_path" ); nil == err {
		log_path = value
	}

	if "" == log_path {
		ERROR( "missing option 'log_path'" )
		return false
	}

	// /data/logs/LOGINAPP/loginapp-1.log
	// /data/logs/TENCENT1/baseapp-1.log

	// 先读取默认配置
	log_level := this.GetOptString( DefaultSection, "log_level", "DEBUG" )

	// 优先使用本级别配置
	if value, err := this.GetString( this.String(), "log_level" ); nil == err {
		log_level = value
	}

	// 文件名使用app的名字 或者使用主程序的名字
	log_file := ToLower( this.String() ) + ".log"

	// 默认使用类型名
	log_name := ToUpper( this.String() )

	// 优先使用自定义名
	if value, err := this.GetString( this.String(), "log_name" ); nil == err {
		log_name = value
	}

	INFO( "初始化日志: %s/%s, %s, %s", log_path, log_file, log_name, log_level )

	// 初始化日志
	InitializeLogger( log_path, log_file, log_name, log_level )

	// 重定向 stdout | stderr 到文件
	RedirectErrorToFile()

	return true
}

func(this *ServerApp) Close( err error ) {
	if this.gpmd != nil {
		this.gpmd.Close( err )
	}
}

func(this *ServerApp) GetComponentBy( key uint32 ) (component *Component) {

	if component = ComponentManagerInstance.Find( key ); nil == component {
		ERROR( "ServerApp: %s, NotFound Component: %s", this.String(), FetchServerName( key ) )
	}

	return
}

func(this *ServerApp) GetComponent( t command.ServerType, zone uint32, numbers...uint32 ) *Component {

	number := topServerNumber

	if len(numbers) > 0 && numbers[0] > 0 {
		number = numbers[0]
	}

	return this.GetComponentBy( FetchServerKey( t, zone, number ) )
}

func(this *ServerApp) Component( t command.ServerType, numbers...uint32 ) *Component {

	number := topServerNumber

	if len(numbers) > 0 && numbers[0] > 0 {
		number = numbers[0]
	}

	return this.GetComponentBy( FetchServerKey( t, this.SvrZone(), number ) )
}

// 默认ITick实现
func(this *ServerApp) TickDelay() time.Duration {
	return time.Second
}

// 默认逻辑帧实现
func(this *ServerApp) OnTick() {
}

// 启动服务
func(this *ServerApp) OnInitialize() bool {
	return true
}

// 停止服务器
func(this *ServerApp) Stop( reason error, force bool ) {
	this.Close( reason )
}

// 默认退出信号实现
func(this *ServerApp) OnSigQuit() xnet.SignalResult {
	this.Stop( xnet.ReceiveQuitSignal, true )
	return xnet.SignalResult_Quit
}

// 默认升级信号实现
func(this *ServerApp) OnSigHup() xnet.SignalResult {
	LOG_WARN( "ServerApp.OnSigHup: SignalResult_Continue %s", this.String() )
	return xnet.SignalResult_Continue
}

// 默认其他类型信号处理
func(this *ServerApp) OnSignal( sig os.Signal ) xnet.SignalResult {
	LOG_WARN( "ServerApp.OnSignal: %s %d", this.String(), sig )
	return xnet.SignalResult_Continue
}

// 默认销毁实现
func(this *ServerApp) OnDestroy() {
	LOG_INFO( "ServerApp.OnDestroy... %s", this.String() )
	this.Stop( xnet.ServiceOnDestory, true )
}

// 登录key
// LOGIN:TENCENT1:PASSPORT
func(this *ServerApp) LoginKey( accid string ) string {
	return Sprintf( "LOGIN:%s:%s", this.ZoneName(), accid )
}

// 重新登录Key
// RELOGIN:TENCENT1:PASSPORT
func(this *ServerApp) ReLoginKey( accid string ) string {
	return Sprintf( "RELOGIN:%s:%s", this.ZoneName(), accid )
}
*/
