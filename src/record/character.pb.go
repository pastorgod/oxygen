// Code generated by protoc-gen-gogo.
// source: record/character.proto
// DO NOT EDIT!

/*
	Package record is a generated protocol buffer package.

	It is generated from these files:
		record/character.proto

	It has these top-level messages:
		EquipAttrib
		RecordCharacter
		UpdateCharacterRequest
		UpdateCharacterResponse
*/
package record

import proto "github.com/gogo/protobuf/proto"
import math "math"
import "base/xnet"

// discarding unused import gogoproto "github.com/gogo/protobuf/gogoproto"

import command "command"
import db "db"

import "time"

import io "io"
import fmt "fmt"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf
var _ = db.C{}

type EquipAttrib struct {
	Damage           int32  `protobuf:"varint,1,opt,name=damage,def=0" json:"damage" form:"damage"`
	Hit              int32  `protobuf:"varint,2,opt,name=hit,def=0" json:"hit" form:"hit"`
	Defense          int32  `protobuf:"varint,3,opt,name=defense,def=0" json:"defense" form:"defense"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *EquipAttrib) Reset()         { *m = EquipAttrib{} }
func (m *EquipAttrib) String() string { return proto.CompactTextString(m) }
func (*EquipAttrib) ProtoMessage()    {}
func (m *EquipAttrib) IsNil() bool    { return nil == m || (*EquipAttrib)(nil) == m }

const Default_EquipAttrib_Damage int32 = 0
const Default_EquipAttrib_Hit int32 = 0
const Default_EquipAttrib_Defense int32 = 0

func (m *EquipAttrib) GetDamage() int32 {
	if m != nil {
		return m.Damage
	}
	return Default_EquipAttrib_Damage
}

func (m *EquipAttrib) GetHit() int32 {
	if m != nil {
		return m.Hit
	}
	return Default_EquipAttrib_Hit
}

func (m *EquipAttrib) GetDefense() int32 {
	if m != nil {
		return m.Defense
	}
	return Default_EquipAttrib_Defense
}

type RecordCharacter struct {
	Uid              uint32       `protobuf:"varint,1,opt,name=uid,def=0" json:"uid" form:"uid"`
	Name             *string      `protobuf:"bytes,2,opt,name=name" json:"name,omitempty" form:"name"`
	Level            int32        `protobuf:"varint,3,opt,name=level,def=0" json:"level" form:"level"`
	Role             int32        `protobuf:"varint,4,opt,name=role,def=0" json:"role" form:"role"`
	Attrib           *EquipAttrib `protobuf:"bytes,5,opt,name=attrib" json:"attrib,omitempty" form:"attrib"`
	XXX_unrecognized []byte       `json:"-"`
	_mask            uint64       `json:"-"`
}

func (m *RecordCharacter) Reset()         { *m = RecordCharacter{} }
func (m *RecordCharacter) String() string { return proto.CompactTextString(m) }
func (*RecordCharacter) ProtoMessage()    {}
func (m *RecordCharacter) IsNil() bool    { return nil == m || (*RecordCharacter)(nil) == m }

// generate db.IRecord implementation
func (*RecordCharacter) Table() string { return "Character" }

func (m *RecordCharacter) Key() db.C {
	return db.C{
		"uid": m.GetUid(),
	}
}

func (*RecordCharacter) IndexKey() []string {
	return []string{
		"uid",
	}
}

type RecordCharacterMask uint64

const (
	RecordCharacterMask_Uid    uint64 = 1
	RecordCharacterMask_Name   uint64 = 2
	RecordCharacterMask_Level  uint64 = 3
	RecordCharacterMask_Role   uint64 = 4
	RecordCharacterMask_Attrib uint64 = 5

	RecordCharacterMask_Keys uint64 = (1 << RecordCharacterMask_Uid)
)

func (this *RecordCharacter) Mask() uint64 {
	return this._mask
}

func (this *RecordCharacter) Dirty(masks ...uint64) {
	for _, t := range masks {
		xnet.Assert(t <= 5, "unknown mask", t)
		this._mask |= 1 << t
	}
}

func (this *RecordCharacter) Recover(mask uint64) {
	this._mask |= mask
}

func (this *RecordCharacter) ClearMask() {
	this._mask = 0
}

func (this *RecordCharacter) Flush() bool {
	// no changes.
	if 0 == this._mask {
		return true
	}
	// update changes to db.
	if UpdateFields(this, this.ToFields(this._mask)) {
		this._mask = 0
		return true
	}
	return false
}

func (this *RecordCharacter) MergeFrom(mask uint64, value db.IRecord) {
	data := value.(*RecordCharacter)
	this.update(mask, data)
	this._mask |= mask
}

func (this *RecordCharacter) CopyFrom(mask uint64, value db.IRecord) {
	data := value.(*RecordCharacter)
	this.update(mask|RecordCharacterMask_Keys, data)
}

func (this *RecordCharacter) update(mask uint64, value *RecordCharacter) {
	for i := uint(1); i <= 5; i++ {
		if 0 == (mask & (1 << i)) {
			continue
		}
		switch uint64(i) {
		case RecordCharacterMask_Uid:
			this.Uid = value.Uid
		case RecordCharacterMask_Name:
			this.Name = value.Name
		case RecordCharacterMask_Level:
			this.Level = value.Level
		case RecordCharacterMask_Role:
			this.Role = value.Role
		case RecordCharacterMask_Attrib:
			this.Attrib = value.Attrib
		default:
			panic(fmt.Sprintf("unknown field: RecordCharacter, %d", i))
		}
	}
}
func (this *RecordCharacter) ToFields(mask uint64) []string {
	var all_fields = [...]string{
		"-",
		"Uid",
		"Name",
		"Level",
		"Role",
		"Attrib",
	}
	fields := make([]string, 0, 5)
	for i := uint(1); i <= 5; i++ {
		if 0 != (mask & (1 << i)) {
			if field := all_fields[i]; "-" != field {
				fields = append(fields, field)
			}
		}
	}
	return fields
}

const Default_RecordCharacter_Uid uint32 = 0
const Default_RecordCharacter_Level int32 = 0
const Default_RecordCharacter_Role int32 = 0

func (m *RecordCharacter) GetUid() uint32 {
	if m != nil {
		return m.Uid
	}
	return Default_RecordCharacter_Uid
}

func (m *RecordCharacter) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

func (m *RecordCharacter) GetLevel() int32 {
	if m != nil {
		return m.Level
	}
	return Default_RecordCharacter_Level
}

func (m *RecordCharacter) GetRole() int32 {
	if m != nil {
		return m.Role
	}
	return Default_RecordCharacter_Role
}

func (m *RecordCharacter) GetAttrib() *EquipAttrib {
	if m != nil {
		return m.Attrib
	}
	return nil
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////
type UpdateCharacterRequest struct {
	Data             *RecordCharacter `protobuf:"bytes,1,opt,name=data" json:"data,omitempty" form:"data"`
	Mask             uint64           `protobuf:"varint,2,opt,name=mask,def=0" json:"mask" form:"mask"`
	XXX_unrecognized []byte           `json:"-"`
}

func (m *UpdateCharacterRequest) Reset()         { *m = UpdateCharacterRequest{} }
func (m *UpdateCharacterRequest) String() string { return proto.CompactTextString(m) }
func (*UpdateCharacterRequest) ProtoMessage()    {}
func (m *UpdateCharacterRequest) IsNil() bool    { return nil == m || (*UpdateCharacterRequest)(nil) == m }

const Default_UpdateCharacterRequest_Mask uint64 = 0

func (m *UpdateCharacterRequest) GetData() *RecordCharacter {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *UpdateCharacterRequest) GetMask() uint64 {
	if m != nil {
		return m.Mask
	}
	return Default_UpdateCharacterRequest_Mask
}

type UpdateCharacterResponse struct {
	XXX_unrecognized []byte `json:"-"`
}

func (m *UpdateCharacterResponse) Reset()         { *m = UpdateCharacterResponse{} }
func (m *UpdateCharacterResponse) String() string { return proto.CompactTextString(m) }
func (*UpdateCharacterResponse) ProtoMessage()    {}
func (m *UpdateCharacterResponse) IsNil() bool {
	return nil == m || (*UpdateCharacterResponse)(nil) == m
}

var character_factory = map[uint32]func() xnet.Message{
	356866186:  func() xnet.Message { return &EquipAttrib{} },
	388761048:  func() xnet.Message { return &RecordCharacter{} },
	1307013523: func() xnet.Message { return &UpdateCharacterRequest{} },
	2114510413: func() xnet.Message { return &UpdateCharacterResponse{} },
}

var character_hash_names = map[uint32]string{
	356866186:  "EquipAttrib",
	388761048:  "RecordCharacter",
	1307013523: "UpdateCharacterRequest",
	2114510413: "UpdateCharacterResponse",
}

var character_name_hashs = map[string]uint32{
	"EquipAttrib":             356866186,
	"RecordCharacter":         388761048,
	"UpdateCharacterRequest":  1307013523,
	"UpdateCharacterResponse": 2114510413,
}

func init() {
	// character.proto
	command.RegisterProtoFactroy(&character_factory, &character_hash_names, &character_name_hashs)
}
func (m *EquipAttrib) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *EquipAttrib) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Damage != Default_EquipAttrib_Damage {
		data[i] = 0x8
		i++
		i = encodeVarintCharacter(data, i, uint64(m.Damage))
	}
	if m.Hit != Default_EquipAttrib_Hit {
		data[i] = 0x10
		i++
		i = encodeVarintCharacter(data, i, uint64(m.Hit))
	}
	if m.Defense != Default_EquipAttrib_Defense {
		data[i] = 0x18
		i++
		i = encodeVarintCharacter(data, i, uint64(m.Defense))
	}
	if m.XXX_unrecognized != nil {
		i += copy(data[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func (m *RecordCharacter) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *RecordCharacter) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Uid != Default_RecordCharacter_Uid {
		data[i] = 0x8
		i++
		i = encodeVarintCharacter(data, i, uint64(m.Uid))
	}
	if m.Name != nil {
		data[i] = 0x12
		i++
		i = encodeVarintCharacter(data, i, uint64(len(*m.Name)))
		i += copy(data[i:], *m.Name)
	}
	if m.Level != Default_RecordCharacter_Level {
		data[i] = 0x18
		i++
		i = encodeVarintCharacter(data, i, uint64(m.Level))
	}
	if m.Role != Default_RecordCharacter_Role {
		data[i] = 0x20
		i++
		i = encodeVarintCharacter(data, i, uint64(m.Role))
	}
	if m.Attrib != nil {
		data[i] = 0x2a
		i++
		i = encodeVarintCharacter(data, i, uint64(m.Attrib.Size()))
		n1, err := m.Attrib.MarshalTo(data[i:])
		if err != nil {
			return 0, err
		}
		i += n1
	}
	if m.XXX_unrecognized != nil {
		i += copy(data[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func (m *UpdateCharacterRequest) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *UpdateCharacterRequest) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Data != nil {
		data[i] = 0xa
		i++
		i = encodeVarintCharacter(data, i, uint64(m.Data.Size()))
		n2, err := m.Data.MarshalTo(data[i:])
		if err != nil {
			return 0, err
		}
		i += n2
	}
	if m.Mask != Default_UpdateCharacterRequest_Mask {
		data[i] = 0x10
		i++
		i = encodeVarintCharacter(data, i, uint64(m.Mask))
	}
	if m.XXX_unrecognized != nil {
		i += copy(data[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func (m *UpdateCharacterResponse) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *UpdateCharacterResponse) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i += copy(data[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func encodeFixed64Character(data []byte, offset int, v uint64) int {
	data[offset] = uint8(v)
	data[offset+1] = uint8(v >> 8)
	data[offset+2] = uint8(v >> 16)
	data[offset+3] = uint8(v >> 24)
	data[offset+4] = uint8(v >> 32)
	data[offset+5] = uint8(v >> 40)
	data[offset+6] = uint8(v >> 48)
	data[offset+7] = uint8(v >> 56)
	return offset + 8
}
func encodeFixed32Character(data []byte, offset int, v uint32) int {
	data[offset] = uint8(v)
	data[offset+1] = uint8(v >> 8)
	data[offset+2] = uint8(v >> 16)
	data[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintCharacter(data []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		data[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	data[offset] = uint8(v)
	return offset + 1
}

type ICharacterService interface {
	Update(context *xnet.Context, in *UpdateCharacterRequest, out *UpdateCharacterResponse) *string
}

// generate CharacterServiceModule to implementation IServiceModule.
type CharacterServiceModule struct {
	ICharacterService
}

func NewCharacterServiceModule(svc ICharacterService) *CharacterServiceModule {
	return &CharacterServiceModule{ICharacterService: svc}
}

func (this *CharacterServiceModule) Name() string {
	return "CharacterService"
}

func (this *CharacterServiceModule) Impl() interface{} {
	return this.ICharacterService
}

func (this *CharacterServiceModule) ServiceCall(ctx *xnet.Context, rpc_id uint32, in xnet.Message) (reply xnet.Message, err *string) {

	switch rpc_id {
	case uint32(1149662859):
		out := &UpdateCharacterResponse{}
		reply, err = out, this.ICharacterService.Update(ctx, in.(*UpdateCharacterRequest), out)

	default:
		err = proto.String("MethodNotFound")
	}

	return
}

// generate CharacterServiceImpl
type CharacterServiceImpl struct {
	*xnet.RpcService
	xnet.IServiceModule
}

// NewCharacterServiceServer returns a new CharacterService Server.
func NewCharacterServiceImpl(rawurl string, impl ICharacterService) (*CharacterServiceImpl, error) {
	var service *xnet.RpcService
	var err error
	if service, err = xnet.ListenRpc(rawurl); err != nil {
		return nil, err
	}
	return NewCharacterServiceImplFrom(service, impl)
}

// NewCharacterServiceServer use proxy.
func NewCharacterServiceImplWithProxy(rawurl, proxy_url string) (*CharacterServiceImpl, error) {

	// dial remote service with proxy_url.
	impl, err := NewCharacterServiceProxy(proxy_url, time.Second*15)
	// dial remote service to failed.
	if err != nil {
		return nil, err
	}

	// create local service with rawurl.
	var service *CharacterServiceImpl
	if service, err = NewCharacterServiceImpl(rawurl, impl); nil != err {
		impl.proxy.Close(err)
		return nil, err
	}

	return service, nil
}

func NewCharacterServiceImplFrom(service *xnet.RpcService, impl ICharacterService) (*CharacterServiceImpl, error) {
	svc_impl := &CharacterServiceImpl{
		RpcService:     service,
		IServiceModule: NewCharacterServiceModule(impl),
	}

	service.RegisterService(svc_impl.IServiceModule)

	return svc_impl, nil
}

func (this *CharacterServiceImpl) Startup() {
	this.Dispatcher().RegisterService(this.IServiceModule)
}

func (this *CharacterServiceImpl) Shutdown() {
	this.Dispatcher().UnregisterService(this.IServiceModule)
}

// define ICharacterServiceClient interface
type ICharacterServiceClient interface {
	xnet.ISession

	Update(*UpdateCharacterRequest) (*UpdateCharacterResponse, *string)
	AsyncUpdate(*UpdateCharacterRequest, func(*string, *UpdateCharacterResponse))
}

// implement ICharacterServiceClient interface.
type CharacterServiceClient struct {
	xnet.ISession
}

// New CharacterServiceClient from ISession.
func NewCharacterServiceClient(conn xnet.ISession) ICharacterServiceClient {
	return &CharacterServiceClient{conn}
}

// DialCharacterService connects to an CharacterService at the specified network address.
func DialCharacterService(rawurl string) (ICharacterServiceClient, error) {
	return DialCharacterServiceTimeout(rawurl, time.Second*15)
}

// DialCharacterService connects to an CharacterService at the specified network address.
func DialCharacterServiceTimeout(rawurl string, timeout time.Duration) (ICharacterServiceClient, error) {
	c, err := xnet.DialRpc(rawurl, timeout)
	if err != nil {
		return nil, err
	}
	return &CharacterServiceClient{c}, nil
}

func (this *CharacterServiceClient) Update(in *UpdateCharacterRequest) (*UpdateCharacterResponse, *string) {
	xnet.Assert(nil != in, "nil pointer: *UpdateCharacterRequest")
	// CharacterService.Update = 1149662859
	// reply, err := this.Call( "CharacterService.Update", in )
	reply, err := this.CallTimeout(1149662859, xnet.RPC_TIMEOUT, in)
	if nil != reply && (*UpdateCharacterResponse)(nil) != reply {
		return reply.(*UpdateCharacterResponse), err
	}
	return nil, err
}

func (this *CharacterServiceClient) AsyncUpdate(in *UpdateCharacterRequest, handler func(*string, *UpdateCharacterResponse)) {
	xnet.Assert(nil != in, "nil pointer: *UpdateCharacterRequest")
	xnet.Assert(nil != handler, "nil handler: func(*string, *UpdateCharacterResponse)")
	// CharacterService.Update = 1149662859
	// this.AsyncCall("CharacterService.Update", in, func(err *string, reply xnet.Message) {
	this.AsyncCallTimeout(1149662859, xnet.RPC_TIMEOUT, in, func(err *string, reply xnet.Message) {
		var output *UpdateCharacterResponse
		if nil != reply && (*UpdateCharacterResponse)(nil) != reply {
			output = reply.(*UpdateCharacterResponse)
		}
		handler(err, output)
	})
}

// generate CharacterServiceProxy to implementation ICharacterService interface.
type CharacterServiceProxy struct {
	proxy ICharacterServiceClient
}

func NewCharacterServiceProxy(rawurl string, timeout time.Duration) (*CharacterServiceProxy, error) {
	client, err := DialCharacterServiceTimeout(rawurl, timeout)
	if nil != err {
		return nil, err
	}
	return &CharacterServiceProxy{proxy: client}, nil
}

// ICharacterService.Update
func (this *CharacterServiceProxy) Update(ctx *xnet.Context, in *UpdateCharacterRequest, out *UpdateCharacterResponse) *string {
	this.proxy.AsyncUpdate(in, func(err *string, resp *UpdateCharacterResponse) {
		ctx.Response(err, resp)
	})
	return ctx.Asynchronized()
}

func (m *EquipAttrib) Size() (n int) {
	var l int
	_ = l
	if m.Damage != Default_EquipAttrib_Damage {
		n += 1 + sovCharacter(uint64(m.Damage))
	}
	if m.Hit != Default_EquipAttrib_Hit {
		n += 1 + sovCharacter(uint64(m.Hit))
	}
	if m.Defense != Default_EquipAttrib_Defense {
		n += 1 + sovCharacter(uint64(m.Defense))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *RecordCharacter) Size() (n int) {
	var l int
	_ = l
	if m.Uid != Default_RecordCharacter_Uid {
		n += 1 + sovCharacter(uint64(m.Uid))
	}
	if m.Name != nil {
		l = len(*m.Name)
		n += 1 + l + sovCharacter(uint64(l))
	}
	if m.Level != Default_RecordCharacter_Level {
		n += 1 + sovCharacter(uint64(m.Level))
	}
	if m.Role != Default_RecordCharacter_Role {
		n += 1 + sovCharacter(uint64(m.Role))
	}
	if m.Attrib != nil {
		l = m.Attrib.Size()
		n += 1 + l + sovCharacter(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *UpdateCharacterRequest) Size() (n int) {
	var l int
	_ = l
	if m.Data != nil {
		l = m.Data.Size()
		n += 1 + l + sovCharacter(uint64(l))
	}
	if m.Mask != Default_UpdateCharacterRequest_Mask {
		n += 1 + sovCharacter(uint64(m.Mask))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *UpdateCharacterResponse) Size() (n int) {
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovCharacter(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozCharacter(x uint64) (n int) {
	return sovCharacter(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *EquipAttrib) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Damage", wireType)
			}
			m.Damage = 0
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				m.Damage |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Hit", wireType)
			}
			m.Hit = 0
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				m.Hit |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Defense", wireType)
			}
			m.Defense = 0
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				m.Defense |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			var sizeOfWire int
			for {
				sizeOfWire++
				wire >>= 7
				if wire == 0 {
					break
				}
			}
			iNdEx -= sizeOfWire
			skippy, err := skipCharacter(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCharacter
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, data[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	return nil
}
func (m *RecordCharacter) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Uid", wireType)
			}
			m.Uid = 0
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				m.Uid |= (uint32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := iNdEx + int(stringLen)
			if stringLen < 0 {
				return ErrInvalidLengthCharacter
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			s := string(data[iNdEx:postIndex])
			m.Name = &s
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Level", wireType)
			}
			m.Level = 0
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				m.Level |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Role", wireType)
			}
			m.Role = 0
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				m.Role |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Attrib", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := iNdEx + msglen
			if msglen < 0 {
				return ErrInvalidLengthCharacter
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Attrib == nil {
				m.Attrib = &EquipAttrib{}
			}
			if err := m.Attrib.Unmarshal(data[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			var sizeOfWire int
			for {
				sizeOfWire++
				wire >>= 7
				if wire == 0 {
					break
				}
			}
			iNdEx -= sizeOfWire
			skippy, err := skipCharacter(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCharacter
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, data[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	return nil
}
func (m *UpdateCharacterRequest) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := iNdEx + msglen
			if msglen < 0 {
				return ErrInvalidLengthCharacter
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Data == nil {
				m.Data = &RecordCharacter{}
			}
			if err := m.Data.Unmarshal(data[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Mask", wireType)
			}
			m.Mask = 0
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				m.Mask |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			var sizeOfWire int
			for {
				sizeOfWire++
				wire >>= 7
				if wire == 0 {
					break
				}
			}
			iNdEx -= sizeOfWire
			skippy, err := skipCharacter(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCharacter
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, data[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	return nil
}
func (m *UpdateCharacterResponse) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		switch fieldNum {
		default:
			var sizeOfWire int
			for {
				sizeOfWire++
				wire >>= 7
				if wire == 0 {
					break
				}
			}
			iNdEx -= sizeOfWire
			skippy, err := skipCharacter(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCharacter
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, data[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	return nil
}
func skipCharacter(data []byte) (n int, err error) {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for {
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if data[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthCharacter
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := data[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipCharacter(data[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthCharacter = fmt.Errorf("proto: negative length found during unmarshaling")
)
