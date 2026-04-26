package app

import (
	"context"
	"fmt"
)

// Container 管理组件集合，提供拓扑排序后的统一生命周期。
type Container struct {
	registered []Component
	sorted     []Component
}

// Register 注册一组组件，返回自身以支持链式调用。
func (c *Container) Register(components ...Component) *Container {
	c.registered = append(c.registered, components...)
	return c
}

// Init 按拓扑顺序初始化所有已注册组件。
func (c *Container) Init(ctx context.Context, configFn ConfigFunc) error {
	sorted, err := topoSort(c.registered)
	if err != nil {
		return err
	}
	c.sorted = sorted
	for _, comp := range c.sorted {
		if err := safeInit(ctx, comp, configFn); err != nil {
			return fmt.Errorf("init %s: %w", comp.Name(), err)
		}
	}
	return nil
}

// Start 按拓扑顺序启动所有组件（非阻塞）。
func (c *Container) Start(ctx context.Context) error {
	for _, comp := range c.sorted {
		if err := comp.Start(ctx); err != nil {
			return fmt.Errorf("start %s: %w", comp.Name(), err)
		}
	}
	return nil
}

// Stop 按逆拓扑顺序关闭所有组件，收集全部错误后一并返回。
func (c *Container) Stop(ctx context.Context) error {
	errs := make([]error, 0)
	for i := len(c.sorted) - 1; i >= 0; i-- {
		comp := c.sorted[i]
		if err := comp.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("stop %s: %w", comp.Name(), err))
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("stop errors: %v", errs)
}

// topoSort 使用 Kahn 算法对组件进行拓扑排序。
// 检测重复名称、缺失依赖、循环依赖。
func topoSort(components []Component) ([]Component, error) {
	byName := make(map[string]Component, len(components))
	for _, c := range components {
		if _, dup := byName[c.Name()]; dup {
			return nil, fmt.Errorf("duplicate component name %q", c.Name())
		}
		byName[c.Name()] = c
	}

	for _, c := range components {
		for _, dep := range c.Dependencies() {
			if _, ok := byName[dep]; !ok {
				return nil, fmt.Errorf("component %q: dependency %q not registered", c.Name(), dep)
			}
		}
	}

	inDegree := make(map[string]int, len(components))
	revEdges := make(map[string][]string) // dep -> list of dependents
	for _, c := range components {
		if _, exists := inDegree[c.Name()]; !exists {
			inDegree[c.Name()] = 0
		}
		for _, dep := range c.Dependencies() {
			inDegree[c.Name()]++
			revEdges[dep] = append(revEdges[dep], c.Name())
		}
	}

	queue := make([]string, 0, len(components))
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}

	result := make([]Component, 0, len(components))
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		result = append(result, byName[name])
		for _, dependent := range revEdges[name] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	if len(result) != len(components) {
		return nil, fmt.Errorf("circular dependency detected among components")
	}
	return result, nil
}

// safeInit 将 Init 过程中的 panic 转换为 error。
func safeInit(ctx context.Context, comp Component, fn ConfigFunc) (retErr error) {
	defer func() {
		if r := recover(); r != nil {
			retErr = fmt.Errorf("panic: %v", r)
		}
	}()
	return comp.Init(ctx, fn)
}
