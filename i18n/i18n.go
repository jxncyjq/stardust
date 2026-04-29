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

func InitConfig(c Config) error {
	if err := c.Validate(); err != nil {
		return err
	}

	mu.Lock()
	cfg = c
	mu.Unlock()
	return nil
}

func SetBundle(b *goi18n.Bundle) {
	mu.Lock()
	bundle = b
	mu.Unlock()
}

func GetConfig() (Config, error) {
	mu.RLock()
	defer mu.RUnlock()
	if cfg.DefaultLanguage == "" && len(cfg.SupportedLanguages) == 0 && cfg.QueryKey == "" && cfg.HeaderKey == "" {
		return Config{}, ErrNotInitialized
	}
	return cfg, nil
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

	if len(langs) == 0 {
		cfg, err := GetConfig()
		if err != nil {
			return nil, err
		}
		langs = append(langs, cfg.DefaultLanguage)
	}
	return goi18n.NewLocalizer(b, langs...), nil
}

func Localize(ctx context.Context, messageID string, data any) (string, error) {
	if localizer, ok := LocalizerFromContext(ctx); ok && localizer != nil {
		return localizer.Localize(&goi18n.LocalizeConfig{
			MessageID:    messageID,
			TemplateData: data,
		})
	}

	cfg, err := GetConfig()
	if err != nil {
		return "", err
	}
	localizer, err := GetLocalizer(cfg.DefaultLanguage)
	if err != nil {
		return "", err
	}
	return localizer.Localize(&goi18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: data,
	})
}
