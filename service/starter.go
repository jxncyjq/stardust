package service

import (
	"go.uber.org/zap"
)

// Startable 可启动的服务器接口（HttpServer 和 GrpcServer 都实现了）
type Startable interface {
	Startup() error
	Stop()
}

// ServerStarter 将 Startable 适配为 Service 接口
type ServerStarter struct {
	server Startable
	logger *zap.Logger
}

func NewServerStarter(server Startable) *ServerStarter {
	return &ServerStarter{
		server: server,
		logger: getLoggerSafe("server_starter"),
	}
}

func (s *ServerStarter) Start() {
	if err := s.server.Startup(); err != nil {
		if s.logger != nil {
			s.logger.Fatal("failed to start server", zap.Error(err))
		}
	}
}

func (s *ServerStarter) Stop() {
	s.server.Stop()
}
