# Casbin Authz Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Casbin-based authorization to stardust as a framework-managed module with component lifecycle support and HTTP middleware integration.

**Architecture:** Create a new `authz` package that wraps Casbin behind a small stardust interface. Add `components.AuthzComponent()` to initialize it through `app.Application`, and add `http_server/middleware.Authz(...)` for Gin route authorization. Keep authentication (`Access`) and authorization (`Authz`) separate.

**Tech Stack:** Go, Gin, stardust `app/components`, Casbin v3 (`github.com/casbin/casbin/v3`), table-driven tests, project reference docs.

---

## File Structure

- Create `authz/config.go`: configuration types and validation.
- Create `authz/authz.go`: public interfaces, global manager, init/get helpers.
- Create `authz/casbin.go`: Casbin enforcer implementation.
- Create `authz/errors.go`: sentinel errors.
- Create `authz/authz_test.go`: config and authorizer tests.
- Create `app/components/authz.go`: component adapter.
- Create `app/components/authz_test.go`: component tests.
- Create `http_server/middleware/authz.go`: Gin authorization middleware.
- Create `http_server/middleware/authz_test.go`: middleware tests.
- Modify `example/main.go`: show `AuthzComponent` and group middleware usage.
- Add `example/config/casbin/model.conf`: minimal RBAC model.
- Add `example/config/casbin/policy.csv`: minimal example policy.
- Create `docs/reference/reference-authz-module-001.md`: module reference.
- Create `docs/reference/components/reference-component-authz-001.md`: component reference.
- Modify `docs/docs-index.md`: register new docs.
- Modify `docs/reference-docs-index.md`: add core doc entry.
- Modify `.codex/skills/go-stardust-rtl/SKILL.md` and references: include authz in reusable skill.
- Modify `go.mod` / `go.sum`: add Casbin v3 dependency.

---

## Task 1: Add authz Config And Errors

**Files:**
- Create: `authz/config.go`
- Create: `authz/errors.go`
- Create: `authz/authz_test.go`

- [ ] **Step 1: Write failing config tests**

Create `authz/authz_test.go`:

```go
package authz

import (
	"errors"
	"testing"
)

func TestConfigSetDefaults(t *testing.T) {
	cfg := Config{}
	cfg.SetDefaults()

	if cfg.SubjectKey != "id" {
		t.Fatalf("SubjectKey = %q, want id", cfg.SubjectKey)
	}
	if cfg.ObjectMode != ObjectModeRoute {
		t.Fatalf("ObjectMode = %q, want %q", cfg.ObjectMode, ObjectModeRoute)
	}
	if cfg.ActionMode != ActionModeMethod {
		t.Fatalf("ActionMode = %q, want %q", cfg.ActionMode, ActionModeMethod)
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want error
	}{
		{
			name: "valid file adapter",
			cfg: Config{
				ModelPath:  "model.conf",
				PolicyPath: "policy.csv",
			},
		},
		{
			name: "missing model",
			cfg:  Config{PolicyPath: "policy.csv"},
			want: ErrInvalidConfig,
		},
		{
			name: "missing file policy",
			cfg:  Config{ModelPath: "model.conf"},
			want: ErrInvalidConfig,
		},
		{
			name: "unsupported adapter",
			cfg: Config{
				ModelPath:  "model.conf",
				PolicyPath: "policy.csv",
				Adapter:    "gorm",
			},
			want: ErrUnsupportedAdapter,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cfg.SetDefaults()
			err := tt.cfg.Validate()
			if tt.want == nil && err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
			if tt.want != nil && !errors.Is(err, tt.want) {
				t.Fatalf("Validate() error = %v, want %v", err, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./authz
```

Expected: FAIL because package `authz` or types are not defined.

- [ ] **Step 3: Implement config and errors**

Create `authz/errors.go`:

