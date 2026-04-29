package i18n

import "errors"

var (
	ErrInvalidConfig   = errors.New("i18n: invalid config")
	ErrNotInitialized  = errors.New("i18n: not initialized")
	ErrUnknownLanguage = errors.New("i18n: unknown language")
)
