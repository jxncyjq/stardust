package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	"github.com/jxncyjq/stardust/metric"
	"github.com/jxncyjq/stardust/nats"
	"github.com/jxncyjq/stardust/redis"
	"github.com/jxncyjq/stardust/service"
	"google.golang.org/grpc"
)

var (
	userCreatedTotal = metric.NewCounter(metric.CounterOpts{
		Namespace: "stardust",
		Subsystem: "example",
		Name:      "user_created_total",
		Help:      "Total number of created users.",
		Labels:    []string{"result"},
	})
	scoreAcceptedTotal = metric.NewCounter(metric.CounterOpts{
		Namespace: "stardust",
		Subsystem: "example",
		Name:      "score_accepted_total",
		Help:      "Total number of accepted score submissions.",
		Labels:    []string{"result"},
	})
)

func main() {
	if os.Getenv("runConfig") == "" {
		os.Setenv("runConfig", "./example/config.toml")
	}
	conf.Init()

	myApp := app.New(conf.Get).
		WithHTTPGroup("v1",
			middleware.I18n(),
			middleware.Metrics(conf.GetAppName()),
			middleware.Tracing(conf.GetAppName()),
			middleware.CircuitBreaker(),
			middleware.Timeout(0),
			middleware.Access(),
			middleware.Authz(),
		)

	userModule := &UserModule{}
	gameModule := &GameModule{}
	businessComponent := components.Business(
		service.NewServiceGroup(),
		userModule,
		gameModule,
	).WithDependencies("logs", "redis", "databases", "nats")

	if err := myApp.Use(
		components.LogsComponent(),
		components.RedisComponent(),
		components.DatabasesComponent(),
		components.NatsComponent(),
		components.TracingComponent(),
		components.I18nComponent(),
		components.AuthzComponent(),
		businessComponent,
		components.HTTPServerFromApp(myApp, businessComponent.SetupHTTP),
		components.GRPCServerComponent(businessComponent.SetupGRPC),
	).Run(context.Background()); err != nil {
		panic(err)
	}
}

// UserModule 负责用户相关 HTTP / gRPC 绑定。
type UserModule struct {
	service *UserService
}

