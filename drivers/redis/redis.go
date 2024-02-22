package redis

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/stremovskyy/cachemar"
)

// RedisCacheService is a service for caching data in Redis
type redisDriver struct {
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

func New(options *Options) cachemar.Cacher {
	client := redis.NewClient(
		&redis.Options{
			Addr:     options.DSN,
			Password: options.Password, // Set password if required
			DB:       options.Database, // Use default database
		},
	)

	return &redisDriver{
		client:   client,
		compress: options.CompressionEnabled,
		prefix:   options.Prefix,
	}
}

func (d *redisDriver) Name() string {
	return "cache"
}

func (d *redisDriver) Init() error {
	statusCmd := d.client.Ping(context.Background())
	if err := statusCmd.Err(); err != nil {
		return err
	}

	return nil
}

func (d *redisDriver) Run(ctx context.Context) error {
	return nil
}

func (d *redisDriver) Stop() error {
	return d.client.Close()
}

func (d *redisDriver) Set(ctx context.Context, key string, value interface{}, ttl time.Duration, tags []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %v", err)
	}

	finalKey := d.keyWithPrefix(key)

	// Optionally compress the data using Gzip if compression is enabled
	if d.compress {
		compressedData, err := compressData(data)
		if err != nil {
			return fmt.Errorf("failed to compress data: %v", err)
		}
		data = compressedData
	}

	err = d.client.Set(ctx, finalKey, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set key-value pair in Redis: %v", err)
	}

	if len(tags) > 0 {
		for _, tag := range tags {
			keyForTags := getTagKey(tag)

			err = d.client.SAdd(ctx, keyForTags, finalKey).Err()
			if err != nil {
				return fmt.Errorf("failed to add key to tag: %v", err)
			}

			err = d.client.Expire(ctx, keyForTags, ttl).Err()
			if err != nil {
				return fmt.Errorf("failed to set tag expiration: %v", err)
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

func (c *redisDriver) Get(ctx context.Context, key string, value interface{}) error {
	finalKey := c.keyWithPrefix(key)

	data, err := c.client.Get(ctx, finalKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return fmt.Errorf("key not found: %s", finalKey)
		}
		return fmt.Errorf("failed to get value from Redis: %v", err)
	}

	// Check if the data is compressed
	isCompressed := false
	if len(data) > 2 {
		if data[0] == 0x1f && data[1] == 0x8b {
			isCompressed = true
		}
	}

	if isCompressed {
		data, err = decompressData(data)
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

func (d *redisDriver) Remove(ctx context.Context, key string) error {
	finalKey := d.keyWithPrefix(key)

	err := d.client.Del(ctx, finalKey).Err()
	if err != nil {
		return fmt.Errorf("failed to remove key from Redis: %v", err)
	}

	return nil
}

func (d *redisDriver) RemoveByTag(ctx context.Context, tag string) error {
	keyForTags := getTagKey(tag)

	keys, err := d.client.SMembers(ctx, keyForTags).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys associated with tag: %v", err)
	}

	for _, key := range keys {
		err := d.client.Del(ctx, key).Err()
		if err != nil {
			return fmt.Errorf("failed to remove key from Redis: %v", err)
		}
	}

	err = d.client.Del(ctx, keyForTags).Err()
	if err != nil {
		return fmt.Errorf("failed to remove tag from Redis: %v", err)
	}

	return nil
}
func (d *redisDriver) Exists(ctx context.Context, key string) (bool, error) {
	finalKey := d.keyWithPrefix(key)

	cmd := d.client.Exists(ctx, finalKey)
	if err := cmd.Err(); err != nil {
		return false, fmt.Errorf("failed to check key existence in Redis: %v", err)
	}
	return cmd.Val() > 0, nil
}

func (d *redisDriver) Increment(ctx context.Context, key string) error {
	finalKey := d.keyWithPrefix(key)

	cmd := d.client.Incr(ctx, finalKey)
	if err := cmd.Err(); err != nil {
		return fmt.Errorf("failed to increment key value in Redis: %v", err)
	}
	return nil
}

func (d *redisDriver) Decrement(ctx context.Context, key string) error {
	finalKey := d.keyWithPrefix(key)

	cmd := d.client.Decr(ctx, finalKey)
	if err := cmd.Err(); err != nil {
		return fmt.Errorf("failed to decrement key value in Redis: %v", err)
	}
	return nil
}

func (d *redisDriver) GetKeysByTag(ctx context.Context, tag string) ([]string, error) {
	keyForTags := getTagKey(tag)

	cmd := d.client.SMembers(ctx, keyForTags)
	if err := cmd.Err(); err != nil {
		return nil, fmt.Errorf("failed to get keys associated with tag: %v", err)
	}
	return cmd.Val(), nil
}

func (d *redisDriver) RemoveByTags(ctx context.Context, tags []string) error {
	for _, tag := range tags {
		err := d.RemoveByTag(ctx, tag)
		if err != nil {
			return fmt.Errorf("failed to remove keys for tag: %v", err)
		}
	}

	return nil
}

func getTagKey(tag string) string {
	return fmt.Sprintf("tag:%s", tag)
}

func (d *redisDriver) keyWithPrefix(key string) string {
	return fmt.Sprintf("%s:%s", d.prefix, key)
}

func (d *redisDriver) Close() error {
	return d.client.Close()
}

func (d *redisDriver) Ping() error {
	ctx := context.Background()
	err := d.client.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to ping Redis: %v", err)
	}
	return nil
}
