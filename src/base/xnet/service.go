/**
* Copyright (c) limpo. All rights reserved.
* @author: limpo,limpo1989@gmail.com
* @create: 2016-03-07 17:26:39
* @brief:
*
**/
package xnet

import "net"
import "sync"
import "time"
import "os"
import . "logger"

type IService interface {
	OnInitialize() bool

	Stop(error)

	OnDestroy()

	OnRelease()
}

type IUpdater interface {
	Restart() bool
}

type ILogicService interface {
	IService
	ITicker
	IUpdater
	ISignalHandler
}

type Service struct {
	acceptor  *Acceptor
	closeflag sync.Once
}

func NewService(l net.Listener) *Service {
	return &Service{acceptor: NewAcceptorWith(l)}
}

func (this *Service) Addr() string {
	return this.acceptor.Addr()
}

func (this *Service) Port() int32 {
	return this.acceptor.Port()
}

func (this *Service) AcceptLoop(callback func(ISession) bool) {

	if err := this.acceptor.AcceptLoop(callback); nil != err {
		this.Stop(err)
	}
}

func (this *Service) Stop(reason error) {

	this.closeflag.Do(func() {
		// close the listener.
		this.acceptor.Close()

		LOG_WARN("server is shutdown... %s", reason)
	})
}

func (this *Service) Restart() bool {
	LOG_ERROR("不支持重启操作!")
	return false
}

func (this *Service) TickDelay() time.Duration {
	return time.Second
}

func (this *Service) OnTick() {
}

//启动服务
func (this *Service) OnInitialize() bool {
	return true
}

// 默认退出信号实现
func (this *Service) OnSigQuit() SignalResult {
	return SignalResult_Quit
}

// 默认升级信号实现
func (this *Service) OnSigHup() SignalResult {
	LOG_WARN("Service.OnSigHup: SignalResult_Continue")
	return SignalResult_Continue
}

// 默认其他类型信号处理
func (this *Service) OnSignal(sig os.Signal) SignalResult {
	LOG_WARN("Service.OnSignal: %d", sig)
	return SignalResult_Continue
}

// 默认销毁实现
func (this *Service) OnDestroy() {
	LOG_INFO("Service.OnDestroy...")
}

func (this *Service) OnRelease() {
}
