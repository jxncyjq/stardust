package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type mongoClient struct {
	client   *mongo.Client
	database string
	timeout  time.Duration
}

func (c *mongoClient) col(name string) *mongo.Collection {
	return c.client.Database(c.database).Collection(name)
}

func (c *mongoClient) InsertOne(ctx context.Context, collection string, doc interface{}) (string, error) {
	res, err := c.col(collection).InsertOne(ctx, doc)
	if err != nil {
		return "", err
	}
	return bsonID(res.InsertedID), nil
}

func (c *mongoClient) InsertMany(ctx context.Context, collection string, docs []interface{}) ([]string, error) {
	res, err := c.col(collection).InsertMany(ctx, docs)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(res.InsertedIDs))
	for _, id := range res.InsertedIDs {
		ids = append(ids, bsonID(id))
	}
	return ids, nil
}

func (c *mongoClient) FindOne(ctx context.Context, collection string, filter interface{}, result interface{}) error {
	err := c.col(collection).FindOne(ctx, filter).Decode(result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return ErrNotFound
	}
	return err
}

func (c *mongoClient) FindMany(ctx context.Context, collection string, filter interface{}, results interface{}, opts ...options.Lister[options.FindOptions]) error {
	cursor, err := c.col(collection).Find(ctx, filter, opts...)
	if err != nil {
		return err
	}
	return cursor.All(ctx, results)
}

func (c *mongoClient) UpdateOne(ctx context.Context, collection string, filter, update interface{}) (int64, error) {
	res, err := c.col(collection).UpdateOne(ctx, filter, update)
	if err != nil {
		return 0, err
	}
	return res.ModifiedCount, nil
}

func (c *mongoClient) UpdateMany(ctx context.Context, collection string, filter, update interface{}) (int64, error) {
	res, err := c.col(collection).UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, err
	}
	return res.ModifiedCount, nil
}

func (c *mongoClient) DeleteOne(ctx context.Context, collection string, filter interface{}) (int64, error) {
	res, err := c.col(collection).DeleteOne(ctx, filter)
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}

func (c *mongoClient) DeleteMany(ctx context.Context, collection string, filter interface{}) (int64, error) {
	res, err := c.col(collection).DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}

func (c *mongoClient) CountDocuments(ctx context.Context, collection string, filter interface{}) (int64, error) {
	return c.col(collection).CountDocuments(ctx, filter)
}

func (c *mongoClient) Collection(name string) *mongo.Collection {
	return c.col(name)
}

func (c *mongoClient) Ping(ctx context.Context) error {
	return c.client.Ping(ctx, readpref.Primary())
}

// compile-time interface check
var _ MongoCli = (*mongoClient)(nil)

// ObjectID 从十六进制字符串解析为 bson.ObjectID，方便业务层使用。
func ObjectID(hex string) (bson.ObjectID, error) {
	return bson.ObjectIDFromHex(hex)
}
