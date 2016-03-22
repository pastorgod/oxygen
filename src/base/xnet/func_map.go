package xnet

import (
	. "logger"
	"reflect"
)

type MemberFunc struct {
	obj reflect.Value
	fn  reflect.Value
}

type funcMap map[uint32]MemberFunc

type FuncMap struct {
	funcs funcMap
}

func NewFuncMap(init_num int) *FuncMap {

	return &FuncMap{
		funcs: make(funcMap, init_num),
	}
}

func (this *FuncMap) Bind(fnid uint32, fn interface{}) {

	if _, ok := this.Exists(fnid); ok {
		FATAL("repeat to bind: %d", fnid)
	}

	fun := reflect.ValueOf(fn)

	if fun.Kind() != reflect.Func {
		FATAL("fn is not a function! %v", fn)
	}

	this.funcs[fnid] = MemberFunc{obj: reflect.Value{}, fn: fun}
}

func (this *FuncMap) Remove(fnid uint32) {
	delete(this.funcs, fnid)
}

func (this *FuncMap) BindName(fnname string, object interface{}, fn reflect.Value) uint32 {

	fnid := Hash(fnname)

	if _, ok := this.Exists(fnid); ok {
		FATAL("repeat to bind: %d, %s", fnid, fnname)
	}

	this.funcs[fnid] = MemberFunc{obj: reflect.ValueOf(object), fn: fn}
	return fnid
}

func (this *FuncMap) Call(fnid uint32, params ...interface{}) []reflect.Value {

	mf, ok := this.Exists(fnid)

	if !ok {
		FATAL("%d func dose not exist.", fnid)
	}

	n := 0

	if mf.obj.IsValid() {
		n = 1
	}

	args := make([]reflect.Value, len(params)+n)

	if n > 0 {
		args[0] = mf.obj
	}

	fn := mf.fn

	for index, param := range params {
		if nil != params {
			args[index+n] = reflect.ValueOf(param)
		} else {
			args[index+n] = reflect.Zero(fn.Type().In(index + n))
		}
	}

	return fn.Call(args)
}

func (this *FuncMap) Exists(fnid uint32) (MemberFunc, bool) {
	fn, ok := this.funcs[fnid]
	return fn, ok
}

func (this *FuncMap) Reset() {
	this.funcs = make(funcMap)
}
