/**
* Copyright (c) limpo. All rights reserved.
* @author: limpo,limpo1989@gmail.com
* @create: 2016-03-07 16:12:10
* @brief:
*
**/
package xnet

import "net"
import "net/url"
import "strings"
import "fmt"
import "os"
import "strconv"

import . "logger"

// tcp://192.168.1.2:4321
// udp://192.168.1.2:4321
// fd://121
func listen(rawurl string) (net.Listener, error) {

	u, err := url.Parse(rawurl)

	if err != nil {
		return nil, err
	}

	scheme := strings.ToLower(u.Scheme)

	switch {
	case strings.Contains(scheme, "tcp"):
		return net.Listen(scheme, u.Host)
	case strings.Contains(scheme, "udp"):
		return ListenUDP(scheme, u.Host)
	case strings.Contains(scheme, "fd"):
		fd, err := strconv.Atoi(u.Host)
		if err != nil {
			return nil, fmt.Errorf("invalid fd: %s, %v", rawurl, err)
		}
		return listen_fd(fd)
	default:
		return nil, fmt.Errorf("invalid scheme: %s", u.Scheme)
	}
}

// listen fd
func listen_fd(fd int) (net.Listener, error) {

	DEBUG("new listener with %d", fd)

	f := os.NewFile(uintptr(fd), "listen socket")
	defer f.Close()

	c, err := net.FileConn(f)

	if err != nil {
		return nil, err
	}

	defer c.Close()

	switch conn := c.(type) {
	case *net.TCPConn:
		return net.FileListener(f)
	case *net.UDPConn:
		return new_udp_listener(conn), nil
	default:
		return nil, fmt.Errorf("invalid conn: %#v", conn)
	}
}

func Listen(rawurl string) (*Service, error) {

	l, err := listen(rawurl)

	if nil != err {
		return nil, err
	}

	return NewService(l), nil
}

func ListenRpc(rawurl string) (*RpcService, error) {

	service, err := Listen(rawurl)

	if nil != err {
		return nil, err
	}

	return NewRpcService(service), nil
}
