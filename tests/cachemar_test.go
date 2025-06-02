package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stremovskyy/cachemar"
	"github.com/stremovskyy/cachemar/drivers/memcached"
	"github.com/stremovskyy/cachemar/drivers/redis"
)

func setupCachemar() cachemar.Manager {
	cachemarService := cachemar.New()

	// Register Redis cache service
	redisOptions := redis.NewSingleInstanceOptions("127.0.0.1:6379", "", 0).
		WithPrefix("test")
	redisCacheService := redis.New(redisOptions)
	cachemarService.Register("redis", redisCacheService)

	// Register Memcached cache service
	memcachedOptions := &memcached.Options{
		Servers: []string{"127.0.0.1:11211"},
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
	if err != nil {
		t.Logf("Redis not available: %v", err)
	} else {
		var value string
		err = cachemarService.Get(ctx, "key", &value)
		assert.NoError(t, err)
		assert.Equal(t, "value", value)
	}

	// Test Memcached service
	cachemarService.SetCurrent("memcached")
	err = cachemarService.Set(ctx, "key", "value", 1*time.Minute, []string{})
	if err != nil {
		t.Logf("Memcached not available: %v", err)
	} else {
		var value string
		err = cachemarService.Get(ctx, "key", &value)
		assert.NoError(t, err)
		assert.Equal(t, "value", value)
	}
}