```go
package authz

import "errors"

var (
	ErrInvalidConfig      = errors.New("authz: invalid config")
	ErrNotInitialized     = errors.New("authz: not initialized")
	ErrUnsupportedAdapter = errors.New("authz: unsupported adapter")
	ErrForbidden          = errors.New("authz: forbidden")
)
```

Create `authz/config.go`:

```go
package authz

import "fmt"

const (
	AdapterFile = "file"

	ObjectModeRoute = "route"
	ObjectModePath  = "path"

	ActionModeMethod = "method"
)

type Config struct {
	ModelPath  string `json:"model_path" yaml:"model_path" toml:"model_path"`
	PolicyPath string `json:"policy_path" yaml:"policy_path" toml:"policy_path"`
	Adapter    string `json:"adapter" yaml:"adapter" toml:"adapter"`

	SubjectKey string `json:"subject_key" yaml:"subject_key" toml:"subject_key"`
	ObjectMode string `json:"object_mode" yaml:"object_mode" toml:"object_mode"`
	ActionMode string `json:"action_mode" yaml:"action_mode" toml:"action_mode"`
}

func (c *Config) SetDefaults() {
	if c.Adapter == "" {
		c.Adapter = AdapterFile
	}
	if c.SubjectKey == "" {
		c.SubjectKey = "id"
	}
	if c.ObjectMode == "" {
		c.ObjectMode = ObjectModeRoute
	}
	if c.ActionMode == "" {
		c.ActionMode = ActionModeMethod
	}
}

func (c Config) Validate() error {
	if c.ModelPath == "" {
		return fmt.Errorf("%w: model_path is required", ErrInvalidConfig)
	}
	if c.Adapter != AdapterFile {
		return fmt.Errorf("%w: %s", ErrUnsupportedAdapter, c.Adapter)
	}
	if c.PolicyPath == "" {
		return fmt.Errorf("%w: policy_path is required for file adapter", ErrInvalidConfig)
	}
	if c.ObjectMode != ObjectModeRoute && c.ObjectMode != ObjectModePath {
		return fmt.Errorf("%w: object_mode must be route or path", ErrInvalidConfig)
	}
	if c.ActionMode != ActionModeMethod {
		return fmt.Errorf("%w: action_mode must be method", ErrInvalidConfig)
	}
	if c.SubjectKey == "" {
		return fmt.Errorf("%w: subject_key is required", ErrInvalidConfig)
	}
	return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
go test ./authz
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add authz/config.go authz/errors.go authz/authz_test.go
git commit -m "feat: add authz config"
```

---

## Task 2: Add Casbin Authorizer

**Files:**
- Modify: `authz/authz.go`
- Create: `authz/casbin.go`
- Modify: `authz/authz_test.go`
- Modify: `go.mod`
- Modify: `go.sum`

- [ ] **Step 1: Add tests for file-backed authorization**

Append to `authz/authz_test.go`:

```go
func TestCasbinAuthorizerEnforce(t *testing.T) {
	dir := t.TempDir()
	modelPath := filepath.Join(dir, "model.conf")
	policyPath := filepath.Join(dir, "policy.csv")

	model := `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && r.act == p.act
`
	policy := "p, admin, /api/v1/users/:id, GET\ng, alice, admin\n"

	if err := os.WriteFile(modelPath, []byte(model), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(policyPath, []byte(policy), 0o644); err != nil {
		t.Fatal(err)
	}

	az, err := NewCasbinAuthorizer(Config{
		ModelPath:  modelPath,
		PolicyPath: policyPath,
	})
	if err != nil {
		t.Fatalf("NewCasbinAuthorizer() error = %v", err)
	}

	allowed, err := az.Enforce(context.Background(), "alice", "/api/v1/users/1001", "GET")
	if err != nil {
		t.Fatalf("Enforce() error = %v", err)
	}
	if !allowed {
		t.Fatal("Enforce() = false, want true")
	}

	allowed, err = az.Enforce(context.Background(), "bob", "/api/v1/users/1001", "GET")
	if err != nil {
		t.Fatalf("Enforce() bob error = %v", err)
	}
	if allowed {
		t.Fatal("Enforce() bob = true, want false")
	}
}
```

