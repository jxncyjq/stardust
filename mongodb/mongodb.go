package mongodb

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

// Config MongoDB 连接配置。
type Config struct {
	Name      string `json:"name"`       // 逻辑名称，多库时用于区分
	URI       string `json:"uri"`        // mongodb://user:pass@host:port/dbname?authSource=admin
	Database  string `json:"database"`   // 默认操作的数据库名
	MaxPool   uint64 `json:"max_pool"`   // 最大连接池大小
	MinPool   uint64 `json:"min_pool"`   // 最小连接池大小
	TimeoutS  int    `json:"timeout_s"`  // 单次操作超时（秒）
}

// SetDefaults 填充未设置的默认值。
func (c *Config) SetDefaults() {
	if c.MaxPool == 0 {
		c.MaxPool = 10
	}
	if c.MinPool == 0 {
		c.MinPool = 2
	}
	if c.TimeoutS == 0 {
		c.TimeoutS = 5
	}
}

// Validate 校验必填字段。
func (c *Config) Validate() error {
	if c.Name == "" {
		return errors.New("mongodb: config name is required")
	}
	if c.URI == "" {
		return errors.New("mongodb: uri is required")
	}
	if c.Database == "" {
		return errors.New("mongodb: database is required")
	}
	return nil
}

var (
	managerOnce   sync.Once
	manager       *MongoManager
	managerConfig []*Config
)

var managerErr error

// Init 解析配置，支持单个对象或数组两种 JSON 格式，与 databases.Init 保持一致。
// 必须在 GetMongoManager() 之前调用。
func Init(config []byte) error {
	if len(config) == 0 {
		return fmt.Errorf("mongodb: config is empty")
	}

	var single Config
	if err := json.Unmarshal(config, &single); err == nil && single.Name != "" {
		managerConfig = []*Config{&single}
		return nil
	}

	var multiple []*Config
	if err := json.Unmarshal(config, &multiple); err != nil {
		return fmt.Errorf("mongodb: parse config failed: %w", err)
	}
	managerConfig = multiple
	return nil
}

// GetMongoManager 返回全局单例 MongoManager，首次调用时初始化所有连接。
func GetMongoManager() (*MongoManager, error) {
	if len(managerConfig) == 0 {
		return nil, fmt.Errorf("mongodb: not initialized, call Init() first")
	}
	managerOnce.Do(func() {
		m, err := newMongoManager(managerConfig)
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

// GetClient 快捷方法：获取指定名称的 MongoCli，未初始化或连接失败时返回 nil。
func GetClient(name string) MongoCli {
	m, err := GetMongoManager()
	if err != nil {
		return nil
	}
	return m.GetClient(name)
}
