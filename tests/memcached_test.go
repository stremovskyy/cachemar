package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stremovskyy/cachemar"
	"github.com/stremovskyy/cachemar/drivers/memcached"
)

const testPrefix = "test"

var memcacheCacheService cachemar.Cacher

func setup() {
	memcacheCacheService = memcached.New(
		&memcached.Options{
			Servers: []string{"127.0.0.1:11211"},
			Prefix:  testPrefix,
		},
	)
}

func TestSetGet(t *testing.T) {
	setup()
	ctx := context.Background()

	err := memcacheCacheService.Set(ctx, "key", "value", 1*time.Second, []string{"tag1", "tag2"})
	if err != nil {
		t.Skipf("Memcached not available: %v", err)
		return
	}

	var value string
	err = memcacheCacheService.Get(ctx, "key", &value)
	assert.NoError(t, err)
	assert.Equal(t, "value", value)

	err = memcacheCacheService.Remove(ctx, "key")
	assert.NoError(t, err)
}

func TestRemoveByTag(t *testing.T) {
	setup()
	ctx := context.Background()

	err := memcacheCacheService.Set(ctx, "key1", "value", 1*time.Second, []string{"tag1"})
	if err != nil {
		t.Skipf("Memcached not available: %v", err)
		return
	}

	err = memcacheCacheService.Set(ctx, "key2", "value", 1*time.Second, []string{"tag2"})
	assert.NoError(t, err)

	err = memcacheCacheService.RemoveByTag(ctx, "tag1")
	assert.NoError(t, err)

	var value string
	err = memcacheCacheService.Get(ctx, "key1", &value)
	assert.Error(t, err)

	err = memcacheCacheService.Get(ctx, "key2", &value)
	assert.NoError(t, err)
	assert.Equal(t, "value", value)

	err = memcacheCacheService.Remove(ctx, "key2")
	assert.NoError(t, err)
}

func TestIncrementDecrement(t *testing.T) {
	setup()
	ctx := context.Background()

	err := memcacheCacheService.Set(ctx, "key", "1", 1*time.Minute, []string{})
	if err != nil {
		t.Skipf("Memcached not available: %v", err)
		return
	}

	err = memcacheCacheService.Increment(ctx, "key")
	assert.NoError(t, err)

	var value string
	err = memcacheCacheService.Get(ctx, "key", &value)
	assert.NoError(t, err)
	assert.Equal(t, "2", value)

	err = memcacheCacheService.Decrement(ctx, "key")
	assert.NoError(t, err)

	err = memcacheCacheService.Get(ctx, "key", &value)
	assert.NoError(t, err)
	assert.Equal(t, "1", value)

	err = memcacheCacheService.Remove(ctx, "key")
	assert.NoError(t, err)
}
