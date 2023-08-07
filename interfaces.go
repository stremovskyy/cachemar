package cachemar

import (
	"context"
	"time"
)

type Cacher interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration, tags []string) error
	Get(ctx context.Context, key string, value interface{}) error
	Remove(ctx context.Context, key string) error
	RemoveByTag(ctx context.Context, tag string) error
	RemoveByTags(ctx context.Context, tags []string) error
	Exists(ctx context.Context, key string) (bool, error)
	Increment(ctx context.Context, key string) error
	Decrement(ctx context.Context, key string) error
	GetKeysByTag(ctx context.Context, tag string) ([]string, error)
}

type Service interface {
	Register(name string, service Cacher)
	Use(name string) Cacher
	Current() Cacher
	SetCurrent(name string)

	Cacher
}
