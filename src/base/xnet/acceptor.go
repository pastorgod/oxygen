/**
* Copyright (c) limpo. All rights reserved.
* @author: limpo,limpo1989@gmail.com
* @create: 2016-03-07 16:44:05
* @brief:
*
**/
package xnet

import "net"
import "os"
import . "logger"

type Acceptor struct {
	l net.Listener
}

func NewAcceptor(rawurl string) (*Acceptor, error) {

	l, err := listen(rawurl)

	if err != nil {
		return nil, err
	}

	return NewAcceptorWith(l), nil
}

func NewAcceptorWith(l net.Listener) *Acceptor {
	return &Acceptor{l: l}
}

func (this *Acceptor) Addr() string {
	return this.l.Addr().String()
}

func (this *Acceptor) Port() int32 {

	switch addr := this.l.Addr().(type) {
	case *net.TCPAddr:
		return int32(addr.Port)
	case *net.UDPAddr:
		return int32(addr.Port)
	default:
		ERROR("incorrent addr: %#v", addr)
		return 0
	}
}

func (this *Acceptor) Accept() (ISession, error) {

	conn, err := this.l.Accept()

	if err != nil {
		return nil, err
	}

	session := new_session_impl(conn)
	go session.recvLoop()
	go session.sendLoop()

	return session, nil
}

func (this *Acceptor) Close() error {
	return this.l.Close()
}

func (this *Acceptor) AcceptLoop(callback func(ISession) bool) error {

	for {
		session, err := this.Accept()

		if err != nil {
			return err
		}

		if !callback(session) {
			session.Close(ConnectionReset)
		}
	}

	return nil
}

func (this *Acceptor) FileDescriptor(noCloseOnExec bool) int {

	var file *os.File
	var err error

	switch listener := this.l.(type) {
	case *net.TCPListener:
		file, err = listener.File()
	case *udp_listener:
		file, err = listener.File()
	default:
		LOG_ERROR("invalid listener: %#v", listener)
		return -1
	}

	if nil != err {
		LOG_ERROR("Conn.File: %v", err)
		return -1
	}

	fd := file.Fd()

	if noCloseOnExec && NoCloseOnExec(fd) {
		LOG_ERROR("NoCloseOnExec: %d", fd)
		return -1
	}

	return int(fd)
}
