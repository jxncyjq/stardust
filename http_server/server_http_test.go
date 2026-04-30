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

func TestHttpServerPatchAndDelete_RegisterRoutes(t *testing.T) {
	config := []byte(`{"port":18080,"address":"127.0.0.1","path":"/","worker_id":1}`)
	server, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("NewHttpServer() error = %v", err)
	}

	handler := NewHandlerRaw[struct{}]("noop", nil, func(c *gin.Context, _ struct{}) {
		c.Status(http.StatusNoContent)
	})

	server.Patch("items/:id", "", handler)
	server.Delete("items/:id", "", handler)
	server.AddGroup("v1")
	server.Patch("users/:id", "v1", handler)
	server.Delete("users/:id", "v1", handler)

	wantRoutes := map[string]bool{
		http.MethodPatch + " /api/items/:id":     false,
		http.MethodDelete + " /api/items/:id":    false,
		http.MethodPatch + " /api/v1/users/:id":  false,
		http.MethodDelete + " /api/v1/users/:id": false,
	}
	assertRoutesRegistered(t, server.engine.Routes(), wantRoutes)
}

func TestHttpServerStandardMethods_RegisterRoutes(t *testing.T) {
	config := []byte(`{"port":18080,"address":"127.0.0.1","path":"/","worker_id":1}`)
	server, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("NewHttpServer() error = %v", err)
	}

	handler := NewHandlerRaw[struct{}]("noop", nil, func(c *gin.Context, _ struct{}) {
		c.Status(http.StatusNoContent)
	})

	server.Head("items/:id", "", handler)
	server.Options("items/:id", "", handler)
	server.Connect("items/:id", "", handler)
	server.Trace("items/:id", "", handler)
	server.AddGroup("v1")
	server.Head("users/:id", "v1", handler)
	server.Options("users/:id", "v1", handler)
	server.Connect("users/:id", "v1", handler)
	server.Trace("users/:id", "v1", handler)

	wantRoutes := map[string]bool{
		http.MethodHead + " /api/items/:id":       false,
		http.MethodOptions + " /api/items/:id":    false,
		http.MethodConnect + " /api/items/:id":    false,
		http.MethodTrace + " /api/items/:id":      false,
		http.MethodHead + " /api/v1/users/:id":    false,
		http.MethodOptions + " /api/v1/users/:id": false,
		http.MethodConnect + " /api/v1/users/:id": false,
		http.MethodTrace + " /api/v1/users/:id":   false,
	}
	assertRoutesRegistered(t, server.engine.Routes(), wantRoutes)
}

func TestBackendHTTPMethods_RegisterRoutes(t *testing.T) {
	config := []byte(`{"port":18080,"address":"127.0.0.1","path":"/","worker_id":1}`)
	backend, err := NewBackend(config)
	if err != nil {
		t.Fatalf("NewBackend() error = %v", err)
	}

	handler := NewHandlerRaw[struct{}]("items/:id", nil, func(c *gin.Context, _ struct{}) {
		c.Status(http.StatusNoContent)
	})

	backend.Put("", handler)
	backend.Patch("", handler)
	backend.Delete("", handler)
	backend.Head("", handler)
	backend.Options("", handler)
	backend.Connect("", handler)
	backend.Trace("", handler)

	wantRoutes := map[string]bool{
		http.MethodPut + " /api/items/:id":     false,
		http.MethodPatch + " /api/items/:id":   false,
		http.MethodDelete + " /api/items/:id":  false,
		http.MethodHead + " /api/items/:id":    false,
		http.MethodOptions + " /api/items/:id": false,
		http.MethodConnect + " /api/items/:id": false,
		http.MethodTrace + " /api/items/:id":   false,
	}
	assertRoutesRegistered(t, backend.httpServer.engine.Routes(), wantRoutes)
}

func assertRoutesRegistered(t *testing.T, routes gin.RoutesInfo, wantRoutes map[string]bool) {
	t.Helper()

	for _, route := range routes {
		key := route.Method + " " + route.Path
		if _, ok := wantRoutes[key]; ok {
			wantRoutes[key] = true
		}
	}
	for route, found := range wantRoutes {
		if !found {
			t.Errorf("routes missing route %q", route)
		}
	}
}
