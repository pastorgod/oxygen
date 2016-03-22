package xnet

import (
	. "logger"
	"sort"
	"sync"
	"time"
)

var DefaultRequestRecorder = NewRequestRecorder()

type RequestRecord struct {
	ReqId               uint32
	ReqName             string
	ReqTotalCalled      int64
	ReqTotalElapsedTime int64
	ReqMinElapsedTime   int64
	ReqMaxElapsedTime   int64
	ReqAvgElapsedTime   int64
}

type RequestRecords []*RequestRecord

func (m RequestRecords) Len() int {
	return len(m)
}

func (m RequestRecords) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m RequestRecords) Less(i, j int) bool {
	//return m[i].ReqName < m[j].ReqName
	return m[i].ReqAvgElapsedTime < m[j].ReqAvgElapsedTime
}

type RequestRecorder struct {
	reqs  map[uint32]*RequestRecord
	mutex *sync.Mutex
}

func NewRequestRecorder() *RequestRecorder {
	return &RequestRecorder{
		reqs:  make(map[uint32]*RequestRecord, 128),
		mutex: &sync.Mutex{},
	}
}

func (this *RequestRecorder) get(req_id uint32) *RequestRecord {

	req, ok := this.reqs[req_id]

	if !ok {
		name, _ := FindMsgNameByCode(req_id)
		req = &RequestRecord{ReqId: req_id, ReqName: name}
		this.reqs[req_id] = req
	}

	return req
}

func (this *RequestRecorder) List() RequestRecords {

	this.mutex.Lock()
	defer this.mutex.Unlock()

	result := make(RequestRecords, 0, len(this.reqs))

	for _, value := range this.reqs {
		result = append(result, value)
	}

	sort.Sort(sort.Reverse(result))

	return result
}

func (this *RequestRecorder) Record(req_id uint32) func() {

	begin := time.Now()

	return func() {

		elapsed := int64(time.Now().Sub(begin) / time.Microsecond)

		this.mutex.Lock()
		defer this.mutex.Unlock()

		req := this.get(req_id)

		req.ReqTotalCalled += 1
		req.ReqTotalElapsedTime += elapsed
		req.ReqAvgElapsedTime = req.ReqTotalElapsedTime / req.ReqTotalCalled

		if 0 == req.ReqMinElapsedTime || elapsed < req.ReqMinElapsedTime {
			req.ReqMinElapsedTime = elapsed
		}

		if elapsed > req.ReqMaxElapsedTime {
			req.ReqMaxElapsedTime = elapsed
		}

		if elapsed > int64(time.Microsecond)*1000*10 {
			LOG_WARN("slow-request: Req: %s, Hash: %d, Elapsed: %dus", req.ReqName, req.ReqId, elapsed)
		}
	}
}

func (this *RequestRecorder) Dump() {

	list := this.List()

	if 0 == len(list) {
		return
	}

	LOG_DEBUG("******************* request-recorder: %d *******************", len(list))
	for _, req := range list {
		LOG_DEBUG("Name:%s, Total: %d op, Elapsed: %dus, Min: %dus, Max: %dus, Avg: %dus",
			req.ReqName,
			req.ReqTotalCalled,
			req.ReqTotalElapsedTime,
			req.ReqMinElapsedTime,
			req.ReqMaxElapsedTime,
			req.ReqAvgElapsedTime)
	}
	LOG_DEBUG("***************************************************************")
}
