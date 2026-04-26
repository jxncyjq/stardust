package clickhouse

import "errors"

var (
	ErrQueryFailed   = errors.New("clickhouse: query failed")
	ErrExecFailed    = errors.New("clickhouse: exec failed")
	ErrNilClient     = errors.New("clickhouse: client is nil")
	ErrInvalidConfig = errors.New("clickhouse: invalid config")
	ErrBatchFailed   = errors.New("clickhouse: batch operation failed")
)
