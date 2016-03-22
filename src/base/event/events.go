package event

import (
	"fmt"
	. "logger"
	"reflect"
)

type EventType int

type eventArray []*eventHandler
type eventMap map[EventType]eventArray

type eventForward []*EventEmitter

type eventHandler struct {
	fn   reflect.Value
	args []reflect.Type
}

type EventEmitter struct {
	events     eventMap
	eventsOnce eventMap
	forwards   eventForward
}

func NewEventEmitter() *EventEmitter {
	return &EventEmitter{events: make(eventMap, 128), eventsOnce: make(eventMap, 128), forwards: make(eventForward, 0, 32)}
}

func genEventHandler(fn interface{}) (handler *eventHandler, err error) {
	// if a handler have been generated before, use it first
	fnValue := reflect.ValueOf(fn)
	handler = new(eventHandler)

	if reflect.TypeOf(fn).Kind() != reflect.Func {
		err = fmt.Errorf("%v is not a function", fn)
		LOG_ERROR("参数不是函数! %s", fn)
		return
	}

	handler.fn = fnValue
	fnType := fnValue.Type()
	nArgs := fnValue.Type().NumIn()
	handler.args = make([]reflect.Type, nArgs)

	for i := 0; i < nArgs; i++ {
		handler.args[i] = fnType.In(i)
	}

	return
}

func (ee *EventEmitter) AddForward(target *EventEmitter) {

	if ee == target {
		FATAL("PARAMS ERROR.")
	}

	ee.RemoveForward(target)
	ee.forwards = append(ee.forwards, target)
}

func (ee *EventEmitter) RemoveForward(target *EventEmitter) {

	for index, e := range ee.forwards {

		if e != target {
			continue
		}

		ee.forwards = append(ee.forwards[:index], ee.forwards[index+1:]...)
		break
	}
}

func (ee *EventEmitter) On(etype EventType, fn interface{}) error {
	handler, err := genEventHandler(fn)
	if err != nil {
		return err
	}

	if _, ok := ee.events[etype]; !ok {
		ee.events[etype] = make(eventArray, 0, 32)
	}

	ee.events[etype] = append(ee.events[etype], handler)
	return nil
}

func (ee *EventEmitter) Once(etype EventType, fn interface{}) error {
	handler, err := genEventHandler(fn)
	if err != nil {
		return err
	}

	if _, ok := ee.events[etype]; !ok {
		ee.events[etype] = make(eventArray, 0, 32)
	}
	ee.eventsOnce[etype] = append(ee.eventsOnce[etype], handler)
	return nil
}

func (ee *EventEmitter) AddListener(etype EventType, fn interface{}) error {
	return ee.On(etype, fn)
}

func (ee *EventEmitter) RemoveListener(etype EventType, fn interface{}) {
	for i, handler := range ee.events[etype] {
		if handler.fn.Pointer() == reflect.ValueOf(fn).Pointer() {
			ee.events[etype] = append(ee.events[etype][0:i], ee.events[etype][i+1:]...)
			break
		}
	}
	if len(ee.events[etype]) == 0 {
		delete(ee.events, etype)
	}
	for i, handler := range ee.eventsOnce[etype] {
		if handler.fn.Pointer() == reflect.ValueOf(fn).Pointer() {
			ee.eventsOnce[etype] = append(ee.eventsOnce[etype][0:i], ee.eventsOnce[etype][i+1:]...)
			break
		}
	}
	if len(ee.eventsOnce[etype]) == 0 {
		delete(ee.eventsOnce, etype)
	}
}

func (ee *EventEmitter) RemoveAllListenersBy(etype EventType) {
	// assign nil?
	delete(ee.events, etype)
	delete(ee.eventsOnce, etype)
}

func (ee *EventEmitter) RemoveAllListeners() {
	ee.events = make(eventMap)
	ee.eventsOnce = make(eventMap)
	ee.forwards = make(eventForward, 0)
}

func (ee *EventEmitter) fetchHandlers(etype EventType) (handlers []*eventHandler) {
	handlers = ee.eventsOnce[etype]
	ee.eventsOnce[etype] = nil
	delete(ee.eventsOnce, etype)
	handlers = append(handlers, ee.events[etype]...)
	return
}

var VoidParam = make([]reflect.Value, 0)

func (ee *EventEmitter) Emit(etype EventType, args ...interface{}) {

	handlers := ee.fetchHandlers(etype)

	callArgs := make([]reflect.Value, len(args))

	for i, arg := range args {
		callArgs[i] = reflect.ValueOf(arg)
	}

	for _, handler := range handlers {

		if len(args) < len(handler.args) {
			FATAL("函数参数数目错误! %s", handler.fn.String())
		}

		if length := len(handler.args); length > 0 {
			handler.fn.Call(callArgs[:])
			//handler.fn.Call( callArgs[:length] )
		} else {
			handler.fn.Call(VoidParam)
		}
	}

	// 转发事件
	for _, em := range ee.forwards {
		em.Emit(etype, args...)
	}
}
