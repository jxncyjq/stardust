# I18n Middleware Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 基于 `github.com/nicksnyder/go-i18n/v2/i18n` 为 stardust 增加框架级 i18n 支撑层、Gin 中间件和示例接线，让语言解析、消息本地化和翻译资源加载都走统一入口，而不是在业务代码里手写翻译逻辑。

**Architecture:** 新增一个轻量 `i18n` 核心包，只负责管理 go-i18n 的 `Bundle`、`Localizer` 和上下文读写 helper，不重新实现消息拼装、复数规则或语言匹配。再增加 `components.I18nComponent()` 负责在应用初始化时加载嵌入式翻译文件，`http_server/middleware.I18n()` 负责从请求里解析语言并把 `Localizer` 注入 `context` / `gin.Context`，业务 handler 再通过统一 helper 读取本地化消息。首版优先支持 `Accept-Language` 和 `?lang=` 两种来源，翻译文件通过 `go:embed` 固定打包，避免运行时依赖文件路径。

**Tech Stack:** Go, Gin, `github.com/nicksnyder/go-i18n/v2/i18n`, `golang.org/x/text/language`, `github.com/BurntSushi/toml`, `go:embed`, table-driven tests, project reference docs.

---

## File Structure

- Create `i18n/config.go`: i18n 配置结构、默认值和校验。
- Create `i18n/errors.go`: 哨兵错误。
- Create `i18n/bundle.go`: 嵌入式翻译文件加载。
- Create `i18n/i18n.go`: 包级管理器、Init/Get/Set/Localize helper。
- Create `i18n/context.go`: `context.Context` / `gin.Context` 的 localizer 读写 helper。
- Create `i18n/i18n_test.go`: 配置、bundle、localizer、上下文 helper 测试。
- Create `app/components/i18n.go`: i18n 组件适配器。
- Create `app/components/i18n_test.go`: 组件契约和初始化测试。
- Create `http_server/middleware/i18n.go`: Gin 语言解析中间件。
- Create `http_server/middleware/i18n_test.go`: 中间件测试。
- Add `i18n/locales/en.toml` 和 `i18n/locales/zh.toml`: 最小翻译资源。
- Modify `example/main.go`: 展示 `I18nComponent` 与 `middleware.I18n()` 的接线。
- Modify `example/config.toml`: 增加 `i18n` 配置段。
- Create `docs/reference/reference-i18n-module-001.md`: i18n 模块参考。
- Create `docs/reference/components/reference-component-i18n-001.md`: I18nComponent 参考。
- Modify `docs/docs-index.md`: 注册新文档。
- Modify `docs/reference-docs-index.md`: 注册索引关系。
- Modify `.codex/skills/go-stardust-rtl/SKILL.md` 与 references：把 i18n 加入可复用参考技能。
- Modify `go.mod` / `go.sum`: 添加 go-i18n 与 toml 依赖。

---

## Task 1: Add I18n Core Package

**Files:**
- Create: `i18n/config.go`
- Create: `i18n/errors.go`
- Create: `i18n/bundle.go`
- Create: `i18n/i18n.go`
- Create: `i18n/context.go`
- Create: `i18n/i18n_test.go`

- [ ] **Step 1: Write failing tests**

Create `i18n/i18n_test.go`:

