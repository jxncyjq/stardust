package components

import (
	"context"

	"github.com/jxncyjq/stardust/app"
	stardi18n "github.com/jxncyjq/stardust/i18n"
)

type i18nComponent struct{}

// I18nComponent 返回国际化组件，依赖 logs。
func I18nComponent() app.Component { return &i18nComponent{} }

func (c *i18nComponent) Name() string { return "i18n" }

func (c *i18nComponent) Dependencies() []string { return []string{"logs"} }

func (c *i18nComponent) Init(_ context.Context, configFn app.ConfigFunc) (retErr error) {
	defer recoverToError(&retErr, "i18n")
	if err := stardi18n.Init(requireConfig(configFn, "i18n")); err != nil {
		return err
	}

	bundle, err := stardi18n.LoadEmbeddedBundle()
	if err != nil {
		return err
	}
	stardi18n.SetBundle(bundle)
	return nil
}

func (c *i18nComponent) Start(_ context.Context) error { return nil }

func (c *i18nComponent) Stop(_ context.Context) error { return nil }
