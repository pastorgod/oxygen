package xnet

import (
	"errors"
	. "logger"
)

var (
	ConnectionReset   = errors.New("ConnectionReset")
	PacketTooLarge    = errors.New("PacketTooLarge")
	InvalidMessage    = errors.New("InvalidMessage")
	ChecksumFail      = errors.New("ChecksumFail")
	KeepAliveTimeout  = errors.New("KeepAliveTimeout")
	InvalidSequence   = errors.New("InvalidSequence")
	ReceiveQuitSignal = errors.New("ReceiveQuitSignal")
	ServerUpgrade     = errors.New("ServerUpgrade")
	ServiceOnDestory  = errors.New("ServiceOnDestory")
	ResendTimeouted   = errors.New("ResendTimeouted")
	InvalidUDPSession = errors.New("InvalidUDPSession")
	HandshakeFailed   = errors.New("HandshakeFailed")
	ClientEOF         = errors.New("ClientEOF")
	RepeatSequence    = errors.New("RepeatSequence")
	ReconnectedClose  = errors.New("ReconnectedClose")

	RequestTimeout = "RequestTimeout"
	NetworkError   = "NetworkError"
)

// 检查mask是否存在
func CheckMask(head, mask uint32) bool {
	return 0 != ((head & HEADER_MASK) & mask)
}

// 头部长度
func HeadSize(head uint32) (length uint16) {

	length = 4

	if CheckMask(head, HEAD_CHK_MASK) {
		length += 4
	}

	if CheckMask(head, HEAD_SEQ_MASK) {
		length += 4
	}

	if CheckMask(head, HEAD_REQ_MASK) {
		length += 4
	}

	if CheckMask(head, HEAD_RPC_MASK) {
		length += 4
	}

	if CheckMask(head, HEAD_MSG_MASK) {
		length += 4
	}

	return length
}

// 消息长度
func MsgSize(head uint32) uint16 {
	return uint16(head & LENGTH_MASK)
}

// 整个包长度
func PacketSize(head uint32) uint16 {
	return HeadSize(head) + uint16(head&LENGTH_MASK)
}

func buildNormalPacket(pb Message, requestId uint32) *Packet {
	return buildPacket(nil, pb, 0, requestId, 0)
}

func buildRpcPacket(pb Message, requestId, rpcId uint32) *Packet {
	return buildPacket(nil, pb, 0, requestId, rpcId)
}

// 构建一个包裹数据
func buildPacket(packet *Packet, pb Message, sequence, requestId, rpcId uint32) *Packet {

	checksum, crypto := false, false

	if nil == packet {
		packet = &Packet{}
	}

	if pb != nil && !pb.IsNil() {

		// 消息大小
		pb_size := uint16(pb.Size())

		if pb_size > MAX_MSG_LEN {
			LOG_ERROR("message to large: %s %d", ToName(pb), pb_size)
			return nil
		}

		// 消息
		packet.head |= uint32(pb_size)

		if code, ok := FindMsgCodeByObject(pb); ok {
			packet.opcode = code
			packet.Msg = pb
			packet.head |= HEAD_MSG_MASK
		} else {
			LOG_ERROR("没有找到这个消息的消息号: %+v", pb)
			return nil
		}

		// 请求
		if 0 != requestId {
			packet.requestId = requestId
			packet.head |= HEAD_REQ_MASK
		}

		// RPC id
		if 0 != rpcId {
			packet.rpcId = rpcId
			packet.head |= HEAD_RPC_MASK
		}

		// 校验选项
		if checksum {
			packet.head |= HEAD_CHK_MASK
		}

		// 加密选项
		if crypto {
			packet.head |= HEAD_CRY_MASK
		}
	}

	// 序列
	if 0 != sequence {
		packet.sequence = sequence
		packet.head |= HEAD_SEQ_MASK
	}

	return packet
}

func buildMessage(cmd Message, sequence, requestId uint32) []byte {

	if packet := buildPacket(nil, cmd, sequence, requestId, 0); packet != nil {
		buf := make([]byte, packet.AllSize())

		if err := packet.MarshalTo(buf); err != nil {
			LOG_ERROR("序列化数据错误: %s", err.Error())
			return nil
		}

		return buf
	}

	LOG_ERROR("buildMessage fail: %s", ToName(cmd))
	return nil
}

func BuildCmdToBytes(cmd Message) []byte {
	return buildMessage(cmd, 0, 0)
}
