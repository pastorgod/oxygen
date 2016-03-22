/**
* Copyright (c) limpo. All rights reserved.
* @author: limpo,limpo1989@gmail.com
* @create: 2016-02-20 18:20:13
* @brief:
*
**/
package xnet

import "net"
import "fmt"
import "time"
import "sync"
import "encoding/binary"

import ikcp "github.com/xtaci/kcp-go"

import . "logger"

func assert(exp bool, args ...interface{}) {
	if !exp {
		panic(fmt.Sprintf("%v", args))
	}
}

var (
	ERR_BROKEN_PIPE = fmt.Errorf("broken pipe")
	ERR_TIMEOUT     = fmt.Errorf("i/o timeout")
	ERR_HANDLESHAKE = fmt.Errorf("invalid handleshake")
	ERR_CLOSED      = fmt.Errorf("socket closed")
)

func iclock() uint32 {
	return uint32((time.Now().UnixNano() / 1000000) & 0xffffffff)
}

type udp_connection struct {
	conv     uint32
	conn     *net.UDPConn
	raddr    *net.UDPAddr
	l        *udp_listener
	kcp      *ikcp.KCP
	signal   chan struct{}
	recvs    chan []byte
	pending  []byte
	mutex    sync.Mutex
	last_err error
	// update frame optimize.
	need_udpate_flag bool
	next_update_time uint32
	// read deadline.
	r_deadline time.Time
	// client handlshake.
	connect_sig      chan struct{}
	in_connect_stage bool
}

// Dial connects to the remote address raddr on the network net, which must be "udp", "udp4", or "udp6".
func DialUDP(snet string, addr string, timeout time.Duration) (net.Conn, error) {

	// parse udp address.
	udp_addr, err := net.ResolveUDPAddr(snet, addr)

	if err != nil {
		return nil, err
	}

	conn := new_udp_conn(nil, 0, nil, udp_addr)
	if err := conn.connect(nil, udp_addr, timeout); err != nil {
		return nil, err
	}
	return conn, nil
}

func new_udp_conn(l *udp_listener, conv uint32, conn *net.UDPConn, raddr *net.UDPAddr) *udp_connection {

	udp_conn := &udp_connection{
		conv:        conv,
		conn:        conn,
		raddr:       raddr,
		l:           l,
		signal:      make(chan struct{}),
		recvs:       make(chan []byte, 128),
		connect_sig: make(chan struct{}),
	}

	if nil != l {
		udp_conn.init_kcp(conv)
	}

	go udp_conn.run()

	return udp_conn
}

func (this *udp_connection) init_kcp(conv uint32) {
	assert(nil == this.kcp, "use error.")
	this.conv = conv
	//DEBUG("udp_connection.init_kcp: server: %v, %v", this.l != nil, this.conv)
	this.kcp = ikcp.NewKCP(this.conv, func(buf []byte, size int) {
		if nil != this.l {
			this.conn.WriteToUDP(buf[:size], this.raddr)
		} else {
			this.conn.Write(buf[:size])
		}
	})

	// set kcp param.
	// use fastmode.
	this.kcp.NoDelay(1, 10, 2, 1)
	// set send/recv wnd size.
	this.kcp.WndSize(128, 128)
}

func (this *udp_connection) connect(laddr, raddr *net.UDPAddr, timeout time.Duration) error {

	udp_conn, err := net.DialUDP("udp4", laddr, raddr)

	if nil != err {
		return err
	}

	if timeout < time.Millisecond*250 {
		timeout = time.Millisecond * 250
	}

	this.conn, this.raddr = udp_conn, raddr

	ticker := time.NewTicker(time.Millisecond * 100)
	defer ticker.Stop()

	this.in_connect_stage = true
	go this.do_read_loop()

	this.conn.Write(handshake_req)

	ch_timeout := time.After(timeout)

	for {
		select {
		case <-this.connect_sig:
			this.in_connect_stage = false
			return nil
		case <-ticker.C:
			this.conn.Write(handshake_req)
		case <-ch_timeout:
			this.Close()
			return ERR_TIMEOUT
		}
	}
}

