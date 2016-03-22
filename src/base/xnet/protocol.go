package xnet

import proto "github.com/gogo/protobuf/proto"

// 消息接口
type Message interface {
	// google protobuf
	proto.Message

	// gogoprotobuf interface.
	// 序列化数据到新的buf
	Marshal() ([]byte, error)

	// 序列化数据到已存在的buf
	MarshalTo([]byte) (int, error)

	// 从buf中反序列化数据
	Unmarshal([]byte) error

	// 当前的对象所需空间
	Size() int

	// 判断对象是否有效
	IsNil() bool
}
