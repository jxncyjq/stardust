package components

import "fmt"

// recoverToError 将 panic 转换为 error，供各组件 Init 方法使用 defer。
func recoverToError(retErr *error, name string) {
	if r := recover(); r != nil {
		*retErr = fmt.Errorf("%s: panic: %v", name, r)
	}
}

// requireConfig 从 configFn 中获取指定 key 的配置，key 不存在时 panic（被 recoverToError 捕获）。
func requireConfig(configFn func(string) []byte, key string) []byte {
	cfg := configFn(key)
	if cfg == nil {
		panic(fmt.Sprintf("config key %q not found", key))
	}
	return cfg
}
