package httpServer

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jxncyjq/stardust/logs"
	"github.com/jxncyjq/stardust/utils"
	"go.uber.org/zap"
)

type Backend struct {
	config     HttpServerConfig
	Ctx        context.Context
	Logger     *zap.Logger
	httpServer *HttpServer
}

type HandlerFunc = gin.HandlerFunc

func NewBackend(config []byte) (*Backend, error) {
	httpServer, err := NewHttpServer(config)
	if err != nil {
		return nil, err
	}
	configStruct, err := utils.Bytes2Struct[HttpServerConfig](config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTTP server configuration: %w", err)
	}
	return &Backend{
		config:     configStruct,
		Ctx:        context.Background(),
		Logger:     logs.GetLogger("http_backend"),
		httpServer: httpServer,
	}, nil
}

func (m *Backend) AddGroup(group string, middleware ...HandlerFunc) {
	m.httpServer.AddGroup(group, middleware...)
}

func (m *Backend) Post(group string, h IHandler) {
	m.httpServer.Post(h.GetName(), group, h)
}

func (m *Backend) Get(group string, h IHandler) {
	m.httpServer.Get(h.GetName(), group, h)
}

func (m *Backend) Put(group string, h IHandler) {
	m.httpServer.Put(h.GetName(), group, h)
}

func (m *Backend) Patch(group string, h IHandler) {
	m.httpServer.Patch(h.GetName(), group, h)
}

func (m *Backend) Delete(group string, h IHandler) {
	m.httpServer.Delete(h.GetName(), group, h)
}

func (m *Backend) Head(group string, h IHandler) {
	m.httpServer.Head(h.GetName(), group, h)
}

func (m *Backend) Options(group string, h IHandler) {
	m.httpServer.Options(h.GetName(), group, h)
}

func (m *Backend) Connect(group string, h IHandler) {
	m.httpServer.Connect(h.GetName(), group, h)
}

func (m *Backend) Trace(group string, h IHandler) {
	m.httpServer.Trace(h.GetName(), group, h)
}

func (m *Backend) AddHandler(method, path string, h IHandler) {
	m.httpServer.Handle(method, path, h)
}

func (m *Backend) AddNativeHandler(method string, path string, handler HandlerFunc) {
	m.httpServer.AddNativeHandler(method, path, handler)
}

func (m *Backend) Start() error {
	return m.httpServer.Startup()
}

func (m *Backend) Stop() {
	if err := m.httpServer.WaitForShutdown(); err != nil {
		m.Logger.Error("http server shutdown error", zap.Error(err))
	}
}
