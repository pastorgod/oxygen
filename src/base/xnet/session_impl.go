package xnet

import (
	"base/timer"
	. "logger"
	"net"
	"sync"
	"time"
)

type session_impl struct {
	*IdAlloctor
	conn    *connection
	rsps    responseMap
	handler IPacketHandler
	mutex   *sync.Mutex
}

func new_session_impl(conn net.Conn) *session_impl {

	session := &session_impl{
		IdAlloctor: NewIdAlloctor(MAX_REQUEST_NUM, 0),
		rsps:       make(responseMap, 16),
		handler:    defaultPacketHandler,
		mutex:      new(sync.Mutex),
	}

	// 挂接一个连接
	session.conn = new_connection(session, conn)

	return session
}

// 地址
func (this *session_impl) Addr() string {
	return this.conn.Addr()
}

func (this *session_impl) RawAddr() string {
	return this.conn.RawAddr()
}

// 切换消息处理者
func (this *session_impl) Attach(handler IPacketHandler) {

	if handler != nil {
		this.handler = handler
	} else {
		this.handler = defaultPacketHandler
	}
}

// 是否还活着
func (this *session_impl) Alive() bool {
	return this.conn.Alive()
}

// 判断是否慢速网络
func (this *session_impl) IsSlowing() bool {
	return this.conn.IsSlowing()
}

// 请求消息号生成
func (this *session_impl) requestId() uint32 {
	return this.Get()
}

// 设置每秒收包速度
func (this *session_impl) SetLimitRatio(num int32) {
	this.conn.SetLimitRatio(num)
}

// 设置超时(超时之后强制关闭连接, 单位: 秒, 0 => 取消设置)
func (this *session_impl) SetLifetime(sec int) {
	this.conn.SetReadTimeout(sec)
}

// 发送消息到对端
func (this *session_impl) SendCmd(cmd Message) bool {

	if bytes := buildMessage(cmd, 0, 0); bytes != nil {
		return this.SendBytes(bytes)
	}

	return false
}

// 发送包数据
func (this *session_impl) SendPacket(packet *Packet) bool {
	return this.conn.SendPacket(packet)
}

// 发送二进制数据
func (this *session_impl) SendBytes(buf []byte) bool {
	return this.conn.SendBytes(buf)
}

// 发送请求 & 回调handler
func (this *session_impl) Request(cmd Message, handler ResponseHandler) bool {

	request_id := this.requestId()

	if packet := buildNormalPacket(cmd, request_id); packet != nil {
		this.putRequest(request_id, REQUEST_TIMEOUT, handler)
		this.SendPacket(packet)
		return true
	}

	return false
}

// 回复请求, 包装消息
func (this *session_impl) Response(err *string, cmd Message, response_id uint32) bool {

	if response := buildResponse(err, cmd, response_id); response != nil {
		return this.SendCmd(response)
	}

	return false
}

func (this *session_impl) putRequest(request_id uint32, timeout time.Duration, handler ResponseHandler) {

	if nil == handler {
		return
	}

	// locl response map.
	this.mutex.Lock()

	// add request-handler pair.
	this.rsps[request_id] = handler

	// unlock response map.
	this.mutex.Unlock()

	if timeout > 0 {
		timer.AfterFunc(timeout, func() {

			if handle := this.popRequest(request_id); handle != nil {
				handle(&RequestTimeout, nil)
			}
		})
	}
}

func (this *session_impl) popRequest(responseId uint32) (handler ResponseHandler) {

	// lock response map.
	this.mutex.Lock()
	var ok bool

	// pop handler.
	if handler, ok = this.rsps[responseId]; ok {
		delete(this.rsps, responseId)
	}

	// unlock.
	this.mutex.Unlock()
	return handler
}

// 收到回复消息
func (this *session_impl) onResponse(responseId uint32, err *string, msg Message) bool {

	if handler := this.popRequest(responseId); handler != nil {
		handler(err, msg)
		return true
	}

	return false
}

// 处理发送数据
func (this *session_impl) sendLoop() {

	this.conn.sendLoop()
}

