package authz

import (
	"fmt"

	"github.com/casbin/casbin/v3"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/jxncyjq/stardust/databases"
	"gorm.io/gorm"
)

var (
	getDatabaseManager = func() (databaseManager, error) {
		return databases.GetDatabaseManager()
	}
	newGormAdapterByDB = func(db *gorm.DB) (interface{}, error) {
		return gormadapter.NewAdapterByDB(db)
	}
	newGormAdapter = func(driver, dsn string) (interface{}, error) {
		return gormadapter.NewAdapter(driver, dsn)
	}
)

type databaseManager interface {
	GetDBDao(string) databases.BaseDao
}

func newPolicyAdapter(cfg Config) (interface{}, error) {
	switch cfg.Adapter {
	case DefaultAdapter:
		return nil, nil
	case GormAdapter:
		return newGormPolicyAdapter(cfg)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedAdapter, cfg.Adapter)
	}
}

func newGormPolicyAdapter(cfg Config) (interface{}, error) {
	if dbName := cfg.DatabaseName; dbName != "" {
		mgr, err := getDatabaseManager()
		if err != nil {
			return nil, fmt.Errorf("authz: get database manager: %w", err)
		}
		dao := mgr.GetDBDao(dbName)
		if dao == nil {
			return nil, fmt.Errorf("%w: database %q not found", ErrInvalidConfig, dbName)
		}
		return newGormAdapterByDB((*gorm.DB)(dao.Native()))
	}

	return newGormAdapter(cfg.Driver, cfg.DSN)
}

func newEnforcer(modelPath string, adapter interface{}, policyPath string) (*casbin.Enforcer, error) {
	if adapter == nil {
		return casbin.NewEnforcer(modelPath, policyPath)
	}
	return casbin.NewEnforcer(modelPath, adapter)
}
