package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockComponent 用于测试的简单组件实现。
type mockComponent struct {
	name string
	deps []string
}

func (m *mockComponent) Name() string          { return m.name }
func (m *mockComponent) Dependencies() []string { return m.deps }
func (m *mockComponent) Init(_ context.Context, _ ConfigFunc) error { return nil }
func (m *mockComponent) Start(_ context.Context) error              { return nil }
func (m *mockComponent) Stop(_ context.Context) error               { return nil }

func mc(name string, deps ...string) Component {
	return &mockComponent{name: name, deps: deps}
}

func TestTopoSort_NoDeps(t *testing.T) {
	result, err := topoSort([]Component{mc("a"), mc("b"), mc("c")})
	require.NoError(t, err)
	assert.Len(t, result, 3)
}

func TestTopoSort_LinearChain(t *testing.T) {
	// logs -> redis -> app
	comps := []Component{mc("app", "redis"), mc("redis", "logs"), mc("logs")}
	result, err := topoSort(comps)
	require.NoError(t, err)
	require.Len(t, result, 3)

	indexOf := func(name string) int {
		for i, c := range result {
			if c.Name() == name {
				return i
			}
		}
		return -1
	}
	assert.Less(t, indexOf("logs"), indexOf("redis"))
	assert.Less(t, indexOf("redis"), indexOf("app"))
}

func TestTopoSort_DiamondDeps(t *testing.T) {
	// logs -> redis, databases -> app
	comps := []Component{
		mc("app", "redis", "databases"),
		mc("redis", "logs"),
		mc("databases", "logs"),
		mc("logs"),
	}
	result, err := topoSort(comps)
	require.NoError(t, err)
	require.Len(t, result, 4)

	indexOf := func(name string) int {
		for i, c := range result {
			if c.Name() == name {
				return i
			}
		}
		return -1
	}
	assert.Less(t, indexOf("logs"), indexOf("redis"))
	assert.Less(t, indexOf("logs"), indexOf("databases"))
	assert.Less(t, indexOf("redis"), indexOf("app"))
	assert.Less(t, indexOf("databases"), indexOf("app"))
}

func TestTopoSort_CircularDep(t *testing.T) {
	comps := []Component{mc("a", "b"), mc("b", "a")}
	_, err := topoSort(comps)
	assert.ErrorContains(t, err, "circular dependency")
}

func TestTopoSort_MissingDep(t *testing.T) {
	comps := []Component{mc("redis", "logs")} // logs 未注册
	_, err := topoSort(comps)
	assert.ErrorContains(t, err, "not registered")
}

func TestTopoSort_DuplicateName(t *testing.T) {
	comps := []Component{mc("logs"), mc("logs")}
	_, err := topoSort(comps)
	assert.ErrorContains(t, err, "duplicate component name")
}

func TestContainer_InitStartStopOrder(t *testing.T) {
	var order []string

	newOrderedComp := func(name string, deps ...string) Component {
		return &orderedComponent{name: name, deps: deps, order: &order}
	}

	c := &Container{}
	c.Register(
		newOrderedComp("redis", "logs"),
		newOrderedComp("logs"),
	)

	ctx := context.Background()
	cfgFn := func(string) []byte { return []byte("{}") }

	require.NoError(t, c.Init(ctx, cfgFn))
	require.NoError(t, c.Start(ctx))
	require.NoError(t, c.Stop(ctx))

	assert.Equal(t, []string{
		"init:logs", "init:redis",
		"start:logs", "start:redis",
		"stop:redis", "stop:logs", // 逆序
	}, order)
}

type orderedComponent struct {
	name  string
	deps  []string
	order *[]string
}

func (c *orderedComponent) Name() string          { return c.name }
func (c *orderedComponent) Dependencies() []string { return c.deps }
func (c *orderedComponent) Init(_ context.Context, _ ConfigFunc) error {
	*c.order = append(*c.order, "init:"+c.name)
	return nil
}
func (c *orderedComponent) Start(_ context.Context) error {
	*c.order = append(*c.order, "start:"+c.name)
	return nil
}
func (c *orderedComponent) Stop(_ context.Context) error {
	*c.order = append(*c.order, "stop:"+c.name)
	return nil
}
