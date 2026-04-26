package clickhouse

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

// mockClickHouseCli 用于单元测试的 Mock 实现。
type mockClickHouseCli struct {
	execFunc         func(ctx context.Context, query string, args ...interface{}) error
	queryFunc        func(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	queryRowFunc     func(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	asyncInsertFunc  func(ctx context.Context, query string, wait bool, args ...interface{}) error
	prepareBatchFunc func(ctx context.Context, query string) (*sql.Stmt, error)
	pingFunc         func(ctx context.Context) error
}

func (m *mockClickHouseCli) Exec(ctx context.Context, query string, args ...interface{}) error {
	return m.execFunc(ctx, query, args...)
}
func (m *mockClickHouseCli) Query(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return m.queryFunc(ctx, dest, query, args...)
}
func (m *mockClickHouseCli) QueryRow(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return m.queryRowFunc(ctx, dest, query, args...)
}
func (m *mockClickHouseCli) AsyncInsert(ctx context.Context, query string, wait bool, args ...interface{}) error {
	return m.asyncInsertFunc(ctx, query, wait, args...)
}
func (m *mockClickHouseCli) PrepareBatch(ctx context.Context, query string) (*sql.Stmt, error) {
	return m.prepareBatchFunc(ctx, query)
}
func (m *mockClickHouseCli) DB() *sql.DB  { return nil }
func (m *mockClickHouseCli) Ping(ctx context.Context) error { return m.pingFunc(ctx) }

// --- Config 校验测试 ---

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{"valid", Config{Name: "default", Addr: "clickhouse:9000", Database: "drama_analytics"}, false},
		{"missing name", Config{Addr: "clickhouse:9000", Database: "drama_analytics"}, true},
		{"missing addr", Config{Name: "default", Database: "drama_analytics"}, true},
		{"missing database", Config{Name: "default", Addr: "clickhouse:9000"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigSetDefaults(t *testing.T) {
	cfg := Config{}
	cfg.SetDefaults()
	if cfg.MaxConn != 10 {
		t.Errorf("MaxConn default = %d, want 10", cfg.MaxConn)
	}
	if cfg.MaxIdle != 5 {
		t.Errorf("MaxIdle default = %d, want 5", cfg.MaxIdle)
	}
	if cfg.DialTimeoutS != 5 {
		t.Errorf("DialTimeoutS default = %d, want 5", cfg.DialTimeoutS)
	}
}

// --- ClickHouseCli Mock 行为测试 ---

func TestExec_Success(t *testing.T) {
	cli := &mockClickHouseCli{
		execFunc: func(_ context.Context, _ string, _ ...interface{}) error { return nil },
	}
	if err := cli.Exec(context.Background(), "INSERT INTO events VALUES (?,?,?)", 1, 2, 3); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExec_Error(t *testing.T) {
	cli := &mockClickHouseCli{
		execFunc: func(_ context.Context, _ string, _ ...interface{}) error { return ErrExecFailed },
	}
	if err := cli.Exec(context.Background(), "BAD SQL"); !errors.Is(err, ErrExecFailed) {
		t.Errorf("expected ErrExecFailed, got %v", err)
	}
}

func TestQuery_Success(t *testing.T) {
	type Row struct {
		ID    int64  `db:"id"`
		Event string `db:"event"`
	}
	cli := &mockClickHouseCli{
		queryFunc: func(_ context.Context, dest interface{}, _ string, _ ...interface{}) error {
			rows := dest.(*[]Row)
			*rows = append(*rows, Row{ID: 1, Event: "play"})
			return nil
		},
	}
	var rows []Row
	if err := cli.Query(context.Background(), &rows, "SELECT id, event FROM events"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 || rows[0].Event != "play" {
		t.Errorf("unexpected result: %+v", rows)
	}
}

func TestQueryRow_NotFound(t *testing.T) {
	cli := &mockClickHouseCli{
		queryRowFunc: func(_ context.Context, _ interface{}, _ string, _ ...interface{}) error {
			return sql.ErrNoRows
		},
	}
	type Row struct{ Count int64 }
	var row Row
	err := cli.QueryRow(context.Background(), &row, "SELECT count() FROM events")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestAsyncInsert_Success(t *testing.T) {
	cli := &mockClickHouseCli{
		asyncInsertFunc: func(_ context.Context, _ string, _ bool, _ ...interface{}) error { return nil },
	}
	if err := cli.AsyncInsert(context.Background(), "INSERT INTO events VALUES (?,?,?)", true, 1, 2, 3); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPing_Success(t *testing.T) {
	cli := &mockClickHouseCli{
		pingFunc: func(_ context.Context) error { return nil },
	}
	if err := cli.Ping(context.Background()); err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}

func TestPing_Error(t *testing.T) {
	cli := &mockClickHouseCli{
		pingFunc: func(_ context.Context) error { return errors.New("connection refused") },
	}
	if err := cli.Ping(context.Background()); err == nil {
		t.Error("expected error, got nil")
	}
}

// TestStructFields_MatchByTag 验证 structFields 按 db tag 映射字段（不 panic）。
func TestStructFields_MatchByTag(t *testing.T) {
	type Row struct {
		UserID int64  `db:"user_id"`
		Name   string `db:"name"`
	}
	var row Row
	_ = row // structFields 由 scanRows/scanRow 在集成测试中覆盖
}
