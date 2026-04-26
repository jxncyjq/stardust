//go:build ignore

// Package main 演示以纯组件形式启动完整微服务，
// 并在 HTTP 业务逻辑中调用 database 和 redis 组件。
//
// 本文件展示两种等价写法：
//
//  1. 【推荐】myApp.HTTPServer(setupHTTP)
//     中间件组在 app 层集中声明（WithHTTPGroup），
//     setupHTTP 只关注服务构造和路由注册。
//
//  2. 【兼容】components.HTTPServerComponent(setupHTTP)
//     在 setupHTTP 内部手动调用 srv.AddGroup。
//
// 启动：
//
//	export runConfig=./example/config.toml
//	go run app_example.go
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
	"github.com/jxncyjq/stardust/logs"
	"github.com/jxncyjq/stardust/redis"
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
			middleware.Metrics(conf.GetAppName()),
			middleware.Tracing(conf.GetAppName()),
			middleware.CircuitBreaker(),
			middleware.Timeout(0),
		)

	if err := myApp.Use(
		components.LogsComponent(),
		components.RedisComponent(),
		components.DatabasesComponent(),
		components.MongoDBComponent(),
		components.ClickhouseComponent(),
		components.NatsComponent(),
		components.TracingComponent(),
		components.HTTPServerFromApp(myApp, setupHTTP), // ← 与 Application 绑定，自动应用 WithHTTPGroup
		components.GRPCServerComponent(setupGRPC),
	).Run(context.Background()); err != nil {
		panic(err)
	}
}

// setupHTTP 在 HTTP 组件 Init 阶段调用，此时 redis/databases 已就绪。
//
// 使用 components.HTTPServerFromApp(myApp, setupHTTP) 时，
// 中间件组由 WithHTTPGroup 在 Init 前自动应用，
// 本函数只需关注两件事：① 构造 Service；② 注册路由。
func setupHTTP(srv *httpServer.HttpServer) {
	// ── 1. 获取已初始化的 Manager ─────────────────────────────
	// 框架保证本函数执行时 redis/databases 组件已完成 Init()。
	redisMgr, err := redis.GetRedisManager()
	if err != nil {
		panic("redis manager unavailable: " + err.Error())
	}
	dbMgr, err := databases.GetDatabaseManager()
	if err != nil {
		panic("database manager unavailable: " + err.Error())
	}

	// ── 2. 构造 Service（依赖注入）────────────────────────────
	logger := logs.GetLogger("user-service")
	userSvc := &UserService{
		dao:   dbMgr.GetDBDao("main"),
		cache: redisMgr.GetRedisView("default", conf.GetRedisKeyPrefix(), logger),
	}

	// ── 3. 注册路由（中间件组 "v1" 已在 app 层声明）──────────
	// GET  /api/v1/user/:id  — 先查缓存，缓存未命中查 DB 并回填
	srv.Get("user/:id", "v1", httpServer.NewHandler(
		"user/:id",
		[]string{"user"},
		userSvc.GetUser,
	))

	// POST /api/v1/user  — 写入 DB，同时更新缓存
	srv.Post("user", "v1", httpServer.NewHandler(
		"user",
		[]string{"user"},
		userSvc.CreateUser,
	))
}

func setupGRPC(s *grpc.Server) {
	// pb.RegisterUserServiceServer(s, &UserServiceImpl{})
	_ = s
}

// ════════════════════════════════════════════════════════════════════
// 方式二（兼容旧写法）：在 setupHTTP 内部手动 AddGroup
// ════════════════════════════════════════════════════════════════════
//
// func main2() {
// 	conf.Init()
// 	if err := app.New(conf.Get).
// 		Use(
// 			components.LogsComponent(),
// 			components.RedisComponent(),
// 			components.DatabasesComponent(),
// 			components.HTTPServerComponent(setupHTTPManual), // 独立函数，无需 myApp
// 			components.GRPCServerComponent(setupGRPC),
// 		).
// 		Run(context.Background()); err != nil {
// 		panic(err)
// 	}
// }
//
// func setupHTTPManual(srv *httpServer.HttpServer) {
// 	srv.AddGroup("v1",
// 		middleware.Metrics(conf.GetAppName()),
// 		middleware.Tracing(conf.GetAppName()),
// 		middleware.CircuitBreaker(),
// 		middleware.Timeout(0),
// 	)
// 	// ... 同样注册路由
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
