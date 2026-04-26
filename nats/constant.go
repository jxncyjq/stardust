package nats

import (
	"errors"
	"github.com/nats-io/nats.go"
)

var (
	errAlreadySubscribed = errors.New("already subscribed to topic")
	errNotSubscribed     = errors.New("not subscribed")
	errEmptyTopic        = errors.New("empty topic")
)

// Config 数据库连接配置
type Config struct {
	Url   string `json:"url"`
	Topic string `json:"topic"`
}

type MNatsMessage nats.Msg
type CNatsConn nats.Conn
