/**
* Copyright (c) limpo. All rights reserved.
* @author: limpo,limpo1989@gmail.com
* @create: 2016-03-07 17:14:10
* @brief:
*
**/
package xnet

import "net"
import "net/url"
import "time"
import "fmt"
import "strings"

func dial(rawurl string, timeout time.Duration) (net.Conn, error) {

	u, err := url.Parse(rawurl)

	if err != nil {
		return nil, err
	}

	scheme := strings.ToLower(u.Scheme)

	switch {
	case strings.Contains(scheme, "tcp"):
		return net.DialTimeout(scheme, u.Host, timeout)
	case strings.Contains(scheme, "udp"):
		return DialUDP(scheme, u.Host, timeout)
	default:
		return nil, fmt.Errorf("invalid scheme: %s", u.Scheme)
	}
}

func Dial(rawurl string, timeout time.Duration) (ISession, error) {

	conn, err := dial(rawurl, timeout)

	if err != nil {
		return nil, err
	}

	session := new_session_impl(conn)
	go session.recvLoop()
	go session.sendLoop()

	return session, nil
}

func DialRpc(rawurl string, timeout time.Duration) (*RpcSession, error) {

	session, err := Dial(rawurl, timeout)

	if err != nil {
		return nil, err
	}

	return NewRpcSession(session, nil), nil
}
