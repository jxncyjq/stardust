package authz

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/jxncyjq/stardust/databases"
	"gorm.io/gorm"
)

func TestConfigSetDefaults(t *testing.T) {
	cfg := Config{}

	cfg.SetDefaults()

	if cfg.Adapter != DefaultAdapter {
		t.Errorf("Config.SetDefaults() Adapter = %q, want %q", cfg.Adapter, DefaultAdapter)
	}
	if cfg.SubjectKey != DefaultSubjectKey {
		t.Errorf("Config.SetDefaults() SubjectKey = %q, want %q", cfg.SubjectKey, DefaultSubjectKey)
	}
	if cfg.ObjectMode != DefaultObjectMode {
		t.Errorf("Config.SetDefaults() ObjectMode = %q, want %q", cfg.ObjectMode, DefaultObjectMode)
	}
	if cfg.ActionMode != DefaultActionMode {
		t.Errorf("Config.SetDefaults() ActionMode = %q, want %q", cfg.ActionMode, DefaultActionMode)
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr error
	}{
		{
			name: "valid file adapter",
			cfg: Config{
				ModelPath:  "model.conf",
				PolicyPath: "policy.csv",
			},
		},
		{
			name: "valid path object mode",
			cfg: Config{
				ModelPath:  "model.conf",
				PolicyPath: "policy.csv",
				ObjectMode: "path",
			},
		},
		{
			name: "valid gorm adapter with database name",
			cfg: Config{
				ModelPath:    "model.conf",
				Adapter:      GormAdapter,
				DatabaseName: "authz",
			},
		},
		{
			name: "valid gorm adapter with dsn",
			cfg: Config{
				ModelPath: "model.conf",
				Adapter:   GormAdapter,
				Driver:    "mysql",
				DSN:       "root:password@tcp(127.0.0.1:3306)/authz",
			},
		},
		{
			name: "missing model path",
			cfg: Config{
				PolicyPath: "policy.csv",
			},
			wantErr: ErrInvalidConfig,
		},
		{
			name: "missing policy path",
			cfg: Config{
				ModelPath: "model.conf",
			},
			wantErr: ErrInvalidConfig,
		},
		{
			name: "unsupported adapter",
			cfg: Config{
				ModelPath:  "model.conf",
				PolicyPath: "policy.csv",
				Adapter:    "database",
			},
			wantErr: ErrUnsupportedAdapter,
		},
		{
			name: "gorm adapter without storage",
			cfg: Config{
				ModelPath: "model.conf",
				Adapter:   GormAdapter,
			},
			wantErr: ErrInvalidConfig,
		},
		{
			name: "blank subject key",
			cfg: Config{
				ModelPath:  "model.conf",
				PolicyPath: "policy.csv",
				SubjectKey: " \t",
				ObjectMode: DefaultObjectMode,
				ActionMode: DefaultActionMode,
			},
			wantErr: ErrInvalidConfig,
		},
		{
			name: "invalid object mode",
			cfg: Config{
				ModelPath:  "model.conf",
				PolicyPath: "policy.csv",
				ObjectMode: "url",
			},
			wantErr: ErrInvalidConfig,
		},
		{
			name: "invalid action mode",
			cfg: Config{
				ModelPath:  "model.conf",
				PolicyPath: "policy.csv",
				ActionMode: "header",
			},
			wantErr: ErrInvalidConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Config.Validate(%+v) error = %v, want nil", tt.cfg, err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Config.Validate(%+v) error = %v, want %v", tt.cfg, err, tt.wantErr)
			}
		})
	}
}

func TestCasbinAuthorizerEnforce(t *testing.T) {
	cfg := writeCasbinFiles(t)
	authorizer, err := NewCasbinAuthorizer(cfg)
	if err != nil {
		t.Fatalf("NewCasbinAuthorizer(%+v) error = %v, want nil", cfg, err)
	}

	allowed, err := authorizer.Enforce(context.Background(), "roleName", "/orders", "GET")
	if err != nil {
		t.Fatalf("Authorizer.Enforce(roleName, /orders, GET) error = %v, want nil", err)
	}
	if !allowed {
		t.Errorf("Authorizer.Enforce(roleName, /orders, GET) = false, want true")
	}

	allowed, err = authorizer.Enforce(context.Background(), "roleName", "/orders", "POST")
	if err != nil {
		t.Fatalf("Authorizer.Enforce(roleName, /orders, POST) error = %v, want nil", err)
	}
	if allowed {
		t.Errorf("Authorizer.Enforce(roleName, /orders, POST) = true, want false")
	}
}

