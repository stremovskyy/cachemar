package cachemar

import (
	"context"
	"time"
)

// cacheMarService is an implementation of the Service interface.
type cacheMarService struct {
	services map[string]Cacher // A map to store registered cache services with their names as keys.
	current  string            // The name of the current cache service being used.
}

// Register adds a cache service to the service manager and assigns it a name.
func (c *cacheMarService) Register(name string, service Cacher) {
	c.services[name] = service
	c.current = name
}

// Use retrieves a registered cache service by its name. Returns nil if the service is not found.
func (c *cacheMarService) Use(name string) Cacher {
	service, ok := c.services[name]
	if !ok {
		return nil
	}

	return service
}

// Current retrieves the current cache service being used by the service manager.
func (c *cacheMarService) Current() Cacher {
	return c.services[c.current]
}

// SetCurrent sets the current cache service the service manager should use.
func (c *cacheMarService) SetCurrent(name string) {
	c.current = name
}

// Set forwards the "Set" operation to the current cache service.
func (c *cacheMarService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration, tags []string) error {
	return c.Current().Set(ctx, key, value, ttl, tags)
}

// Get forwards the "Get" operation to the current cache service.
func (c *cacheMarService) Get(ctx context.Context, key string, value interface{}) error {
	return c.Current().Get(ctx, key, value)
}

// Remove forwards the "Remove" operation to the current cache service.
func (c *cacheMarService) Remove(ctx context.Context, key string) error {
	return c.Current().Remove(ctx, key)
}

// RemoveByTag forwards the "RemoveByTag" operation to the current cache service.
func (c *cacheMarService) RemoveByTag(ctx context.Context, tag string) error {
	return c.Current().RemoveByTag(ctx, tag)
}

// RemoveByTags forwards the "RemoveByTags" operation to the current cache service.
func (c *cacheMarService) RemoveByTags(ctx context.Context, tags []string) error {
	return c.Current().RemoveByTags(ctx, tags)
}

// Exists forwards the "Exists" operation to the current cache service.
func (c *cacheMarService) Exists(ctx context.Context, key string) (bool, error) {
	return c.Current().Exists(ctx, key)
}

// Increment forwards the "Increment" operation to the current cache service.
func (c *cacheMarService) Increment(ctx context.Context, key string) error {
	return c.Current().Increment(ctx, key)
}

// Decrement forwards the "Decrement" operation to the current cache service.
func (c *cacheMarService) Decrement(ctx context.Context, key string) error {
	return c.Current().Decrement(ctx, key)
}

// GetKeysByTag forwards the "GetKeysByTag" operation to the current cache service.
func (c *cacheMarService) GetKeysByTag(ctx context.Context, tag string) ([]string, error) {
	return c.Current().GetKeysByTag(ctx, tag)
}

// New creates and returns a new instance of the cacheMarService.
func New() Service {
	return &cacheMarService{
		services: make(map[string]Cacher),
	}
}
