package clickhouse

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

// Config ClickHouse 连接配置。
type Config struct {
	Name         string `json:"name"`           // 逻辑名称
	Addr         string `json:"addr"`           // host:port（Native 协议，默认 9000）
	Database     string `json:"database"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	MaxConn      int    `json:"max_conn"`       // 最大连接数
	MaxIdle      int    `json:"max_idle"`       // 最大空闲连接数
	DialTimeoutS int    `json:"dial_timeout_s"` // 连接超时（秒）
}

// SetDefaults 填充未设置的默认值。
func (c *Config) SetDefaults() {
	if c.MaxConn == 0 {
		c.MaxConn = 10
	}
	if c.MaxIdle == 0 {
		c.MaxIdle = 5
	}
	if c.DialTimeoutS == 0 {
		c.DialTimeoutS = 5
	}
}

// Validate 校验必填字段。
func (c *Config) Validate() error {
	if c.Name == "" {
		return errors.New("clickhouse: config name is required")
	}
	if c.Addr == "" {
		return errors.New("clickhouse: addr is required")
	}
	if c.Database == "" {
		return errors.New("clickhouse: database is required")
	}
	return nil
}

var (
	managerOnce   sync.Once
	manager       *ClickHouseManager
	managerConfig []*Config
	managerErr    error
)

// Init 解析配置，支持单个对象或数组两种 JSON 格式，与 databases.Init 保持一致。
func Init(config []byte) error {
	if len(config) == 0 {
		return fmt.Errorf("clickhouse: config is empty")
	}

	var single Config
	if err := json.Unmarshal(config, &single); err == nil && single.Name != "" {
		managerConfig = []*Config{&single}
		return nil
	}

	var multiple []*Config
	if err := json.Unmarshal(config, &multiple); err != nil {
		return fmt.Errorf("clickhouse: parse config failed: %w", err)
	}
	managerConfig = multiple
	return nil
}

// GetClickHouseManager 返回全局单例 ClickHouseManager，首次调用时初始化所有连接。
func GetClickHouseManager() (*ClickHouseManager, error) {
	if len(managerConfig) == 0 {
		return nil, fmt.Errorf("clickhouse: not initialized, call Init() first")
	}
	managerOnce.Do(func() {
		m, err := newClickHouseManager(managerConfig)
		if err != nil {
			managerErr = err
			return
		}
		manager = m
	})
	if managerErr != nil {
		return nil, managerErr
	}
	return manager, nil
}

// GetClient 快捷方法：获取指定名称的 ClickHouseCli，未初始化或连接失败时返回 nil。
func GetClient(name string) ClickHouseCli {
	m, err := GetClickHouseManager()
	if err != nil {
		return nil
	}
	return m.GetClient(name)
}