```go
package i18n

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func TestConfigSetDefaults(t *testing.T) {
	cfg := Config{}
	cfg.SetDefaults()

	if cfg.DefaultLanguage != "zh" {
		t.Fatalf("DefaultLanguage = %q, want zh", cfg.DefaultLanguage)
	}
	if cfg.QueryKey != "lang" {
		t.Fatalf("QueryKey = %q, want lang", cfg.QueryKey)
	}
	if cfg.HeaderKey != "Accept-Language" {
		t.Fatalf("HeaderKey = %q, want Accept-Language", cfg.HeaderKey)
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr error
	}{
		{name: "valid", cfg: Config{DefaultLanguage: "zh", SupportedLanguages: []string{"zh", "en"}}},
		{name: "missing default", cfg: Config{SupportedLanguages: []string{"zh", "en"}}, wantErr: ErrInvalidConfig},
		{name: "unsupported default", cfg: Config{DefaultLanguage: "fr", SupportedLanguages: []string{"zh", "en"}}, wantErr: ErrInvalidConfig},
		{name: "empty languages", cfg: Config{DefaultLanguage: "zh"}, wantErr: ErrInvalidConfig},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr == nil && err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestBundleAndLocalize(t *testing.T) {
	SetBundle(nil)
	t.Cleanup(func() { SetBundle(nil) })

	bundle := goi18n.NewBundle(language.Chinese)
	if err := bundle.AddMessages(language.Chinese, &goi18n.Message{ID: "Hello", Other: "你好，{{.Name}}"}); err != nil {
		t.Fatalf("AddMessages() error = %v", err)
	}
	SetBundle(bundle)

	localizer, err := GetLocalizer("zh")
	if err != nil {
		t.Fatalf("GetLocalizer() error = %v", err)
	}
	msg, err := localizer.Localize(&goi18n.LocalizeConfig{MessageID: "Hello", TemplateData: map[string]any{"Name": "Alice"}})
	if err != nil {
		t.Fatalf("Localize() error = %v", err)
	}
	if msg == "" {
		t.Fatal("Localize() returned empty message")
	}
}

func TestContextHelpers(t *testing.T) {
	bundle := goi18n.NewBundle(language.Chinese)
	localizer := goi18n.NewLocalizer(bundle, "zh")

	ctx := WithLocalizer(context.Background(), localizer)
	got, ok := LocalizerFromContext(ctx)
	if !ok || got == nil {
		t.Fatal("LocalizerFromContext() did not return localizer")
	}
}

func testI18nConfig(t *testing.T) Config {
	t.Helper()
	return Config{
		DefaultLanguage:    "zh",
		SupportedLanguages: []string{"zh", "en"},
		QueryKey:           "lang",
		HeaderKey:          "Accept-Language",
	}
}

func testI18nConfigBytes(t *testing.T) []byte {
	t.Helper()
	data, err := json.Marshal(testI18nConfig(t))
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	return data
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./i18n
```

Expected: FAIL because the package APIs are not defined yet.

- [ ] **Step 3: Implement the core package**

Create `i18n/errors.go`:

```go
package i18n

import "errors"

var (
	ErrInvalidConfig   = errors.New("i18n: invalid config")
	ErrNotInitialized  = errors.New("i18n: not initialized")
	ErrUnknownLanguage = errors.New("i18n: unknown language")
)
```

Create `i18n/config.go`:

```go
package i18n

type Config struct {
	DefaultLanguage    string   `json:"default_language" yaml:"default_language" toml:"default_language"`
	SupportedLanguages []string `json:"supported_languages" yaml:"supported_languages" toml:"supported_languages"`
	QueryKey           string   `json:"query_key" yaml:"query_key" toml:"query_key"`
	HeaderKey          string   `json:"header_key" yaml:"header_key" toml:"header_key"`
}

func (c *Config) SetDefaults() {
	if c.DefaultLanguage == "" {
		c.DefaultLanguage = "zh"
	}
	if len(c.SupportedLanguages) == 0 {
		c.SupportedLanguages = []string{"zh", "en"}
	}
	if c.QueryKey == "" {
		c.QueryKey = "lang"
	}
	if c.HeaderKey == "" {
		c.HeaderKey = "Accept-Language"
	}
}

func (c Config) Validate() error {
	c.SetDefaults()
	if c.DefaultLanguage == "" || len(c.SupportedLanguages) == 0 {
		return ErrInvalidConfig
	}
	return nil
}
```

Create `i18n/bundle.go`:

