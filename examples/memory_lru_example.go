package main

import (
	"context"
	"fmt"
	"time"

	"github.com/stremovskyy/cachemar/drivers/memory"
)

func main() {
	ctx := context.Background()

	// Create a memory cache with LRU eviction and max size of 3 items
	cache := memory.NewWithConfig(memory.Config{MaxSize: 3})

	fmt.Println("=== Memory Cache with LRU Support Example ===")
	fmt.Println()

	// Add initial items
	fmt.Println("Adding 3 items to cache (max capacity = 3):")
	cache.Set(ctx, "user:1", map[string]interface{}{"name": "John", "age": 30}, time.Hour, []string{"user"})
	cache.Set(ctx, "user:2", map[string]interface{}{"name": "Jane", "age": 25}, time.Hour, []string{"user"})
	cache.Set(ctx, "user:3", map[string]interface{}{"name": "Bob", "age": 35}, time.Hour, []string{"user"})

	fmt.Println("  - user:1 (John)")
	fmt.Println("  - user:2 (Jane)")
	fmt.Println("  - user:3 (Bob)")
	fmt.Println()

	// Check that all items exist
	fmt.Println("Verifying all items exist:")
	for i := 1; i <= 3; i++ {
		key := fmt.Sprintf("user:%d", i)
		var user map[string]interface{}
		if err := cache.Get(ctx, key, &user); err != nil {
			fmt.Printf("  - %s: NOT FOUND\n", key)
		} else {
			fmt.Printf("  - %s: %v\n", key, user)
		}
	}
	fmt.Println()

	// Access user:1 to make it most recently used
	fmt.Println("Accessing user:1 to make it most recently used...")
	var user map[string]interface{}
	cache.Get(ctx, "user:1", &user)
	fmt.Printf("  Retrieved: %v\n", user)
	fmt.Println()

	// Add a 4th item, which should evict the LRU item (user:2)
	fmt.Println("Adding 4th item (should evict least recently used - user:2):")
	cache.Set(ctx, "user:4", map[string]interface{}{"name": "Alice", "age": 28}, time.Hour, []string{"user"})
	fmt.Println("  - user:4 (Alice)")
	fmt.Println()

	// Check which items still exist
	fmt.Println("Current cache contents after LRU eviction:")
	for i := 1; i <= 4; i++ {
		key := fmt.Sprintf("user:%d", i)
		var user map[string]interface{}
		if err := cache.Get(ctx, key, &user); err != nil {
			fmt.Printf("  - %s: EVICTED (LRU)\n", key)
		} else {
			fmt.Printf("  - %s: %v\n", key, user)
		}
	}
	fmt.Println()

	// Demonstrate unlimited cache (default behavior)
	fmt.Println("=== Unlimited Cache (Default Behavior) ===")
	unlimitedCache := memory.New() // Default is unlimited

	fmt.Println("Adding 1000 items to unlimited cache...")
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("item:%d", i)
		value := fmt.Sprintf("value-%d", i)
		unlimitedCache.Set(ctx, key, value, time.Hour, nil)
	}

	// Verify all items still exist
	allExist := true
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("item:%d", i)
		var value string
		if err := unlimitedCache.Get(ctx, key, &value); err != nil {
			allExist = false
			break
		}
	}

	if allExist {
		fmt.Println("✓ All 1000 items still exist (no eviction)")
	} else {
		fmt.Println("✗ Some items were evicted (unexpected)")
	}
	fmt.Println()

	// Demonstrate LRU with counter operations
	fmt.Println("=== LRU with Counter Operations ===")
	counterCache := memory.NewWithConfig(memory.Config{MaxSize: 2})

	counterCache.Set(ctx, "counter1", 10, time.Hour, nil)
	counterCache.Set(ctx, "counter2", 20, time.Hour, nil)

	// Increment counter1 (makes it most recently used)
	counterCache.Increment(ctx, "counter1")

	// Add counter3 (should evict counter2)
	counterCache.Set(ctx, "counter3", 30, time.Hour, nil)

	fmt.Println("After incrementing counter1 and adding counter3:")
	for i := 1; i <= 3; i++ {
		key := fmt.Sprintf("counter%d", i)
		var value int
		if err := counterCache.Get(ctx, key, &value); err != nil {
			fmt.Printf("  - %s: EVICTED\n", key)
		} else {
			fmt.Printf("  - %s: %d\n", key, value)
		}
	}

	fmt.Println()
	fmt.Println("=== Summary ===")
	fmt.Println("✓ LRU eviction prevents unlimited memory growth")
	fmt.Println("✓ Backward compatibility maintained (unlimited by default)")
	fmt.Println("✓ Access patterns (Get, Set, Increment, Decrement) update LRU order")
	fmt.Println("✓ Configurable max cache size prevents application crashes")
}
