//go:build ignore

// Package main 旧的 build-ignored 参考示例。
//
// 现在新的 canonical 示例放在 example/main.go：
// 它按最新 go-stardust-rtl skills 演示 logs / tracing / redis / databases /
// nats / authz / i18n / http_server / business 的组合方式。
//
// 这个文件保留为历史参考，不再作为首选启动入口。
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jxncyjq/stardust/app"
	"github.com/jxncyjq/stardust/app/components"
	"github.com/jxncyjq/stardust/conf"
	"github.com/jxncyjq/stardust/databases"
	httpServer "github.com/jxncyjq/stardust/http_server"
	"github.com/jxncyjq/stardust/http_server/middleware"
	stardi18n "github.com/jxncyjq/stardust/i18n"
	"github.com/jxncyjq/stardust/logs"
	"github.com/jxncyjq/stardust/redis"
	"github.com/jxncyjq/stardust/service"
	"google.golang.org/grpc"
)

func main() {
	if os.Getenv("runConfig") == "" {
		os.Setenv("runConfig", "./example/config.toml")
	}
	conf.Init()

	// ── 方式一（推荐）：在 app 层集中声明中间件组 ─────────────
	// WithHTTPGroup 将中间件配置从 setupHTTP 中剥离，
	// 集中放在 main 里，便于统一查阅和替换。
	myApp := app.New(conf.Get).
		WithHTTPGroup("v1",
			middleware.I18n(),
			middleware.Access(),
			middleware.Metrics(conf.GetAppName()),
			middleware.Tracing(conf.GetAppName()),
			middleware.CircuitBreaker(),
			middleware.Timeout(0),
			middleware.Authz(),
		)
	userModule := &UserModule{}
	gameModule := &GameModule{}
	businessComponent := components.Business(
		service.NewServiceGroup(),
		userModule,
		gameModule,
	).WithDependencies("logs", "redis", "databases")

	if err := myApp.Use(
		components.LogsComponent(),
		components.RedisComponent(),
		components.DatabasesComponent(),
		components.MongoDBComponent(),
		components.ClickhouseComponent(),
		components.NatsComponent(),
		components.TracingComponent(),
		components.I18nComponent(),
		components.AuthzComponent(),
		businessComponent,                                                // ← 用 service.Manager 统一管理 UserModule、GameModule 的生命周期
		components.HTTPServerFromApp(myApp, businessComponent.SetupHTTP), // ← 与 Application 绑定，自动应用 WithHTTPGroup
		components.GRPCServerComponent(businessComponent.SetupGRPC),
	).Run(context.Background()); err != nil {
		panic(err)
	}
}

// UserModule 演示一个业务模块同时绑定 HTTP 和 gRPC。
//
// sample/backend/services/admin/AdminService.Start(app) 也是这个模式：
// 同一个业务 Service 负责注册 HTTP 路由，也负责把 gRPC 实现注册到 grpc.Server。
type UserModule struct {
	service *UserService
}

// Service 返回共享业务 Service。HTTP handler 和 gRPC implementation 应复用同一个实例，
// 避免 HTTP、gRPC 各自构造一套 DB/Redis 依赖。
func (m *UserModule) Service() *UserService {
	if m.service != nil {
		return m.service
	}

	// 框架保证 HTTP/gRPC setup 执行时 redis/databases 组件已完成 Init()。
	redisMgr, err := redis.GetRedisManager()
	if err != nil {
		panic("redis manager unavailable: " + err.Error())
	}
	dbMgr, err := databases.GetDatabaseManager()
	if err != nil {
		panic("database manager unavailable: " + err.Error())
	}

	logger := logs.GetLogger("user-service")
	m.service = &UserService{
		dao:   dbMgr.GetDBDao("main"),
		cache: redisMgr.GetRedisView("default", conf.GetRedisKeyPrefix(), logger),
	}
	return m.service
}

// Start 实现 service.Service，启动用户业务模块。
func (m *UserModule) Start() {
	_ = m.Service()
	logs.GetLogger("user-service").Info("UserModule started")
}

// Stop 实现 service.Service，停止用户业务模块。
func (m *UserModule) Stop() {
	logs.GetLogger("user-service").Info("UserModule stopped")
}

// SetupHTTP 在 HTTP 组件 Init 阶段调用，此时 redis/databases 已就绪。
//
// 使用 components.HTTPServerFromApp(myApp, userModule.SetupHTTP) 时，
// 中间件组由 WithHTTPGroup 在 Init 前自动应用，
// 本函数只需关注两件事：① 获取共享 Service；② 注册 HTTP 路由。
func (m *UserModule) SetupHTTP(srv *httpServer.HttpServer) {
	// ── 注册路由（中间件组 "v1" 已在 app 层声明）──────────────
	// GET  /api/v1/user/:id  — 先查缓存，缓存未命中查 DB 并回填
	srv.Get("user/:id", "v1", httpServer.NewHandler(
		"user/:id",
		[]string{"user"},
		m.GetUser,
	))

	// POST /api/v1/user  — 写入 DB，同时更新缓存
	srv.Post("user", "v1", httpServer.NewHandler(
		"user",
		[]string{"user"},
		m.CreateUser,
	))

	srv.Get("hello/:name", "v1", httpServer.NewHandler(
		"hello/:name",
		[]string{"greet"},
		m.GetGreeting,
	))
}