func (m *UserModule) Service() *UserService {
	if m.service != nil {
		return m.service
	}

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

func (m *UserModule) Start() {
	_ = m.Service()
	logs.GetLogger("user-service").Info("UserModule started")
}

func (m *UserModule) Stop() {
	logs.GetLogger("user-service").Info("UserModule stopped")
}

func (m *UserModule) SetupHTTP(srv *httpServer.HttpServer) {
	srv.Get("user/:id", "v1", httpServer.NewHandler(
		"user/:id",
		[]string{"user"},
		m.GetUser,
	))

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

func (m *UserModule) SetupGRPC(s *grpc.Server) {
	_ = s
}

func (m *UserModule) GetUser(c *gin.Context, req struct{}) (UserResp, error) {
	return m.Service().GetUser(c, req)
}

func (m *UserModule) CreateUser(c *gin.Context, req CreateUserReq) (UserResp, error) {
	return m.Service().CreateUser(c, req)
}

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

// GameModule 负责游戏相关 HTTP / gRPC 绑定。
type GameModule struct {
	service *GameService
}

func (m *GameModule) Service() *GameService {
	if m.service != nil {
		return m.service
	}
	m.service = &GameService{}
	return m.service
}

func (m *GameModule) Start() {
	_ = m.Service()
	logs.GetLogger("game-service").Info("GameModule started")
}

func (m *GameModule) Stop() {
	logs.GetLogger("game-service").Info("GameModule stopped")
}

func (m *GameModule) SetupHTTP(srv *httpServer.HttpServer) {
	srv.Get("game/:id", "v1", httpServer.NewHandler(
		"game/:id",
		[]string{"game"},
		m.GetGame,
	))

	srv.Post("game/score", "v1", httpServer.NewHandler(
		"game/score",
		[]string{"game"},
		m.SubmitScore,
	))
}

func (m *GameModule) SetupGRPC(s *grpc.Server) {
	_ = s
}

func (m *GameModule) GetGame(c *gin.Context, req struct{}) (GameResp, error) {
	return m.Service().GetGame(c, req)
}

func (m *GameModule) SubmitScore(c *gin.Context, req SubmitScoreReq) (SubmitScoreResp, error) {
	return m.Service().SubmitScore(c, req)
}

type User struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"column:name;not null"`
	Email     string    `json:"email" gorm:"column:email;uniqueIndex"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (User) TableName() string { return "users" }

type CreateUserReq struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

type UserResp struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	FromCache bool   `json:"from_cache,omitempty"`
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
	Score  int   `json:"score" binding:"required"`
}

type SubmitScoreResp struct {
	UserID int64  `json:"user_id"`
	GameID int64  `json:"game_id"`
	Score  int    `json:"score"`
	Status string `json:"status"`
}

type UserService struct {
	dao   databases.BaseDao
	cache redis.RedisCli
}

func (s *UserService) cacheKey(id int64) string {
	return fmt.Sprintf("user:%d", id)
}

func (s *UserService) GetUser(c *gin.Context, _ struct{}) (UserResp, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return UserResp{}, fmt.Errorf("invalid user id: %w", err)
	}

	ctx := c.Request.Context()
	key := s.cacheKey(id)

	cached, err := s.cache.Get(ctx, key)
	if err == nil && cached != nil {
		var user User
		if jsonErr := json.Unmarshal(cached, &user); jsonErr == nil {
			return toUserResp(user, true), nil
		}
	}

	var user User
	found, err := s.dao.FindById(id, &user)
	if err != nil {
		return UserResp{}, fmt.Errorf("db query: %w", err)
	}
	if !found {
		return UserResp{}, errors.New("user not found")
	}

	if data, jsonErr := json.Marshal(user); jsonErr == nil {
		_ = s.cache.Set(ctx, key, data, "5m")
	}

	return toUserResp(user, false), nil
}

func (s *UserService) CreateUser(c *gin.Context, req CreateUserReq) (UserResp, error) {
	ctx := c.Request.Context()
	user := User{
		Name:  req.Name,
		Email: req.Email,
	}

	if _, err := s.dao.InsertOne(&user); err != nil {
		userCreatedTotal.Inc("error")
		return UserResp{}, fmt.Errorf("db insert: %w", err)
	}

	if data, err := json.Marshal(user); err == nil {
		_ = s.cache.Set(ctx, s.cacheKey(user.ID), data, "5m")
	}

	if err := publishEvent("user.created", map[string]any{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
	}); err != nil {
		return UserResp{}, err
	}

	userCreatedTotal.Inc("success")
	return toUserResp(user, false), nil
}

type GameService struct{}

func (s *GameService) GetGame(c *gin.Context, _ struct{}) (GameResp, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return GameResp{}, fmt.Errorf("invalid game id: %w", err)
	}

	return GameResp{
		ID:     id,
		Name:   fmt.Sprintf("game-%d", id),
		Status: "online",
	}, nil
}

func (s *GameService) SubmitScore(_ *gin.Context, req SubmitScoreReq) (SubmitScoreResp, error) {
	resp := SubmitScoreResp{
		UserID: req.UserID,
		GameID: req.GameID,
		Score:  req.Score,
		Status: "accepted",
	}

	if err := publishEvent("game.score.accepted", resp); err != nil {
		scoreAcceptedTotal.Inc("error")
		return SubmitScoreResp{}, err
	}

	scoreAcceptedTotal.Inc("success")
	return resp, nil
}

func publishEvent(subject string, payload any) error {
	mgr, err := nats.GetNatsManager()
	if err != nil {
		return err
	}
	conn, ok := mgr.GetClient("default")
	if !ok {
		return fmt.Errorf("nats client %q not found", "default")
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return conn.Publish(subject, data)
}

func toUserResp(u User, fromCache bool) UserResp {
	return UserResp{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		FromCache: fromCache,
	}
}