Add imports:

```go
import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./authz
```

Expected: FAIL because `NewCasbinAuthorizer` and `Authorizer` are missing.

- [ ] **Step 3: Add Casbin dependency**

Run:

```bash
go get github.com/casbin/casbin/v3
```

Expected: `go.mod` and `go.sum` updated.

- [ ] **Step 4: Implement authorizer**

Create `authz/authz.go`:

```go
package authz

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

type Authorizer interface {
	Enforce(ctx context.Context, sub, obj, act string) (bool, error)
}

var (
	managerMu sync.RWMutex
	manager   Authorizer
)

func Init(config []byte) error {
	if len(config) == 0 {
		return fmt.Errorf("%w: config is empty", ErrInvalidConfig)
	}

	var cfg Config
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("%w: parse config failed: %v", ErrInvalidConfig, err)
	}

	authorizer, err := NewCasbinAuthorizer(cfg)
	if err != nil {
		return err
	}

	managerMu.Lock()
	manager = authorizer
	managerMu.Unlock()
	return nil
}

func GetAuthorizer() (Authorizer, error) {
	managerMu.RLock()
	defer managerMu.RUnlock()
	if manager == nil {
		return nil, ErrNotInitialized
	}
	return manager, nil
}
```

Create `authz/casbin.go`:

```go
package authz

import (
	"context"

	casbin "github.com/casbin/casbin/v3"
)

type CasbinAuthorizer struct {
	enforcer *casbin.Enforcer
}

func NewCasbinAuthorizer(cfg Config) (*CasbinAuthorizer, error) {
	cfg.SetDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	enforcer, err := casbin.NewEnforcer(cfg.ModelPath, cfg.PolicyPath)
	if err != nil {
		return nil, err
	}
	return &CasbinAuthorizer{enforcer: enforcer}, nil
}

func (a *CasbinAuthorizer) Enforce(_ context.Context, sub, obj, act string) (bool, error) {
	return a.enforcer.Enforce(sub, obj, act)
}
```

Remove `validateForInit`, `errMissingFile`, and `errFileExists` from `authz/config.go` if unused.

- [ ] **Step 5: Run tests**

Run:

```bash
go test ./authz
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add authz go.mod go.sum
git commit -m "feat: add casbin authorizer"
```

---

## Task 3: Add Authz App Component

**Files:**
- Create: `app/components/authz.go`
- Create: `app/components/authz_test.go`

- [ ] **Step 1: Write component tests**

Create `app/components/authz_test.go`:

```go
package components

import (
	"context"
	"testing"
)

func TestAuthzComponentContract(t *testing.T) {
	component := AuthzComponent()

	if component.Name() != "authz" {
		t.Fatalf("Name() = %q, want authz", component.Name())
	}

	deps := component.Dependencies()
	if len(deps) != 1 || deps[0] != "logs" {
		t.Fatalf("Dependencies() = %#v, want [logs]", deps)
	}

	if err := component.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := component.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./app/components -run TestAuthzComponentContract
```

Expected: FAIL because `AuthzComponent` is undefined.

- [ ] **Step 3: Implement component**

Create `app/components/authz.go`:

```go
package components

import (
	"context"

	"github.com/jxncyjq/stardust/app"
	"github.com/jxncyjq/stardust/authz"
)

type authzComponent struct{}

func AuthzComponent() app.Component { return &authzComponent{} }

func (c *authzComponent) Name() string { return "authz" }

func (c *authzComponent) Dependencies() []string { return []string{"logs"} }

func (c *authzComponent) Init(_ context.Context, configFn app.ConfigFunc) (retErr error) {
	defer recoverToError(&retErr, "authz")
	return authz.Init(requireConfig(configFn, "authz"))
}

func (c *authzComponent) Start(_ context.Context) error { return nil }

func (c *authzComponent) Stop(_ context.Context) error { return nil }
```

