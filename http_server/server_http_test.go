package httpServer

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHttpServerAddNativeHandler_PreservesGinPathParams(t *testing.T) {
	config := []byte(`{"port":18080,"address":"127.0.0.1","path":"/","worker_id":1}`)
	server, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("NewHttpServer() error = %v", err)
	}

	server.AddNativeHandler(http.MethodGet, "v1/admin/content/entries/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"id": c.Param("id")})
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/api/v1/admin/content/entries/101", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	server.engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("AddNativeHandler(%q) HTTP status = %d, want %d, body=%s", "/api/v1/admin/content/entries/101", w.Code, http.StatusOK, w.Body.String())
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal(response) error = %v, body=%s", err, w.Body.String())
	}
	if got := body["id"]; got != "101" {
		t.Errorf("AddNativeHandler(%q) param id = %q, want %q", "/api/v1/admin/content/entries/101", got, "101")
	}
}
