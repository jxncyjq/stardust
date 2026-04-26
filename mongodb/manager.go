package mongodb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// MongoManager 管理多个 MongoDB 客户端实例（线程安全）。
type MongoManager struct {
	mu      sync.RWMutex
	clients map[string]MongoCli
}

func newMongoManager(configs []*Config) (*MongoManager, error) {
	m := &MongoManager{
		clients: make(map[string]MongoCli),
	}
	for _, cfg := range configs {
		cfg.SetDefaults()
		if err := cfg.Validate(); err != nil {
			return nil, fmt.Errorf("mongodb: invalid config [%s]: %w", cfg.Name, err)
		}
		cli, err := newMongoClient(cfg)
		if err != nil {
			return nil, fmt.Errorf("mongodb: connect failed [%s]: %w", cfg.Name, err)
		}
		m.clients[cfg.Name] = cli
	}
	return m, nil
}

// GetClient 返回指定名称的 MongoCli，不存在时返回 nil。
func (m *MongoManager) GetClient(name string) MongoCli {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.clients[name]
}

// newMongoClient 建立 MongoDB 连接并验证可达性。
func newMongoClient(cfg *Config) (MongoCli, error) {
	timeout := time.Duration(cfg.TimeoutS) * time.Second

	opts := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPool).
		SetMinPoolSize(cfg.MinPool).
		SetConnectTimeout(timeout).
		SetTimeout(timeout)

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return &mongoClient{
		client:   client,
		database: cfg.Database,
		timeout:  timeout,
	}, nil
}
