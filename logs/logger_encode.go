package logs

import (
	"encoding/json"
	"runtime/debug"
	"strings"

	"go.uber.org/zap"
)

func StacktraceField() zap.Field {
	b := debug.Stack()
	result := ""
	var jsonData interface{}
	if err := json.Unmarshal(b, &jsonData); err != nil {
		// 不是有效的 JSON，原样输出
		for k, line := range strings.Split(string(b), "\n") {
			if k == 0 {
				result += line
			} else {
				result += "\n\t" + line
			}
		}
	} else {
		// 格式化为标准 JSON 字符串
		formatted, _ := json.MarshalIndent(jsonData, "", "  ")
		result = string(formatted)
	}
	return zap.Any("stacktrace", result) // 直接写入 JSON 数组
}
