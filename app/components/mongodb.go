package components

import (
	"context"

	"github.com/jxncyjq/stardust/app"
	"github.com/jxncyjq/stardust/mongodb"
)

type mongodbComponent struct{}

// MongoDBComponent 返回 MongoDB 组件，依赖 logs。
func MongoDBComponent() app.Component { return &mongodbComponent{} }

func (c *mongodbComponent) Name() string          { return "mongodb" }
func (c *mongodbComponent) Dependencies() []string { return []string{"logs"} }

func (c *mongodbComponent) Init(_ context.Context, configFn app.ConfigFunc) (retErr error) {
	defer recoverToError(&retErr, "mongodb")
	if err := mongodb.Init(requireConfig(configFn, "mongodb")); err != nil {
		return err
	}
	_, err := mongodb.GetMongoManager() // 触发懒初始化，提前暴露连接错误
	return err
}

func (c *mongodbComponent) Start(_ context.Context) error { return nil }
func (c *mongodbComponent) Stop(_ context.Context) error  { return nil }
