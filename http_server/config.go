package httpServer

import "errors"

// ServerMode 服务器模式
type ServerMode string

const (
	ModeGin  ServerMode = "gin"
	ModeGrpc ServerMode = "grpc"
)

type HttpServerConfig struct {
	Port       int        `json:"port" yaml:"port"`               // HTTP服务器端口
	Address    string     `json:"address" yaml:"address"`         // HTTP服务器主机名
	Path       string     `json:"path" yaml:"path"`               // HTTP服务器路径
	Cors       bool       `json:"cors" yaml:"cors"`               // 是否启用CORS
	RequestLog bool       `json:"request_log" yaml:"request_log"` // 是否启用请求日志
	Access     bool       `json:"access" yaml:"access"`           // 是否启用访问日志
	WorkerID   int64      `json:"worker_id" yaml:"worker_id"`     // Snowflake workerID (0-1023)，用于生成唯一sessionId
	Mode       ServerMode `json:"mode" yaml:"mode"`               // 服务器模式: gin 或 grpc
}

func (c *HttpServerConfig) Validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return errors.New("invalid port: must be between 1 and 65535")
	}
	if c.Path != "" && c.Path[0] != '/' {
		return errors.New("path must start with /")
	}
	return nil
}

type HttpWebSocketConfig struct {
	ReadBufferSize    int `json:"read_buffer_size" yaml:"read_buffer_size"`   // WebSocket读取缓冲区大小
	WriteBufferSize   int `json:"write_buffer_size" yaml:"write_buffer_size"` // WebSocket写入缓冲区大小
	TickerTime        int `json:"ticker_time" yaml:"ticker_time"`             // WebSocket心跳间隔时间
	WriteDeadlineTime int `json:"deadline_time" yaml:"deadline_time"`         // WebSocket连接超时时间
	ReadLimit         int `json:"read_limit" yaml:"read_limit"`               // WebSocket读取限制
	HandshakeTimeout  int `json:"handshake_timeout" yaml:"handshake_timeout"` // Web
}
