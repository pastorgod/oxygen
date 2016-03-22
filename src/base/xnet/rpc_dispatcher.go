package xnet

import (
	. "logger"
	"reflect"
)

type IServiceModule interface {
	// name of service.
	Name() string
	// call service method by rpc_id.
	ServiceCall(ctx *Context, rpc_id uint32, in Message) (reply Message, err *string)
	// service implementation.
	Impl() interface{}
}

type RpcDispatcher struct {
	methods map[uint32]IServiceModule
}

func NewRpcDispatcher() *RpcDispatcher {
	return &RpcDispatcher{
		methods: make(map[uint32]IServiceModule, 128),
	}
}

func (this *RpcDispatcher) RegisterService(svc IServiceModule) {

	t := reflect.TypeOf(svc.Impl())

	// use service name.
	name := svc.Name()

	for i, size := 0, t.NumMethod(); i < size; i++ {

		method := t.Method(i)

		if this.checkFunc(method.Type, 1) {
			sname := Sprintf("%s.%s", name, method.Name)
			rpcid := Hash(sname)

			if svc, ok := this.methods[rpcid]; ok {
				FATAL("repeat method: %s => %s", sname, svc.Name())
				return
			}

			this.methods[rpcid] = svc

			DEBUG("RpcDispatcher.Registe: %s => %d", sname, rpcid)
		}
	}
}

func (this *RpcDispatcher) UnregisterService(svc IServiceModule) {

	t := reflect.TypeOf(svc.Impl())

	// use service name.
	name := svc.Name()

	for i, size := 0, t.NumMethod(); i < size; i++ {

		method := t.Method(i)

		if this.checkFunc(method.Type, 1) {
			sname := Sprintf("%s.%s", name, method.Name)
			delete(this.methods, Hash(sname))
		}
	}
}

func (this *RpcDispatcher) ServiceCall(ctx *Context, rpc_id uint32, in Message) (reply Message, err *string) {

	if svc, ok := this.methods[rpc_id]; ok {
		return svc.ServiceCall(ctx, rpc_id, in)
	}

	err = &MethodNotFound
	return
}

// like this: func( *Context, in, out Message ) *string
func (this *RpcDispatcher) checkFunc(fn reflect.Type, n int) bool {

	if fn.NumIn() != 3+n {
		//DEBUG( "参数数目不对: %d", 2 + n )
		return false
	}

	retType := reflect.TypeOf((*string)(nil))

	if fn.NumOut() != 1 || !fn.Out(0).ConvertibleTo(retType) {
		//DEBUG( "返回值错误!" )
		return false
	}

	msgType := reflect.TypeOf((*Message)(nil)).Elem()

	if !fn.In(1 + n).Implements(msgType) {
		//DEBUG( "输入参数错误" )
		return false
	}

	if !fn.In(2 + n).Implements(msgType) {
		//DEBUG( "输出参数错误" )
		return false
	}

	return true
}