- [ ] **Step 4: Run component tests**

Run:

```bash
go test ./app/components
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add app/components/authz.go app/components/authz_test.go
git commit -m "feat: add authz component"
```

---

## Task 4: Add Gin Authz Middleware

**Files:**
- Create: `http_server/middleware/authz.go`
- Create: `http_server/middleware/authz_test.go`

- [ ] **Step 1: Write middleware tests**

Create `http_server/middleware/authz_test.go`:

```go
package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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

func TestAuthzMiddlewareAllowsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	az := &fakeAuthorizer{allowed: true}

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("id", "alice")
		c.Next()
	})
	r.GET("/api/v1/users/:id", Authz(az))
	r.GET("/api/v1/users/:id", func(c *gin.Context) {
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

func TestAuthzMiddlewareRejectsMissingSubject(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/users/:id", Authz(&fakeAuthorizer{allowed: true}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/1001", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthzMiddlewareRejectsForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("id", "bob")
		c.Next()
	})
	r.GET("/api/v1/users/:id", Authz(&fakeAuthorizer{allowed: false}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/1001", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./http_server/middleware -run Authz
```

Expected: FAIL because `Authz` middleware is undefined.

- [ ] **Step 3: Implement middleware**

Create `http_server/middleware/authz.go`:

```go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jxncyjq/stardust/authz"
)

func Authz(authorizer authz.Authorizer) gin.HandlerFunc {
	return func(c *gin.Context) {
		if authorizer == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"errCode": 2, "errMsg": "授权器未初始化"})
			c.Abort()
			return
		}

		sub, ok := c.Get("id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"errCode": 2, "errMsg": "用户未认证"})
			c.Abort()
			return
		}

		subject, ok := sub.(string)
		if !ok || subject == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"errCode": 2, "errMsg": "用户信息无效"})
			c.Abort()
			return
		}

		obj := c.FullPath()
		if obj == "" {
			obj = c.Request.URL.Path
		}
		act := c.Request.Method

		allowed, err := authorizer.Enforce(c.Request.Context(), subject, obj, act)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"errCode": 2, "errMsg": "授权检查失败"})
			c.Abort()
			return
		}
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"errCode": 2, "errMsg": "无访问权限"})
			c.Abort()
			return
		}

		c.Next()
	}
}
```

- [ ] **Step 4: Run middleware tests**

Run:

```bash
go test ./http_server/middleware -run Authz
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add http_server/middleware/authz.go http_server/middleware/authz_test.go
git commit -m "feat: add authz middleware"
```

---

## Task 5: Add Example Wiring And Casbin Files

**Files:**
- Modify: `example/main.go`
- Create: `example/config/casbin/model.conf`
- Create: `example/config/casbin/policy.csv`

- [ ] **Step 1: Add Casbin model**

Create `example/config/casbin/model.conf`:

```ini
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && r.act == p.act
```

- [ ] **Step 2: Add example policy**

Create `example/config/casbin/policy.csv`:

```csv
p, admin, /api/v1/user/:id, GET
p, admin, /api/v1/game/:id, GET
p, admin, /api/v1/game/:id/score, POST
g, alice, admin
```

- [ ] **Step 3: Update example app imports**

In `example/main.go`, add imports:

```go
	"github.com/jxncyjq/stardust/authz"
```

Only add it if the example directly calls `authz.GetAuthorizer()`.

- [ ] **Step 4: Wire component and middleware in example**

In the component list, add:

```go
components.AuthzComponent(),
```

When building protected HTTP groups, get the authorizer after component init is available through setup time:

```go
authorizer, err := authz.GetAuthorizer()
if err != nil {
    logger.Warn("authz not initialized", zap.Error(err))
} else {
    srv.AddGroup("secure", middleware.Access(), middleware.Authz(authorizer))
}
```

