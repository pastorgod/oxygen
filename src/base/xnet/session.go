package xnet

import (
	"time"
)

type ResponseHandler func(*string, Message)
type responseMap map[uint32]ResponseHandler

const MAX_REQUEST_NUM = 1<<32 - 1
const REQUEST_TIMEOUT = time.Second * 15
const RPC_TIMEOUT = 30

type ISession interface {
	//public:
	Addr() string
	Attach(IPacketHandler)
	SetLimitRatio(int32)
	SetLifetime(int)
	SendPacket(*Packet) bool
	SendBytes([]byte) bool
	SendCmd(Message) bool
	Request(Message, ResponseHandler) bool
	Response(*string, Message, uint32) bool
	Process(*Packet)
	Close(error)
	Reconnect() bool
	Alive() bool
	IsSlowing() bool

	Call(string, Message) (Message, *string)
	AsyncCall(string, Message, ResponseHandler)
	CallTimeout(uint32, int, Message) (Message, *string)
	AsyncCallTimeout(uint32, int, Message, ResponseHandler)

	SetDispatcher(*RpcDispatcher)
	Dispatcher() *RpcDispatcher

	//virtual:
	OnPacket(ISession, *Packet)
	OnRequest(ISession, *Packet) bool
	OnClosed(ISession, error)

	//protected:
	requestId() uint32
	putRequest(uint32, time.Duration, ResponseHandler)
	popRequest(uint32) ResponseHandler
	onResponse(uint32, *string, Message) bool
}
