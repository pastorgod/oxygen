package command

import (
	"base/xnet"
	. "logger"
	"reflect"
)

func toName(obj interface{}) string {
	return reflect.TypeOf(obj).Elem().Name()
}

var CommandFactoryInstance = &CommandFactory{
	proto_factory: make(map[uint32]func() xnet.Message, 256),
	hash_names:    make(map[uint32]string, 256),
	name_hashs:    make(map[string]uint32, 256),
}

type CommandFactory struct {
	proto_factory map[uint32]func() xnet.Message
	hash_names    map[uint32]string
	name_hashs    map[string]uint32
}

func (this *CommandFactory) insert(key uint32, value func() xnet.Message) bool {
	if _, ok := this.proto_factory[key]; ok {
		return false
	}

	this.proto_factory[key] = value
	return true
}

func (this *CommandFactory) add_hash_name(key uint32, value string) {
	this.hash_names[key] = value
}

func (this *CommandFactory) add_name_hash(key string, value uint32) {
	this.name_hashs[key] = value
}

func RegisterProtoFactroy(factory *map[uint32]func() xnet.Message, hash_names *map[uint32]string, name_hashs *map[string]uint32) {

	// registe proto message factory
	for key, val := range *factory {
		if !CommandFactoryInstance.insert(key, val) {
			FATAL("already proto key: %d", key)
		}
	}

	for key, val := range *hash_names {
		CommandFactoryInstance.add_hash_name(key, val)
	}

	for key, val := range *name_hashs {
		CommandFactoryInstance.add_name_hash(key, val)
	}

	*factory = nil
	*hash_names = nil
	*name_hashs = nil
}

// implelement of ICommandFactory

func (this *CommandFactory) FindMsgNameByCode(opcode uint32) (string, bool) {

	if name, ok := this.hash_names[opcode]; ok {
		return name, true
	}

	return "<unknown>", false
}

func (this *CommandFactory) FindMsgCodeByName(name string) (uint32, bool) {

	if code, ok := this.name_hashs[name]; ok {
		return code, true
	}

	return 0, false
}

func (this *CommandFactory) NewMsgObjectByCode(opcode uint32) (xnet.Message, bool) {

	if fn, ok := this.proto_factory[opcode]; ok {
		return fn(), true
	}

	return nil, false
}

func (this *CommandFactory) ForEach(callback func(uint32, string)) {

	for code, name := range this.hash_names {
		callback(code, name)
	}
}

func (this *CommandFactory) GetResponseCode() uint32 {

	if code, ok := this.FindMsgCodeByName("Response"); ok {
		return code
	}

	FATAL("Invalid Response")
	return 0
}

func (this *CommandFactory) BuildResponse(errStr *string, pb xnet.Message, response_id uint32) xnet.Message {

	var dataBuf []byte
	var res_type = int32(-1)

	if pb != nil && !pb.IsNil() {

		if code, ok := CommandFactoryInstance.FindMsgCodeByName(toName(pb)); ok {
			res_type = int32(code)
		} else {
			LOG_ERROR("buildResponse.FindMsgCodeByObject %v", pb)
			return nil
		}

		var err error
		if dataBuf, err = pb.Marshal(); err != nil {
			LOG_ERROR("buildResponse: %s, %s", toName(pb), err.Error())
			return nil
		}
	}

	resp := &Response{
		Error: errStr,
		Bin:   dataBuf,
		Mtype: res_type,
		Id:    response_id,
	}

	return resp

}

func (this *CommandFactory) ParseResponse(resp xnet.Message) (uint32, *string, xnet.Message) {

	response := resp.(*Response)

	if -1 == response.GetMtype() {
		return response.GetId(), response.Error, nil
	}

	pb, ok := CommandFactoryInstance.NewMsgObjectByCode(uint32(response.GetMtype()))

	if !ok {
		LOG_ERROR("parseResponse notFound: %d", response.GetMtype())
		return response.GetId(), response.Error, nil
	}

	if err := pb.Unmarshal(response.Bin); err != nil {
		LOG_ERROR("parseResponse: %d %s, %s", response.GetMtype(), toName(pb), err.Error())
		return response.GetId(), response.Error, nil
	}

	return response.GetId(), response.Error, pb
}

func init() {
	xnet.Init(CommandFactoryInstance)
}