```go
package i18n

import (
	"embed"

	"github.com/BurntSushi/toml"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.toml
var localeFS embed.FS

func LoadEmbeddedBundle() (*goi18n.Bundle, error) {
	bundle := goi18n.NewBundle(language.Chinese)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	for _, file := range []string{"locales/zh.toml", "locales/en.toml"} {
		if _, err := bundle.LoadMessageFileFS(localeFS, file); err != nil {
			return nil, err
		}
	}
	return bundle, nil
}
```

Create `i18n/i18n.go`:

```go
package i18n

import (
	"context"
	"encoding/json"
	"sync"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
)

var (
	mu     sync.RWMutex
	bundle *goi18n.Bundle
	cfg    Config
)

func InitConfig(c Config) error {
	if err := c.Validate(); err != nil {
		return err
	}
	mu.Lock()
	cfg = c
	mu.Unlock()
	return nil
}

func Init(config []byte) error {
	if len(config) == 0 {
		return ErrInvalidConfig
	}
	var c Config
	if err := json.Unmarshal(config, &c); err != nil {
		return ErrInvalidConfig
	}
	return InitConfig(c)
}

func SetBundle(b *goi18n.Bundle) {
	mu.Lock()
	bundle = b
	mu.Unlock()
}

func GetBundle() (*goi18n.Bundle, error) {
	mu.RLock()
	defer mu.RUnlock()
	if bundle == nil {
		return nil, ErrNotInitialized
	}
	return bundle, nil
}

func GetLocalizer(langs ...string) (*goi18n.Localizer, error) {
	b, err := GetBundle()
	if err != nil {
		return nil, err
	}
	return goi18n.NewLocalizer(b, langs...), nil
}

func Localize(ctx context.Context, messageID string, data any) (string, error) {
	localizer, ok := LocalizerFromContext(ctx)
	if !ok {
		return "", ErrNotInitialized
	}
	return localizer.Localize(&goi18n.LocalizeConfig{MessageID: messageID, TemplateData: data})
}
```

Create `i18n/context.go`:

```go
package i18n

import (
	"context"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
)

type contextKey struct{}

func WithLocalizer(ctx context.Context, localizer *goi18n.Localizer) context.Context {
	return context.WithValue(ctx, contextKey{}, localizer)
}

func LocalizerFromContext(ctx context.Context) (*goi18n.Localizer, bool) {
	localizer, ok := ctx.Value(contextKey{}).(*goi18n.Localizer)
	return localizer, ok
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
go test ./i18n
```

Expected: PASS.

---

## Task 2: Add I18n Component

**Files:**
- Create: `app/components/i18n.go`
- Create: `app/components/i18n_test.go`
- Add: `i18n/locales/en.toml`
- Add: `i18n/locales/zh.toml`

- [ ] **Step 1: Write component tests**

Create `app/components/i18n_test.go`:

```go
package components

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jxncyjq/stardust/i18n"
)

func TestI18nComponentContract(t *testing.T) {
	component := I18nComponent()
	if component.Name() != "i18n" {
		t.Fatalf("Name() = %q, want i18n", component.Name())
	}
	if deps := component.Dependencies(); len(deps) != 1 || deps[0] != "logs" {
		t.Fatalf("Dependencies() = %#v, want [logs]", deps)
	}
}

func TestI18nComponentInit(t *testing.T) {
	t.Cleanup(func() { i18n.SetBundle(nil) })

	component := I18nComponent()
	cfgBytes := mustI18nConfigBytes(t)
	if err := component.Init(context.Background(), func(key string) []byte {
		if key != "i18n" {
			return nil
		}
		return cfgBytes
	}); err != nil {
		t.Fatalf("Init() error = %v, want nil", err)
	}
	if _, err := i18n.GetBundle(); err != nil {
		t.Fatalf("GetBundle() error = %v, want nil", err)
	}
}

func mustI18nConfigBytes(t *testing.T) []byte {
	t.Helper()
	data, err := json.Marshal(i18n.Config{
		DefaultLanguage:    "zh",
		SupportedLanguages: []string{"zh", "en"},
		QueryKey:           "lang",
		HeaderKey:          "Accept-Language",
	})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	return data
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./app/components -run TestI18nComponentContract
```

