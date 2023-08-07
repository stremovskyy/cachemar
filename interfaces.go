package cachemar

import (
	"context"
	"time"
)

// Cacher is an interface that defines all operations a cache service should support.
type Cacher interface {
	// Set stores a key-value pair in the cache with the specified ttl (time-to-live) duration and tags.
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration, tags []string) error

	// Get retrieves a value based on its key from the cache, and unmarshals it into the provided variable.
	Get(ctx context.Context, key string, value interface{}) error

	// Remove deletes a key-value pair from the cache using the key.
	Remove(ctx context.Context, key string) error

	// RemoveByTag deletes all key-value pairs associated with the given tag from the cache.
	RemoveByTag(ctx context.Context, tag string) error

	// RemoveByTags deletes all key-value pairs associated with the given set of tags from the cache.
	RemoveByTags(ctx context.Context, tags []string) error

	// Exists checks if a key exists in the cache.
	Exists(ctx context.Context, key string) (bool, error)

	// Increment increments the integer value of a key in the cache by one.
	Increment(ctx context.Context, key string) error

	// Decrement decrements the integer value of a key in the cache by one.
	Decrement(ctx context.Context, key string) error

	// GetKeysByTag retrieves all keys associated with a given tag.
	GetKeysByTag(ctx context.Context, tag string) ([]string, error)
}

// Service is an interface that defines all operations a cache service manager should support.
type Service interface {
	// Register adds a cache service to the service manager and assigns it a name.
	Register(name string, service Cacher)

	// Use retrieves a registered cache service by its name.
	Use(name string) Cacher

	// Current retrieves the current cache service being used by the service manager.
	Current() Cacher

	// SetCurrent sets the current cache service the service manager should use.
	SetCurrent(name string)

	// Cacher is embedded to allow the service manager to act as a Cacher itself, proxying calls to the current cache service.
	Cacher
}
