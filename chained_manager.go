package cachemar

import (
	"context"
	"fmt"
	"time"
)

type chained struct {
	m        *manager
	chain    []string
	fallback string
}

func newChained(m *manager) ChainedManager {
	return &chained{
		m:     m,
		chain: make([]string, 0),
	}
}

func (c *chained) SetFallback(name string) {
	c.fallback = name
}

func (c *chained) AddToChain(name string) {
	c.chain = append(c.chain, name)
}

func (c *chained) RemoveFromChain(name string) {
	for i, managerName := range c.chain {
		if managerName == name {
			c.chain = append(c.chain[:i], c.chain[i+1:]...)
			break
		}
	}
}

// Implementing the Manager interface methods

func (c *chained) Register(name string, manager Cacher) {
	c.m.Register(name, manager)
}

func (c *chained) Use(name string) Cacher {
	return c.m.Use(name)
}

func (c *chained) Current() Cacher {
	return c.m.Current()
}

func (c *chained) SetCurrent(name string) {
	c.m.SetCurrent(name)
}

func (c *chained) Ping() error {
	return c.m.Ping()
}

func (c *chained) Close() error {
	return c.m.Close()
}

func (c *chained) Chain() ChainedManager {
	return c
}

// Implementing the Cacher interface methods with chaining logic

func (c *chained) Set(ctx context.Context, key string, value interface{}, ttl time.Duration, tags []string) error {
	var errors []error
	for _, managerName := range c.chain {
		manager := c.m.managers[managerName]
		err := manager.Set(ctx, key, value, ttl, tags)
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("errors occurred while setting value in chain: %v", errors)
	}
	return nil
}

func (c *chained) Get(ctx context.Context, key string, value interface{}) error {
	for _, managerName := range c.chain {
		manager := c.m.managers[managerName]
		err := manager.Get(ctx, key, value)
		if err == nil {
			return nil
		}
	}
	if c.fallback != "" {
		return c.m.managers[c.fallback].Get(ctx, key, value)
	}
	return fmt.Errorf("value not found in any cache manager")
}

// ... [Previous code]

func (c *chained) Remove(ctx context.Context, key string) error {
	var errors []error
	for _, managerName := range c.chain {
		manager := c.m.managers[managerName]
		err := manager.Remove(ctx, key)
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("errors occurred while removing key in chain: %v", errors)
	}
	return nil
}

func (c *chained) RemoveByTag(ctx context.Context, tag string) error {
	var errors []error
	for _, managerName := range c.chain {
		manager := c.m.managers[managerName]
		err := manager.RemoveByTag(ctx, tag)
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("errors occurred while removing by tag in chain: %v", errors)
	}
	return nil
}

func (c *chained) RemoveByTags(ctx context.Context, tags []string) error {
	var errors []error
	for _, managerName := range c.chain {
		manager := c.m.managers[managerName]
		err := manager.RemoveByTags(ctx, tags)
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("errors occurred while removing by tags in chain: %v", errors)
	}
	return nil
}

func (c *chained) Exists(ctx context.Context, key string) (bool, error) {
	for _, managerName := range c.chain {
		manager := c.m.managers[managerName]
		exists, err := manager.Exists(ctx, key)
		if err == nil && exists {
			return true, nil
		}
	}
	if c.fallback != "" {
		return c.m.managers[c.fallback].Exists(ctx, key)
	}
	return false, fmt.Errorf("key not found in any cache manager")
}

func (c *chained) Increment(ctx context.Context, key string) error {
	var errors []error
	for _, managerName := range c.chain {
		manager := c.m.managers[managerName]
		err := manager.Increment(ctx, key)
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("errors occurred while incrementing key in chain: %v", errors)
	}
	return nil
}

func (c *chained) Decrement(ctx context.Context, key string) error {
	var errors []error
	for _, managerName := range c.chain {
		manager := c.m.managers[managerName]
		err := manager.Decrement(ctx, key)
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("errors occurred while decrementing key in chain: %v", errors)
	}
	return nil
}

func (c *chained) GetKeysByTag(ctx context.Context, tag string) ([]string, error) {
	var allKeys []string
	for _, managerName := range c.chain {
		manager := c.m.managers[managerName]
		keys, err := manager.GetKeysByTag(ctx, tag)
		if err == nil {
			allKeys = append(allKeys, keys...)
		}
	}
	if len(allKeys) == 0 && c.fallback != "" {
		return c.m.managers[c.fallback].GetKeysByTag(ctx, tag)
	}
	return allKeys, nil
}

// Override method to create a new chain with the given names and use it as the current call
func (c *chained) Override(names ...string) ChainedManager {
	newChain := &chained{
		m:        c.m,
		chain:    names,
		fallback: c.fallback,
	}

	return newChain
}
