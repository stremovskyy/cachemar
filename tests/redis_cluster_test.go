package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stremovskyy/cachemar/drivers/redis"
)

func TestRedisClusterCacheService(t *testing.T) {
	clusterOptions := redis.NewClusterOptions(
		[]string{"localhost:7000", "localhost:7001", "localhost:7002"},
		"",
	).WithCompression().WithPrefix("test-cluster")

	cacheService := redis.New(clusterOptions)

	err := cacheService.Set(context.Background(), "testKey", "testValue", time.Minute, []string{"tag1", "tag2"})
	if err != nil {
		t.Skipf("Redis cluster not available: %v", err)
		return
	}

	var val string
	err = cacheService.Get(context.Background(), "testKey", &val)
	assert.NoError(t, err)
	assert.Equal(t, "testValue", val)

	exists, err := cacheService.Exists(context.Background(), "testKey")
	assert.NoError(t, err)
	assert.True(t, exists)

	intKey := "testIntKey"
	err = cacheService.Increment(context.Background(), intKey)
	assert.NoError(t, err)

	err = cacheService.Decrement(context.Background(), intKey)
	assert.NoError(t, err)

	keys, err := cacheService.GetKeysByTag(context.Background(), "tag1")
	assert.NoError(t, err)
	assert.Contains(t, keys, "test-cluster:testKey")

	err = cacheService.RemoveByTag(context.Background(), "tag1")
	assert.NoError(t, err)

	exists, err = cacheService.Exists(context.Background(), "testKey")
	assert.NoError(t, err)
	assert.False(t, exists)

	err = cacheService.Close()
	assert.NoError(t, err)
}

func TestRedisClusterWithCustomConfig(t *testing.T) {
	customClusterOptions := &redis.ClusterOptions{
		MaxRedirects:   5,
		ReadOnly:       false,
		RouteByLatency: true,
		RouteRandomly:  false,
		PoolSize:       20,
		PoolTimeout:    time.Second * 60,
		MinIdleConns:   2,
		MaxIdleConns:   5,
		DialTimeout:    time.Second * 10,
		ReadTimeout:    time.Second * 10,
		WriteTimeout:   time.Second * 10,
	}

	options := redis.NewClusterOptions(
		[]string{"localhost:7000", "localhost:7001", "localhost:7002"},
		"",
	).WithClusterConfig(customClusterOptions).WithPrefix("custom-cluster")

	cacheService := redis.New(options)

	err := cacheService.Set(context.Background(), "customKey", "customValue", time.Minute, nil)
	if err != nil {
		t.Skipf("Redis cluster not available: %v", err)
		return
	}

	var val string
	err = cacheService.Get(context.Background(), "customKey", &val)
	assert.NoError(t, err)
	assert.Equal(t, "customValue", val)

	err = cacheService.Remove(context.Background(), "customKey")
	assert.NoError(t, err)

	err = cacheService.Close()
	assert.NoError(t, err)
}

func TestRedisBackwardCompatibility(t *testing.T) {

	options := redis.NewSingleInstanceOptions("localhost:6379", "", 0).
		WithCompression().
		WithPrefix("compat-test")

	cacheService := redis.New(options)

	err := cacheService.Set(context.Background(), "compatKey", "compatValue", time.Minute, []string{"compat-tag"})
	if err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}

	var val string
	err = cacheService.Get(context.Background(), "compatKey", &val)
	assert.NoError(t, err)
	assert.Equal(t, "compatValue", val)

	keys, err := cacheService.GetKeysByTag(context.Background(), "compat-tag")
	assert.NoError(t, err)
	assert.Contains(t, keys, "compat-test:compatKey")

	err = cacheService.RemoveByTag(context.Background(), "compat-tag")
	assert.NoError(t, err)

	err = cacheService.Close()
	assert.NoError(t, err)
}

func TestClusterOptionsBuilder(t *testing.T) {

	options := redis.NewClusterOptions(
		[]string{"localhost:7000", "localhost:7001"},
		"mypassword",
	).WithCompression().
		WithPrefix("builder-test").
		WithClusterConfig(
			&redis.ClusterOptions{
				MaxRedirects: 10,
				PoolSize:     50,
			},
		)

	assert.Equal(t, []string{"localhost:7000", "localhost:7001"}, options.ClusterAddrs)
	assert.Equal(t, "mypassword", options.Password)
	assert.True(t, options.CompressionEnabled)
	assert.Equal(t, "builder-test", options.Prefix)
	assert.Equal(t, 10, options.ClusterOptions.MaxRedirects)
	assert.Equal(t, 50, options.ClusterOptions.PoolSize)
}

func TestDefaultClusterOptions(t *testing.T) {
	defaults := redis.DefaultClusterOptions()

	assert.Equal(t, 3, defaults.MaxRedirects)
	assert.False(t, defaults.ReadOnly)
	assert.False(t, defaults.RouteByLatency)
	assert.False(t, defaults.RouteRandomly)
	assert.Equal(t, 10, defaults.PoolSize)
	assert.Equal(t, time.Second*30, defaults.PoolTimeout)
	assert.Equal(t, 1, defaults.MinIdleConns)
	assert.Equal(t, 3, defaults.MaxIdleConns)
	assert.Equal(t, time.Second*5, defaults.DialTimeout)
	assert.Equal(t, time.Second*5, defaults.ReadTimeout)
	assert.Equal(t, time.Second*5, defaults.WriteTimeout)
}
