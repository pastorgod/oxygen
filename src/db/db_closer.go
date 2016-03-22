package db

import (
	"sync"
)

var closer_notifer = make(chan int)
var wait_quit = &sync.WaitGroup{}

type ICloser interface {

	// close on destroy.
	Close()
}

func onDestroy(closer ICloser) {
	go func() {
		wait_quit.Add(1)
		<-closer_notifer
		closer.Close()
		wait_quit.Done()
	}()
}

func Destroy() {
	if nil != closer_notifer {
		close(closer_notifer)
		wait_quit.Wait()
	}
}
