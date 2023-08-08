package cachemar

import (
	"context"
	"fmt"
	"time"
)

// manager is an implementation of the Manager interface.
type manager struct {
	managers      map[string]Cacher // A map to store registered cache managers with their names as keys.
	current       string            // The name of the current cache manager being used.
	chainInstance ChainedManager    // The chained manager instance.
}

// New creates and returns a new instance of the manager.
func New() Manager {
	return &manager{
		managers: make(map[string]Cacher),
	}
}

// Register adds a cache manager to the manager  and assigns it a name.
func (c *manager) Register(name string, manager Cacher) {
	c.managers[name] = manager
	c.current = name
}

// Use retrieves a registered cache manager by its name. Returns nil if the manager is not found.
func (c *manager) Use(name string) Cacher {
	manager, ok := c.managers[name]
	if !ok {
		return nil
	}

	return manager
}

// Current retrieves the current cache manager being used by the manager .
func (c *manager) Current() Cacher {
	return c.managers[c.current]
}

// SetCurrent sets the current cache manager the manager  should use.
func (c *manager) SetCurrent(name string) {
	c.current = name
}

// Set forwards the "Set" operation to the current cache manager.
func (c *manager) Set(ctx context.Context, key string, value interface{}, ttl time.Duration, tags []string) error {
	return c.Current().Set(ctx, key, value, ttl, tags)
}

// Get forwards the "Get" operation to the current cache manager.
func (c *manager) Get(ctx context.Context, key string, value interface{}) error {
	return c.Current().Get(ctx, key, value)
}

// Remove forwards the "Remove" operation to the current cache manager.
func (c *manager) Remove(ctx context.Context, key string) error {
	return c.Current().Remove(ctx, key)
}

// RemoveByTag forwards the "RemoveByTag" operation to the current cache manager.
func (c *manager) RemoveByTag(ctx context.Context, tag string) error {
	return c.Current().RemoveByTag(ctx, tag)
}

// RemoveByTags forwards the "RemoveByTags" operation to the current cache manager.
func (c *manager) RemoveByTags(ctx context.Context, tags []string) error {
	return c.Current().RemoveByTags(ctx, tags)
}

// Exists forwards the "Exists" operation to the current cache manager.
func (c *manager) Exists(ctx context.Context, key string) (bool, error) {
	return c.Current().Exists(ctx, key)
}

// Increment forwards the "Increment" operation to the current cache manager.
func (c *manager) Increment(ctx context.Context, key string) error {
	return c.Current().Increment(ctx, key)
}

// Decrement forwards the "Decrement" operation to the current cache manager.
func (c *manager) Decrement(ctx context.Context, key string) error {
	return c.Current().Decrement(ctx, key)
}

// GetKeysByTag forwards the "GetKeysByTag" operation to the current cache manager.
func (c *manager) GetKeysByTag(ctx context.Context, tag string) ([]string, error) {
	return c.Current().GetKeysByTag(ctx, tag)
}

// Ping forwards the "Ping" operation to the current cache manager.
func (c *manager) Ping() error {
	errors := make([]error, 0)

	for _, manager := range c.managers {
		err := manager.Ping()
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors: %v", errors)
	}

	return nil
}

// Close forwards the "Close" operation to the current cache manager.
func (d *manager) Close() error {
	errors := make([]error, 0)

	for _, manager := range d.managers {
		err := manager.Close()
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors: %v", errors)
	}

	return nil
}

// Chain returns a ChainedManager instance.
func (c *manager) Chain() ChainedManager {
	if c.chainInstance == nil {
		c.chainInstance = newChained(c)
	}
	return c.chainInstance
}
