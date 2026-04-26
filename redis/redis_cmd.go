package redis

import (
	"crypto/tls"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config represents Redis configuration.
type Config struct {
	Name         string        `json:"name"`       //实例名称
	Addrs        []string      `json:"addrs"`      //集群地址
	Password     string        `json:"password"`   //密码
	KeyPrefix    string        `json:"key_prefix"` //key前缀
	DbIndex      int           `json:"db_index"`   //数据索引
	DialTimeout  time.Duration `json:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	ReadOnly     bool          `json:"read_only"`
	// PoolSize applies per cluster node and not for the whole cluster.
	PoolSize           int           `json:"pool_size"`
	PoolTimeout        time.Duration `json:"pool_timeout"`
	IdleTimeout        time.Duration `json:"idle_timeout"`
	IdleCheckFrequency time.Duration `json:"idle_check_frequency"`
	UseCluster         bool          `json:"use_cluster"`
	TLSConfig          *tls.Config   `json:"tls_config"`
}

type RedisCmd interface {
	redis.Cmdable
}

// NewRedisCmd creates a Redis client from a single config without mutating the
// package-level default instance.
func NewRedisCmd(c *Config) (RedisCmd, error) {
	if len(c.Addrs) == 0 {
		return nil, ErrRedisAddrsEmpty
	}
	// 设置默认连接池配置
	poolSize := c.PoolSize
	if poolSize <= 0 {
		poolSize = 10
	}
	poolTimeout := c.PoolTimeout
	if poolTimeout <= 0 {
		poolTimeout = 4 * time.Second
	}
	if c.UseCluster {
		return redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        c.Addrs,
			ReadOnly:     c.ReadOnly,
			Password:     c.Password,
			DialTimeout:  c.DialTimeout,
			ReadTimeout:  c.ReadTimeout,
			WriteTimeout: c.WriteTimeout,
			PoolSize:     poolSize,
			PoolTimeout:  poolTimeout,
			TLSConfig:    c.TLSConfig,
		}), nil
	}
	return redis.NewClient(&redis.Options{
		Addr:         c.Addrs[0],
		Password:     c.Password,
		DB:           c.DbIndex,
		DialTimeout:  c.DialTimeout,
		ReadTimeout:  c.ReadTimeout,
		WriteTimeout: c.WriteTimeout,
		PoolSize:     poolSize,
		PoolTimeout:  poolTimeout,
		TLSConfig:    c.TLSConfig,
	}), nil
}

var redisCon RedisCmd

// GetRedisDb returns the default Redis client.
//
// The default client is the first entry in the config passed to Init and is
// created through the same initialization path as GetRedisManager. If Init has
// not loaded any config yet, GetRedisDb returns nil.
func GetRedisDb() RedisCmd {
	_ = ensureRedisInitialized()
	return redisCon
}
