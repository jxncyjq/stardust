package authz

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v3"
)

// CasbinAuthorizer authorizes requests with a Casbin enforcer.
type CasbinAuthorizer struct {
	enforcer *casbin.Enforcer
	adapter  interface{}
}

var _ Authorizer = (*CasbinAuthorizer)(nil)

// NewCasbinAuthorizer creates a Casbin authorizer using the file adapter.
func NewCasbinAuthorizer(cfg Config) (*CasbinAuthorizer, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	adapter, err := newPolicyAdapter(cfg)
	if err != nil {
		return nil, err
	}

	enforcer, err := newEnforcer(cfg.ModelPath, adapter, cfg.PolicyPath)
	if err != nil {
		return nil, fmt.Errorf("authz: create casbin enforcer: %w", err)
	}
	return &CasbinAuthorizer{enforcer: enforcer, adapter: adapter}, nil
}

// Enforce checks whether sub can perform act on obj.
func (a *CasbinAuthorizer) Enforce(ctx context.Context, sub, obj, act string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return a.enforcer.Enforce(sub, obj, act)
}

// Close releases resources held by the policy adapter, such as an underlying
// database connection opened by the gorm adapter. It is safe to call when the
// adapter holds no closable resources (e.g. the default file adapter).
func (a *CasbinAuthorizer) Close() error {
	if closer, ok := a.adapter.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}
