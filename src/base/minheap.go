package base

import (
	"container/heap"
)

type HeapElement struct {
	// user data.
	Value interface{}
	// The priority of the item in the queue.
	Priority int64
	// The index of the item in the heap.
	index int
}

// A PriorityQueue implements heap.Interface and holds Items.
type MinHeap []*HeapElement

func NewMinHeap(capcity int) *MinHeap {
	mh := make(MinHeap, 0, capcity)
	heap.Init(&mh)
	return &mh
}

// implements heap.Interface
func (this MinHeap) Len() int { return len(this) }

func (this MinHeap) Less(i, j int) bool {
	return this[i].Priority < this[j].Priority
}

func (this MinHeap) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
	this[i].index = i
	this[j].index = j
}

func (this *MinHeap) Push(x interface{}) {
	n := len(*this)
	item := x.(*HeapElement)
	item.index = n
	*this = append(*this, item)
}

func (this *MinHeap) Pop() interface{} {
	old := *this
	n := len(old)
	item := old[n-1]
	item.index = -1
	*this = old[:n-1]
	return item
}

func (this *MinHeap) PushEl(el *HeapElement) {
	heap.Push(this, el)
}

func (this *MinHeap) PopEl() *HeapElement {
	el := heap.Pop(this)
	return el.(*HeapElement)
}

func (this *MinHeap) RemoveEl(el *HeapElement) {
	if -1 == el.index {
		panic("min-heap error.")
	}
	heap.Remove(this, el.index)
}

func (this *MinHeap) PeekEl() *HeapElement {
	items := *this
	return items[0]
}

func (this *MinHeap) UpdateEl(el *HeapElement, priority int64) {
	heap.Remove(this, el.index)
	el.Priority = priority
	heap.Push(this, el)
}
