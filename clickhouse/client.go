package clickhouse

import (
	"context"
	"database/sql"
)

// ClickHouseCli 定义 ClickHouse 的操作接口。
type ClickHouseCli interface {
	// Exec 执行 DDL 或 INSERT/UPDATE 等不返回行的语句。
	Exec(ctx context.Context, query string, args ...interface{}) error

	// Query 执行查询并将结果扫描到 dest（需传切片指针，元素为 struct）。
	// 使用 sqlx 风格：dest 必须是 *[]T 或 *[]*T。
	Query(ctx context.Context, dest interface{}, query string, args ...interface{}) error

	// QueryRow 执行查询并将单行结果扫描到 dest（需传 struct 指针）。
	QueryRow(ctx context.Context, dest interface{}, query string, args ...interface{}) error

	// AsyncInsert 异步写入，wait=true 时等待服务端确认落盘。
	AsyncInsert(ctx context.Context, query string, wait bool, args ...interface{}) error

	// PrepareBatch 准备批量写入，返回 *sql.Stmt 后由调用方 Exec/Close。
	// 高吞吐场景（埋点/事件）推荐使用此方法。
	PrepareBatch(ctx context.Context, query string) (*sql.Stmt, error)

	// DB 返回底层 *sql.DB，用于事务或复杂场景。
	DB() *sql.DB

	// Ping 检查连接健康状态。
	Ping(ctx context.Context) error
}
