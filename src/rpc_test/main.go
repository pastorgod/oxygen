/**
* Copyright (c) limpo. All rights reserved.
* @author: limpo,limpo1989@gmail.com
* @create: 2016-03-07 15:27:26
* @brief:
*
**/
package main

import "base/xnet"
import "command"

import _ "record"

type MathService struct {
}

func (*MathService) Add(context *xnet.Context, in *command.AddRequest, out *command.AddReply) *string {

	out.X = in.A + in.B
	return nil

	/*
		go func() {
			// 模拟耗时操作
			time.Sleep(time.Second * 10)
			out.X = in.A + in.B
			context.Response(nil, out)
		}()

		return context.Asynchronized()
	*/
}

func main() {

	// remote service.
	{
		math_service, err := command.NewMathServiceImpl("tcp://192.168.1.2:4444", &MathService{})
		xnet.Assert(nil == err, err)

		go math_service.AcceptLoop(func(session xnet.ISession) bool {
			return true
		})
	}

	// local proxy service.
	{
		proxy_service, err := command.NewMathServiceImplWithProxy("udp://192.168.1.2:4455", "tcp://192.168.1.2:4444")
		xnet.Assert(nil == err, err)

		go proxy_service.AcceptLoop(func(session xnet.ISession) bool {
			return true
		})
	}

	// dial to local proxy service.
	//client, err := command.DialMathService("tcp://192.168.1.2:4444")
	client, err := command.DialMathService("udp://192.168.1.2:4455")
	xnet.Assert(nil == err, err)

	reply, serr := client.Add(&command.AddRequest{A: 1, B: 2})
	xnet.Assert(nil == serr, serr)

	xnet.Assert(3 == reply.X, "result error.", reply)
}