// SetupGRPC 在 gRPC 组件 Init 阶段调用，用同一个 UserService 注册 protobuf 服务。
func (m *UserModule) SetupGRPC(s *grpc.Server) {
	// 示例项目没有提交 protobuf 生成代码，因此这里保留注册位置。
	// 实际服务中应像 sample 的 AdminService.registerGRPC 一样，把共享 service
	// 注入 gRPC implementation 后注册到 grpc.Server：
	//
	// pb.RegisterUserServiceServer(s, &UserServiceImpl{service: m.Service()})
	_ = s
}

// GetUser 将 HTTP handler 委托给共享 UserService。
func (m *UserModule) GetUser(c *gin.Context, req struct{}) (UserResp, error) {
	return m.Service().GetUser(c, req)
}

// CreateUser 将 HTTP handler 委托给共享 UserService。
func (m *UserModule) CreateUser(c *gin.Context, req CreateUserReq) (UserResp, error) {
	return m.Service().CreateUser(c, req)
}

// GetGreeting 返回本地化问候语。
func (m *UserModule) GetGreeting(c *gin.Context, req struct{}) (GreetingResp, error) {
	_ = req
	message, err := stardi18n.Localize(c.Request.Context(), "Hello", map[string]any{
		"Name": c.Param("name"),
	})
	if err != nil {
		return GreetingResp{}, err
	}
	return GreetingResp{Message: message}, nil
}

// GameModule 演示第二个业务模块，体现 service.Manager 管理多个服务。
type GameModule struct {
	service *GameService
}

// Service 返回共享 GameService。
func (m *GameModule) Service() *GameService {
	if m.service != nil {
		return m.service
	}
	m.service = &GameService{}
	return m.service
}

// Start 实现 service.Service，启动游戏业务模块。
func (m *GameModule) Start() {
	_ = m.Service()
	logs.GetLogger("game-service").Info("GameModule started")
}

// Stop 实现 service.Service，停止游戏业务模块。
func (m *GameModule) Stop() {
	logs.GetLogger("game-service").Info("GameModule stopped")
}

// SetupHTTP 注册游戏业务 HTTP 路由。
func (m *GameModule) SetupHTTP(srv *httpServer.HttpServer) {
	// GET /api/v1/game/:id
	srv.Get("game/:id", "v1", httpServer.NewHandler(
		"game/:id",
		[]string{"game"},
		m.GetGame,
	))

	// POST /api/v1/game/score
	srv.Post("game/score", "v1", httpServer.NewHandler(
		"game/score",
		[]string{"game"},
		m.SubmitScore,
	))
}

// SetupGRPC 注册游戏业务 gRPC 服务。
func (m *GameModule) SetupGRPC(s *grpc.Server) {
	// 示例项目没有提交 protobuf 生成代码，因此这里保留注册位置：
	//
	// pb.RegisterGameServiceServer(s, &GameServiceImpl{service: m.Service()})
	_ = s
}

// GetGame 将 HTTP handler 委托给共享 GameService。
func (m *GameModule) GetGame(c *gin.Context, req struct{}) (GameResp, error) {
	return m.Service().GetGame(c, req)
}

// SubmitScore 将 HTTP handler 委托给共享 GameService。
func (m *GameModule) SubmitScore(c *gin.Context, req SubmitScoreReq) (SubmitScoreResp, error) {
	return m.Service().SubmitScore(c, req)
}

// ════════════════════════════════════════════════════════════════════
// 方式二（兼容旧写法）：在 setupHTTP 内部手动 AddGroup
// ════════════════════════════════════════════════════════════════════
//
// func main2() {
// 	conf.Init()
// 	userModule := &UserModule{}
// 	gameModule := &GameModule{}
// 	businessComponent := components.Business(
// 		service.NewServiceGroup(),
// 		userModule,
// 		gameModule,
// 	).WithDependencies("logs", "redis", "databases")
// 	if err := app.New(conf.Get).
// 		Use(
// 			components.LogsComponent(),
// 			components.RedisComponent(),
// 			components.DatabasesComponent(),
// 			components.AuthzComponent(),
// 			businessComponent,
// 			components.HTTPServerComponent(businessComponent.SetupHTTPManual), // 独立函数，无需 myApp
// 			components.GRPCServerComponent(businessComponent.SetupGRPC),
// 		).
// 		Run(context.Background()); err != nil {
// 		panic(err)
// 	}
// }
//
// func (b *BusinessComponent) SetupHTTPManual(srv *httpServer.HttpServer) {
// 	srv.AddGroup("v1",
// 		middleware.Metrics(conf.GetAppName()),
// 		middleware.Tracing(conf.GetAppName()),
// 		middleware.CircuitBreaker(),
// 		middleware.Timeout(0),
// 		middleware.Authz(),
// 	)
// 	b.SetupHTTP(srv)
// }

