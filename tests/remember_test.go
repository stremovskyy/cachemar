package tests

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/stremovskyy/cachemar"
)

type stubCache struct {
	mu      sync.Mutex
	store   map[string]any
	lastTTL time.Duration
	getErr  error
	setErr  error
}

func newStubCache() *stubCache {
	return &stubCache{
		store: make(map[string]any),
	}
}

func (c *stubCache) Get(_ context.Context, key string, dest any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.getErr != nil {
		return c.getErr
	}

	val, ok := c.store[key]
	if !ok {
		return cachemar.ErrNotFound
	}

	dstVal := reflect.ValueOf(dest)
	if dstVal.Kind() != reflect.Pointer {
		return fmt.Errorf("dest must be a pointer, got %T", dest)
	}

	dstVal.Elem().Set(reflect.ValueOf(val))
	return nil
}

func (c *stubCache) Set(_ context.Context, key string, value any, ttl time.Duration, _ []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.setErr != nil {
		return c.setErr
	}

	c.lastTTL = ttl
	c.store[key] = value
	return nil
}

func TestRememberCachesWithJitter(t *testing.T) {
	cache := newStubCache()
	ctx := context.Background()
	loaderCalls := 0

	var dest string
	err := cachemar.Remember(
		ctx,
		cache,
		"remember-key",
		&dest,
		2*time.Second,
		nil,
		func(ctx context.Context) (any, error) {
			loaderCalls++
			return "value", nil
		},
		cachemar.WithTTLJitter(500*time.Millisecond),
	)
	require.NoError(t, err)
	require.Equal(t, "value", dest)
	require.Equal(t, 1, loaderCalls)
	require.GreaterOrEqual(t, cache.lastTTL, 2*time.Second)
	require.LessOrEqual(t, cache.lastTTL, 2500*time.Millisecond)

	var destCached string
	err = cachemar.Remember(
		ctx,
		cache,
		"remember-key",
		&destCached,
		2*time.Second,
		nil,
		func(ctx context.Context) (any, error) {
			loaderCalls++
			return "stale", nil
		},
	)
	require.NoError(t, err)
	require.Equal(t, "value", destCached)
	require.Equal(t, 1, loaderCalls, "second call should be served from cache")
}

func TestRememberFallsBackOnCacheErrors(t *testing.T) {
	cache := newStubCache()
	cache.getErr = errors.New("cache unavailable")
	cache.setErr = errors.New("persist failure")

	ctx := context.Background()
	loaderCalls := 0

	var dest string
	err := cachemar.Remember(
		ctx,
		cache,
		"fallback-key",
		&dest,
		time.Minute,
		nil,
		func(ctx context.Context) (any, error) {
			loaderCalls++
			return "fresh", nil
		},
		cachemar.WithLocalFallback(time.Second),
	)
	require.NoError(t, err)
	require.Equal(t, "fresh", dest)
	require.Equal(t, 1, loaderCalls)

	cache.getErr = errors.New("still unavailable")

	var destFallback string
	err = cachemar.Remember(
		ctx,
		cache,
		"fallback-key",
		&destFallback,
		time.Minute,
		nil,
		func(ctx context.Context) (any, error) {
			loaderCalls++
			return "fresh-again", nil
		},
		cachemar.WithLocalFallback(time.Second),
	)
	require.NoError(t, err)
	require.Equal(t, "fresh", destFallback)
	require.Equal(t, 1, loaderCalls, "second call should be served from local fallback")
}

func TestRememberRespectsFailOnSetError(t *testing.T) {
	cache := newStubCache()
	cache.setErr = errors.New("write failure")

	ctx := context.Background()
	var dest string
	err := cachemar.Remember(
		ctx,
		cache,
		"fail-on-set",
		&dest,
		time.Minute,
		nil,
		func(ctx context.Context) (any, error) {
			return "value", nil
		},
		cachemar.WithFailOnSetError(true),
	)
	require.Error(t, err)
	require.Empty(t, dest)
}
