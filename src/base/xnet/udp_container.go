/**
* Copyright (c) limpo. All rights reserved.
* @author: limpo,limpo1989@gmail.com
* @create: 2016-02-20 16:43:29
* @brief:
*
**/

package xnet

import "net"
import "sync/atomic"

type container map[uint32]*udp_connection

type udp_container struct {
	conns      container
	conv_index uint32
}

func new_udp_container(capcity int) *udp_container {
	return &udp_container{
		conns: make(container, capcity),
	}
}

func (this *udp_container) find_by_conv(conv uint32) *udp_connection {
	ptr, _ := this.conns[conv]
	return ptr
}

func (this *udp_container) stop(conv uint32) {
	delete(this.conns, conv)
}

func (this *udp_container) stop_all() {
	this.conns = make(container, 1024)
}

func (this *udp_container) new_connection(l *udp_listener, conv uint32, c *net.UDPConn, addr *net.UDPAddr) *udp_connection {
	conn := new_udp_conn(l, conv, c, addr)
	this.conns[conv] = conn
	return conn
}

func (this *udp_container) input(conv uint32, data []byte, raddr *net.UDPAddr) {

	if conn := this.find_by_conv(conv); nil != conn {
		conn.input(data, raddr)
	}
}

func (this *udp_container) get_new_conv() uint32 {
	if atomic.CompareAndSwapUint32(&this.conv_index, 1<<32-1, 0) {
		// idx fold back
	}
	return atomic.AddUint32(&this.conv_index, 1)
}
