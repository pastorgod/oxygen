package xnet

import (
	"sync/atomic"
)

type IdAlloctor struct {
	inc   uint32
	max   uint32
	start uint32
}

func NewIdAlloctor(max, start uint32) *IdAlloctor {

	return &IdAlloctor{
		inc:   start,
		max:   max,
		start: start,
	}
}

func (this *IdAlloctor) Get() uint32 {
	atomic.CompareAndSwapUint32(&this.inc, this.max, this.start)
	return atomic.AddUint32(&this.inc, 1)
}

func (this *IdAlloctor) Reset() {
	atomic.StoreUint32(&this.inc, this.start)
}
