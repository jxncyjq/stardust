package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	stardi18n "github.com/jxncyjq/stardust/i18n"
)

func TestI18nMiddlewareUsesQueryLanguage(t *testing.T) {
	initTestI18n(t)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(I18n())
	r.GET("/hello", func(c *gin.Context) {
		msg, err := stardi18n.Localize(c.Request.Context(), "Hello", map[string]any{"Name": "Alice"})
		if err != nil {
			t.Fatalf("Localize() error = %v", err)
		}
		c.JSON(http.StatusOK, gin.H{"message": msg})
	})

	req := httptest.NewRequest(http.MethodGet, "/hello?lang=en", nil)
	req.Header.Set("Accept-Language", "zh")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if !strings.HasPrefix(resp["message"], "Hello, Alice") {
		t.Fatalf("message = %q, want english translation", resp["message"])
	}
}

func TestI18nMiddlewareUsesContextLocalizer(t *testing.T) {
	initTestI18n(t)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(I18n())
	r.GET("/hello", func(c *gin.Context) {
		localizer, ok := stardi18n.LocalizerFromContext(c.Request.Context())
		if !ok || localizer == nil {
			t.Fatal("missing localizer in context")
		}
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestI18nMiddlewareRejectsWhenNotInitialized(t *testing.T) {
	stardi18n.SetBundle(nil)
	t.Cleanup(func() { stardi18n.SetBundle(nil) })

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(I18n())
	r.GET("/hello", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func initTestI18n(t *testing.T) {
	t.Helper()

	stardi18n.SetBundle(nil)
	t.Cleanup(func() { stardi18n.SetBundle(nil) })

	cfgBytes := mustI18nConfigBytes(t)
	if err := stardi18n.Init(cfgBytes); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	bundle, err := stardi18n.LoadEmbeddedBundle()
	if err != nil {
		t.Fatalf("LoadEmbeddedBundle() error = %v", err)
	}
	stardi18n.SetBundle(bundle)
}

func mustI18nConfigBytes(t *testing.T) []byte {
	t.Helper()

	data, err := json.Marshal(stardi18n.Config{
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