// ════════════════════════════════════════════════════════════════
// 数据模型
// ════════════════════════════════════════════════════════════════

// User 对应 users 表
type User struct {
	ID        int64     `json:"id"         gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name"       gorm:"column:name;not null"`
	Email     string    `json:"email"      gorm:"column:email;uniqueIndex"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (User) TableName() string { return "users" }

// ════════════════════════════════════════════════════════════════
// 请求 / 响应结构体
// ════════════════════════════════════════════════════════════════

type CreateUserReq struct {
	Name  string `json:"name"  binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

type UserResp struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	FromCache bool   `json:"from_cache,omitempty"` // 标识是否来自缓存
}

type GreetingResp struct {
	Message string `json:"message"`
}

type GameResp struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type SubmitScoreReq struct {
	UserID int64 `json:"user_id" binding:"required"`
	GameID int64 `json:"game_id" binding:"required"`
	Score  int   `json:"score"   binding:"required"`
}

type SubmitScoreResp struct {
	UserID int64  `json:"user_id"`
	GameID int64  `json:"game_id"`
	Score  int    `json:"score"`
	Status string `json:"status"`
}

// ════════════════════════════════════════════════════════════════
// UserService：封装 DB + Redis 的业务逻辑
// ════════════════════════════════════════════════════════════════

type UserService struct {
	dao   databases.BaseDao
	cache redis.RedisCli
}

// cacheKey 生成用户缓存 key（带全局前缀，前缀由 RedisCli 内部处理）
func (s *UserService) cacheKey(id int64) string {
	return fmt.Sprintf("user:%d", id)
}

// GetUser GET /api/v1/user/:id
// 流程：读 Redis → 命中直接返回；未命中查 MySQL → 写回 Redis
func (s *UserService) GetUser(c *gin.Context, _ struct{}) (UserResp, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return UserResp{}, err
	}

	ctx := c.Request.Context()
	key := s.cacheKey(id)

	// ── 1. 查 Redis 缓存 ──────────────────────────────────────
	cached, err := s.cache.Get(ctx, key)
	if err == nil && cached != nil {
		var user User
		if jsonErr := json.Unmarshal(cached, &user); jsonErr == nil {
			return toUserResp(user, true), nil
		}
	}

	// ── 2. 缓存未命中：查 MySQL ───────────────────────────────
	var user User
	user.ID = id
	found, err := s.dao.FindById(id, &user)
	if err != nil {
		return UserResp{}, fmt.Errorf("db query: %w", err)
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return UserResp{}, errors.New("not found")
	}

	// ── 3. 回填缓存（5 分钟 TTL）─────────────────────────────
	if data, jsonErr := json.Marshal(user); jsonErr == nil {
		_ = s.cache.Set(ctx, key, data, "5m")
	}

	return toUserResp(user, false), nil
}

// CreateUser POST /api/v1/user
// 流程：写入 MySQL → 写入 Redis 缓存
func (s *UserService) CreateUser(c *gin.Context, req CreateUserReq) (UserResp, error) {
	ctx := c.Request.Context()

	user := User{
		Name:  req.Name,
		Email: req.Email,
	}

	// ── 1. 写入 MySQL ─────────────────────────────────────────
	if _, err := s.dao.InsertOne(&user); err != nil {
		return UserResp{}, fmt.Errorf("db insert: %w", err)
	}

	// ── 2. 写入 Redis 缓存（5 分钟 TTL）──────────────────────
	key := s.cacheKey(user.ID)
	if data, err := json.Marshal(user); err == nil {
		_ = s.cache.Set(ctx, key, data, "5m")
	}

	c.JSON(http.StatusCreated, toUserResp(user, false))
	return UserResp{}, nil
}

// ════════════════════════════════════════════════════════════════
// GameService：第二个业务服务，用于演示多服务统一管理
// ════════════════════════════════════════════════════════════════

type GameService struct{}

// GetGame GET /api/v1/game/:id
func (s *GameService) GetGame(c *gin.Context, _ struct{}) (GameResp, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return GameResp{}, err
	}

	return GameResp{
		ID:     id,
		Name:   fmt.Sprintf("game-%d", id),
		Status: "online",
	}, nil
}

// SubmitScore POST /api/v1/game/score
func (s *GameService) SubmitScore(_ *gin.Context, req SubmitScoreReq) (SubmitScoreResp, error) {
	return SubmitScoreResp{
		UserID: req.UserID,
		GameID: req.GameID,
		Score:  req.Score,
		Status: "accepted",
	}, nil
}

// ════════════════════════════════════════════════════════════════
// 辅助函数
// ════════════════════════════════════════════════════════════════

func toUserResp(u User, fromCache bool) UserResp {
	return UserResp{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		FromCache: fromCache,
	}
}
