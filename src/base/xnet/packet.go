package xnet

import (
	"hash/adler32"
	. "logger"
)

const (
	HEAD_CRY_MASK uint32 = 0x80000000 // 加密选项
	HEAD_CMP_MASK uint32 = 0x40000000 // 压缩选项
	HEAD_CHK_MASK uint32 = 0x20000000 // 校验
	HEAD_SEQ_MASK uint32 = 0x10000000 // 序列ID
	HEAD_REQ_MASK uint32 = 0x08000000 // 请求ID
	HEAD_MSG_MASK uint32 = 0x04000000 // 消息类型
	HEAD_RPC_MASK uint32 = 0x02000000 // RPC调用选项

	HEADER_MASK uint32 = 0xFFFF0000
	LENGTH_MASK uint32 = 0x0000FFFF

	HEAD_SIZE uint32 = 4

	MAX_MSG_LEN uint16 = 1<<16 - 20
	MAX_PKT_LEN uint16 = 1<<16 - 1

	// keep-alive
	SEQUENCE_KEEPALIVE     = 1
	SEQUENCE_KEEPALIVE_ACK = 2

	// handshake
	SEQUENCE_HANDSHAKE     = 3
	SEQUENCE_HANDSHAKE_ACK = 4

	SEQUENCE_KICK    = 5
	SEQUENCE_CONNECT = 6

	SEQUENCE_INTERNAL = 100
)

type Packet struct {
	head      uint32 // uint16(head) + uint16(len)
	checksum  uint32 // adler32
	sequence  uint32 // sequence
	requestId uint32 // request id
	rpcId     uint32 // rpc id => rpc router
	opcode    uint32 // operate code
	Msg       Message
}

func (this *Packet) setMask(mask uint32) {
	this.head |= (mask & HEADER_MASK)
}

func (this *Packet) CheckMask(mask uint32) bool {
	return CheckMask(this.head, mask)
}

func (this *Packet) IsRpc() bool {
	return this.CheckMask(HEAD_RPC_MASK)
}

// 消息名
func (this *Packet) Name() string {
	if this.CheckMask(HEAD_MSG_MASK) {
		name, _ := FindMsgNameByCode(this.MsgType())
		return name
	}

	if this.Msg != nil {
		return ToName(this.Msg)
	}

	return "<nil-packet>"
}

// 序列
func (this *Packet) Sequence() uint32 {
	if this.CheckMask(HEAD_SEQ_MASK) {
		return this.sequence
	}
	return 0
}

// 请求号
func (this *Packet) RequestId() uint32 {
	if this.CheckMask(HEAD_REQ_MASK) {
		return this.requestId
	}
	return 0
}

// Rpc Id
func (this *Packet) RpcId() uint32 {
	return this.rpcId
}

// 消息号
func (this *Packet) MsgType() uint32 {

	if this.CheckMask(HEAD_MSG_MASK) {
		return this.opcode
	}

	return 0
}

// 消息长度
func (this *Packet) MsgSize() uint16 {

	if this.CheckMask(HEAD_MSG_MASK) {
		return uint16(this.head & LENGTH_MASK)
	}

	if this.Msg != nil {
		return uint16(this.Msg.Size())
	}

	return 0
}

// 整个包长度( 头部 + 消息 )
func (this *Packet) AllSize() uint16 {
	return PacketSize(this.head)
}

// 输出packet结构
func (this *Packet) String() string {

	return Sprintf(
		`Packet of %d bytes {
		CRYPTO		: %v,
		COMPRESS	: %v,
		CHECKSUM	: %v(%d),
		SEQUENCE	: %v(%d),
		REQUESTID	: %v(%d),
		RPC_CALL	: %v(%d),
		MSG_TYPE	: %d,
		MSG_LEN		: %d byte,
		MSG_NAME	: %s,
		MSG		: %+v,
	}`,
		this.AllSize(),
		this.CheckMask(HEAD_CRY_MASK),
		this.CheckMask(HEAD_CMP_MASK),
		this.CheckMask(HEAD_CHK_MASK),
		this.checksum,
		this.CheckMask(HEAD_SEQ_MASK),
		this.sequence,
		this.CheckMask(HEAD_REQ_MASK),
		this.requestId,
		this.CheckMask(HEAD_RPC_MASK),
		this.rpcId,
		this.MsgType(),
		this.MsgSize(),
		this.Name(),
		this.Msg)
}

