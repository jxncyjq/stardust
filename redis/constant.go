package redis

import "errors"

type InsertOP int32

const (
	InsertOP_BEFORE InsertOP = 0
	InsertOP_AFTER  InsertOP = 1
)

var ErrRedisAddrsEmpty = errors.New("redis addrs is empty")
