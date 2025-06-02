package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/stremovskyy/cachemar/drivers/redis"
)

func main() {
	fmt.Println("=== Single Redis Instance Example ===")
	singleInstance()

	fmt.Println("\n=== Redis Cluster Example ===")
	clusterExample()

	fmt.Println("\n=== Redis Cluster with Custom Config ===")
	clusterWithCustomConfig()
}

func singleInstance() {
	options := redis.NewSingleInstanceOptions("localhost:6379", "", 0).
		WithCompression().
		WithPrefix("app")

	cache := redis.New(options)
	defer cache.Close()

	// Test connectivity
	if err := cache.Ping(); err != nil {
		log.Printf("Single instance not available: %v", err)
		return
	}

	// Store some data
	ctx := context.Background()
	err := cache.Set(
		ctx, "user:123", map[string]interface{}{
			"name": "John Doe",
		}, time.Hour, []string{"users", "active"},
	)

	if err != nil {
		log.Printf("Error setting data: %v", err)
		return
	}

	// Retrieve data
	var user map[string]interface{}
	err = cache.Get(ctx, "user:123", &user)
	if err != nil {
		log.Printf("Error getting data: %v", err)
		return
	}

	fmt.Printf("Retrieved user: %+v\n", user)
}

func clusterExample() {
	// Create options for Redis cluster with default settings
	options := redis.NewClusterOptions(
		[]string{
			"localhost:7000",
			"localhost:7001",
			"localhost:7002",
		},
		"", // no password
	).WithCompression().WithPrefix("cluster-app")

	cache := redis.New(options)
	defer cache.Close()

	// Test connectivity
	if err := cache.Ping(); err != nil {
		log.Printf("Cluster not available: %v", err)
		return
	}

	ctx := context.Background()

	// Store data across the cluster
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("item:%d", i)
		value := fmt.Sprintf("value-%d", i)
		tags := []string{"items", fmt.Sprintf("batch-%d", i/5)}

		err := cache.Set(ctx, key, value, time.Hour, tags)
		if err != nil {
			log.Printf("Error setting %s: %v", key, err)
			continue
		}
	}

	// Retrieve data by tag
	keys, err := cache.GetKeysByTag(ctx, "batch-0")
	if err != nil {
		log.Printf("Error getting keys by tag: %v", err)
		return
	}

	fmt.Printf("Keys in batch-0: %v\n", keys)

	// Test increment operation
	err = cache.Increment(ctx, "counter")
	if err != nil {
		log.Printf("Error incrementing counter: %v", err)
		return
	}

	fmt.Println("Successfully stored and retrieved data from cluster")
}

func clusterWithCustomConfig() {
	customConfig := &redis.ClusterOptions{
		MaxRedirects:   5,                // Allow more redirects
		ReadOnly:       false,            // Enable write operations
		RouteByLatency: true,             // Route to fastest node
		RouteRandomly:  false,            // Don't use random routing
		PoolSize:       20,               // Larger connection pool
		PoolTimeout:    time.Minute,      // Longer pool timeout
		MinIdleConns:   5,                // More idle connections
		MaxIdleConns:   10,               // Higher idle connection limit
		DialTimeout:    time.Second * 10, // Longer dial timeout
		ReadTimeout:    time.Second * 10, // Longer read timeout
		WriteTimeout:   time.Second * 10, // Longer write timeout
	}

	options := redis.NewClusterOptions(
		[]string{
			"localhost:7000",
			"localhost:7001",
			"localhost:7002",
		},
		"my-cluster-password", // with password
	).WithClusterConfig(customConfig).
		WithPrefix("high-perf").
		WithCompression()

	cache := redis.New(options)
	defer cache.Close()

	if err := cache.Ping(); err != nil {
		log.Printf("Custom cluster not available: %v", err)
		return
	}

	ctx := context.Background()

	start := time.Now()
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("perf-test:%d", i)
		value := map[string]interface{}{
			"id":        i,
			"timestamp": time.Now().Unix(),
			"data":      fmt.Sprintf("large-data-payload-%d", i),
		}

		err := cache.Set(ctx, key, value, time.Hour, []string{"performance"})
		if err != nil {
			log.Printf("Error in performance test: %v", err)
			break
		}
	}
	duration := time.Since(start)

	fmt.Printf(
		"Stored 1000 items in %v (%.2f ops/sec)\n",
		duration, 1000.0/duration.Seconds(),
	)
}