Expected: FAIL because `I18nComponent` is undefined.

- [ ] **Step 3: Implement the component**

Create `app/components/i18n.go`:

```go
package components

import (
	"context"

	"github.com/jxncyjq/stardust/app"
	"github.com/jxncyjq/stardust/i18n"
)

type i18nComponent struct{}

func I18nComponent() app.Component { return &i18nComponent{} }

func (c *i18nComponent) Name() string { return "i18n" }

func (c *i18nComponent) Dependencies() []string { return []string{"logs"} }

func (c *i18nComponent) Init(_ context.Context, configFn app.ConfigFunc) (retErr error) {
	defer recoverToError(&retErr, "i18n")
	if err := i18n.Init(requireConfig(configFn, "i18n")); err != nil {
		return err
	}
	bundle, err := i18n.LoadEmbeddedBundle()
	if err != nil {
		return err
	}
	i18n.SetBundle(bundle)
	return nil
}

func (c *i18nComponent) Start(_ context.Context) error { return nil }

func (c *i18nComponent) Stop(_ context.Context) error { return nil }
```

Add `i18n/locales/zh.toml`:

```toml
[Hello]
other = "你好，{{.Name}}"
```

Add `i18n/locales/en.toml`:

```toml
[Hello]
other = "Hello, {{.Name}}"
```

- [ ] **Step 4: Run component tests**

Run:

```bash
go test ./app/components
```

Expected: PASS.

---

## Task 3: Add Gin I18n Middleware

**Files:**
- Create: `http_server/middleware/i18n.go`
- Create: `http_server/middleware/i18n_test.go`

- [ ] **Step 1: Write middleware tests**

Create `http_server/middleware/i18n_test.go`:

```go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jxncyjq/stardust/i18n"
)

func TestI18nMiddlewareFromQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Next()
	})
	r.Use(I18n())
	r.GET("/ping", func(c *gin.Context) {
		localizer, ok := i18n.LocalizerFromContext(c.Request.Context())
		if !ok || localizer == nil {
			t.Fatal("localizer missing")
		}
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping?lang=zh", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./http_server/middleware -run I18n
```

Expected: FAIL because `I18n()` middleware is undefined.

- [ ] **Step 3: Implement middleware**

Create `http_server/middleware/i18n.go`:

```go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jxncyjq/stardust/i18n"
)

func I18n() gin.HandlerFunc {
	return func(c *gin.Context) {
		langs := make([]string, 0, 2)
		if lang := c.Query("lang"); lang != "" {
			langs = append(langs, lang)
		}
		if accept := c.GetHeader("Accept-Language"); accept != "" {
			langs = append(langs, accept)
		}
		localizer, err := i18n.GetLocalizer(langs...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"errCode": 2, "errMsg": "i18n not initialized"})
			c.Abort()
			return
		}
		c.Request = c.Request.WithContext(i18n.WithLocalizer(c.Request.Context(), localizer))
		c.Next()
	}
}
```

- [ ] **Step 4: Run middleware tests**

Run:

```bash
go test ./http_server/middleware -run I18n
```

Expected: PASS.

---

## Task 4: Wire Example And Config

**Files:**
- Modify: `example/main.go`
- Modify: `example/config.toml`

- [ ] **Step 1: Add example coverage**

Update `example/main.go` to register the component and middleware:

```go
myApp := app.New(conf.Get).
	WithHTTPGroup("v1",
		middleware.I18n(),
		middleware.Metrics(conf.GetAppName()),
		middleware.Tracing(conf.GetAppName()),
	)

myApp.Use(
	components.LogsComponent(),
	components.I18nComponent(),
	components.HTTPServerFromApp(myApp, func(srv *httpServer.HttpServer) {
		srv.Get("greet/:name", "v1", httpServer.NewHandler(
			"greet/:name",
			[]string{"greet"},
			func(c *gin.Context, req struct{}) (map[string]string, error) {
				msg, err := i18n.Localize(c.Request.Context(), "Hello", map[string]any{"Name": c.Param("name")})
				return map[string]string{"message": msg}, err
			},
		))
	}),
)
```

