package codec

import (
	"github.com/jxncyjq/stardust/utils"
)

// IMessage 消息接口
type IMessage interface {
	GetType() string
	GetData() interface{}
}

// Message 消息结构体
type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func (m *Message) GetType() string    { return m.Type }
func (m *Message) GetData() interface{} { return m.Data }

// IMessageProcessor 消息处理器接口
type IMessageProcessor interface {
	HandlerMessage(message IMessage) (string, error)
}

type ICodec interface {
	Decode(data []byte) (IMessage, error)
	Encode(message IMessage) (string, error)
}

type JsonCodec struct{}

func NewJsonCodec() ICodec {
	return &JsonCodec{}
}

func (c *JsonCodec) Decode(data []byte) (IMessage, error) {
	msg, err := utils.Bytes2Struct[Message](data)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (c *JsonCodec) Encode(message IMessage) (string, error) {
	return utils.Struct2Bytes(message)
}
