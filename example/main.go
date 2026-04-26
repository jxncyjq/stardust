// Package main 演示手动初始化 stardust.mini 全部组件的完整启动流程。
//
// 启动前准备：
//  1. 复制 config.toml 并按实际环境修改连接参数
//  2. 设置配置文件路径：export runConfig=./example/config.toml
//  3. 启动依赖服务（Redis / MySQL / Jaeger 等）
//
// 初始化顺序：conf → logs → tracing → redis → databases → mongodb → clickhouse → nats → http/grpc
package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jxncyjq/stardust/clickhouse"
	"github.com/jxncyjq/stardust/conf"
	"github.com/jxncyjq/stardust/databases"
	httpServer "github.com/jxncyjq/stardust/http_server"
	"github.com/jxncyjq/stardust/http_server/middleware"
	"github.com/jxncyjq/stardust/logs"
	"github.com/jxncyjq/stardust/mongodb"
	"github.com/jxncyjq/stardust/nats"
	"github.com/jxncyjq/stardust/redis"
	"github.com/jxncyjq/stardust/service"
	"github.com/jxncyjq/stardust/tracing"
	"go.uber.org/zap"
)

func main() {
	// ── 0. 配置文件路径 ──────────────────────────────────────
	// 优先级：环境变量 runConfig > devConf（ISDEBUG=1）> prodConf
	// 此处仅作演示兜底，实际通过 export runConfig=./example/config.toml 设置
	if os.Getenv("runConfig") == "" {
		os.Setenv("runConfig", "./example/config.toml")
	}

	// ── 1. 加载配置 ──────────────────────────────────────────
	// conf.Init() 读取 TOML/YAML 配置文件，后续通过 conf.Get(section) 取各段 JSON
	conf.Init()

	// ── 2. 日志（最先初始化，其他组件依赖它）────────────────
	// conf.Get("logs") 返回 [logs] 段的 JSON 字节
	if err := logs.Init(conf.Get("logs")); err != nil {
		panic("日志初始化失败: " + err.Error())
	}
	logger := logs.GetLogger("main")
	logger.Info("配置加载完成",
		zap.String("app", conf.GetAppName()),
		zap.String("version", conf.GetAppVersion()),
	)

	// ── 3. 链路追踪（Jaeger OTLP HTTP）──────────────────────
	// conf.Get("tracing") 返回 JaegerConfig JSON
	tracer, err := tracing.NewJaegerTracer(conf.Get("tracing"))
	if err != nil {
		logger.Fatal("链路追踪初始化失败", zap.Error(err))
	}
	defer tracer.Shutdown(context.Background())
	logger.Info("链路追踪已就绪")

	// ── 4. Redis（多实例）────────────────────────────────────
	// conf.Get("redis") 返回 [{...},{...}] 数组 JSON（TOML 中使用 [[redis]]）
	// 单实例时也兼容 {...} 对象格式
	if err := redis.Init(conf.Get("redis")); err != nil {
		logger.Fatal("Redis 初始化失败", zap.Error(err))
	}
	redisMgr, err := redis.GetRedisManager()
	if err != nil {
		logger.Fatal("Redis 连接失败", zap.Error(err))
	}
	// GetRedisView(name, keyPrefix, logger) 返回带前缀封装的 RedisCli
	// name 对应 config.toml 中各 [[redis]] 段的 name 字段
	defaultCache := redisMgr.GetRedisView("default", conf.GetRedisKeyPrefix(), logger)
	sessionCache := redisMgr.GetRedisView("session", conf.GetRedisKeyPrefix(), logger)
	_, _ = defaultCache, sessionCache
	logger.Info("Redis 已就绪（2 个实例：default, session）")

	// ── 5. 关系型数据库（多库，MySQL + PostgreSQL）──────────
	// conf.Get("databases") 返回 [{...},{...}] 数组 JSON（TOML 中使用 [[databases]]）
	// 单库时也兼容 {...} 对象格式
	// Config 字段：name / db_type(mysql|postgres) / master / slaves / use_master_slave
	if err := databases.Init(conf.Get("databases")); err != nil {
		logger.Fatal("数据库初始化失败", zap.Error(err))
	}
	dbMgr, err := databases.GetDatabaseManager()
	if err != nil {
		logger.Fatal("数据库连接失败", zap.Error(err))
	}
	// GetDBDao(name) 返回封装了 GORM 的 BaseDao，name 对应各 [[databases]] 的 name 字段
	// main  库（MySQL，主从读写分离）：写操作走 master，读操作由 GORM DBResolver 自动路由从库
	mainDao := dbMgr.GetDBDao("main")
	// archive 库（PostgreSQL，单机）
	archiveDao := dbMgr.GetDBDao("archive")
	_, _ = mainDao, archiveDao
	logger.Info("数据库已就绪（2 个实例：main[MySQL+主从], archive[PostgreSQL]）")

	// ── 6. MongoDB ───────────────────────────────────────────
	// conf.Get("mongodb") 支持单库 {} 或多库 [{},{}] 两种 JSON 格式
	if err := mongodb.Init(conf.Get("mongodb")); err != nil {
		logger.Fatal("MongoDB 初始化失败", zap.Error(err))
	}
	mongoMgr, err := mongodb.GetMongoManager()
	if err != nil {
		logger.Fatal("MongoDB 连接失败", zap.Error(err))
	}
	// 获取客户端：mongoMgr.GetClient("default")
	_ = mongoMgr
	logger.Info("MongoDB 已就绪")

	// ── 7. ClickHouse ────────────────────────────────────────
	// conf.Get("clickhouse") 支持单实例 {} 或多实例 [{},{}] 两种 JSON 格式
	if err := clickhouse.Init(conf.Get("clickhouse")); err != nil {
		logger.Fatal("ClickHouse 初始化失败", zap.Error(err))
	}
	chMgr, err := clickhouse.GetClickHouseManager()
	if err != nil {
		logger.Fatal("ClickHouse 连接失败", zap.Error(err))
	}
	// 获取客户端：chMgr.GetClient("default")
	_ = chMgr
	logger.Info("ClickHouse 已就绪")

	// ── 8. NATS / JetStream ──────────────────────────────────
	// conf.Get("nats") 必须为数组格式（TOML 中使用 [[nats]]）
	if err := nats.Init(conf.Get("nats")); err != nil {
		logger.Fatal("NATS 初始化失败", zap.Error(err))
	}
	natsMgr, err := nats.GetNatsManager()
	if err != nil {
		logger.Fatal("NATS 连接失败", zap.Error(err))
	}
	go natsMgr.StartAll() // 非阻塞：启动重连监控 goroutine
	defer natsMgr.CloseAll()

	// 消息发布示例（可在业务代码中使用）：
	// conn, _ := natsMgr.GetClient("default")
	// conn.Publish(ctx, "my-service.event", payload)
	logger.Info("NATS 已就绪")

	// ── 9. HTTP 服务器（Gin）─────────────────────────────────
	// conf.Get("http_server") 返回 HttpServerConfig JSON
	httpSrv, err := httpServer.NewHttpServer(conf.Get("http_server"))
	if err != nil {
		logger.Fatal("HTTP 服务器创建失败", zap.Error(err))
	}

	// 注册中间件组（中间件应用于 /api/<group>/* 路径）
	httpSrv.AddGroup("v1",
		middleware.Metrics(conf.GetAppName()),  // Prometheus 指标
		middleware.Tracing(conf.GetAppName()),  // OTel 链路追踪
		middleware.CircuitBreaker(),            // 熔断
		middleware.Timeout(0),                  // 超时（0 使用默认值）
		middleware.Access(),                    // JWT 鉴权（/api/v1/* 路径）
	)

	// 注册路由：Post(path, group, handler)
	// path  = handler 名称 / 子路径，实际路由 = /api/<group>/<path>
	// group = AddGroup 时注册的名称（空字符串表示不挂载到 group）
	httpSrv.Post("hello", "v1", httpServer.NewHandler(
		"hello",
		[]string{"demo"},
		func(c *gin.Context, req HelloRequest) (HelloResponse, error) {
			return HelloResponse{Message: "Hello, " + req.Name}, nil
		},
	))

	// 无鉴权的 health check 路由
	httpSrv.Get("health", "", httpServer.NewHandler(
		"health",
		[]string{"monitor"},
		func(c *gin.Context, _ struct{}) (map[string]string, error) {
			return map[string]string{"status": "ok"}, nil
		},
	))

	// ── 10. gRPC 服务器 ──────────────────────────────────────
	// conf.Get("grpc_server") 返回 GrpcServerConfig JSON
	var grpcConf httpServer.GrpcServerConfig
	if err := json.Unmarshal(conf.Get("grpc_server"), &grpcConf); err != nil {
		logger.Fatal("gRPC 配置解析失败", zap.Error(err))
	}
	grpcSrv, err := httpServer.NewGrpcServer(grpcConf)
	if err != nil {
		logger.Fatal("gRPC 服务器创建失败", zap.Error(err))
	}
	// 注册 protobuf 服务（取消注释并替换为实际生成的代码）：
	// pb.RegisterUserServiceServer(grpcSrv.Server(), &UserServiceImpl{})

	// ── 11. 服务组：统一生命周期管理 ─────────────────────────
	// ServiceGroup 监听 SIGINT / SIGTERM，收到信号后逆序关闭所有服务
	sg := service.NewServiceGroup()
	sg.Add(grpcSrv)                           // GrpcServer 直接实现 Service 接口
	sg.Add(service.NewServerStarter(httpSrv)) // HttpServer 通过 ServerStarter 适配

	logger.Info("服务启动中...",
		zap.String("http", "0.0.0.0:8080"),
		zap.String("grpc", "0.0.0.0:9090"),
	)
	sg.Start() // 阻塞，直到收到退出信号
}

// ── 业务请求/响应结构体示例 ──────────────────────────────────

// HelloRequest POST /api/v1/hello 请求体
type HelloRequest struct {
	Name string `json:"name" form:"name" binding:"required"`
}

// HelloResponse POST /api/v1/hello 响应体
type HelloResponse struct {
	Message string `json:"message"`
}
