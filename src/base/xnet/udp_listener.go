/**
* Copyright (c) limpo. All rights reserved.
* @author: limpo,limpo1989@gmail.com
* @create: 2016-02-29 12:08:47
* @brief:
*
**/

package xnet

import "os"
import "net"
import "time"
import "errors"
import "encoding/binary"

// handleshake packet.
var handshake_req = []byte{0, 0, 0, 0}

type udp_listener struct {
	conn      *net.UDPConn
	mapping   map[string]uint32
	container *udp_container
	accepts   chan net.Conn
	closed    chan string
	signal    chan struct{}
	stoped    bool
}

// Listen announces on the local network address laddr.
func ListenUDP(snet, laddr string) (net.Listener, error) {

	// parse udp address.
	udp_addr, err := net.ResolveUDPAddr(snet, laddr)

	if err != nil {
		return nil, err
	}

	// listen udp to UDPAddr.
	udp_conn, err := net.ListenUDP(snet, udp_addr)

	if err != nil {
		return nil, err
	}

	return new_udp_listener(udp_conn), nil
}

func new_udp_listener(udp_conn *net.UDPConn) *udp_listener {

	listener := &udp_listener{
		conn:      udp_conn,
		mapping:   make(map[string]uint32, 1024),
		container: new_udp_container(1024),
		accepts:   make(chan net.Conn, 64),
		closed:    make(chan string, 128),
		signal:    make(chan struct{}),
	}

	go listener.run()
	return listener
}

// Accept waits for and returns the next connection to the listener.
func (this *udp_listener) Accept() (net.Conn, error) {

	select {
	case <-this.signal:
		return nil, errors.New("listener closed.")
	case conn := <-this.accepts:
		return conn, nil
	}
}

// Close closes the listener.
func (this *udp_listener) Close() error {
	/*
		if err := this.conn.Close(); nil != err {
			return err
		}
		close(this.signal)
	*/
	this.stoped = true
	return nil
}

// Addr returns the listener's network address.
func (this *udp_listener) Addr() net.Addr {
	return this.conn.LocalAddr()
}

func (this *udp_listener) File() (*os.File, error) {
	return this.conn.File()
}

// read udp packet from listen conn.
func (this *udp_listener) run() {

	// max buffer = 1400(MTU) * 255 fragments.
	buffer := make([]byte, 1400*255)

	// handshake ack.
	handshake_ack := []byte{0, 0, 0, 0}

	for {
		this.conn.SetReadDeadline(time.Now().Add(time.Second))
		if recv_bytes, raddr, err := this.conn.ReadFromUDP(buffer); nil == err && recv_bytes > 0 {

			// read packet head.
			if conv := binary.LittleEndian.Uint32(buffer[:recv_bytes]); 0 == conv {
				// new udp connection.
				// accept from remote addr.
				if cache_conv, ok := this.mapping[raddr.String()]; ok {
					conv = cache_conv
				} else {

					// stoped.
					if this.stoped {
						goto RETRY
					}

					conv = this.container.get_new_conv()
					this.mapping[raddr.String()] = conv
					// put new udp connection.
					this.accepts <- this.container.new_connection(this, conv, this.conn, raddr)
				}

				binary.LittleEndian.PutUint32(handshake_ack, conv)
				this.conn.WriteToUDP(handshake_ack, raddr)
			} else {
				//TODO: get conv from mapping ?
				// if mapping[raddr.String()] != conv then
				// 		invalid conv
				// end
				this.container.input(conv, buffer[:recv_bytes], raddr)
			}
		}

	RETRY:
		select {
		case addr := <-this.closed:
			if conv, ok := this.mapping[addr]; ok {
				this.container.stop(conv)
			}
			delete(this.mapping, addr)
			break RETRY
		case <-this.signal:
			return
		default:
		}
	}
}
