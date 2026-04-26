package conf

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

var (
	config         map[string]interface{}
	vp             *viper.Viper
	mu             sync.RWMutex
	initOnce       sync.Once
	initErr        error
	AppName        string
	AppVersion     string
	RedisKeyPrefix string
)

func GetAppName() string        { return AppName }
func GetAppVersion() string     { return AppVersion }
func GetRedisKeyPrefix() string { return RedisKeyPrefix }

func resolveConfigPath() string {
	runConfig := os.Getenv("runConfig")
	if runConfig != "" {
		return runConfig
	}

	isDebug := os.Getenv("ISDEBUG")
	if isDebug != "" && isDebug != "0" {
		if devConf := os.Getenv("devConf"); devConf != "" {
			return devConf
		}
	}

	return os.Getenv("prodConf")
}

func Init() {
	initOnce.Do(func() {
		configPath := resolveConfigPath()
		if configPath == "" {
			initErr = errors.New("config path is empty, set runConfig or devConf/prodConf")
			return
		}

		v := viper.New()
		v.SetConfigFile(configPath)
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()

		err := v.ReadInConfig()
		if err != nil {
			initErr = err
			return
		}

		globalInfo := v.GetStringMap("global")
		if len(globalInfo) == 0 {
			initErr = errors.New("global configuration is missing in the config file")
			return
		}

		if !v.IsSet("global.app_name") {
			initErr = errors.New("app name is missing in the global configuration")
			return
		}
		appName := v.GetString("global.app_name")

		if !v.IsSet("global.app_version") {
			initErr = errors.New("app version is missing in the global configuration")
			return
		}
		appVersion := v.GetString("global.app_version")

		if !v.IsSet("global.redis_key_prefix") {
			initErr = errors.New("redis key prefix is missing in the global configuration")
			return
		}
		redisKeyPrefix := v.GetString("global.redis_key_prefix")

		mu.Lock()
		vp = v
		config = v.AllSettings()
		AppName = appName
		AppVersion = appVersion
		RedisKeyPrefix = redisKeyPrefix
		mu.Unlock()
	})

	if initErr != nil {
		panic(initErr)
	}
}

func Get(key string) []byte {
	mu.RLock()
	cfg := config
	v := vp
	mu.RUnlock()

	if cfg == nil {
		Init()
		mu.RLock()
		cfg = config
		v = vp
		mu.RUnlock()
	}

	if value, exists := cfg[key]; exists {
		if strValue, ok := value.(string); ok {
			return []byte(strValue)
		}
		bytes, err := json.Marshal(value)
		if err != nil {
			return nil
		}
		return bytes
	}

	if v != nil && v.IsSet(key) {
		value := v.Get(key)
		if strValue, ok := value.(string); ok {
			return []byte(strValue)
		}
		bytes, err := json.Marshal(value)
		if err != nil {
			return nil
		}
		return bytes
	}

	return nil
}

func GetConfigInstance() *viper.Viper {
	mu.RLock()
	v := vp
	mu.RUnlock()
	if v == nil {
		Init()
		mu.RLock()
		v = vp
		mu.RUnlock()
	}
	return v
}
