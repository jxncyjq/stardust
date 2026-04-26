package mongodb

import (
	"context"
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// mockMongoCli 用于单元测试的 Mock 实现。
type mockMongoCli struct {
	insertOneFunc      func(ctx context.Context, collection string, doc interface{}) (string, error)
	insertManyFunc     func(ctx context.Context, collection string, docs []interface{}) ([]string, error)
	findOneFunc        func(ctx context.Context, collection string, filter interface{}, result interface{}) error
	findManyFunc       func(ctx context.Context, collection string, filter interface{}, results interface{}) error
	updateOneFunc      func(ctx context.Context, collection string, filter, update interface{}) (int64, error)
	updateManyFunc     func(ctx context.Context, collection string, filter, update interface{}) (int64, error)
	deleteOneFunc      func(ctx context.Context, collection string, filter interface{}) (int64, error)
	deleteManyFunc     func(ctx context.Context, collection string, filter interface{}) (int64, error)
	countDocumentsFunc func(ctx context.Context, collection string, filter interface{}) (int64, error)
	pingFunc           func(ctx context.Context) error
}

func (m *mockMongoCli) InsertOne(ctx context.Context, collection string, doc interface{}) (string, error) {
	return m.insertOneFunc(ctx, collection, doc)
}
func (m *mockMongoCli) InsertMany(ctx context.Context, collection string, docs []interface{}) ([]string, error) {
	return m.insertManyFunc(ctx, collection, docs)
}
func (m *mockMongoCli) FindOne(ctx context.Context, collection string, filter interface{}, result interface{}) error {
	return m.findOneFunc(ctx, collection, filter, result)
}
func (m *mockMongoCli) FindMany(ctx context.Context, collection string, filter interface{}, results interface{}, _ ...options.Lister[options.FindOptions]) error {
	return m.findManyFunc(ctx, collection, filter, results)
}
func (m *mockMongoCli) UpdateOne(ctx context.Context, collection string, filter, update interface{}) (int64, error) {
	return m.updateOneFunc(ctx, collection, filter, update)
}
func (m *mockMongoCli) UpdateMany(ctx context.Context, collection string, filter, update interface{}) (int64, error) {
	return m.updateManyFunc(ctx, collection, filter, update)
}
func (m *mockMongoCli) DeleteOne(ctx context.Context, collection string, filter interface{}) (int64, error) {
	return m.deleteOneFunc(ctx, collection, filter)
}
func (m *mockMongoCli) DeleteMany(ctx context.Context, collection string, filter interface{}) (int64, error) {
	return m.deleteManyFunc(ctx, collection, filter)
}
func (m *mockMongoCli) CountDocuments(ctx context.Context, collection string, filter interface{}) (int64, error) {
	return m.countDocumentsFunc(ctx, collection, filter)
}
func (m *mockMongoCli) Collection(_ string) *mongo.Collection { return nil }
func (m *mockMongoCli) Ping(ctx context.Context) error        { return m.pingFunc(ctx) }

// --- Config 校验测试 ---

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{"valid", Config{Name: "default", URI: "mongodb://localhost:27017", Database: "test"}, false},
		{"missing name", Config{URI: "mongodb://localhost:27017", Database: "test"}, true},
		{"missing uri", Config{Name: "default", Database: "test"}, true},
		{"missing database", Config{Name: "default", URI: "mongodb://localhost:27017"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigSetDefaults(t *testing.T) {
	cfg := Config{}
	cfg.SetDefaults()
	if cfg.MaxPool != 10 {
		t.Errorf("MaxPool default = %d, want 10", cfg.MaxPool)
	}
	if cfg.MinPool != 2 {
		t.Errorf("MinPool default = %d, want 2", cfg.MinPool)
	}
	if cfg.TimeoutS != 5 {
		t.Errorf("TimeoutS default = %d, want 5", cfg.TimeoutS)
	}
}

// --- MongoCli Mock 行为测试 ---

func TestInsertOne_Success(t *testing.T) {
	cli := &mockMongoCli{
		insertOneFunc: func(_ context.Context, _ string, _ interface{}) (string, error) {
			return "507f1f77bcf86cd799439011", nil
		},
	}
	id, err := cli.InsertOne(context.Background(), "users", bson.M{"name": "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "507f1f77bcf86cd799439011" {
		t.Errorf("id = %q, want 507f1f77bcf86cd799439011", id)
	}
}

func TestInsertOne_Error(t *testing.T) {
	cli := &mockMongoCli{
		insertOneFunc: func(_ context.Context, _ string, _ interface{}) (string, error) {
			return "", ErrInsertFailed
		},
	}
	_, err := cli.InsertOne(context.Background(), "users", bson.M{"name": "test"})
	if !errors.Is(err, ErrInsertFailed) {
		t.Errorf("expected ErrInsertFailed, got %v", err)
	}
}

func TestFindOne_NotFound(t *testing.T) {
	cli := &mockMongoCli{
		findOneFunc: func(_ context.Context, _ string, _ interface{}, _ interface{}) error {
			return ErrNotFound
		},
	}
	var result bson.M
	err := cli.FindOne(context.Background(), "users", bson.M{"_id": "nonexistent"}, &result)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestFindOne_Success(t *testing.T) {
	cli := &mockMongoCli{
		findOneFunc: func(_ context.Context, _ string, _ interface{}, result interface{}) error {
			m := result.(*bson.M)
			*m = bson.M{"name": "alice"}
			return nil
		},
	}
	var result bson.M
	if err := cli.FindOne(context.Background(), "users", bson.M{}, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["name"] != "alice" {
		t.Errorf("name = %v, want alice", result["name"])
	}
}

func TestUpdateOne_Success(t *testing.T) {
	cli := &mockMongoCli{
		updateOneFunc: func(_ context.Context, _ string, _, _ interface{}) (int64, error) {
			return 1, nil
		},
	}
	n, err := cli.UpdateOne(context.Background(), "users", bson.M{"_id": "1"}, bson.M{"$set": bson.M{"name": "bob"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Errorf("modified count = %d, want 1", n)
	}
}

func TestDeleteOne_Success(t *testing.T) {
	cli := &mockMongoCli{
		deleteOneFunc: func(_ context.Context, _ string, _ interface{}) (int64, error) {
			return 1, nil
		},
	}
	n, err := cli.DeleteOne(context.Background(), "users", bson.M{"_id": "1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Errorf("deleted count = %d, want 1", n)
	}
}

func TestPing_Success(t *testing.T) {
	cli := &mockMongoCli{
		pingFunc: func(_ context.Context) error { return nil },
	}
	if err := cli.Ping(context.Background()); err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}

func TestPing_Error(t *testing.T) {
	cli := &mockMongoCli{
		pingFunc: func(_ context.Context) error { return errors.New("connection refused") },
	}
	if err := cli.Ping(context.Background()); err == nil {
		t.Error("expected error, got nil")
	}
}

// --- ObjectID 辅助函数测试 ---

func TestObjectID_Valid(t *testing.T) {
	hex := "507f1f77bcf86cd799439011"
	id, err := ObjectID(hex)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.Hex() != hex {
		t.Errorf("hex = %q, want %q", id.Hex(), hex)
	}
}

func TestObjectID_Invalid(t *testing.T) {
	_, err := ObjectID("not-valid-hex")
	if err == nil {
		t.Error("expected error for invalid hex, got nil")
	}
}