Register one route into `secure` group to demonstrate authorization. Keep existing public routes unchanged to avoid breaking the example.

- [ ] **Step 5: Run example build test**

Run:

```bash
go test ./example
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add example/main.go example/config/casbin/model.conf example/config/casbin/policy.csv
git commit -m "example: wire authz middleware"
```

---

## Task 6: Add Reference Docs And Skill Updates

**Files:**
- Create: `docs/reference/reference-authz-module-001.md`
- Create: `docs/reference/components/reference-component-authz-001.md`
- Modify: `docs/docs-index.md`
- Modify: `docs/reference-docs-index.md`
- Modify: `.codex/skills/go-stardust-rtl/SKILL.md`
- Modify: `.codex/skills/go-stardust-rtl/references/components.md`
- Modify: `.codex/skills/go-stardust-rtl/references/microservice.md`

- [ ] **Step 1: Create module reference**

Create `docs/reference/reference-authz-module-001.md`:

```markdown
---
id: "reference-authz-module-001"
title: "authz 模块使用参考"
aliases: ["authz模块", "Casbin授权", "权限中间件"]
type: "reference"
category: "backend/library"
tags: ["authz", "casbin", "rbac", "middleware", "authorization"]
version: "1.0.0"
created: "2026-04-27"
updated: "2026-04-27"
author: "jxncyjq"
status: "published"
parent: "reference-component-authz-001"
children: []
related_docs:
  - id: "reference-component-authz-001"
    relation: "depends_on"
    path: "./components/reference-component-authz-001.md"
---

# authz 模块使用参考

## 概述

`authz` 模块基于 Casbin v3 提供授权能力。认证仍由 `middleware.Access()` 等模块负责，`authz` 只判断已认证主体是否允许访问资源。

## 配置

```toml
[authz]
model_path = "example/config/casbin/model.conf"
policy_path = "example/config/casbin/policy.csv"
adapter = "file"
subject_key = "id"
object_mode = "route"
action_mode = "method"
```

## 使用

```go
authz.Init(conf.Get("authz"))
authorizer, err := authz.GetAuthorizer()
allowed, err := authorizer.Enforce(ctx, "alice", "/api/v1/users/:id", "GET")
```

## HTTP 中间件

```go
srv.AddGroup("secure", middleware.Access(), middleware.Authz(authorizer))
```

## 注意事项

- Go 依赖使用 `github.com/casbin/casbin/v3`。
- 第一版只支持 file adapter。
- 中间件默认使用 `gin.Context` 中的 `id` 作为 subject。
- object 默认使用 `c.FullPath()`，也就是路由模板。
```

- [ ] **Step 2: Create component reference**

Create `docs/reference/components/reference-component-authz-001.md`:

```markdown
---
id: "reference-component-authz-001"
title: "AuthzComponent 使用说明"
aliases: ["AuthzComponent", "授权组件", "Casbin组件"]
type: "reference"
category: "backend/library/components"
tags: ["app", "components", "authz", "casbin", "authorization"]
version: "1.0.0"
created: "2026-04-27"
updated: "2026-04-27"
author: "jxncyjq"
status: "published"
parent: "guide-app-components-module-001"
children:
  - "reference-authz-module-001"
related_docs:
  - id: "guide-app-components-module-001"
    relation: "depends_on"
    path: "../../guides/guide-app-components-module-001.md"
  - id: "reference-authz-module-001"
    relation: "parent_of"
    path: "../reference-authz-module-001.md"
---

# AuthzComponent 使用说明

## 概述

`components.AuthzComponent()` 读取配置 key `authz`，调用 `authz.Init(...)` 初始化 Casbin 授权器。

## 组件契约

| 项 | 值 |
| --- | --- |
| 构造函数 | `components.AuthzComponent()` |
| 组件名 | `authz` |
| 依赖 | `logs` |
| 配置 key | `authz` |
| Init | 初始化 Casbin authorizer |
| Start | 无操作 |
| Stop | 无操作 |

## 使用

```go
app.New(conf.Get).
    Use(
        components.LogsComponent(),
        components.AuthzComponent(),
    )
