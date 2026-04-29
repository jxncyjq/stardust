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
