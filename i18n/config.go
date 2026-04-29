package i18n

import (
	"fmt"
	"strings"

	"golang.org/x/text/language"
)

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

	c.DefaultLanguage = strings.TrimSpace(c.DefaultLanguage)
	c.QueryKey = strings.TrimSpace(c.QueryKey)
	c.HeaderKey = strings.TrimSpace(c.HeaderKey)

	if c.DefaultLanguage == "" || c.QueryKey == "" || c.HeaderKey == "" {
		return fmt.Errorf("%w: language keys are required", ErrInvalidConfig)
	}
	if len(c.SupportedLanguages) == 0 {
		return fmt.Errorf("%w: supported languages are required", ErrInvalidConfig)
	}

	defaultTag, err := language.Parse(c.DefaultLanguage)
	if err != nil {
		return fmt.Errorf("%w: default language %q", ErrUnknownLanguage, c.DefaultLanguage)
	}
	foundDefault := false
	for _, raw := range c.SupportedLanguages {
		lang := strings.TrimSpace(raw)
		if lang == "" {
			return fmt.Errorf("%w: empty supported language", ErrInvalidConfig)
		}
		tag, err := language.Parse(lang)
		if err != nil {
			return fmt.Errorf("%w: supported language %q", ErrUnknownLanguage, lang)
		}
		if tag == defaultTag {
			foundDefault = true
		}
	}
	if !foundDefault {
		return fmt.Errorf("%w: default language %q must be in supported_languages", ErrInvalidConfig, c.DefaultLanguage)
	}
	return nil
}
