package xnet

import (
	. "logger"
)

var gs_command_factory ICommandFactory = nil
var RESPONSE_CODE uint32

func Init(factory ICommandFactory) {
	gs_command_factory = factory
	RESPONSE_CODE = gs_command_factory.GetResponseCode()
}

type ICommandFactory interface {

	// 根据消息号获得消息名字
	FindMsgNameByCode(uint32) (string, bool)

	// 根据消息名字获得消息号
	FindMsgCodeByName(string) (uint32, bool)

	// 根据消息号创建一个新对象
	NewMsgObjectByCode(uint32) (Message, bool)

	// 获取的回应包编码
	GetResponseCode() uint32

	// 构建回应包
	BuildResponse(*string, Message, uint32) Message

	// 解析回应数据
	ParseResponse(Message) (uint32, *string, Message)
}

// 根据消息号获得消息名字
func FindMsgNameByCode(opcode uint32) (string, bool) {
	return gs_command_factory.FindMsgNameByCode(opcode)
}

// 根据消息名字获得消息号
func FindMsgCodeByName(name string) (uint32, bool) {
	return gs_command_factory.FindMsgCodeByName(name)
}

// 根据消息对象获得消息号
func FindMsgCodeByObject(object Message) (uint32, bool) {
	// 通过反射获得消息名 然后通过消息名查找消息号
	return FindMsgCodeByName(ToName(object))
}

// 根据消息号创建一个新对象
func NewMsgObjectByCode(opcode uint32) (Message, bool) {

	object, ok := gs_command_factory.NewMsgObjectByCode(opcode)

	if !ok || nil == object {
		LOG_ERROR("没有找到消息号对应的消息! %d", opcode)
		ERROR("没有找到消息号对应的消息! %d", opcode)
		return nil, false
	}

	return object, ok
}

// 根据消息名字创建一个新对象
func NewMsgObjectByName(name string) (Message, bool) {
	// 先找消息号
	if code, ok := FindMsgCodeByName(name); ok {
		// 然后通过消息号创建对象
		return NewMsgObjectByCode(code)
	}
	return nil, false
}

// 包装回应
func buildResponse(errStr *string, pb Message, response_id uint32) Message {
	return gs_command_factory.BuildResponse(errStr, pb, response_id)
}

// 从回应消息里面解析实际消息体
func parseResponse(response Message) (uint32, *string, Message) {
	return gs_command_factory.ParseResponse(response)
}
