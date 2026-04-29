package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jxncyjq/stardust/authz"
)

type fakeAuthorizer struct {
	allowed bool
	err     error
	sub     string
	obj     string
	act     string
}

func (f *fakeAuthorizer) Enforce(_ context.Context, sub, obj, act string) (bool, error) {
	f.sub = sub
	f.obj = obj
	f.act = act
	return f.allowed, f.err
}

func TestAuthzAllowsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	az := &fakeAuthorizer{allowed: true}

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("id", "alice")
		c.Next()
	})
	r.GET("/api/v1/users/:id", Authz(az), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/1001", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if az.sub != "alice" {
		t.Fatalf("sub = %q, want alice", az.sub)
	}
	if az.obj != "/api/v1/users/:id" {
		t.Fatalf("obj = %q, want route pattern", az.obj)
	}
	if az.act != http.MethodGet {
		t.Fatalf("act = %q, want GET", az.act)
	}
}

func TestAuthzRejectsMissingSubject(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/users/:id", Authz(&fakeAuthorizer{allowed: true}), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/1001", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthzRejectsForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("id", "bob")
		c.Next()
	})
	r.GET("/api/v1/users/:id", Authz(&fakeAuthorizer{allowed: false}), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/1001", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestAuthzRejectsAuthorizerError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("id", "bob")
		c.Next()
	})
	r.GET("/api/v1/users/:id", Authz(&fakeAuthorizer{err: context.Canceled}), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/1001", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestAuthzUsesGlobalAuthorizerWhenUnset(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { authz.Set(nil) })

	cfgBytes := mustGlobalAuthzConfigBytes(t)
	if err := authz.Init(cfgBytes); err != nil {
		t.Fatalf("authz.Init() error = %v, want nil", err)
	}

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("id", "alice")
		c.Next()
	})
	r.GET("/api/v1/users/:id", Authz(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/1001", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func mustGlobalAuthzConfigBytes(t *testing.T) []byte {
	t.Helper()

	dir := t.TempDir()
	modelPath := filepath.Join(dir, "model.conf")
	policyPath := filepath.Join(dir, "policy.csv")

	model := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
`
	policy := "p, alice, /api/v1/users/:id, GET\n"

	if err := os.WriteFile(modelPath, []byte(model), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v, want nil", modelPath, err)
	}
	if err := os.WriteFile(policyPath, []byte(policy), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v, want nil", policyPath, err)
	}

	cfg := authz.Config{
		ModelPath:  modelPath,
		PolicyPath: policyPath,
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("json.Marshal(%+v) error = %v, want nil", cfg, err)
	}
	return data
}
