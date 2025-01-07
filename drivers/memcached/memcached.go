package memcached

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"

	"github.com/stremovskyy/cachemar"
)

type memcached struct {
	client *memcache.Client
	prefix string
}

type Options struct {
	Servers []string
	Prefix  string
}

func New(options *Options) cachemar.Cacher {
	client := memcache.New(options.Servers...)

	return &memcached{
		client: client,
		prefix: options.Prefix,
	}
}

func (d *memcached) Set(ctx context.Context, key string, value interface{}, ttl time.Duration, tags []string) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %v", err)
	}

	finalKey := d.keyWithPrefix(key)
	item := &memcache.Item{
		Key:        finalKey,
		Value:      data,
		Expiration: int32(ttl.Seconds()),
	}

	err = d.client.Set(item)
	if err != nil {
		return fmt.Errorf("failed to set key-value pair in Memcached: %v", err)
	}

	if len(tags) > 0 {
		for _, tag := range tags {
			tagKey := d.getTagKey(tag)
			tagValueItem, err := d.client.Get(tagKey)
			if err != nil && err != memcache.ErrCacheMiss {
				return err
			}
			tagValue := make([]string, 0)
			if err != memcache.ErrCacheMiss {
				if err := json.Unmarshal(tagValueItem.Value, &tagValue); err != nil {
					return err
				}
			}
			tagValue = append(tagValue, key)
			tagValueBytes, err := json.Marshal(tagValue)
			if err != nil {
				return err
			}
			err = d.client.Set(&memcache.Item{Key: tagKey, Value: tagValueBytes})
			if err != nil {
				return fmt.Errorf("failed to set tag key-value pair in Memcached: %v", err)
			}
		}

	}

	return nil
}

func (d *memcached) Get(ctx context.Context, key string, value interface{}) error {
	finalKey := d.keyWithPrefix(key)

	item, err := d.client.Get(finalKey)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return fmt.Errorf("key not found: %s", finalKey)
		}
		return fmt.Errorf("failed to get value from Memcached: %v", err)
	}

	err = json.Unmarshal(item.Value, value)
	if err != nil {
		return fmt.Errorf("failed to deserialize value: %v", err)
	}

	return nil
}

func (d *memcached) Remove(ctx context.Context, key string) error {
	finalKey := d.keyWithPrefix(key)

	err := d.client.Delete(finalKey)
	if err != nil {
		return fmt.Errorf("failed to remove key from Memcached: %v", err)
	}

	return nil
}

func (d *memcached) RemoveByTag(ctx context.Context, tag string) error {
	keyForTags := getTagKey(tag)

	item, err := d.client.Get(keyForTags)
	if err != nil {
		return fmt.Errorf("failed to get keys associated with tag: %v", err)
	}

	keys := strings.Split(string(item.Value), ",")
	for _, key := range keys {
		err := d.client.Delete(key)
		if err != nil {
			return fmt.Errorf("failed to remove key from Memcached: %v", err)
		}
	}

	return nil
}

func (d *memcached) RemoveByTags(ctx context.Context, tags []string) error {
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

func (d *memcached) keyWithPrefix(key string) string {
	return fmt.Sprintf("%s:%s", d.prefix, key)
}

func (d *memcached) Exists(ctx context.Context, key string) (bool, error) {
	finalKey := d.keyWithPrefix(key)
	_, err := d.client.Get(finalKey)

	if err == memcache.ErrCacheMiss {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to check key existence in Memcached: %v", err)
	}

	return true, nil
}

func (d *memcached) Increment(ctx context.Context, key string) error {
	finalKey := d.keyWithPrefix(key)

	_, err := d.client.Increment(finalKey, 1)
	if err != nil {
		return fmt.Errorf("failed to increment key value in Memcached: %v", err)
	}

	return nil
}

func (d *memcached) Decrement(ctx context.Context, key string) error {
	finalKey := d.keyWithPrefix(key)

	_, err := d.client.Decrement(finalKey, 1)
	if err != nil {
		return fmt.Errorf("failed to decrement key value in Memcached: %v", err)
	}

	return nil
}
func (d *memcached) GetKeysByTag(ctx context.Context, tag string) ([]string, error) {
	tagKey := d.getTagKey(tag)
	item, err := d.client.Get(tagKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get keys associated with tag: %v", err)
	}

	var keys []string
	if err := json.Unmarshal(item.Value, &keys); err != nil {
		return nil, err
	}

	return keys, nil
}

func (d *memcached) getTagKey(tag string) string {
	return fmt.Sprintf("tag:%s", tag)
}

func (d *memcached) Close() error {
	return d.client.Close()
}

func (d *memcached) Ping() error {
	err := d.client.Set(&memcache.Item{Key: "selfcheck", Value: []byte("selfcheck")})
	if err != nil {
		return err
	}
	_, err = d.client.Get("selfcheck")
	if err != nil {
		return fmt.Errorf("failed to get value from Memcached: %v", err)
	}

	return nil
}
