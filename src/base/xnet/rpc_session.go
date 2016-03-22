package xnet

import (
	. "logger"
	"time"
)

var (
	MethodNotFound = "MethodNotFound"
	RpcLocalError  = "RpcLocalError"
	RpcParamError  = "RpcParamError"
)

type AsynResult struct {
	err *string
	out Message
}

type RpcSession struct {
	ISession
	dispatcher *RpcDispatcher
}

func NewRpcSession(session ISession, dispatcher *RpcDispatcher) *RpcSession {

	if nil == dispatcher {
		dispatcher = NewRpcDispatcher()
	}

	rpc_session := &RpcSession{
		ISession:   session,
		dispatcher: dispatcher,
	}

	// 挂接RPC包处理器
	session.Attach(rpc_session)

	return rpc_session
}

func (this *RpcSession) SetDispatcher(dispatcher *RpcDispatcher) {
	Assert(nil != dispatcher, "dispatcher is nil.")
	this.dispatcher = dispatcher
}

func (this *RpcSession) Dispatcher() *RpcDispatcher {
	return this.dispatcher
}

// 默认超时 RPC同步调用
func (this *RpcSession) Call(method string, input Message) (output Message, err *string) {
	return this.CallTimeout(Hash(method), RPC_TIMEOUT, input)
}

// RPC 同步调用
// @param method 	远程方法名
// @param timeout	超时时间(秒)
// @param input		输入参数
// return 返回结果 & 错误信息
func (this *RpcSession) CallTimeout(method_hash uint32, timeout int, input Message) (output Message, err *string) {

	// 网络不可用
	if !this.Alive() {
		return nil, &NetworkError
	}

	requestId := this.requestId()

	// 构建一个rpc请求包
	packet := buildRpcPacket(input, requestId, method_hash)

	if nil == packet {
		return nil, &RpcLocalError
	}

	// 准备阻塞接受结果
	retChan := make(chan AsynResult, 1)

	// 等待对端返回并且设置一个超时
	this.putRequest(requestId, time.Second*time.Duration(timeout), func(err *string, out Message) {
		// 等待IO线程写入并激活chan
		// 超时处理线程激活
		retChan <- AsynResult{err: err, out: out}
	})

	// 发送RPC调用请求
	this.SendPacket(packet)

	// 等待调用返回或者超时
	ret := <-retChan

	// 关闭这个chan
	close(retChan)

	return ret.out, ret.err
}

// 默认超时 RPC异步调用
func (this *RpcSession) AsyncCall(method string, input Message, handler ResponseHandler) {
	this.AsyncCallTimeout(Hash(method), RPC_TIMEOUT, input, handler)
}

// RPC 异步调用
// @param method	远程方法名
// @param timeout	超时时间(秒)
// @param input		输入参数
// @param handler	收到回应时的回调函数
func (this *RpcSession) AsyncCallTimeout(method_hash uint32, timeout int, input Message, handler ResponseHandler) {

	if !this.Alive() {
		handler(&NetworkError, nil)
		return
	}

	requestId := this.requestId()

	packet := buildRpcPacket(input, requestId, method_hash)

	if nil == packet {
		handler(&RpcLocalError, nil)
		return
	}

	if handler != nil {
		this.putRequest(requestId, time.Second*time.Duration(timeout), func(err *string, out Message) {
			// run this handler on logic thread.
			PushTask(func() { handler(err, out) })
		})
	}

	this.SendPacket(packet)
}

// 收到请求
func (this *RpcSession) OnRequest(session ISession, packet *Packet) bool {

	// 不是RPC包那么转到默认处理器处理
	// 上层可以overwrite此函数做自己的处理
	if !packet.IsRpc() {
		return defaultPacketHandler.OnRequest(this, packet)
	}

	// 构建本次调用的上下文
	context := NewContext(this, packet)

	// 调用相应的RPC方法并得到结果
	rsp, err := this.dispatcher.ServiceCall(context, packet.RpcId(), packet.Msg)

	// 如果本次调用被转换为异步操作那么这里不回应
	// 仍然为同步操作那么直接返回结果信息
	if !context.isAsynced() {
		context.Response(err, rsp)
	}

	return true
}

func (this *RpcSession) OnClosed(session ISession, err error) {
	DEBUG("RpcSession.OnClosed: %s %v", session.Addr(), err)
}
