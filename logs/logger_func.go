package logs

import (
	"time"

	"go.uber.org/zap"
)

func String(k, v string) zap.Field {
	return zap.String(k, v)
}

func Int(k string, v int) zap.Field {
	return zap.Int(k, v)
}

func Int64(k string, v int64) zap.Field {
	return zap.Int64(k, v)
}

func Int32(k string, v int32) zap.Field {
	return zap.Int32(k, v)
}

func ErrorInfo(v error) zap.Field {
	if v == nil {
		return zap.Skip()
	}
	return zap.Error(v)
}

func Duration(k string, v time.Duration) zap.Field {
	return zap.Duration(k, v)
}
