package redis

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/stremovskyy/cachemar"
	"io"
	"strings"
	"sync"
	"time"
)

const compressedDataPrefix = "COMPRESSED:"

// RedisCacheService is a service for caching data in Redis
type redisCacheService struct {
	mu       sync.Mutex
	client   *redis.Client
	prefix   string
	compress bool // New field to enable/disable Gzip compression
}

type Options struct {
	DSN                string
	Password           string
	Database           int
	CompressionEnabled bool
	Prefix             string
}

func NewCacheService(options *Options) cachemar.Cacher {
	client := redis.NewClient(&redis.Options{
		Addr:     options.DSN,
		Password: options.Password, // Set password if required
		DB:       options.Database, // Use default database
	})

	return &redisCacheService{
		client:   client,
		compress: options.CompressionEnabled,
		prefix:   options.Prefix,
	}
}

func (c *redisCacheService) Name() string {
	return "cache"
}

func (c *redisCacheService) Init() error {
	statusCmd := c.client.Ping(context.Background())
	if err := statusCmd.Err(); err != nil {
		return err
	}

	return nil
}

func (c *redisCacheService) Run(ctx context.Context) error {
	return nil
}

func (c *redisCacheService) Stop() error {
	return c.client.Close()
}

func (c *redisCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration, tags []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %v", err)
	}

	finalKey := c.keyWithPrefix(key)

	// Optionally compress the data using Gzip if compression is enabled
	if c.compress {
		compressedData, err := compressData(data)
		if err != nil {
			return fmt.Errorf("failed to compress data: %v", err)
		}
		data = append([]byte(compressedDataPrefix), compressedData...)
	}

	err = c.client.Set(ctx, finalKey, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set key-value pair in Redis: %v", err)
	}

	if len(tags) > 0 {
		for _, tag := range tags {
			keyForTags := getTagKey(tag)

			err = c.client.SAdd(ctx, keyForTags, finalKey).Err()
			if err != nil {
				return fmt.Errorf("failed to add key to tag: %v", err)
			}
		}
	}

	return nil
}

func compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *redisCacheService) Get(ctx context.Context, key string, value interface{}) error {
	finalKey := c.keyWithPrefix(key)

	data, err := c.client.Get(ctx, finalKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return fmt.Errorf("key not found: %s", finalKey)
		}
		return fmt.Errorf("failed to get value from Redis: %v", err)
	}

	if strings.HasPrefix(string(data), compressedDataPrefix) {
		data, err = decompressData(data[len(compressedDataPrefix):])
		if err != nil {
			return fmt.Errorf("failed to decompress data: %v", err)
		}
	}

	err = json.Unmarshal(data, value)
	if err != nil {
		return fmt.Errorf("failed to deserialize value: %v", err)
	}

	return nil
}

func decompressData(compressedData []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, gz); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (c *redisCacheService) Remove(ctx context.Context, key string) error {
	finalKey := c.keyWithPrefix(key)

	err := c.client.Del(ctx, finalKey).Err()
	if err != nil {
		return fmt.Errorf("failed to remove key from Redis: %v", err)
	}

	return nil
}

func (c *redisCacheService) RemoveByTag(ctx context.Context, tag string) error {
	keyForTags := getTagKey(tag)

	keys, err := c.client.SMembers(ctx, keyForTags).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys associated with tag: %v", err)
	}

	for _, key := range keys {
		err := c.client.Del(ctx, key).Err()
		if err != nil {
			return fmt.Errorf("failed to remove key from Redis: %v", err)
		}
	}

	return nil
}
func (c *redisCacheService) Exists(ctx context.Context, key string) (bool, error) {
	finalKey := c.keyWithPrefix(key)

	cmd := c.client.Exists(ctx, finalKey)
	if err := cmd.Err(); err != nil {
		return false, fmt.Errorf("failed to check key existence in Redis: %v", err)
	}
	return cmd.Val() > 0, nil
}

func (c *redisCacheService) Increment(ctx context.Context, key string) error {
	finalKey := c.keyWithPrefix(key)

	cmd := c.client.Incr(ctx, finalKey)
	if err := cmd.Err(); err != nil {
		return fmt.Errorf("failed to increment key value in Redis: %v", err)
	}
	return nil
}

func (c *redisCacheService) Decrement(ctx context.Context, key string) error {
	finalKey := c.keyWithPrefix(key)

	cmd := c.client.Decr(ctx, finalKey)
	if err := cmd.Err(); err != nil {
		return fmt.Errorf("failed to decrement key value in Redis: %v", err)
	}
	return nil
}

func (c *redisCacheService) GetKeysByTag(ctx context.Context, tag string) ([]string, error) {
	keyForTags := getTagKey(tag)

	cmd := c.client.SMembers(ctx, keyForTags)
	if err := cmd.Err(); err != nil {
		return nil, fmt.Errorf("failed to get keys associated with tag: %v", err)
	}
	return cmd.Val(), nil
}

func (c *redisCacheService) RemoveByTags(ctx context.Context, tags []string) error {
	for _, tag := range tags {
		err := c.RemoveByTag(ctx, tag)
		if err != nil {
			return fmt.Errorf("failed to remove keys for tag: %v", err)
		}
	}

	return nil
}

func getTagKey(tag string) string {
	return fmt.Sprintf("tag:%s", tag)
}

func (c *redisCacheService) keyWithPrefix(key string) string {
	return fmt.Sprintf("%s:%s", c.prefix, key)
}
