package codec

import (
	"errors"

	"google.golang.org/protobuf/proto"
)

// ProtoMessage protobuf消息接口
type ProtoMessage interface {
	proto.Message
	GetType() string
	GetData() interface{}
}

// ProtobufCodec protobuf编解码器
type ProtobufCodec struct{}

func NewProtobufCodec() *ProtobufCodec {
	return &ProtobufCodec{}
}

// DecodeProto 解码protobuf消息
func (c *ProtobufCodec) DecodeProto(data []byte, msg proto.Message) error {
	return proto.Unmarshal(data, msg)
}

// EncodeProto 编码protobuf消息
func (c *ProtobufCodec) EncodeProto(msg proto.Message) ([]byte, error) {
	if msg == nil {
		return nil, errors.New("message is nil")
	}
	return proto.Marshal(msg)
}

// Decode 实现ICodec接口 - 需要具体的proto消息类型
func (c *ProtobufCodec) Decode(data []byte) (IMessage, error) {
	return nil, errors.New("use DecodeProto with specific proto.Message type")
}

// Encode 实现ICodec接口
func (c *ProtobufCodec) Encode(message IMessage) (string, error) {
	if pm, ok := message.(proto.Message); ok {
		data, err := proto.Marshal(pm)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	return "", errors.New("message does not implement proto.Message")
}