// 处理接收数据 & 处理包
func (this *session_impl) recvLoop() {

	// 接收 & 处理消息
	this.conn.recvLoop(func(packet *Packet) {

		// 投递消息给上层处理
		this.handler.OnPacket(this, packet)
	})
}

// 包的默认处理模式
func (this *session_impl) OnPacket(session ISession, packet *Packet) {
	//TODO: FastReject
	// 快速拒绝掉超过服务自身处理能力的请求，即使在过载时，也能稳定地提供有效输出
	session.Process(packet)
}

// 默认请求处理
func (this *session_impl) OnRequest(session ISession, packet *Packet) bool {
	WARN("丢弃的数据包: %s %s", session.Addr(), packet.Name())
	return false
}

// 默认连接关闭时处理
func (this *session_impl) OnClosed(session ISession, err error) {
	DEBUG("session_impl.OnClosed: %s %v", session.Addr(), err)
}

// 处理消息包
func (this *session_impl) Process(packet *Packet) {

	//recorder := DefaultRequestRecorder.Record(packet.MsgType())

	defer func() {
		if err := recover(); err != nil {
			PrintStack("session_impl.Process Exception: %s %v, %s,\n %s => %+v",
				packet.Name(), err, packet.String(), ToName(this.handler), this.handler)
		}

		// 结束记录
		//recorder()
	}()

	// 先处理回应消息
	if packet.MsgType() == RESPONSE_CODE {

		// 从回应中解析出数据
		respId, err, message := parseResponse(packet.Msg)

		// 处理回应
		this.onResponse(respId, err, message)

		return
	}

	// 处理请求消息
	if !this.handler.OnRequest(this, packet) {
		ERROR("未处理的消息: %s %d %s", this.Addr(), packet.MsgType(), packet.Name())
	}
}

func (this *session_impl) clearRequests(err *string) {

	for responseId, handler := range this.rsps {
		delete(this.rsps, responseId)

		if nil != handler {
			handler(err, nil)
		}
	}
}

// 关闭连接
func (this *session_impl) Close(err error) {
	this.conn.Close(err)
}

// 连接被关闭
func (this *session_impl) onClosed(err error) {

	// 清理网络断开时还未完成的请求
	this.clearRequests(&NetworkError)
	// 通知上层断开了
	this.handler.OnClosed(this, err)
	// 重置处理器(解除循环引用)
	this.Attach(nil)
}

func (this *session_impl) Reconnect() bool {

	INFO("重连服务器中... %s", this.RawAddr())

	// 尝试重连服务器
	if conn := this.conn.reconnect(time.Second); conn != nil {
		// 连成功了, 挂接这个连接
		this.conn = conn

		// 之后开始读写操作
		go this.recvLoop()
		go this.sendLoop()

		INFO("重连服务器成功: %s", this.RawAddr())
		return true
	}

	WARN("重连服务器失败: %s", this.RawAddr())
	return false
}

func (this *session_impl) SetDispatcher(*RpcDispatcher) {
	FATAL("NotImplementedException: ISession.SetDispatcher")
}

func (this *session_impl) Dispatcher() *RpcDispatcher {
	FATAL("NotImplementedException: ISession.Dispatcher")
	return nil
}

func (this *session_impl) Call(method string, msg Message) (Message, *string) {
	FATAL("NotImplementedException: Call( %s, %s )", method, ToName(msg))
	return nil, nil
}

func (this *session_impl) CallTimeout(method_hash uint32, timeout int, msg Message) (Message, *string) {
	FATAL("NotImplementedException: CallTimeout(%d, %d, %s)", method_hash, timeout, ToName(msg))
	return nil, nil
}

func (this *session_impl) AsyncCall(method string, msg Message, handler ResponseHandler) {
	FATAL("NotImplementedException: AsyncCall( %s, %s )", method, ToName(msg))
}

func (this *session_impl) AsyncCallTimeout(method_hash uint32, timeout int, msg Message, handler ResponseHandler) {
	FATAL("NotImplementedException: AsyncCallTimeout( %d, %d, %s )", method_hash, timeout, ToName(msg))
}