func TestCasbinAuthorizerEnforceWithGormSQLite(t *testing.T) {
	dir := t.TempDir()
	modelPath := filepath.Join(dir, "model.conf")
	if err := os.WriteFile(modelPath, []byte(casbinModel), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v, want nil", modelPath, err)
	}

	cfg := Config{
		ModelPath: modelPath,
		Adapter:   GormAdapter,
		Driver:    "sqlite3",
		DSN:       filepath.Join(dir, "authz.db"),
	}

	authorizer, err := NewCasbinAuthorizer(cfg)
	if err != nil {
		t.Fatalf("NewCasbinAuthorizer(%+v) error = %v, want nil", cfg, err)
	}

	allowed, err := authorizer.Enforce(context.Background(), "roleName", "/orders", "GET")
	if err != nil {
		t.Fatalf("Authorizer.Enforce(roleName, /orders, GET) error = %v, want nil", err)
	}
	if allowed {
		t.Fatalf("Authorizer.Enforce(roleName, /orders, GET) = true, want false")
	}

	if _, err := authorizer.enforcer.AddPolicy("roleName", "/orders", "GET"); err != nil {
		t.Fatalf("Enforcer.AddPolicy() error = %v, want nil", err)
	}
	if err := authorizer.enforcer.SavePolicy(); err != nil {
		t.Fatalf("Enforcer.SavePolicy() error = %v, want nil", err)
	}

	authorizer, err = NewCasbinAuthorizer(cfg)
	if err != nil {
		t.Fatalf("NewCasbinAuthorizer(%+v) reload error = %v, want nil", cfg, err)
	}

	allowed, err = authorizer.Enforce(context.Background(), "roleName", "/orders", "GET")
	if err != nil {
		t.Fatalf("Authorizer.Enforce(roleName, /orders, GET) after SavePolicy error = %v, want nil", err)
	}
	if !allowed {
		t.Fatalf("Authorizer.Enforce(roleName, /orders, GET) after SavePolicy = false, want true")
	}
}

func TestInitGetSetLifecycle(t *testing.T) {
	Set(nil)
	t.Cleanup(func() { Set(nil) })

	if _, err := Get(); !errors.Is(err, ErrNotInitialized) {
		t.Fatalf("Get() before init error = %v, want %v", err, ErrNotInitialized)
	}

	cfg := writeCasbinFiles(t)
	configBytes := marshalConfig(t, cfg)
	if err := Init(configBytes); err != nil {
		t.Fatalf("Init(%s) error = %v, want nil", configBytes, err)
	}

	authorizer, err := Get()
	if err != nil {
		t.Fatalf("Get() after init error = %v, want nil", err)
	}
	allowed, err := authorizer.Enforce(context.Background(), "roleName", "/orders", "GET")
	if err != nil {
		t.Fatalf("Authorizer.Enforce(roleName, /orders, GET) error = %v, want nil", err)
	}
	if !allowed {
		t.Errorf("Authorizer.Enforce(roleName, /orders, GET) = false, want true")
	}

	Set(denyAuthorizer{})
	authorizer, err = Get()
	if err != nil {
		t.Fatalf("Get() after Set() error = %v, want nil", err)
	}
	allowed, err = authorizer.Enforce(context.Background(), "roleName", "/orders", "GET")
	if err != nil {
		t.Fatalf("Authorizer.Enforce(roleName, /orders, GET) after Set() error = %v, want nil", err)
	}
	if allowed {
		t.Errorf("Authorizer.Enforce(roleName, /orders, GET) after Set() = true, want false")
	}
}

func TestInitConfigLifecycle(t *testing.T) {
	Set(nil)
	t.Cleanup(func() { Set(nil) })

	cfg := writeCasbinFiles(t)
	if err := InitConfig(cfg); err != nil {
		t.Fatalf("InitConfig(%+v) error = %v, want nil", cfg, err)
	}
	if _, err := Get(); err != nil {
		t.Fatalf("Get() after InitConfig() error = %v, want nil", err)
	}
}

func TestInitInvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config []byte
	}{
		{
			name:   "empty config",
			config: nil,
		},
		{
			name:   "invalid json",
			config: []byte("{"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Set(nil)
			err := Init(tt.config)
			if !errors.Is(err, ErrInvalidConfig) {
				t.Errorf("Init(%s) error = %v, want %v", tt.config, err, ErrInvalidConfig)
			}
		})
	}
}

func TestNewGormPolicyAdapterWithDatabaseName(t *testing.T) {
	origGetDatabaseManager := getDatabaseManager
	origNewGormAdapterByDB := newGormAdapterByDB
	defer func() {
		getDatabaseManager = origGetDatabaseManager
		newGormAdapterByDB = origNewGormAdapterByDB
	}()

	var gotDBName string
	getDatabaseManager = func() (databaseManager, error) {
		return fakeDatabaseManager{dao: fakeBaseDao{}}, nil
	}
	newGormAdapterByDB = func(db *gorm.DB) (interface{}, error) {
		if db == nil {
			t.Fatal("newGormAdapterByDB received nil db")
		}
		gotDBName = "called"
		return fakeAdapter{}, nil
	}

	adapter, err := newGormPolicyAdapter(Config{
		Adapter:      GormAdapter,
		DatabaseName: "main",
		ModelPath:    "model.conf",
	})
	if err != nil {
		t.Fatalf("newGormPolicyAdapter() error = %v, want nil", err)
	}
	if _, ok := adapter.(fakeAdapter); !ok {
		t.Fatalf("newGormPolicyAdapter() adapter = %#v, want fakeAdapter", adapter)
	}
	if gotDBName != "called" {
		t.Fatal("newGormPolicyAdapter() did not call database adapter constructor")
	}
}

