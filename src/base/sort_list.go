package base

import (
	"container/list"
)

type ISortInterface interface {
	Less(item interface{}) bool
}

type IEqualInterface interface {
	Equal(key interface{}) bool
}

type IUnmarshaler interface {
	UnmarshalArray([]byte) ([]interface{}, error)
}

type SortList struct {
	list      *list.List
	unmarshal IUnmarshaler
}

func NewSortList() *SortList {
	return NewSortListEx(nil)
}

func NewSortListEx(unmarshaler IUnmarshaler) *SortList {
	return &SortList{
		list:      list.New(),
		unmarshal: unmarshaler,
	}
}

func (this *SortList) Front() interface{} {
	return this.list.Front().Value
}

func (this *SortList) Back() interface{} {
	return this.list.Back().Value
}

func (this *SortList) Len() int {
	return this.list.Len()
}

func (this *SortList) Push(val interface{}) int {

	pos := 0

	for elem := this.list.Front(); elem != nil; elem = elem.Next() {

		v := elem.Value.(ISortInterface)

		if v.Less(val) {
			this.list.InsertBefore(val, elem)
			return pos
		}

		pos += 1
	}

	this.list.PushBack(val)
	return pos
}

func (this *SortList) Remove(key interface{}) interface{} {

	for elem := this.list.Front(); elem != nil; elem = elem.Next() {
		v := elem.Value.(IEqualInterface)

		if v.Equal(key) {
			this.list.Remove(elem)
			return elem.Value
		}
	}

	return nil
}

func (this *SortList) Find(key interface{}) interface{} {

	for elem := this.list.Front(); elem != nil; elem = elem.Next() {
		v := elem.Value.(IEqualInterface)

		if v.Equal(key) {
			return elem.Value
		}
	}

	return nil
}

func (this *SortList) FindIndex(key interface{}) int {

	index := 0

	for elem := this.list.Front(); elem != nil; elem = elem.Next() {
		v := elem.Value.(IEqualInterface)

		if v.Equal(key) {
			return index
		}

		index++
	}

	return -1
}

func (this *SortList) PopFront() interface{} {
	elem := this.list.Front()
	this.list.Remove(elem)

	return elem.Value
}

func (this *SortList) PopBack() interface{} {
	elem := this.list.Back()
	this.list.Remove(elem)

	return elem.Value
}

func (this *SortList) ForEach(fun func(int32, interface{})) {

	index := int32(0)
	for elem := this.list.Front(); elem != nil; elem = elem.Next() {
		fun(index, elem.Value)
		index++
	}
}

func (this *SortList) ForEachBreak(fun func(int32, interface{}) bool) {

	index := int32(0)
	for elem := this.list.Front(); elem != nil; elem = elem.Next() {
		if !fun(index, elem.Value) {
			break
		}
		index++
	}
}

func (this *SortList) ForEachBy(min, max int32, fun func(int32, interface{})) {

	index := int32(0)
	for elem := this.list.Front(); elem != nil; elem = elem.Next() {
		if index < min {
			continue
		}

		if index >= max {
			break
		}
		fun(index, elem.Value)
		index++
	}
}

func (this *SortList) Clear() {
	this.list.Init()
}

func (this *SortList) MarshalJSON() ([]byte, error) {

	list := make([]interface{}, 0, this.Len())

	this.ForEach(func(index int32, item interface{}) {
		list = append(list, item)
	})

	return ToJsonData(list)
}

func (this *SortList) UnmarshalJSON(data []byte) error {

	list, err := this.unmarshal.UnmarshalArray(data)

	if err != nil {
		return err
	}

	for _, item := range list {
		this.list.PushBack(item)
	}

	return nil
}
