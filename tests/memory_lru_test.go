package tests_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stremovskyy/cachemar"
	"github.com/stremovskyy/cachemar/drivers/memory"
)

func TestMemoryLRU(t *testing.T) {
	ctx := context.Background()

	t.Run("LRU Eviction", func(t *testing.T) {
		// Create cache with max size of 3
		cache := memory.NewWithConfig(memory.Config{MaxSize: 3})

		// Add 4 items, the first one should be evicted
		_ = cache.Set(ctx, "key1", "value1", time.Hour, nil)
		_ = cache.Set(ctx, "key2", "value2", time.Hour, nil)
		_ = cache.Set(ctx, "key3", "value3", time.Hour, nil)
		_ = cache.Set(ctx, "key4", "value4", time.Hour, nil)

		// key1 should be evicted (LRU)
		var value string
		err := cache.Get(ctx, "key1", &value)
		if err != cachemar.ErrNotFound {
			t.Errorf("Expected key1 to be evicted, but got: %v", err)
		}

		// Other keys should still exist
		err = cache.Get(ctx, "key2", &value)
		if err != nil {
			t.Errorf("Expected key2 to exist, but got error: %v", err)
		}

		err = cache.Get(ctx, "key3", &value)
		if err != nil {
			t.Errorf("Expected key3 to exist, but got error: %v", err)
		}

		err = cache.Get(ctx, "key4", &value)
		if err != nil {
			t.Errorf("Expected key4 to exist, but got error: %v", err)
		}
	})

	t.Run("LRU Update Order on Get", func(t *testing.T) {
		cache := memory.NewWithConfig(memory.Config{MaxSize: 3})

		// Add 3 items
		_ = cache.Set(ctx, "key1", "value1", time.Hour, nil)
		_ = cache.Set(ctx, "key2", "value2", time.Hour, nil)
		_ = cache.Set(ctx, "key3", "value3", time.Hour, nil)

		// Access key1 to make it most recently used
		var value string
		_ = cache.Get(ctx, "key1", &value)

		// Add key4, key2 should be evicted (was LRU)
		_ = cache.Set(ctx, "key4", "value4", time.Hour, nil)

		// key2 should be evicted
		err := cache.Get(ctx, "key2", &value)
		if err != cachemar.ErrNotFound {
			t.Errorf("Expected key2 to be evicted, but got: %v", err)
		}

		// key1 should still exist (was accessed)
		err = cache.Get(ctx, "key1", &value)
		if err != nil {
			t.Errorf("Expected key1 to exist, but got error: %v", err)
		}

		// key3 should still exist
		err = cache.Get(ctx, "key3", &value)
		if err != nil {
			t.Errorf("Expected key3 to exist, but got error: %v", err)
		}

		// key4 should exist
		err = cache.Get(ctx, "key4", &value)
		if err != nil {
			t.Errorf("Expected key4 to exist, but got error: %v", err)
		}
	})

	t.Run("LRU Update Order on Set Existing Key", func(t *testing.T) {
		cache := memory.NewWithConfig(memory.Config{MaxSize: 3})

		// Add 3 items
		_ = cache.Set(ctx, "key1", "value1", time.Hour, nil)
		_ = cache.Set(ctx, "key2", "value2", time.Hour, nil)
		_ = cache.Set(ctx, "key3", "value3", time.Hour, nil)

		// Update key1 to make it most recently used
		_ = cache.Set(ctx, "key1", "new_value1", time.Hour, nil)

		// Add key4, key2 should be evicted (was LRU)
		_ = cache.Set(ctx, "key4", "value4", time.Hour, nil)

		// key2 should be evicted
		var value string
		err := cache.Get(ctx, "key2", &value)
		if err != cachemar.ErrNotFound {
			t.Errorf("Expected key2 to be evicted, but got: %v", err)
		}

		// key1 should still exist with updated value
		err = cache.Get(ctx, "key1", &value)
		if err != nil {
			t.Errorf("Expected key1 to exist, but got error: %v", err)
		}
		if value != "new_value1" {
			t.Errorf("Expected key1 value to be 'new_value1', got: %s", value)
		}
	})

	t.Run("Unlimited Cache (Backward Compatibility)", func(t *testing.T) {
		cache := memory.New() // Default should be unlimited

		// Add many items - none should be evicted
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("key%d", i)
			value := fmt.Sprintf("value%d", i)
			_ = cache.Set(ctx, key, value, time.Hour, nil)
		}

		// All items should still exist
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("key%d", i)
			var value string
			err := cache.Get(ctx, key, &value)
			if err != nil {
				t.Errorf("Expected %s to exist, but got error: %v", key, err)
				break
			}
		}
	})

	t.Run("LRU with Increment/Decrement", func(t *testing.T) {
		cache := memory.NewWithConfig(memory.Config{MaxSize: 3})

		// Add 3 integer items
		_ = cache.Set(ctx, "key1", 1, time.Hour, nil)
		_ = cache.Set(ctx, "key2", 2, time.Hour, nil)
		_ = cache.Set(ctx, "key3", 3, time.Hour, nil)

		// Increment key1 to make it most recently used
		_ = cache.Increment(ctx, "key1")

		// Add key4, key2 should be evicted
		_ = cache.Set(ctx, "key4", 4, time.Hour, nil)

		// key2 should be evicted
		var value int
		err := cache.Get(ctx, "key2", &value)
		if err != cachemar.ErrNotFound {
			t.Errorf("Expected key2 to be evicted, but got: %v", err)
		}

		// key1 should still exist with incremented value
		err = cache.Get(ctx, "key1", &value)
		if err != nil {
			t.Errorf("Expected key1 to exist, but got error: %v", err)
		}
		if value != 2 {
			t.Errorf("Expected key1 value to be 2, got: %d", value)
		}
	})
}
