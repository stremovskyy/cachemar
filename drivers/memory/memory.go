package memory

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/gob"
	"errors"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/stremovskyy/cachemar"
)

type Item struct {
	Value      []byte
	Tags       []string
	ExpiryTime time.Time
}

type memoryCacheService struct {
	mu    sync.Mutex
	items map[string]Item
}

func NewMemoryCacheService() cachemar.Cacher {
	return &memoryCacheService{
		items: make(map[string]Item),
	}
}

func (c *memoryCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration, tags []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(value); err != nil {
		return err
	}

	compressedValue, err := compressData(buf.Bytes())
	if err != nil {
		return err
	}

	c.items[key] = Item{
		Value:      compressedValue,
		Tags:       tags,
		ExpiryTime: time.Now().Add(ttl),
	}
	return nil
}

func compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)

	_, err := zw.Write(data)
	if err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (c *memoryCacheService) Get(ctx context.Context, key string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.items[key]
	if !exists || item.ExpiryTime.Before(time.Now()) {
		return cachemar.ErrNotFound
	}

	decompressedValue, err := decompressData(item.Value)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(decompressedValue)
	dec := gob.NewDecoder(buf)

	if err := dec.Decode(value); err != nil {
		return err
	}

	return nil
}

func decompressData(data []byte) ([]byte, error) {
	zr, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	decompressedData, err := io.ReadAll(zr)
	if err != nil {
		return nil, err
	}

	if err := zr.Close(); err != nil {
		return nil, err
	}

	return decompressedData, nil
}

func (c *memoryCacheService) Remove(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	return nil
}

func (c *memoryCacheService) RemoveByTag(ctx context.Context, tag string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, item := range c.items {
		for _, itemTag := range item.Tags {
			if itemTag == tag {
				delete(c.items, key)
				break
			}
		}
	}
	return nil
}

func (c *memoryCacheService) RemoveByTags(ctx context.Context, tags []string) error {
	for _, tag := range tags {
		if err := c.RemoveByTag(ctx, tag); err != nil {
			return err
		}
	}
	return nil
}

func (c *memoryCacheService) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.items[key]
	if !exists || item.ExpiryTime.Before(time.Now()) {
		return false, nil
	}
	return true, nil
}

func (m *memoryCacheService) Increment(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, found := m.items[key]
	if !found {
		return errors.New("key not found")
	}

	intValue, err := strconv.Atoi(string(item.Value))
	if err != nil {
		return errors.New("value is not an integer")
	}

	intValue++
	item.Value = []byte(strconv.Itoa(intValue))

	return nil
}

func (m *memoryCacheService) Decrement(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, found := m.items[key]
	if !found {
		return errors.New("key not found")
	}

	intValue, err := strconv.Atoi(string(item.Value))
	if err != nil {
		return errors.New("value is not an integer")
	}

	intValue--
	item.Value = []byte(strconv.Itoa(intValue))

	return nil
}

func (c *memoryCacheService) GetKeysByTag(ctx context.Context, tag string) ([]string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var activeKeys []string
	for key, item := range c.items {
		for _, itemTag := range item.Tags {
			if itemTag == tag {
				activeKeys = append(activeKeys, key)
				break
			}
		}
	}

	return activeKeys, nil
}
