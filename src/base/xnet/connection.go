package xnet

import (
	"base/timer"
	. "logger"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type messageQueue chan []byte

var (
	KEEPALIVE_IDLE    = time.Second * 60       // 闲置超时
	KEEPALIVE_TIMEOUT = time.Millisecond * 500 // 响应超时
	KEEPALIVE_COUNT   = int32(5)               // 尝试次数

	CLEAR_TIMEOUT = time.Time{}

	KEEPALIVE_REQ_PACKET = []byte{0, 0, 0, 16, 1, 0, 0, 0}
	KEEPALIVE_ACK_PACKET = []byte{0, 0, 0, 16, 2, 0, 0, 0}
	HANDSHAKE_REQ_PACKET = []byte{0, 0, 0, 16, 3, 0, 0, 0}
	HANDSHAKE_ACK_PACKET = []byte{0, 0, 0, 16, 4, 0, 0, 0}
)

const DEFAULT_SEND_QUEUE = 256

type connection struct {
	conn   net.Conn
	stream IPacketStream

	// packet send queue.
	sendByteQueue messageQueue

	// about close.
	closeChan chan int
	closeFlag sync.Once
	closed    bool

	// notifier on closed
	closeNotify ICloseNotifier

	// about keep-alive
	keepCount    int32
	keepTimer    *timer.Timer
	recentlyTime int64
	keepMutex    *sync.Mutex

	receivedPackets int32
	perPackets      int32
	oneSecRecvTime  int64
	perLimitRatio   int32

	// sequence id.
	sequenceId uint32
}

func new_connection(notify ICloseNotifier, conn net.Conn) *connection {

	// enable Nagle's algorithm
	//if tcp_conn, ok := conn.(*net.TCPConn); ok {
	//	tcp_conn.SetNoDelay(true)
	//}

	return &connection{
		conn:          conn,
		stream:        NewPacketStream(conn, 1024, 1024),
		sendByteQueue: make(messageQueue, DEFAULT_SEND_QUEUE),
		closeChan:     make(chan int),
		closeNotify:   notify,
		sequenceId:    SEQUENCE_INTERNAL,
	}
}

func (this *connection) Addr() string {
	return this.conn.RemoteAddr().String()
}

func (this *connection) RawAddr() string {

	switch addr := this.conn.RemoteAddr().(type) {
	// tcp://192.168.1.2:1234
	case *net.TCPAddr:
		return Sprintf("%s://%s", addr.Network(), addr.String())
	// udp://192.168.1.2:1234
	case *net.UDPAddr:
		return Sprintf("%s://%s", addr.Network(), addr.String())
	default:
		LOG_ERROR("invalid addr: %#v", addr)
		return addr.String()
	}
}

func (this *connection) SendBytes(bytes []byte) bool {

	select {
	case <-this.closeChan:
		return false
	default:
		this.sendByteQueue <- bytes
	}

	return true
}

func (this *connection) SendPacket(packet *Packet) bool {
	bytes, err := packet.Marshal()
	if nil != err {
		FATAL("packet.Marshal: %v", err)
		return false
	}
	return this.SendBytes(bytes)
}

func (this *connection) SendBytesTimeout(bytes []byte, timeout time.Duration) bool {

	select {
	case <-this.closeChan:
		return false
	default:

		select {
		case this.sendByteQueue <- bytes:
			return true
		case <-time.After(timeout):
			LOG_WARN("SendBytesTimeout: %d bytes", len(bytes))
			return false
		}
	}

	return false
}

func (this *connection) Alive() bool {
	return !this.closed
}

func (this *connection) IsSlowing() bool {
	// 网络断开了那么直接返回
	if !this.Alive() {
		return true
	}
	// 待发送队列如果超过一半等待发送那么就认为是慢速网络
	return len(this.sendByteQueue) > (cap(this.sendByteQueue) / 2)
}

func (this *connection) SetLimitRatio(num int32) {
	this.perLimitRatio = num
}

func (this *connection) SetReadTimeout(sec int) {

	if !this.Alive() {
		return
	}

	var next = CLEAR_TIMEOUT

	if sec > 0 {
		next = time.Now().Add(time.Second * time.Duration(sec))
	}

	this.conn.SetReadDeadline(next)
}

func (this *connection) sendLoop() {

	defer func() {
		if err := recover(); err != nil {
			this.Close(ToError("%v", err))
			PrintStack("connection.sendLoop: %v", err)
		}
	}()

	for {
		select {
		case bytes, ok := <-this.sendByteQueue:
			if !ok {
				return
			}

			if err := this.stream.WriteChan(bytes, this.sendByteQueue); nil != err {
				this.Close(err)
				return
			}
		case <-this.closeChan:
			return
		}
	}

}

func (this *connection) recvLoop(handler func(*Packet)) {

	defer func() {
		if err := recover(); err != nil {
			this.Close(ToError("%v", err))
			PrintStack("connection.recvLoop: %v", err)
		}
	}()

	this.startKeepAlive()
	defer this.stopKeepAlive()

	for {

		packet := &Packet{}

		if err := this.stream.ReadPacket(packet); nil != err {
			this.Close(err)
			return
		}

		if err := this.handle_packet(packet, handler); nil != err {
			this.Close(err)
			return
		}
	}
}

func (this *connection) handle_packet(packet *Packet, handler func(*Packet)) error {

	// 处理 sequence
	if packet.CheckMask(HEAD_SEQ_MASK) {

		// check sequence
		if err := this.onSequence(packet); err != nil {
			return err
		}

		// 内部序号不作处理
		if packet.Sequence() < SEQUENCE_INTERNAL {
			return nil
		}
	}

	// 处理消息
	if packet.CheckMask(HEAD_MSG_MASK) {
		// handler packet
		handler(packet)

		// 记录一下最后一次收到数据包是什么时候, 优化keep-alive
		atomic.StoreInt64(&this.recentlyTime, time.Now().Unix())

		// 累计收包数
		this.receivedPackets += 1

		// 一秒钟重置一次
		if now := time.Now().Unix(); now != this.oneSecRecvTime {
			this.oneSecRecvTime = now
			this.perPackets = 0
		}

		// 累计每秒收包速率
		this.perPackets += 1

		// 如果每秒的收包速率太高则通知上层处理
		if this.perLimitRatio > 0 && this.perPackets > this.perLimitRatio {
			LOG_WARN("收包速率超过阈值: %s [%d > %d] %d",
				this.Addr(), this.perPackets, this.perLimitRatio, this.receivedPackets)
		}
	}

	return nil
}

func (this *connection) startKeepAlive() {

	if nil != this.keepTimer {
		panic("keepAlive: keepTimer non nil.")
	}

	// init mutex.
	this.keepMutex = &sync.Mutex{}

	// 设置一下超时
	this.keepTimer = timer.AfterFunc(KEEPALIVE_IDLE, func() {

		// 保护定时器
		this.keepMutex.Lock()
		defer this.keepMutex.Unlock()

		if nil == this.keepTimer {
			return
		}

		// 如果超时时间内有收到过包那么就不用发包验证了, 仅验证一下空闲的连接
		if time.Now().Unix()-atomic.LoadInt64(&this.recentlyTime) < int64(KEEPALIVE_IDLE/time.Second) {
			// 继续等待下一次超时
			this.keepTimer.Reset(KEEPALIVE_IDLE)
			return
		}

		// 超时的时候检查一下
		if count := atomic.AddInt32(&this.keepCount, 1); count <= KEEPALIVE_COUNT {
			// 等待客户端回应超时
			this.keepTimer.Reset(KEEPALIVE_TIMEOUT)
			// 启动一次探测
			this.SendBytesTimeout(KEEPALIVE_REQ_PACKET, KEEPALIVE_TIMEOUT)
			return
		}

		//		LOG_WARN( "connection: keep-alive fail. %s", this.Addr() )

		// 全部超时，关闭这个连接
		this.Close(KeepAliveTimeout)
	})
}

func (this *connection) stopKeepAlive() {

	// 保护定时器
	this.keepMutex.Lock()
	defer this.keepMutex.Unlock()

	if nil != this.keepTimer {
		this.keepTimer.Stop()
		this.keepTimer.Destory()
		this.keepTimer = nil
	}
}

func (this *connection) onSequence(packet *Packet) error {

	switch packet.Sequence() {
	// 收到探测请求了, 给个回应
	case SEQUENCE_KEEPALIVE:
		this.SendBytesTimeout(KEEPALIVE_ACK_PACKET, KEEPALIVE_TIMEOUT)
	// 收到探测回应了, 重设一下超时次数
	case SEQUENCE_KEEPALIVE_ACK:
		atomic.StoreInt32(&this.keepCount, 0)
		this.keepTimer.Reset(KEEPALIVE_IDLE)
	// 收到握手请求了, 给个回应
	case SEQUENCE_HANDSHAKE:
		//LOG_DEBUG( "connection.onSequence SEQUENCE_HANDSHAKE %s", this.Addr() )
		this.SendBytesTimeout(HANDSHAKE_ACK_PACKET, KEEPALIVE_TIMEOUT)
	// 收到握手回应了, 确认这个连接
	case SEQUENCE_HANDSHAKE_ACK:
		//TODO: switch state
		break
	default:
		break
	}

	return this.verifySequence(packet.Sequence())
}

func (this *connection) verifySequence(seq uint32) error {

	// 内部保留序号
	if seq < SEQUENCE_INTERNAL {
		// LOG_WARN("InternalSequence: %s %d", this.Addr(), seq)
		return nil
	}

	// 阻止重放攻击
	if seq > this.sequenceId {
		this.sequenceId = seq
		return nil
	}

	LOG_ERROR("verifySequence fail, got: %d, want: %d, from: %s", seq, this.sequenceId+1, this.Addr())

	// 无效的序列号
	return InvalidSequence
}

func (this *connection) Close(err error) {

	this.closeFlag.Do(func() {

		//DEBUG( "connection %s closed, reason: %s", this.Addr(), err.Error() )
		this.closed = true

		// notify sendLoop/recvLoop to exit.
		close(this.closeChan)

		// close the socket.
		this.conn.Close()

		// close send queue.
		close(this.sendByteQueue)

		// notify close.
		this.closeNotify.onClosed(err)
		this.closeNotify = nil
	})
}

func (this *connection) clone(conn net.Conn, closeNotify ICloseNotifier) *connection {

	// 正常情况下应该已经调用过Close了
	defer this.Close(ReconnectedClose)

	return &connection{
		conn:            conn,
		stream:          NewPacketStream(conn, 1024, 1024),
		sendByteQueue:   make(messageQueue, DEFAULT_SEND_QUEUE),
		closeChan:       make(chan int),
		closeNotify:     closeNotify,
		sequenceId:      SEQUENCE_INTERNAL,
		receivedPackets: this.receivedPackets,
		perPackets:      this.perPackets,
		perLimitRatio:   this.perLimitRatio,
	}
}

func (this *connection) reconnect_to(rawurl string, timeout time.Duration) *connection {

	if conn, err := dial(rawurl, timeout); nil == err {
		LOG_INFO("网络重连成功: %s", rawurl)
		return this.clone(conn, this.closeNotify)
	} else {
		LOG_ERROR("重连网络失败: %s, %v", rawurl, err)
	}

	return nil
}

func (this *connection) reconnect(timeout time.Duration) *connection {
	return this.reconnect_to(this.RawAddr(), timeout)
}
