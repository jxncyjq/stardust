package databases

import (
	"encoding/json"
	"fmt"
	"sync"
)

type DatabaseManager struct {
	mu sync.RWMutex
	db map[string]DBConn
}

var (
	manager       *DatabaseManager
	managerOnce   sync.Once
	managerConfig []*Config
	managerErr    error
)

func Init(config []byte) error {
	if len(config) == 0 {
		return fmt.Errorf("database config is empty")
	}

	var single Config
	if err := json.Unmarshal(config, &single); err == nil && single.Name != "" {
		managerConfig = []*Config{&single}
		return nil
	}

	var multiple []*Config
	if err := json.Unmarshal(config, &multiple); err != nil {
		return fmt.Errorf("parse database config failed: %w", err)
	}
	managerConfig = multiple
	return nil
}

func GetDatabaseManager() (*DatabaseManager, error) {
	if len(managerConfig) == 0 {
		return nil, fmt.Errorf("databases: not initialized, call Init() first")
	}
	managerOnce.Do(func() {
		m := &DatabaseManager{db: make(map[string]DBConn)}
		for _, cfg := range managerConfig {
			cfg.SetDefaults()
			if err := cfg.Validate(); err != nil {
				managerErr = fmt.Errorf("invalid database config [%s]: %w", cfg.Name, err)
				return
			}
			db, err := NewDBConn(cfg)
			if err != nil {
				managerErr = fmt.Errorf("database connect failed [%s]: %w", cfg.Name, err)
				return
			}
			m.db[cfg.Name] = db
		}
		if managerErr == nil {
			manager = m
		}
	})
	if managerErr != nil {
		return nil, managerErr
	}
	return manager, nil
}

func (m *DatabaseManager) GetDBDao(name string) BaseDao {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if db, exists := m.db[name]; exists {
		return NewBaseDao(db)
	}
	return nil
}

func (m *DatabaseManager) GetDbInterface(name string) DBConn {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if db, exists := m.db[name]; exists {
		return db
	}
	return nil
}
