package tests

import (
	"context"
	"github.com/stremovskyy/cachemar"
	"github.com/stremovskyy/cachemar/drivers/memcached"
	"github.com/stremovskyy/cachemar/drivers/redis"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func setupCachemar() cachemar.Manager {
	cachemarService := cachemar.New()

	// Register Redis cache service
	redisOptions := &redis.Options{
		DSN:      "localhost:6379",
		Password: "",
		Database: 0,
		Prefix:   "test",
	}
	redisCacheService := redis.New(redisOptions)
	cachemarService.Register("redis", redisCacheService)

	// Register Memcached cache service
	memcachedOptions := &memcached.Options{
		Servers: []string{"localhost:11211"},
		Prefix:  "test",
	}
	memcachedCacheService := memcached.New(memcachedOptions)
	cachemarService.Register("memcached", memcachedCacheService)

	return cachemarService
}

func TestCacheMarService(t *testing.T) {
	ctx := context.Background()
	cachemarService := setupCachemar()

	// Test Redis service
	cachemarService.SetCurrent("redis")
	err := cachemarService.Set(ctx, "key", "value", 1*time.Minute, []string{})
	assert.NoError(t, err)

	var value string
	err = cachemarService.Get(ctx, "key", &value)
	assert.NoError(t, err)
	assert.Equal(t, "value", value)

	// Test Memcached service
	cachemarService.SetCurrent("memcached")
	err = cachemarService.Set(ctx, "key", "value", 1*time.Minute, []string{})
	assert.NoError(t, err)

	value = ""
	err = cachemarService.Get(ctx, "key", &value)
	assert.NoError(t, err)
	assert.Equal(t, "value", value)
}
