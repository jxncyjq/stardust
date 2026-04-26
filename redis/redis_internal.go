package redis

import (
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

// expandKey 用于使用前缀扩展为完整的 key 路径
func (r *redisView) expandKey(suffix string) string {
	return r.prefix + RedisKeySep + suffix
}

// truncateKey 用于移除 key 的前缀
// nolint:unused
func (r *redisView) truncateKey(full string) string {
	return strings.Join(strings.Split(full, RedisKeySep)[1:], RedisKeySep)
}

func wrapResult(call func() (interface{}, error)) error {
	result, err := call()
	if nil != err {
		return err
	}
	switch result.(type) {
	case bool:
		if result != true {
			return ErrorResultNotTrue
		}
	case string:
		if result != "OK" {
			return ErrorResultNotOK
		}
	}

	return nil
}
func wrapSliceStringToSliceBytes(call func() ([]string, error)) ([][]byte, error) {
	results, err := call()
	if nil != err {
		return nil, err
	}
	var out [][]byte
	for _, result := range results {
		out = append(out, []byte(result))
	}
	return out, nil
}
func toRangeZMembers(err error, members []string, zSlice []redis.Z) ([]*ZMember, error) {
	if nil != err {
		return nil, err
	}
	var zMembers []*ZMember
	if len(members) > 0 {
		for _, m := range members {
			member := ZMember{
				Member: []byte(m),
			}
			zMembers = append(zMembers, &member)
		}
	}
	if len(zSlice) > 0 { // 修复：原为 len(members)，导致 zSlice 永远不被处理
		for _, m := range zSlice {
			var memberBytes []byte
			switch v := m.Member.(type) { // 安全类型断言，避免运行时 panic
			case string:
				memberBytes = []byte(v)
			case []byte:
				memberBytes = v
			default:
				memberBytes = []byte(fmt.Sprintf("%v", v))
			}
			member := ZMember{
				Score:  m.Score,
				Member: memberBytes,
			}
			zMembers = append(zMembers, &member)
		}
	}
	return zMembers, nil
}
