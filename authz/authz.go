package authz

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// Authorizer checks whether a subject can perform an action on an object.
type Authorizer interface {
	Enforce(ctx context.Context, sub, obj, act string) (bool, error)
}

var (
	authorizerMu sync.RWMutex
	authorizer   Authorizer
)

// Init initializes the package global authorizer from JSON config bytes.
func Init(config []byte) error {
	if len(config) == 0 {
		return fmt.Errorf("%w: config is empty", ErrInvalidConfig)
	}

	var cfg Config
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("%w: parse config: %v", ErrInvalidConfig, err)
	}
	return InitConfig(cfg)
}

// InitConfig initializes the package global authorizer from Config.
func InitConfig(cfg Config) error {
	a, err := NewCasbinAuthorizer(cfg)
	if err != nil {
		return err
	}
	Set(a)
	return nil
}

// Set replaces the package global authorizer. Passing nil clears it.
func Set(a Authorizer) {
	authorizerMu.Lock()
	defer authorizerMu.Unlock()
	authorizer = a
}

// Get returns the package global authorizer.
func Get() (Authorizer, error) {
	authorizerMu.RLock()
	defer authorizerMu.RUnlock()
	if authorizer == nil {
		return nil, ErrNotInitialized
	}
	return authorizer, nil
}

// Enforce checks a request with the package global authorizer.
func Enforce(ctx context.Context, sub, obj, act string) (bool, error) {
	a, err := Get()
	if err != nil {
		return false, err
	}
	return a.Enforce(ctx, sub, obj, act)
}
