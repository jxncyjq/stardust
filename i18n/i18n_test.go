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
		{name: "missing default uses fallback", cfg: Config{SupportedLanguages: []string{"zh", "en"}}},
		{name: "unsupported default", cfg: Config{DefaultLanguage: "fr", SupportedLanguages: []string{"zh", "en"}}, wantErr: ErrInvalidConfig},
		{name: "empty languages uses fallback", cfg: Config{DefaultLanguage: "zh"}},
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

	SetBundle(goi18n.NewBundle(language.Chinese))
	cfg := Config{
		DefaultLanguage:    "zh",
		SupportedLanguages: []string{"zh", "en"},
		QueryKey:           "lang",
		HeaderKey:          "Accept-Language",
	}
	if err := InitConfig(cfg); err != nil {
		t.Fatalf("InitConfig() error = %v", err)
	}

	bundle, err := LoadEmbeddedBundle()
	if err != nil {
		t.Fatalf("LoadEmbeddedBundle() error = %v", err)
	}
	SetBundle(bundle)

	localizer, err := GetLocalizer("zh")
	if err != nil {
		t.Fatalf("GetLocalizer() error = %v", err)
	}
	msg, err := localizer.Localize(&goi18n.LocalizeConfig{
		MessageID:    "Hello",
		TemplateData: map[string]any{"Name": "Alice"},
	})
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

func TestInitFromJSON(t *testing.T) {
	SetBundle(nil)
	t.Cleanup(func() { SetBundle(nil) })

	data, err := json.Marshal(Config{
		DefaultLanguage:    "zh",
		SupportedLanguages: []string{"zh", "en"},
		QueryKey:           "lang",
		HeaderKey:          "Accept-Language",
	})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if err := Init(data); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
}
