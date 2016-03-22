/**
* Copyright (c) limpo. All rights reserved.
* @author: limpo,limpo1989@gmail.com
* @create: 2016-03-07 17:33:29
* @brief:
*
**/
package xnet

type RpcService struct {
	*Service
	dispatcher *RpcDispatcher
}

func NewRpcService(service *Service) *RpcService {
	return &RpcService{Service: service, dispatcher: NewRpcDispatcher()}
}

func (this *RpcService) Dispatcher() *RpcDispatcher {
	return this.dispatcher
}

func (this *RpcService) SetDispatcher(dispatcher *RpcDispatcher) {
	this.dispatcher = dispatcher
}

func (this *RpcService) RegisterService(service IServiceModule) {
	this.dispatcher.RegisterService(service)
}

func (this *RpcService) AcceptLoop(callback func(ISession) bool) {

	this.Service.AcceptLoop(func(session ISession) bool {
		return callback(NewRpcSession(session, this.Dispatcher()))
	})
}
