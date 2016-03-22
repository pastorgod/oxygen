package xnet

import "io"
import "bufio"
import "fmt"

type IPacketStream interface {
	// read packet from reader.
	ReadPacket(*Packet) error
	// write packet to writer.
	WritePacket(*Packet) error
	// write packet bytes.
	WriteBytes([]byte) error
	// write packet chan.
	WriteChan([]byte, <-chan []byte) error
}

type PacketStream struct {
	reader  *bufio.Reader
	writer  *bufio.Writer
	snd_buf []byte
	rcv_buf []byte
}

func NewPacketStream(rw io.ReadWriter, rcv_buf_size, snd_buf_size int) *PacketStream {
	Assert(rcv_buf_size >= 32, "rcv_buf_size > 32")
	Assert(snd_buf_size >= 32, "snd_buf_size > 32")

	return &PacketStream{
		reader:  bufio.NewReaderSize(rw, 1024*16),
		writer:  bufio.NewWriterSize(rw, 1024*16),
		snd_buf: make([]byte, snd_buf_size),
		rcv_buf: make([]byte, rcv_buf_size),
	}
}

func (this *PacketStream) readHead() (uint32, error) {
	if _, err := io.ReadFull(this.reader, this.rcv_buf[0:HEAD_SIZE]); err != nil {
		return 0, err
	}
	return LittleEndian.ReadUInt32(this.rcv_buf[0:HEAD_SIZE]), nil
}

func (this *PacketStream) ReadPacket(packet *Packet) error {

	head, err := this.readHead()

	if nil != err {
		return err
	}

	// 检查数据包是否有那么大
	size := PacketSize(head)

	if size >= MAX_PKT_LEN {
		return PacketTooLarge
	}

	// 如果缓存不足则重新分配新的缓存
	if int(size) > len(this.rcv_buf) {

		// 重新分配内存
		this.rcv_buf = make([]byte, size+32)

		// 写回head
		LittleEndian.WriteUInt32(this.rcv_buf[0:HEAD_SIZE], head)
	}

	// 读取剩下的包长度
	if _, err := io.ReadFull(this.reader, this.rcv_buf[HEAD_SIZE:size]); nil != err {
		return err
	}

	// 重置包状态
	packet.Reset()

	// 解析数据包
	return packet.Unmarshal(this.rcv_buf[:size])
}

func (this *PacketStream) WritePacket(packet *Packet) error {

	size := int(packet.AllSize())

	if len(this.snd_buf) < size {
		this.snd_buf = make([]byte, size+32)
	}

	//序列化数据包
	if err := packet.MarshalTo(this.snd_buf[:size]); nil != err {
		return err
	}

	return this.WriteBytes(this.snd_buf[:size])
}

func (this *PacketStream) WriteBytes(buf []byte) error {

	for len(buf) > 0 {
		n, err := this.writer.Write(buf)
		if err != nil {
			return err
		}
		buf = buf[n:]
	}

	return nil
}

func (this *PacketStream) WriteChan(bytes []byte, send_chan <-chan []byte) error {

	var err error
	var ok = true

LOOP:
	if !ok {
		return fmt.Errorf("send chan closed.")
	}

	if err = this.WriteBytes(bytes); nil != err {
		return err
	}

	select {
	case bytes, ok = <-send_chan:
		goto LOOP
	default:
	}

	return this.writer.Flush()
}
