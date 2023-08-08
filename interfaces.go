package cachemar

import (
	"context"
	"time"
)

// Cacher is an interface that defines all operations a cache manager should support.
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
	// Ping checks if the cache manager is up and running.
	Ping() error
	// Close closes the cache manager.
	Close() error
}

// Manager is an interface that defines all operations a cache  manager should support.
type Manager interface {
	// Register adds a cache manager to the  manager and assigns it a name.
	Register(name string, manager Cacher)

	// Use retrieves a registered cache manager by its name.
	Use(name string) Cacher

	// Current retrieves the current cache manager being used by the  manager.
	Current() Cacher

	// SetCurrent sets the current cache manager the  manager should use.
	SetCurrent(name string)

	// Ping checks ALL cache managers are up and running.
	Ping() error

	// Close closes ALL cache managers.
	Close() error

	// Chain creates a new ChainedManager that can be used to chain multiple cache managers together.
	Chain() ChainedManager

	// Cacher is embedded to allow the manager  to act as a Cacher itself, proxying calls to the current cache manager.
	Cacher
}

// ChainedManager is a cache manager that allows multiple cache managers to be chained together.
type ChainedManager interface {
	Manager

	SetFallback(name string)
	AddToChain(name string)
	RemoveFromChain(name string)
	Override(names ...string) ChainedManager
}