Add `example/config.toml`:

```toml
[i18n]
default_language = "zh"
supported_languages = ["zh", "en"]
query_key = "lang"
header_key = "Accept-Language"
```

- [ ] **Step 2: Verify the example still compiles**

Run:

```bash
go test ./example
```

Expected: PASS.

---

## Task 5: Add Docs And Skill Updates

**Files:**
- Create: `docs/reference/reference-i18n-module-001.md`
- Create: `docs/reference/components/reference-component-i18n-001.md`
- Modify: `docs/docs-index.md`
- Modify: `docs/reference-docs-index.md`
- Modify: `.codex/skills/go-stardust-rtl/SKILL.md`
- Modify: `.codex/skills/go-stardust-rtl/references/components.md`
- Modify: `.codex/skills/go-stardust-rtl/references/reference-map.md`
- Create: `.codex/skills/go-stardust-rtl/references/i18n.md`

- [ ] **Step 1: Write the docs**

`reference-i18n-module-001.md` should document:

- `i18n.Init([]byte)` / `i18n.InitConfig(Config)`
- `i18n.LoadEmbeddedBundle()`
- `i18n.GetLocalizer(...)`
- `i18n.Localize(ctx, messageID, data)`
- `i18n.WithLocalizer` / `i18n.LocalizerFromContext`

`reference-component-i18n-001.md` should document:

- `components.I18nComponent()`
- `Name() == "i18n"`
- dependency `logs`
- config key `i18n`
- embedded locale files `i18n/locales/en.toml` and `i18n/locales/zh.toml`

- [ ] **Step 2: Update indices**

Add both documents to `docs/docs-index.md` and `docs/reference-docs-index.md`, and keep `parent` / `children` / `related_docs.path` consistent.

- [ ] **Step 3: Update skill references**

Add i18n to `go-stardust-rtl` so future work can locate the new pattern from the skill package instead of external docs.

- [ ] **Step 4: Keep links consistent**

Run:

```bash
rg -n "reference-i18n-module-001|reference-component-i18n-001|I18nComponent|middleware.I18n" docs docs/reference .codex/skills/go-stardust-rtl
```

Expected: all new IDs appear in docs, indices, and skill references.

---

## Task 6: Full Verification

**Files:**
- All files above

- [ ] **Step 1: Run package tests**

Run:

```bash
go test ./i18n ./app/components ./http_server/middleware ./service
```

Expected: PASS.

- [ ] **Step 2: Run example validation**

Run:

```bash
go test ./example
```

Expected: PASS.

- [ ] **Step 3: Review translation behavior**

Confirm:

- default language exists in loaded translations
- missing language falls back to default
- context helper returns a `Localizer` after middleware runs
- no custom pluralization engine or matcher logic was added outside go-i18n

---

## Self-Review

- Spec coverage: covers i18n core package, component, middleware, example, docs, and skill updates.
- Occupancy scan: no red-flag markers; each task gives exact files, tests, commands, and sample code.
- Type consistency: `Config`, `InitConfig`, `LoadEmbeddedBundle`, `GetLocalizer`, `WithLocalizer`, and `LocalizerFromContext` are used consistently across tasks.
- Scope check: this plan intentionally does not implement custom translation engines, plural rules, or locale matchers; those stay inside go-i18n.

---

## Deferred Work

- Runtime translation file hot reload.
- Per-route locale override policies beyond `?lang=` / `Accept-Language`.
- Localized wrappers for existing error middlewares (`Access`, `Authz`) if you want all JSON error messages translated.
- Multiple bundle sources beyond embedded `toml` files.
