package databases

import (
	"fmt"
	"time"
)

type ztime time.Time

func (t ztime) MarshalJSON() ([]byte, error) {
	var stamp = fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02 15:04:05"))
	return []byte(stamp), nil
}

// Time 获取time.Time类型，方便拓展方法
func (t ztime) Time() time.Time {
	return time.Time(t)
}

// Format 格式化
func (t ztime) Format() string {
	return time.Time(t).Format("2006-01-02 15:04:05")
}

// FormatSimple 简单格式化
func (t ztime) FormatSimple() string {
	return time.Time(t).Format("2006-01-02")
}
