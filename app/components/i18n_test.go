package components

import (
	"context"
	"encoding/json"
	"testing"

	stardi18n "github.com/jxncyjq/stardust/i18n"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
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
	t.Cleanup(func() { stardi18n.SetBundle(nil) })

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

	localizer, err := stardi18n.GetLocalizer("zh")
	if err != nil {
		t.Fatalf("GetLocalizer() error = %v, want nil", err)
	}
	msg, err := localizer.Localize(&goi18n.LocalizeConfig{
		MessageID:    "Hello",
		TemplateData: map[string]any{"Name": "Alice"},
	})
	if err != nil {
		t.Fatalf("Localize() error = %v, want nil", err)
	}
	if msg == "" {
		t.Fatal("Localize() returned empty message")
	}
}

func TestI18nComponentInitMissingConfig(t *testing.T) {
	component := I18nComponent()
	if err := component.Init(context.Background(), func(string) []byte { return nil }); err == nil {
		t.Fatal("Init() error = nil, want non-nil")
	}
}

func mustI18nConfigBytes(t *testing.T) []byte {
	t.Helper()

	cfg := stardi18n.Config{
		DefaultLanguage:    "zh",
		SupportedLanguages: []string{"zh", "en"},
		QueryKey:           "lang",
		HeaderKey:          "Accept-Language",
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	return data
}
