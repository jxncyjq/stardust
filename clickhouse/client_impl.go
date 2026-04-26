package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"

	_ "github.com/ClickHouse/clickhouse-go/v2" // register driver
)

type clickhouseClient struct {
	db *sql.DB
}

func (c *clickhouseClient) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := c.db.ExecContext(ctx, query, args...)
	return err
}

// Query 执行 SELECT 并将结果扫描到 dest（*[]T 或 *[]*T）。
func (c *clickhouseClient) Query(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	return scanRows(rows, dest)
}

// QueryRow 执行 SELECT 并将单行结果扫描到 dest（struct 指针）。
func (c *clickhouseClient) QueryRow(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}
	return scanRow(rows, dest)
}

func (c *clickhouseClient) AsyncInsert(ctx context.Context, query string, wait bool, args ...interface{}) error {
	suffix := "ASYNC INSERT"
	if wait {
		suffix = "ASYNC INSERT WAIT"
	}
	_, err := c.db.ExecContext(ctx, fmt.Sprintf("%s %s", suffix, query), args...)
	return err
}

func (c *clickhouseClient) PrepareBatch(ctx context.Context, query string) (*sql.Stmt, error) {
	return c.db.PrepareContext(ctx, query)
}

func (c *clickhouseClient) DB() *sql.DB {
	return c.db
}

func (c *clickhouseClient) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

// compile-time interface check
var _ ClickHouseCli = (*clickhouseClient)(nil)

// scanRows 通用行扫描：dest 必须是 *[]T 或 *[]*T，T 为 struct。
func scanRows(rows *sql.Rows, dest interface{}) error {
	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Ptr || destVal.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("clickhouse: dest must be a pointer to a slice")
	}

	sliceVal := destVal.Elem()
	elemType := sliceVal.Type().Elem()
	isPtr := elemType.Kind() == reflect.Ptr
	if isPtr {
		elemType = elemType.Elem()
	}

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	for rows.Next() {
		elem := reflect.New(elemType)
		fields := structFields(elem.Elem(), cols)
		if err = rows.Scan(fields...); err != nil {
			return err
		}
		if isPtr {
			sliceVal.Set(reflect.Append(sliceVal, elem))
		} else {
			sliceVal.Set(reflect.Append(sliceVal, elem.Elem()))
		}
	}
	return rows.Err()
}

// scanRow 扫描单行到 struct 指针。
func scanRow(rows *sql.Rows, dest interface{}) error {
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Ptr {
		return fmt.Errorf("clickhouse: dest must be a pointer")
	}
	fields := structFields(destVal.Elem(), cols)
	return rows.Scan(fields...)
}

// structFields 按列名顺序返回 struct 字段的指针（匹配 `db` tag 或字段名）。
func structFields(v reflect.Value, cols []string) []interface{} {
	t := v.Type()
	fieldMap := make(map[string]reflect.Value, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("db")
		if tag == "" {
			tag = f.Name
		}
		fieldMap[tag] = v.Field(i)
	}
	ptrs := make([]interface{}, len(cols))
	for i, col := range cols {
		if fv, ok := fieldMap[col]; ok {
			ptrs[i] = fv.Addr().Interface()
		} else {
			var discard interface{}
			ptrs[i] = &discard
		}
	}
	return ptrs
}
