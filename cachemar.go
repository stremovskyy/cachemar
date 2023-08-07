package cachemar

import (
	"context"
	"time"
)

type cacheMarService struct {
	services map[string]Cacher
	current  string
}

func (c *cacheMarService) Register(name string, service Cacher) {
	c.services[name] = service
	c.current = name
}

func (c *cacheMarService) Use(name string) Cacher {
	service, ok := c.services[name]
	if !ok {
		return nil
	}

	return service
}

func (c *cacheMarService) Current() Cacher {
	return c.services[c.current]
}

func (c *cacheMarService) SetCurrent(name string) {
	c.current = name
}

func (c *cacheMarService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration, tags []string) error {
	return c.Current().Set(ctx, key, value, ttl, tags)
}

func (c *cacheMarService) Get(ctx context.Context, key string, value interface{}) error {
	return c.Current().Get(ctx, key, value)
}

func (c *cacheMarService) Remove(ctx context.Context, key string) error {
	return c.Current().Remove(ctx, key)
}

func (c *cacheMarService) RemoveByTag(ctx context.Context, tag string) error {
	return c.Current().RemoveByTag(ctx, tag)
}

func (c *cacheMarService) RemoveByTags(ctx context.Context, tags []string) error {
	return c.Current().RemoveByTags(ctx, tags)
}

func (c *cacheMarService) Exists(ctx context.Context, key string) (bool, error) {
	return c.Current().Exists(ctx, key)
}

func (c *cacheMarService) Increment(ctx context.Context, key string) error {
	return c.Current().Increment(ctx, key)
}

func (c *cacheMarService) Decrement(ctx context.Context, key string) error {
	return c.Current().Decrement(ctx, key)
}

func (c *cacheMarService) GetKeysByTag(ctx context.Context, tag string) ([]string, error) {
	return c.Current().GetKeysByTag(ctx, tag)
}

func New() Service {
	return &cacheMarService{
		services: make(map[string]Cacher),
	}
}
