package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// MongoCli 定义 MongoDB 的操作接口，业务代码依赖此接口而非具体实现，便于 Mock。
type MongoCli interface {
	// InsertOne 插入单个文档，返回插入文档的 _id 字符串。
	InsertOne(ctx context.Context, collection string, doc interface{}) (string, error)

	// InsertMany 批量插入文档，返回各文档的 _id 字符串列表。
	InsertMany(ctx context.Context, collection string, docs []interface{}) ([]string, error)

	// FindOne 查询单个文档，结果反序列化到 result（需传指针）。未找到时返回 ErrNotFound。
	FindOne(ctx context.Context, collection string, filter interface{}, result interface{}) error

	// FindMany 查询多个文档，结果反序列化到 results（需传切片指针）。
	FindMany(ctx context.Context, collection string, filter interface{}, results interface{}, opts ...options.Lister[options.FindOptions]) error

	// UpdateOne 更新第一个匹配文档，返回实际修改的文档数。
	UpdateOne(ctx context.Context, collection string, filter, update interface{}) (int64, error)

	// UpdateMany 更新所有匹配文档，返回实际修改的文档数。
	UpdateMany(ctx context.Context, collection string, filter, update interface{}) (int64, error)

	// DeleteOne 删除第一个匹配文档，返回实际删除的文档数。
	DeleteOne(ctx context.Context, collection string, filter interface{}) (int64, error)

	// DeleteMany 删除所有匹配文档，返回实际删除的文档数。
	DeleteMany(ctx context.Context, collection string, filter interface{}) (int64, error)

	// CountDocuments 统计匹配文档数量。
	CountDocuments(ctx context.Context, collection string, filter interface{}) (int64, error)

	// Collection 返回原生 *mongo.Collection，用于 Aggregate 等复杂场景。
	Collection(name string) *mongo.Collection

	// Ping 检查连接健康状态。
	Ping(ctx context.Context) error
}

// bsonID 将 InsertOneResult 中的 _id 转为字符串。
func bsonID(id interface{}) string {
	switch v := id.(type) {
	case bson.ObjectID:
		return v.Hex()
	default:
		return ""
	}
}
