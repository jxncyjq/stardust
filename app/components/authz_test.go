package components

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jxncyjq/stardust/authz"
)

func TestAuthzComponentContract(t *testing.T) {
	component := AuthzComponent()

	if component.Name() != "authz" {
		t.Fatalf("Name() = %q, want authz", component.Name())
	}
	deps := component.Dependencies()
	if len(deps) != 1 || deps[0] != "logs" {
		t.Fatalf("Dependencies() = %#v, want [logs]", deps)
	}
}

func TestAuthzComponentInit(t *testing.T) {
	t.Cleanup(func() { authz.Set(nil) })

	component := AuthzComponent()
	configBytes := mustAuthzConfigBytes(t)
	cfgFn := func(key string) []byte {
		if key != "authz" {
			return nil
		}
		return configBytes
	}

	if err := component.Init(context.Background(), cfgFn); err != nil {
		t.Fatalf("Init() error = %v, want nil", err)
	}

	a, err := authz.Get()
	if err != nil {
		t.Fatalf("authz.Get() error = %v, want nil", err)
	}

	allowed, err := a.Enforce(context.Background(), "roleName", "/orders", "GET")
	if err != nil {
		t.Fatalf("Enforce() error = %v, want nil", err)
	}
	if !allowed {
		t.Fatal("Enforce() = false, want true")
	}
}

func TestAuthzComponentInitMissingConfig(t *testing.T) {
	component := AuthzComponent()

	err := component.Init(context.Background(), func(string) []byte { return nil })
	if err == nil {
		t.Fatal("Init() error = nil, want non-nil")
	}
}

func mustAuthzConfigBytes(t *testing.T) []byte {
	t.Helper()

	dir := t.TempDir()
	modelPath := filepath.Join(dir, "model.conf")
	policyPath := filepath.Join(dir, "policy.csv")

	model := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
`
	policy := "p, roleName, /orders, GET\n"

	if err := os.WriteFile(modelPath, []byte(model), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v, want nil", modelPath, err)
	}
	if err := os.WriteFile(policyPath, []byte(policy), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v, want nil", policyPath, err)
	}

	cfg := authz.Config{
		ModelPath:  modelPath,
		PolicyPath: policyPath,
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("json.Marshal(%+v) error = %v, want nil", cfg, err)
	}
	return data
}
