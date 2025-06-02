package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stremovskyy/cachemar/drivers/redis"
	"github.com/stretchr/testify/assert"
)

func TestRedisCacheService(t *testing.T) {
	options := redis.NewSingleInstanceOptions("127.0.0.1:6379", "", 0).
		WithCompression().
		WithPrefix("prefix")

	cacheService := redis.New(options)

	// Test Set
	err := cacheService.Set(context.Background(), "testKey", "testValue", time.Minute, []string{"tag1", "tag2"})
	if err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}

	// Test Get
	var val string
	err = cacheService.Get(context.Background(), "testKey", &val)
	assert.NoError(t, err)
	assert.Equal(t, "testValue", val)

	// Test Exists
	exists, err := cacheService.Exists(context.Background(), "testKey")
	assert.NoError(t, err)
	assert.True(t, exists)

	var intKey string = "testIntKey"

	// Test Increment
	err = cacheService.Increment(context.Background(), intKey)
	assert.NoError(t, err)

	// Test Decrement
	err = cacheService.Decrement(context.Background(), intKey)
	assert.NoError(t, err)

	// Test GetKeysByTag
	keys, err := cacheService.GetKeysByTag(context.Background(), "tag1")
	assert.NoError(t, err)
	assert.Contains(t, keys, "prefix:testKey")

	// Test RemoveByTag
	err = cacheService.RemoveByTag(context.Background(), "tag1")
	assert.NoError(t, err)

	// Test RemoveByTags
	err = cacheService.RemoveByTags(context.Background(), []string{"tag2"})
	assert.NoError(t, err)

	// Test Remove
	err = cacheService.Remove(context.Background(), "testKey")
	assert.NoError(t, err)

	// Test key not found after remove
	exists, err = cacheService.Exists(context.Background(), "testKey")
	assert.NoError(t, err)
	assert.False(t, exists)
}
