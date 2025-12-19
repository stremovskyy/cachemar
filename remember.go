package cachemar

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"
)

// GetterSetter captures the minimal cache contract Remember depends on.
type GetterSetter interface {
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, ttl time.Duration, tags []string) error
}

// RememberOption configures the cache-aside helper.
type RememberOption func(*rememberOptions)

type rememberOptions struct {
	codec               Codec
	ttlJitter           time.Duration
	localFallbackTTL    time.Duration
	localFallbackMaxEnt int
	failOnSetError      bool
	group               *flightGroup
	fallback            *fallbackCache
}

func defaultRememberOptions() rememberOptions {
	return rememberOptions{
		codec:               JSONCodec{},
		ttlJitter:           0,
		localFallbackTTL:    0,
		localFallbackMaxEnt: 4096,
		failOnSetError:      false,
		group:               &defaultFlightGroup,
		fallback:            defaultFallbackCache, // lazy init
	}
}

func WithTTLJitter(max time.Duration) RememberOption {
	return func(o *rememberOptions) { o.ttlJitter = max }
}

// WithLocalFallback enables an in-memory fallback for the provided TTL to absorb cache/backend failures.
func WithLocalFallback(ttl time.Duration) RememberOption {
	return func(o *rememberOptions) { o.localFallbackTTL = ttl }
}

// WithLocalFallbackMaxEntries configures the max entries for the in-memory fallback buffer.
func WithLocalFallbackMaxEntries(n int) RememberOption {
	return func(o *rememberOptions) {
		if n > 0 {
			o.localFallbackMaxEnt = n
		}
	}
}

// WithCodec allows swapping the serialization format for the fallback cache and singleflight responses.
func WithCodec(c Codec) RememberOption {
	return func(o *rememberOptions) {
		if c != nil {
			o.codec = c
		}
	}
}

// WithFailOnSetError makes Remember return an error if writing to the cache fails.
func WithFailOnSetError(v bool) RememberOption {
	return func(o *rememberOptions) { o.failOnSetError = v }
}

// Remember implements a safe cache-aside:
// - Get -> hit => return
// - miss/error -> singleflight (per instance) -> loader()
// - Set with TTL + jitter
// - (optional) local fallback for N seconds to avoid overwhelming the DB on cache errors
func Remember(
	ctx context.Context,
	cache GetterSetter,
	key string,
	dest any,
	ttl time.Duration,
	tags []string,
	loader func(context.Context) (any, error),
	opts ...RememberOption,
) error {
	if cache == nil {
		return errors.New("cachemar: cache is nil")
	}
	if key == "" {
		return errors.New("cachemar: key is empty")
	}
	if dest == nil {
		return errors.New("cachemar: dest is nil")
	}
	if t := reflect.TypeOf(dest); t == nil || t.Kind() != reflect.Pointer {
		return fmt.Errorf("cachemar: dest must be a pointer, got %T", dest)
	}
	if loader == nil {
		return errors.New("cachemar: loader is nil")
	}

	o := defaultRememberOptions()
	for _, fn := range opts {
		if fn != nil {
			fn(&o)
		}
	}

	// init fallback cache once with desired max entries
	if o.fallback == nil {
		o.fallback = newFallbackCache(o.localFallbackMaxEnt)
	} else {
		o.fallback.SetMax(o.localFallbackMaxEnt)
	}

	// 1) Trying from cache
	if err := cache.Get(ctx, key, dest); err == nil {
		return nil
	}

	// 2) cache miss -> fallback
	if o.localFallbackTTL > 0 && o.fallback != nil {
		if b, ok := o.fallback.Get(key); ok {
			if uErr := o.codec.Unmarshal(b, dest); uErr == nil {
				return nil
			}
		}
	}

	// 3) Singleflight for launched instance
	b, err := o.group.Do(
		ctx, key, func(ctx context.Context) ([]byte, error) {
			// double-check in case another goroutine warmed the cache meanwhile
			tmp, tmpErr := newValueLike(dest)
			if tmpErr == nil {
				if err := cache.Get(ctx, key, tmp); err == nil {
					mb, mErr := o.codec.Marshal(tmp)
					if mErr != nil {
						return nil, mErr
					}
					return mb, nil
				}
			}

			// loader (DB)
			v, lErr := loader(ctx)
			if lErr != nil {
				return nil, lErr
			}

			mb, mErr := o.codec.Marshal(v)
			if mErr != nil {
				return nil, mErr
			}

			effTTL := applyTTLJitter(ttl, o.ttlJitter)
			if sErr := cache.Set(ctx, key, v, effTTL, tags); sErr != nil && o.failOnSetError {
				return nil, sErr
			}

			return mb, nil
		},
	)
	if err != nil {
		return err
	}

	if err := o.codec.Unmarshal(b, dest); err != nil {
		return fmt.Errorf("cachemar: unmarshal into dest failed: %w", err)
	}

	if o.localFallbackTTL > 0 && o.fallback != nil {
		o.fallback.Set(key, b, o.localFallbackTTL)
	}

	return nil
}

func newValueLike(dest any) (any, error) {
	t := reflect.TypeOf(dest)
	if t == nil || t.Kind() != reflect.Pointer {
		return nil, fmt.Errorf("cachemar: dest must be a pointer, got %T", dest)
	}
	return reflect.New(t.Elem()).Interface(), nil
}

type flightGroup struct {
	mu sync.Mutex
	m  map[string]*flightCall
}

type flightCall struct {
	done chan struct{}
	val  []byte
	err  error
}

var defaultFlightGroup flightGroup

func (g *flightGroup) Do(ctx context.Context, key string, fn func(context.Context) ([]byte, error)) ([]byte, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*flightCall)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		select {
		case <-c.done:
			return c.val, c.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	c := &flightCall{done: make(chan struct{})}
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn(ctx)
	close(c.done)

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}

type fallbackCache struct {
	mu       sync.RWMutex
	max      int
	items    map[string]fallbackItem
	sweepCtr uint64
}

type fallbackItem struct {
	exp time.Time
	b   []byte
}

var defaultFallbackCache = newFallbackCache(4096)

func newFallbackCache(max int) *fallbackCache {
	if max <= 0 {
		max = 4096
	}
	return &fallbackCache{
		max:   max,
		items: make(map[string]fallbackItem, max),
	}
}

func (c *fallbackCache) SetMax(max int) {
	if max <= 0 {
		return
	}
	c.mu.Lock()
	c.max = max
	c.mu.Unlock()
}

func (c *fallbackCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	it, ok := c.items[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(it.exp) {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return nil, false
	}

	// // copy to avoid accidental mutations from outside!
	out := make([]byte, len(it.b))
	copy(out, it.b)
	return out, true
}

func (c *fallbackCache) Set(key string, b []byte, ttl time.Duration) {
	if ttl <= 0 {
		return
	}
	exp := time.Now().Add(ttl)

	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.items) >= c.max {
		for k := range c.items {
			delete(c.items, k)
			break
		}
	}

	cp := make([]byte, len(b))
	copy(cp, b)
	c.items[key] = fallbackItem{exp: exp, b: cp}
}
