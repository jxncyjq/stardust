package redis

import (
	"sync"
	"testing"
)

func TestGetRedisDbSharesFirstManagerClient(t *testing.T) {
	resetRedisTestState(t)
	managerConfig = []*Config{
		{Name: "default", Addrs: []string{"127.0.0.1:6379"}},
		{Name: "secondary", Addrs: []string{"127.0.0.1:6380"}},
	}

	defaultCmd := GetRedisDb()
	if defaultCmd == nil {
		t.Fatal("GetRedisDb() = nil, want initialized redis client")
	}

	m, err := GetRedisManager()
	if err != nil {
		t.Fatal("GetRedisManager() error:", err)
	}
	if m == nil {
		t.Fatal("GetRedisManager() = nil, want manager instance")
	}

	gotDefault := m.GetRedisCmd("default")
	if gotDefault == nil {
		t.Fatal("GetRedisManager().GetRedisCmd(default) = nil, want initialized default client")
	}
	if gotDefault != defaultCmd {
		t.Errorf("GetRedisManager().GetRedisCmd(default) = %v, want same client as GetRedisDb() = %v", gotDefault, defaultCmd)
	}

	if gotSecondary := m.GetRedisCmd("secondary"); gotSecondary == nil {
		t.Error("GetRedisManager().GetRedisCmd(secondary) = nil, want initialized secondary client")
	}
}

func TestGetRedisDbBeforeInitCanInitializeLater(t *testing.T) {
	resetRedisTestState(t)

	if got := GetRedisDb(); got != nil {
		t.Errorf("GetRedisDb() before Init = %v, want nil", got)
	}

	managerConfig = []*Config{{Name: "default", Addrs: []string{"127.0.0.1:6379"}}}

	got := GetRedisDb()
	if got == nil {
		t.Fatal("GetRedisDb() after Init = nil, want initialized redis client")
	}

	m, err := GetRedisManager()
	if err != nil {
		t.Fatal("GetRedisManager() error:", err)
	}
	if m == nil {
		t.Fatal("GetRedisManager() after Init = nil, want manager instance")
	}
	if managerCmd := m.GetRedisCmd("default"); managerCmd != got {
		t.Errorf("GetRedisManager().GetRedisCmd(default) = %v, want same client as GetRedisDb() = %v", managerCmd, got)
	}
}

func TestGetRedisManagerBeforeInitReturnsError(t *testing.T) {
	resetRedisTestState(t)

	_, err := GetRedisManager()
	if err == nil {
		t.Fatal("GetRedisManager() before Init = nil error, want error")
	}

	managerConfig = []*Config{{Name: "default", Addrs: []string{"127.0.0.1:6379"}}}
	managerOnce = sync.Once{}

	m, err := GetRedisManager()
	if err != nil {
		t.Fatal("GetRedisManager() after Init error:", err)
	}
	if m == nil {
		t.Fatal("GetRedisManager() after Init = nil, want initialized manager")
	}
	if got := m.GetRedisCmd("default"); got == nil {
		t.Error("GetRedisManager().GetRedisCmd(default) after Init = nil, want initialized redis client")
	}
}

func resetRedisTestState(t *testing.T) {
	t.Helper()
	redisCon = nil
	manager = nil
	managerConfig = nil
	managerOnce = sync.Once{}
	managerErr = nil

	t.Cleanup(func() {
		closeRedisTestState()
		redisCon = nil
		manager = nil
		managerConfig = nil
		managerOnce = sync.Once{}
		managerErr = nil
	})
}

func closeRedisTestState() {
	seen := make(map[RedisCmd]struct{})
	if redisCon != nil {
		seen[redisCon] = struct{}{}
		closeRedisCmd(redisCon)
	}
	if manager == nil {
		return
	}
	for _, cmd := range manager.redisCmds {
		if _, ok := seen[cmd]; ok {
			continue
		}
		seen[cmd] = struct{}{}
		closeRedisCmd(cmd)
	}
}

func closeRedisCmd(cmd RedisCmd) {
	if cmd == nil {
		return
	}
	if closer, ok := cmd.(interface{ Close() error }); ok {
		_ = closer.Close()
	}
}