func TestNewGormPolicyAdapterWithDSN(t *testing.T) {
	origNewGormAdapter := newGormAdapter
	defer func() { newGormAdapter = origNewGormAdapter }()

	newGormAdapter = func(driver, dsn string) (interface{}, error) {
		if driver != "mysql" || dsn != "dsn" {
			t.Fatalf("newGormAdapter(%q, %q) unexpected args", driver, dsn)
		}
		return fakeAdapter{}, nil
	}

	adapter, err := newGormPolicyAdapter(Config{
		Adapter: GormAdapter,
		Driver:  "mysql",
		DSN:     "dsn",
	})
	if err != nil {
		t.Fatalf("newGormPolicyAdapter() error = %v, want nil", err)
	}
	if _, ok := adapter.(fakeAdapter); !ok {
		t.Fatalf("newGormPolicyAdapter() adapter = %#v, want fakeAdapter", adapter)
	}
}

type denyAuthorizer struct{}

type fakeAdapter struct{}

type fakeDatabaseManager struct {
	dao fakeBaseDao
}

func (m fakeDatabaseManager) GetDBDao(string) databases.BaseDao {
	return m.dao
}

type fakeBaseDao struct{}

func (fakeBaseDao) Count(interface{}) (int64, error)                       { return 0, nil }
func (fakeBaseDao) Exists(interface{}) (bool, error)                       { return false, nil }
func (fakeBaseDao) InsertOne(interface{}) (int64, error)                   { return 0, nil }
func (fakeBaseDao) InsertMany(...interface{}) (int64, error)               { return 0, nil }
func (fakeBaseDao) Update(interface{}, ...interface{}) (int64, error)      { return 0, nil }
func (fakeBaseDao) UpdateById(interface{}, interface{}) (int64, error)     { return 0, nil }
func (fakeBaseDao) Upsert(interface{}, interface{}) (int64, error)         { return 0, nil }
func (fakeBaseDao) UpsertById(interface{}, interface{}) (int64, error)     { return 0, nil }
func (fakeBaseDao) UpsertMany([]interface{}, []interface{}) (int64, error) { return 0, nil }
func (fakeBaseDao) Delete(interface{}) (int64, error)                      { return 0, nil }
func (fakeBaseDao) DeleteById(interface{}, interface{}) (int64, error)     { return 0, nil }
func (fakeBaseDao) GetDBMetas() (map[string]interface{}, error)            { return nil, nil }
func (fakeBaseDao) GetTableMetas(string) ([]map[string]interface{}, error) { return nil, nil }
func (fakeBaseDao) FindById(interface{}, interface{}) (bool, error)        { return false, nil }
func (fakeBaseDao) FindOne(interface{}) (bool, error)                      { return false, nil }
func (fakeBaseDao) FindMany(interface{}, string, ...interface{}) error     { return nil }
func (fakeBaseDao) FindAndCount(interface{}, databases.Pageable, ...interface{}) (int64, error) {
	return 0, nil
}
func (fakeBaseDao) Query(interface{}, string, ...interface{}) error { return nil }
func (fakeBaseDao) CallProcedure(string, ...interface{}) ([][]map[string]interface{}, error) {
	return nil, nil
}
func (fakeBaseDao) Native() databases.DBConn         { return databases.DBConn(&gorm.DB{}) }
func (fakeBaseDao) Migrations([]interface{}) error   { return nil }
func (fakeBaseDao) NewSession() databases.SessionDao { return fakeSessionDao{} }

type fakeSessionDao struct{ fakeBaseDao }

func (fakeSessionDao) DB() *gorm.DB    { return nil }
func (fakeSessionDao) Begin() error    { return nil }
func (fakeSessionDao) Commit() error   { return nil }
func (fakeSessionDao) Rollback() error { return nil }
func (fakeSessionDao) Close()          {}

func (denyAuthorizer) Enforce(context.Context, string, string, string) (bool, error) {
	return false, nil
}

func marshalConfig(t *testing.T, cfg Config) []byte {
	t.Helper()

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("json.Marshal(%+v) error = %v, want nil", cfg, err)
	}
	return data
}

func writeCasbinFiles(t *testing.T) Config {
	t.Helper()

	dir := t.TempDir()
	modelPath := filepath.Join(dir, "model.conf")
	policyPath := filepath.Join(dir, "policy.csv")

	policy := "p, roleName, /orders, GET\n"

	if err := os.WriteFile(modelPath, []byte(casbinModel), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v, want nil", modelPath, err)
	}
	if err := os.WriteFile(policyPath, []byte(policy), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v, want nil", policyPath, err)
	}

	return Config{
		ModelPath:  modelPath,
		PolicyPath: policyPath,
	}
}

const casbinModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
`
