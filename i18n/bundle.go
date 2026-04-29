package i18n

import (
	"embed"
	"fmt"
	"path"

	"github.com/BurntSushi/toml"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.toml
var localeFS embed.FS

func LoadEmbeddedBundle() (*goi18n.Bundle, error) {
	cfg, err := GetConfig()
	if err != nil {
		return nil, err
	}

	defaultTag, err := language.Parse(cfg.DefaultLanguage)
	if err != nil {
		return nil, fmt.Errorf("%w: default language %q", ErrUnknownLanguage, cfg.DefaultLanguage)
	}

	bundle := goi18n.NewBundle(defaultTag)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	for _, lang := range cfg.SupportedLanguages {
		if _, err := bundle.LoadMessageFileFS(localeFS, path.Join("locales", lang+".toml")); err != nil {
			return nil, err
		}
	}
	return bundle, nil
}
