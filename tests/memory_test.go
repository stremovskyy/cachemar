package tests_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stremovskyy/cachemar"
	"github.com/stremovskyy/cachemar/drivers/memory"
)

func TestMemoryCache(t *testing.T) {
	ctx := context.Background()
	cache := memory.New()

	t.Run(
		"Set and Get", func(t *testing.T) {
			key := "test_key"
			value := "test_value"
			if err := cache.Set(ctx, key, value, time.Minute, nil); err != nil {
				t.Fatalf("Set failed: %v", err)
			}

			var retrieved string
			if err := cache.Get(ctx, key, &retrieved); err != nil {
				t.Fatalf("Get failed: %v", err)
			}

			if retrieved != value {
				t.Errorf("Expected value %s, got %s", value, retrieved)
			}
		},
	)

	t.Run(
		"Get non-existent key", func(t *testing.T) {
			var retrieved string
			err := cache.Get(ctx, "non_existent_key", &retrieved)
			if !errors.Is(err, cachemar.ErrNotFound) {
				t.Errorf("Expected ErrNotFound, got %v", err)
			}
		},
	)

	t.Run(
		"Set with TTL", func(t *testing.T) {
			key := "ttl_key"
			value := "ttl_value"
			if err := cache.Set(ctx, key, value, time.Second, nil); err != nil {
				t.Fatalf("Set failed: %v", err)
			}

			time.Sleep(2 * time.Second)

			var retrieved string
			err := cache.Get(ctx, key, &retrieved)
			if !errors.Is(err, cachemar.ErrNotFound) {
				t.Errorf("Expected ErrNotFound for expired key, got %v", err)
			}
		},
	)

	t.Run(
		"Exists", func(t *testing.T) {
			key := "exists_key"
			value := "exists_value"
			_ = cache.Set(ctx, key, value, time.Minute, nil)

			exists, err := cache.Exists(ctx, key)
			if err != nil || !exists {
				t.Errorf("Expected key to exist, got exists=%v, err=%v", exists, err)
			}
		},
	)

	t.Run(
		"Remove", func(t *testing.T) {
			key := "remove_key"
			value := "remove_value"
			_ = cache.Set(ctx, key, value, time.Minute, nil)

			if err := cache.Remove(ctx, key); err != nil {
				t.Fatalf("Remove failed: %v", err)
			}

			var retrieved string
			err := cache.Get(ctx, key, &retrieved)
			if !errors.Is(err, cachemar.ErrNotFound) {
				t.Errorf("Expected ErrNotFound, got %v", err)
			}
		},
	)

	t.Run(
		"Remove by Tag", func(t *testing.T) {
			key := "tagged_key"
			value := "tagged_value"
			tag := "test_tag"
			_ = cache.Set(ctx, key, value, time.Minute, []string{tag})

			if err := cache.RemoveByTag(ctx, tag); err != nil {
				t.Fatalf("RemoveByTag failed: %v", err)
			}

			var retrieved string
			err := cache.Get(ctx, key, &retrieved)
			if !errors.Is(err, cachemar.ErrNotFound) {
				t.Errorf("Expected ErrNotFound, got %v", err)
			}
		},
	)

	t.Run(
		"Increment and Decrement", func(t *testing.T) {
			key := "counter"
			initialValue := 10
			if err := cache.Set(ctx, key, initialValue, time.Minute, nil); err != nil {
				t.Fatalf("Set failed: %v", err)
			}

			// Increment
			if err := cache.Increment(ctx, key); err != nil {
				t.Fatalf("Increment failed: %v", err)
			}

			var retrieved int
			if err := cache.Get(ctx, key, &retrieved); err != nil {
				t.Fatalf("Get after Increment failed: %v", err)
			}
			if retrieved != 11 {
				t.Errorf("Expected value 11 after increment, got %d", retrieved)
			}

			// Decrement
			if err := cache.Decrement(ctx, key); err != nil {
				t.Fatalf("Decrement failed: %v", err)
			}

			if err := cache.Get(ctx, key, &retrieved); err != nil {
				t.Fatalf("Get after Decrement failed: %v", err)
			}
			if retrieved != 10 {
				t.Errorf("Expected value 10 after decrement, got %d", retrieved)
			}
		},
	)

	t.Run(
		"Ping", func(t *testing.T) {
			if err := cache.Ping(); err != nil {
				t.Errorf("Ping failed: %v", err)
			}
		},
	)
}
