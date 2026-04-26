package redis

import (
	"encoding/json"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

type RedisManager struct {
	mu        sync.Mutex
	redisCmds map[string]RedisCmd
}

var (
	manager       *RedisManager
	managerOnce   sync.Once
	managerConfig []*Config
	managerErr    error
)

func Init(config []byte) error {
	if len(config) == 0 {
		return fmt.Errorf("redis config is empty")
	}

	var single Config
	if err := json.Unmarshal(config, &single); err == nil && single.Name != "" {
		managerConfig = []*Config{&single}
		return nil
	}

	var multiple []*Config
	if err := json.Unmarshal(config, &multiple); err != nil {
		return fmt.Errorf("parse redis config failed: %w", err)
	}
	managerConfig = multiple
	return nil
}

func ensureRedisInitialized() error {
	if len(managerConfig) == 0 {
		return nil
	}
	managerOnce.Do(func() {
		redisCmds := make(map[string]RedisCmd, len(managerConfig))
		for index, cfg := range managerConfig {
			cli, err := NewRedisCmd(cfg)
			if err != nil {
				managerErr = err
				return
			}
			if index == 0 {
				redisCon = cli
			}
			redisCmds[cfg.Name] = cli
		}
		manager = &RedisManager{redisCmds: redisCmds}
	})
	return managerErr
}

// GetRedisManager 返回全局 Redis 管理器，未初始化或连接失败时返回 error。
func GetRedisManager() (*RedisManager, error) {
	if err := ensureRedisInitialized(); err != nil {
		return nil, err
	}
	if manager == nil {
		return nil, fmt.Errorf("redis: not initialized, call Init() first")
	}
	return manager, nil
}

func (m *RedisManager) GetRedisCmd(name string) RedisCmd {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cmd, ok := m.redisCmds[name]; ok {
		return cmd
	}
	return nil
}

func (m *RedisManager) GetRedisView(name, prefix string, logger *zap.Logger) RedisCli {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cmd, ok := m.redisCmds[name]; ok {
		return NewRedisView(cmd, prefix, logger)
	}
	return nil
}