func (this *Packet) Reset() {
	this.head = 0
	this.checksum = 0
	this.sequence = 0
	this.requestId = 0
	this.rpcId = 0
	this.opcode = 0
	this.Msg = nil
}

func (this *Packet) Marshal() ([]byte, error) {

	buf := make([]byte, this.AllSize())

	if err := this.MarshalTo(buf); err != nil {
		return nil, err
	}

	return buf, nil
}

// 序列化数据
func (this *Packet) MarshalTo(buf []byte) error {

	if len(buf) != int(this.AllSize()) {
		panic("invalid buf.")
	}

	// required headMask
	LittleEndian.WriteUInt32(buf[0:], this.head)

	offset, checksum_offset := int(HEAD_SIZE), 0

	// optional checksum
	if this.CheckMask(HEAD_CHK_MASK) && len(buf) > int(HEAD_SIZE) {
		offset += LittleEndian.WriteUInt32(buf[offset:], 0)

		checksum_offset = offset
	}

	// optional sequence
	if this.CheckMask(HEAD_SEQ_MASK) {
		offset += LittleEndian.WriteUInt32(buf[offset:], this.sequence)
	}

	// optional requestId
	if this.CheckMask(HEAD_REQ_MASK) {
		offset += LittleEndian.WriteUInt32(buf[offset:], this.requestId)
	}

	// optional rpc id
	if this.CheckMask(HEAD_RPC_MASK) {
		offset += LittleEndian.WriteUInt32(buf[offset:], this.rpcId)
	}

	// optional msgMask
	if this.CheckMask(HEAD_MSG_MASK) {
		offset += LittleEndian.WriteUInt32(buf[offset:], this.opcode)

		// append message body
		if int(this.MsgSize()) != len(buf[offset:]) {
			LOG_FATAL("MsgSize != buf size, %d != %d", this.MsgSize(), len(buf[offset:]))
		}

		if _, err := this.Msg.MarshalTo(buf[offset:]); err != nil {
			return err
		}
	}

	// compute checksum
	if this.CheckMask(HEAD_CHK_MASK) && checksum_offset != 0 {
		checksum := adler32.Checksum(buf[checksum_offset:])
		// write checksum valu
		LittleEndian.WriteUInt32(buf[HEAD_SIZE:], checksum)
	}

	// crypto
	if this.CheckMask(HEAD_CRY_MASK) && len(buf) > int(HEAD_SIZE) {
	}

	return nil
}

// 反序列化数据
func (this *Packet) Unmarshal(buf []byte) error {

	// uint32 head mask.
	this.head = LittleEndian.ReadUInt32(buf)

	if len(buf) < int(this.AllSize()) {
		LOG_ERROR("Packet.Unmarshal: %v", buf)
		PrintStack("Packet.Unmarshal error, invalid packet.")
		return InvalidMessage
	}

	// uncrypto
	if this.CheckMask(HEAD_CRY_MASK) {
	}

	offset := int(HEAD_SIZE)

	// optional checksum
	if this.CheckMask(HEAD_CHK_MASK) {
		this.checksum = LittleEndian.ReadUInt32(buf[offset:])
		offset += 4

		checksum := adler32.Checksum(buf[offset:])

		if this.checksum != checksum {
			ERROR("checksum fail. %d %d, [%d, %d]", this.checksum, checksum, this.AllSize(), len(buf))
			return ChecksumFail
		}
	}

	// optional sequence
	if this.CheckMask(HEAD_SEQ_MASK) {
		this.sequence = LittleEndian.ReadUInt32(buf[offset:])
		offset += 4
	}

	// optional requestId
	if this.CheckMask(HEAD_REQ_MASK) {
		this.requestId = LittleEndian.ReadUInt32(buf[offset:])
		offset += 4
	}

	// optional rpc id
	if this.CheckMask(HEAD_RPC_MASK) {
		this.rpcId = LittleEndian.ReadUInt32(buf[offset:])
		offset += 4
	}

	// optional opcode
	if this.CheckMask(HEAD_MSG_MASK) {
		this.opcode = LittleEndian.ReadUInt32(buf[offset:])
		offset += 4

		// new pb object by msg code.
		var ok bool = false
		if this.Msg, ok = NewMsgObjectByCode(this.MsgType()); ok {
			// unmarshal message form buf.
			if err := this.Msg.Unmarshal(buf[offset:]); err != nil {
				return err
			}
		} else {
			LOG_ERROR("NotFound Message code: %d", this.MsgType())
			this.Msg = nil
			this.head &= ^HEAD_MSG_MASK
		}
	}

	return nil
}