```
```

- [ ] **Step 3: Update indexes**

In `docs/docs-index.md`, add rows for:

```markdown
| [[reference-authz-module-001]] | authz 模块使用参考 | reference | `docs/reference/reference-authz-module-001.md` | authz, casbin, rbac, middleware, authorization |
| [[reference-component-authz-001]] | AuthzComponent 使用说明 | reference | `docs/reference/components/reference-component-authz-001.md` | app, components, authz, casbin, authorization |
```

Add keyword rows:

```markdown
| authz 模块 / Casbin 授权 | [[reference-authz-module-001]] |
| AuthzComponent / 授权组件 | [[reference-component-authz-001]] |
```

In `docs/reference-docs-index.md`, add both docs to current core documents.

- [ ] **Step 4: Update skill references**

In `.codex/skills/go-stardust-rtl/SKILL.md`, update description and task routing to mention `authz`.

In `.codex/skills/go-stardust-rtl/references/components.md`, add `AuthzComponent()` to the component table.

In `.codex/skills/go-stardust-rtl/references/microservice.md`, add a short `authz` section:

```markdown
## authz

基于 Casbin v3 做授权。认证和授权分离：`Access()` 负责身份，`Authz(authorizer)` 负责权限。

```go
srv.AddGroup("secure", middleware.Access(), middleware.Authz(authorizer))
```

第一版只支持 file adapter。
```
```

- [ ] **Step 5: Validate docs relationships**

Run a relationship check equivalent to:

```bash
rg -n "reference-authz-module-001|reference-component-authz-001" docs/docs-index.md docs/reference-docs-index.md docs/reference .codex/skills/go-stardust-rtl
```

Expected: both IDs appear in the index and relevant docs.

- [ ] **Step 6: Commit**

```bash
git add docs/reference/reference-authz-module-001.md docs/reference/components/reference-component-authz-001.md docs/docs-index.md docs/reference-docs-index.md .codex/skills/go-stardust-rtl
git commit -m "docs: add authz reference"
```

---

## Task 7: Full Verification

**Files:**
- No new files.

- [ ] **Step 1: Run focused package tests**

Run:

```bash
go test ./authz ./app/components ./http_server/middleware
```

Expected: PASS.

- [ ] **Step 2: Run integration-adjacent tests**

Run:

```bash
go test ./http_server ./example
```

Expected: PASS.

- [ ] **Step 3: Run broader validation if dependencies changed cleanly**

Run:

```bash
go test ./...
```

Expected: PASS. If unrelated packages fail due to existing environment requirements, record exact failures in the final handoff.

- [ ] **Step 4: Review dependency impact**

Run:

```bash
go mod tidy
git diff -- go.mod go.sum
```

Expected: only Casbin and its required transitive dependencies are added.

- [ ] **Step 5: Final commit if verification changed files**

```bash
git add go.mod go.sum
git commit -m "chore: tidy casbin dependencies"
```

Only commit if `go mod tidy` changed dependency files.

---

## Self-Review

- Spec coverage: covers authz package, component, middleware, example, docs, skill update, and dependency validation.
- Ambiguity scan: no deferred markers; task code snippets define concrete signatures and paths.
- Type consistency: `Authorizer.Enforce(ctx, sub, obj, act)` is used consistently by package and middleware.
- Scope check: first version is intentionally file-adapter only; GORM adapter and distributed watcher are deferred.

---

## Deferred Work

- GORM adapter backed by `databases` manager.
- Policy reload interval.
- Redis/NATS watcher for distributed policy sync.
- Subject extraction options beyond `gin.Context` key `id`.
- Route-level helper for object/action customization.
