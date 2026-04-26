package httpServer

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/jxncyjq/stardust/http_server/middleware"
	"github.com/jxncyjq/stardust/logs"
	"github.com/jxncyjq/stardust/utils"
	"github.com/jxncyjq/stardust/uuid"
	"go.uber.org/zap"
)

type HttpServer struct {
	ctx      context.Context
	cancel   context.CancelFunc
	stopOnce sync.Once
	addr     string
	path     string
	logger   *zap.Logger
	engine   *gin.Engine
	server   *http.Server
	group    map[string]*StarDustGroup
}

func NewHttpServer(configByte []byte) (*HttpServer, error) {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	config, err := utils.Bytes2Struct[HttpServerConfig](configByte)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTTP server configuration: %w", err)
	}

	// 初始化 workerID，用于生成唯一 sessionId
	uuid.InitWorker(config.WorkerID)

	// 设置自定义验证器
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v // validator is ready
	}

	addr := fmt.Sprintf("%s:%d", config.Address, config.Port)
	if config.Cors {
		engine.Use(middleware.Cors())
	}
	if config.RequestLog {
		engine.Use(middleware.Request())
	}
	if config.Access {
		engine.Use(middleware.Access())
	}

	if config.Path != "" && config.Path[0] != '/' {
		return nil, errors.New("the http.path must start with a /")
	}

	ctx, cancel := context.WithCancel(context.Background())
	srv := &HttpServer{
		ctx:    ctx,
		cancel: cancel,
		logger: logs.GetLogger("httpServer"),
		engine: engine,
		group:  make(map[string]*StarDustGroup),
		addr:   addr,
		path:   config.Path,
	}
	return srv, nil
}

func (m *HttpServer) Engine() *gin.Engine {
	return m.engine
}

func (m *HttpServer) Use(middleware ...gin.HandlerFunc) *HttpServer {
	m.engine.Use(middleware...)
	return m
}

func (m *HttpServer) Startup() error {
	m.logger.Info("http server listened on:", zap.String("addr", m.addr))
	// 打印路由
	for _, route := range m.engine.Routes() {
		m.logger.Info("http route registered:", logs.String("method", route.Method), logs.String("path", route.Path))
	}

	m.server = &http.Server{
		Addr:    m.addr,
		Handler: m.engine,
	}

	go func() {
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			m.logger.Error("http server error:", zap.Error(err))
		}
	}()
	go func() {
		<-m.ctx.Done()
		m.Stop()
	}()
	return nil
}

func (m *HttpServer) Stop() {
	m.stopOnce.Do(func() {
		m.cancel() // 通知监听 ctx 的 goroutine 退出
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := m.server.Shutdown(ctx); err != nil {
			m.logger.Error("shutdown http server:", zap.Error(err))
			return
		}
		m.logger.Info("http server shutdown gracefully")
	})
}

// Handle registers a new route with the HTTP server.
func (m *HttpServer) Handle(method string, path string, handler IHandler) {
	path, _ = url.JoinPath(m.path, "api", path)
	m.engine.Handle(method, path, handler.GetFunc())
}

func (m *HttpServer) Internal(method string, path string, handler IHandler) {
	path, _ = url.JoinPath(m.path, "internal", path)
	m.engine.Handle(method, path, handler.GetFunc())
}

func (m *HttpServer) AddGroup(path string, middleware ...gin.HandlerFunc) {
	url_path, _ := url.JoinPath(m.path, "api", path)
	m.group[path] = NewStarDustGroup(path, m.engine.Group(url_path, middleware...))
	m.logger.Info("http group registered:", logs.String("path", url_path))
}

func (m *HttpServer) Get(path string, group string, handler IHandler) {
	if group != "" {
		if _, exists := m.group[group]; !exists {
			panic(fmt.Sprintf("http server: group %q not found, call AddGroup first", group))
		}
		m.group[group].Group.GET(fmt.Sprintf("/%s", path), handler.GetFunc())
		m.logger.Info("http handler registered to group:", logs.String("path", path), logs.String("prefix", m.group[group].Prefix))
		return
	}
	m.Handle(http.MethodGet, path, handler)
}

func (m *HttpServer) Post(path string, group string, handler IHandler) {
	if group != "" {
		if _, exists := m.group[group]; !exists {
			panic(fmt.Sprintf("http server: group %q not found, call AddGroup first", group))
		}
		m.group[group].Group.POST(fmt.Sprintf("/%s", path), handler.GetFunc())
		m.logger.Info("http handler registered to group:", logs.String("path", path), logs.String("prefix", m.group[group].Prefix))
		return
	}
	m.Handle(http.MethodPost, path, handler)
}

func (m *HttpServer) Put(path string, group string, handler IHandler) {
	if group != "" {
		if _, exists := m.group[group]; !exists {
			panic(fmt.Sprintf("http server: group %q not found, call AddGroup first", group))
		}
		m.group[group].Group.PUT(fmt.Sprintf("/%s", path), handler.GetFunc())
		m.logger.Info("http handler registered to group:", logs.String("path", path), logs.String("prefix", m.group[group].Prefix))
		return
	}
	m.Handle(http.MethodPut, path, handler)
}

func joinHTTPPath(parts ...string) string {
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.Trim(part, "/")
		if trimmed == "" {
			continue
		}
		segments = append(segments, trimmed)
	}
	if len(segments) == 0 {
		return "/"
	}
	return "/" + strings.Join(segments, "/")
}

func (m *HttpServer) AddNativeHandler(method string, path string, handler gin.HandlerFunc) {
	path = joinHTTPPath(m.path, "api", path)
	m.engine.Handle(method, path, handler)
	m.logger.Info("http native handler registered:", logs.String("method", method), logs.String("path", path))
}

// RegisterHealthCheck 注册健康检查接口
func (m *HttpServer) RegisterHealthCheck() {
	m.engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	m.logger.Info("health check endpoint registered: /health")
}

// WaitForShutdown 等待关闭信号并执行优雅关闭
func (m *HttpServer) WaitForShutdown() error {
	// 创建信号通道
	quit := make(chan os.Signal, 1)

	// 监听指定的信号量：SIGINT (Ctrl+C), SIGTERM
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	sig := <-quit
	m.logger.Info("received shutdown signal:", zap.String("signal", sig.String()))

	// 创建带超时的上下文用于优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	m.logger.Info("shutting down server...")

	// 执行优雅关闭
	if err := m.server.Shutdown(ctx); err != nil {
		m.logger.Error("server forced to shutdown:", zap.Error(err))
		return err
	}

	m.logger.Info("server exited gracefully")
	return nil
}
