package httpServer

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestGetFuncEmptyBodyDelete 复现「无请求体的 DELETE 请求携带 Content-Type: application/json
// 时返回 EOF」的问题：客户端删除设备/解绑打印机时，body 为空但 header 为 json，
// gin 选择 JSON 绑定并对空 body 解码，返回 io.EOF，被包装成 errCode=40000、errMsg="EOF"。
func TestGetFuncEmptyBodyDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	h := NewHandler("test.delete", []string{"test"}, func(c *gin.Context, _ struct{}) (any, error) {
		return gin.H{"ok": true}, nil
	})
	r.Handle(http.MethodDelete, "/x/:id", h.GetFunc())

	req := httptest.NewRequest(http.MethodDelete, "/x/1", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp BaseResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v, body=%s", err, w.Body.String())
	}
	if resp.ErrCode != 0 {
		t.Fatalf("空 body 的 DELETE 期望 errCode=0，实际 errCode=%d errMsg=%q", resp.ErrCode, resp.ErrMsg)
	}
}
