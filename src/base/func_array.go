package base

import (
	. "logger"
	"reflect"
)

type funcArray []reflect.Value

type FuncArray struct {
	funcs funcArray
}

func NewFuncArray(init_num int) *FuncArray {

	return &FuncArray{
		funcs: make(funcArray, init_num),
	}
}

func (this *FuncArray) reserve(index int) {

	if index >= len(this.funcs) {
		tmp := this.funcs
		this.funcs = make(funcArray, index+len(tmp)/2)
		copy(this.funcs, tmp)
	}
}

func (this *FuncArray) Bind(index int, fn interface{}) {

	this.reserve(index)

	if !this.funcs[index].IsNil() {
		FATAL("repeat to bind: %d", index)
	}

	fun := reflect.ValueOf(fn)

	if fun.Kind() != reflect.Func {
		FATAL("fn is not a function! %v", fn)
	}

	this.funcs[index] = fun
}

func (this *FuncArray) Call(index int, params ...interface{}) []reflect.Value {

	if !this.Exists(index) {
		FATAL("%d func dose not exist.", index)
	}

	args := make([]reflect.Value, len(params))

	for index, param := range params {
		args[index] = reflect.ValueOf(param)
	}

	return this.funcs[index].Call(args)

}

func (this *FuncArray) Exists(index int) bool {
	return index < len(this.funcs) && !this.funcs[index].IsNil()
}

func (this *FuncArray) Reset() {
	this.funcs = make(funcArray, len(this.funcs))
}
