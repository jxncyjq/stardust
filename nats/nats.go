package nats

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/jxncyjq/stardust/logs"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type Subscription = nats.Subscription
type Msg = nats.Msg

type NatsConnection struct {
	conn        *nats.Conn
	config      *NatsConfig
	name        string
	jsMu        sync.RWMutex
	js          nats.JetStreamContext
	streamInfo  *nats.StreamInfo
	subject     []*nats.Subscription
	useStream   bool
	logger      *zap.Logger
	url         string
	stopChan    chan struct{}
	ctx         context.Context
	cancel      context.CancelFunc
	messageChan chan *nats.Msg
	handlersMu  sync.Mutex
	handlers    map[string]func(*nats.Msg)
}

// NatsConfig NATS配置结构体
type NatsConfig struct {
	Name       string   `json:"name"`
	Url        string   `json:"url"`
	UseStream  bool     `json:"use_stream"`
	StreamName string   `json:"stream_name"`
	Type       string   `json:"type"`
	Subject    []string `json:"subjects"`
	Username   string   `json:"username"` // 新增用户名
	Password   string   `json:"password"` // 新增密码
}

func (c *NatsConfig) Validate() error {
	if c.Url == "" {
		return errors.New("nats url is required")
	}
	if c.UseStream && c.StreamName == "" {
		return errors.New("stream name is required when use_stream is enabled")
	}
	return nil
}

func NewNatsConnect(name string, natsConfig *NatsConfig) (*NatsConnection, error) {
	if natsConfig == nil {
		return nil, errors.New("NATS 配置未初始化")
	}
	var opts []nats.Option
	opts = append(opts,
		nats.MaxReconnects(10),
		nats.ReconnectWait(5*time.Second),
	)
	if natsConfig.Username != "" && natsConfig.Password != "" {
		opts = append(opts, nats.UserInfo(natsConfig.Username, natsConfig.Password))
	}
	conn, err := nats.Connect(natsConfig.Url, opts...)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	sub := &NatsConnection{
		name:        name,
		conn:        conn,
		config:      natsConfig,
		useStream:   natsConfig.UseStream,
		logger:      logs.GetLogger("nats"),
		url:         natsConfig.Url,
		stopChan:    make(chan struct{}),
		ctx:         ctx,
		cancel:      cancel,
		messageChan: make(chan *nats.Msg, 64),
		handlers:    make(map[string]func(*nats.Msg)),
	}
	if natsConfig.UseStream {
		js, err := conn.JetStream(nats.PublishAsyncMaxPending(100))
		if err != nil {
			conn.Close()
			return nil, err
		}
		sub.js = js
	}
	return sub, nil
}

func (s *NatsConnection) IsConnected() bool {
	return s.conn.IsConnected()
}

func (s *NatsConnection) GetConfig() *NatsConfig {
	return s.config
}

// getJS 线程安全地获取 JetStream 上下文。
func (s *NatsConnection) getJS() nats.JetStreamContext {
	s.jsMu.RLock()
	defer s.jsMu.RUnlock()
	return s.js
}

func (s *NatsConnection) EnsureStream() error {
	if !s.useStream {
		return nil // 如果未启用 JetStream，则无需创建 Stream
	}
	return s.ensureStreamSubjects(s.config.StreamName, s.config.Subject)
}
