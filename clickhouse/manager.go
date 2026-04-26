package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	chsql "github.com/ClickHouse/clickhouse-go/v2"
)

// ClickHouseManager 管理多个 ClickHouse 客户端实例（线程安全）。
type ClickHouseManager struct {
	mu      sync.RWMutex
	clients map[string]ClickHouseCli
}

func newClickHouseManager(configs []*Config) (*ClickHouseManager, error) {
	m := &ClickHouseManager{
		clients: make(map[string]ClickHouseCli),
	}
	for _, cfg := range configs {
		cfg.SetDefaults()
		if err := cfg.Validate(); err != nil {
			return nil, fmt.Errorf("clickhouse: invalid config [%s]: %w", cfg.Name, err)
		}
		cli, err := newClickHouseClient(cfg)
		if err != nil {
			return nil, fmt.Errorf("clickhouse: connect failed [%s]: %w", cfg.Name, err)
		}
		m.clients[cfg.Name] = cli
	}
	return m, nil
}

// GetClient 返回指定名称的 ClickHouseCli，不存在时返回 nil。
func (m *ClickHouseManager) GetClient(name string) ClickHouseCli {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.clients[name]
}

// newClickHouseClient 通过 clickhouse.OpenDB 建立连接（database/sql 兼容）。
func newClickHouseClient(cfg *Config) (ClickHouseCli, error) {
	dialTimeout := time.Duration(cfg.DialTimeoutS) * time.Second

	db := chsql.OpenDB(&chsql.Options{
		Addr: []string{cfg.Addr},
		Auth: chsql.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		DialTimeout: dialTimeout,
	})
	db.SetMaxOpenConns(cfg.MaxConn)
	db.SetMaxIdleConns(cfg.MaxIdle)
	db.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("clickhouse: ping failed: %w", err)
	}

	return &clickhouseClient{db: db}, nil
}

// OpenDB 返回底层 *sql.DB，供 clickhouseClient 使用。
func openDB(cfg *Config) *sql.DB {
	cli, err := newClickHouseClient(cfg)
	if err != nil {
		return nil
	}
	return cli.(*clickhouseClient).db
}