// for client side.
func (this *udp_connection) do_read_loop() {

	// max buffer 1400(MTU) * 255
	buffer := make([]byte, 1400*255)

	var recv_bytes int
	var err error

	for {
		if recv_bytes, err = this.conn.Read(buffer); recv_bytes > 0 {

			// read conv from udp packet.
			conv := binary.LittleEndian.Uint32(buffer[:recv_bytes])

			if this.in_connect_stage {
				this.init_kcp(conv)
				this.in_connect_stage = false
				close(this.connect_sig)
			}

			// kcp packet.
			this.input(buffer[:recv_bytes], this.raddr)
		}

		select {
		case <-this.signal:
			return
		default:
			if nil != err {
				this.last_err = err
				close(this.signal)
				return
			}
		}
	}
}

func (this *udp_connection) run() {
	ticker := time.NewTicker(time.Millisecond * 10)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			this.update_kcp(iclock())
		case <-this.signal:
			close(this.recvs)
			if nil != this.l {
				this.l.closed <- this.raddr.String()
			}
			return
		}
	}
}

func (this *udp_connection) update_kcp(current uint32) {

	if nil == this.kcp {
		return
	}

	if this.need_udpate_flag || current >= this.next_update_time {
		this.mutex.Lock()
		defer this.mutex.Unlock()

		// update kcp.
		this.kcp.Update(current)
		// get next update time.
		this.next_update_time = this.kcp.Check(current)
		// reset flag.
		this.need_udpate_flag = false

		//println("current: ", current, " next:", this.next_update_time, "diff:", this.next_update_time-current)
	}
}

func (this *udp_connection) input(data []byte, raddr *net.UDPAddr) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	// update udp addr.
	this.raddr = raddr

	// raw data.
	this.kcp.Input(data)
	this.need_udpate_flag = true

	// recv user packet.
	for size := this.kcp.PeekSize(); size > 0; size = this.kcp.PeekSize() {
		buffer := make([]byte, size)
		if received := this.kcp.Recv(buffer); received > 0 {
			select {
			case <-this.signal:
				WARN("recv mesasge on onclsed. %d bytes", received)
				return
			default:
				this.recvs <- buffer[:received]
			}
		}
	}
}

// udp_connection is an implementation of the Conn interface for UDP network connections.
func (this *udp_connection) Read(b []byte) (int, error) {

	if len(this.pending) > 0 {
		n := copy(b, this.pending)
		this.pending = this.pending[n:]
		return n, nil
	}

	if this.r_deadline.IsZero() {
		select {
		case <-this.signal:
			return 0, ERR_BROKEN_PIPE
		case buf := <-this.recvs:
			n := copy(b, buf)
			this.pending = buf[n:]
			return n, this.last_err
		}
	}

	if time.Now().After(this.r_deadline) {
		return 0, ERR_TIMEOUT
	}

	select {
	case <-this.signal:
		return 0, ERR_BROKEN_PIPE
	case <-time.After(this.r_deadline.Sub(time.Now())):
		return 0, ERR_TIMEOUT
	case buf := <-this.recvs:
		n := copy(b, buf)
		this.pending = buf[n:]
		return n, this.last_err
	}
}

func (this *udp_connection) Write(b []byte) (int, error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	const max = (ikcp.IKCP_MTU_DEF - ikcp.IKCP_OVERHEAD) * 255

	for {
		if len(b) <= max {
			this.kcp.Send(b)
			break
		}

		this.kcp.Send(b[:max])
		b = b[max:]
	}

	this.need_udpate_flag = true
	return len(b), this.last_err
}

func (this *udp_connection) Close() error {

	select {
	case <-this.signal:
		return ERR_BROKEN_PIPE
	default:
		this.last_err = ERR_CLOSED
		close(this.signal)
		if nil == this.l {
			return this.conn.Close()
		}
	}

	return this.last_err
}

func (this *udp_connection) LocalAddr() net.Addr {
	return this.conn.LocalAddr()
}

func (this *udp_connection) RemoteAddr() net.Addr {
	return this.raddr
}

func (this *udp_connection) SetDeadline(t time.Time) error {
	this.r_deadline = t
	return nil
}

func (this *udp_connection) SetReadDeadline(t time.Time) error {
	this.r_deadline = t
	return nil
}

func (this *udp_connection) SetWriteDeadline(t time.Time) error {
	return nil
}
