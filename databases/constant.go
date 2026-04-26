package databases

import "errors"

// Config 数据库连接配置
type Config struct {
	Name           string   `json:"name" hcl:"name"`
	ShowSql        bool     `json:"show_sql" hcl:"show_sql"`
	MaxIdle        int      `json:"max_idle" hcl:"max_idle"`
	MaxConn        int      `json:"max_conn" hcl:"max_conn"`
	Master         string   `json:"master" hcl:"master"`
	Slaves         []string `json:"slaves" hcl:"slaves"`
	UseMasterSlave bool     `json:"use_master_slave" hcl:"use_master_slave"`
	DbType         string   `json:"db_type" hcl:"db_type"`
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Name == "" {
		return errors.New("database name is required")
	}
	if c.Master == "" {
		return errors.New("master connection string is required")
	}
	if c.MaxIdle < 0 {
		return errors.New("max_idle must be >= 0")
	}
	if c.MaxConn <= 0 {
		return errors.New("max_conn must be > 0")
	}
	if c.MaxIdle > c.MaxConn {
		return errors.New("max_idle cannot be greater than max_conn")
	}
	return nil
}

// SetDefaults 设置默认值
func (c *Config) SetDefaults() {
	if c.MaxIdle == 0 {
		c.MaxIdle = 10
	}
	if c.MaxConn == 0 {
		c.MaxConn = 100
	}
	if c.DbType == "" {
		c.DbType = "mysql"
	}
}

var ErrGetEmpty = errors.New("found 0 rows")
var ErrUpdatedEmpty = errors.New("update affected 0 rows")
var ErrDeletedEmpty = errors.New("delete affected 0 rows")
var ErrInsertedEmpty = errors.New("insert affected 0 rows")
var ErrMigrateTableIDEmpty = errors.New("migrate table id nil")
var ErrMigrateTableNameEmpty = errors.New("migrate table name nil")

const (
	SessionDoctorCommit   SessionDoctor = 0
	SessionDoctorRollback SessionDoctor = 1
)
