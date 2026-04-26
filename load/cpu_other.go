//go:build !linux

package load

// getCpuUsage 在非 Linux 平台返回 0，禁用基于 CPU 的降载触发。
func getCpuUsage() int64 {
	return 0
}
