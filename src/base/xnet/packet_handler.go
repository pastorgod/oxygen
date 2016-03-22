package xnet

import (
	. "logger"
	"sync"
)

// session, responseId, cmd
type MessageHandler func(ISession, uint32, Message)

// 网络断开
type ICloseNotifier interface {

	// on closed.
	onClosed(error)
}

// 网络重连成功
type IConnNotifier interface {

	// on connected.
	OnConnected()
}

// 网络包处理
type IPacketHandler interface {

	// 第一步处理
	OnPacket(ISession, *Packet)

	// 分离请求
	OnRequest(ISession, *Packet) bool

	// 网络断开
	OnClosed(ISession, error)
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// 消息默认处理器

var defaultPacketHandler = &DefaultPacketHandler{handlers: make(map[uint32]MessageHandler, 64), mutex: new(sync.RWMutex)}

func GetDefaultHandler() *DefaultPacketHandler {
	return defaultPacketHandler
}

type DefaultPacketHandler struct {
	handlers map[uint32]MessageHandler
	mutex    *sync.RWMutex
}

func (this *DefaultPacketHandler) register(name string, handler MessageHandler) {

	this.mutex.Lock()
	defer this.mutex.Unlock()

	if code, ok := FindMsgCodeByName(name); ok {
		if _, ok := this.handlers[code]; ok {
			FATAL("repeat to registe: %s : %d", name, code)
		}
		this.handlers[code] = handler
		return
	}

	FATAL("registe command hander to failed.(invalid command): %s", name)
}

// 默认收到网络包直接处理这个包
func (this *DefaultPacketHandler) OnPacket(session ISession, packet *Packet) {
	session.Process(packet)
}

// 处理请求消息
func (this *DefaultPacketHandler) OnRequest(session ISession, packet *Packet) bool {

	this.mutex.RLock()
	handler, ok := this.handlers[packet.MsgType()]
	this.mutex.RUnlock()

	// 处理绑定消息
	if ok && handler != nil {
		handler(session, packet.RequestId(), packet.Msg)
		return true
	}

	LOG_WARN("DefaultPacketHandler.OnRequest: %s %s", session.Addr(), packet.Name())
	return false
}

// 网络断开
func (this *DefaultPacketHandler) OnClosed(session ISession, err error) {
	DEBUG("连接断开: %s, %v", session.Addr(), err)
}

//--------------------------------------------------------------------------------------------------------------
// 注册默认处理函数
func Register(name string, handler MessageHandler) {
	defaultPacketHandler.register(name, handler)
}

// 注册线程安全的逻辑消息处理
func RegisterLogic(name string, handler MessageHandler) {
	defaultPacketHandler.register(name, func(conn ISession, requestId uint32, msg Message) {
		PushTask(func() {
			handler(conn, requestId, msg)
		})
	})
}
