package codec

import (
	"errors"

	flatbuffers "github.com/google/flatbuffers/go"
)

// FlatBufferBuilder flatbuffers构建器别名
type FlatBufferBuilder = flatbuffers.Builder

// FlatbufferCodec flatbuffer编解码器
type FlatbufferCodec struct{}

func NewFlatbufferCodec() *FlatbufferCodec {
	return &FlatbufferCodec{}
}

// NewBuilder 创建新的flatbuffer构建器
func (c *FlatbufferCodec) NewBuilder(size int) *flatbuffers.Builder {
	return flatbuffers.NewBuilder(size)
}

// EncodeFlatbuffer 编码flatbuffer消息
func (c *FlatbufferCodec) EncodeFlatbuffer(builder *flatbuffers.Builder) []byte {
	return builder.FinishedBytes()
}

// DecodeFlatbuffer 获取flatbuffer根表
func (c *FlatbufferCodec) DecodeFlatbuffer(data []byte, offset flatbuffers.UOffsetT) flatbuffers.Table {
	var tab flatbuffers.Table
	tab.Bytes = data
	tab.Pos = offset
	return tab
}

// Decode 实现ICodec接口 - flatbuffer需要具体schema
func (c *FlatbufferCodec) Decode(data []byte) (IMessage, error) {
	return nil, errors.New("use DecodeFlatbuffer with specific schema type")
}

// Encode 实现ICodec接口
func (c *FlatbufferCodec) Encode(message IMessage) (string, error) {
	return "", errors.New("use EncodeFlatbuffer with FlatBufferBuilder")
}
