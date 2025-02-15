package memory

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/gob"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/stremovskyy/cachemar"
)

type Item struct {
	Value      []byte
	Tags       []string
	ExpiryTime time.Time
}

type memory struct {
	mu    sync.Mutex
	items map[string]Item
}

func New() cachemar.Cacher {
	return &memory{
		items: make(map[string]Item),
	}
}

func uniqueTags(tags []string) []string {
	tagSet := make(map[string]struct{})
	for _, tag := range tags {
		tagSet[tag] = struct{}{}
	}

	unique := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		unique = append(unique, tag)
	}
	return unique
}

func (d *memory) Set(ctx context.Context, key string, value interface{}, ttl time.Duration, tags []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	tags = uniqueTags(tags)
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(value); err != nil {
		return err
	}

	compressedValue, err := compressData(buf.Bytes())
	if err != nil {
		return err
	}

	d.items[key] = Item{
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

func (d *memory) Get(ctx context.Context, key string, value interface{}) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	item, exists := d.items[key]
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

func (d *memory) Remove(ctx context.Context, key string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.items, key)
	return nil
}

func (d *memory) RemoveByTag(ctx context.Context, tag string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for key, item := range d.items {
		if item.ExpiryTime.Before(time.Now()) {
			delete(d.items, key)
			continue
		}
		for _, itemTag := range item.Tags {
			if itemTag == tag {
				delete(d.items, key)
				break
			}
		}
	}
	return nil
}

func (d *memory) RemoveByTags(ctx context.Context, tags []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, tag := range tags {
		for key, item := range d.items {
			for _, itemTag := range item.Tags {
				if itemTag == tag {
					delete(d.items, key)
					break
				}
			}
		}
	}
	return nil
}

func (d *memory) Exists(ctx context.Context, key string) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	item, exists := d.items[key]
	if !exists || item.ExpiryTime.Before(time.Now()) {
		return false, nil
	}
	return true, nil
}

func (d *memory) Increment(ctx context.Context, key string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	item, exists := d.items[key]
	if !exists || item.ExpiryTime.Before(time.Now()) {
		return errors.New("key not found or expired")
	}

	// Decompress the value
	decompressedValue, err := decompressData(item.Value)
	if err != nil {
		return err
	}

	// Decode the value into an integer
	var intValue int
	buf := bytes.NewBuffer(decompressedValue)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&intValue); err != nil {
		return errors.New("value is not an integer")
	}

	// Increment the value
	intValue++

	// Re-encode and compress the value
	var newBuf bytes.Buffer
	enc := gob.NewEncoder(&newBuf)
	if err := enc.Encode(intValue); err != nil {
		return err
	}

	compressedValue, err := compressData(newBuf.Bytes())
	if err != nil {
		return err
	}

	// Update the item in the cache
	item.Value = compressedValue
	d.items[key] = item

	return nil
}

func (d *memory) Decrement(ctx context.Context, key string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	item, exists := d.items[key]
	if !exists || item.ExpiryTime.Before(time.Now()) {
		return errors.New("key not found or expired")
	}

	// Decompress the value
	decompressedValue, err := decompressData(item.Value)
	if err != nil {
		return err
	}

	// Decode the value into an integer
	var intValue int
	buf := bytes.NewBuffer(decompressedValue)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&intValue); err != nil {
		return errors.New("value is not an integer")
	}

	// Decrement the value
	intValue--

	// Re-encode and compress the value
	var newBuf bytes.Buffer
	enc := gob.NewEncoder(&newBuf)
	if err := enc.Encode(intValue); err != nil {
		return err
	}

	compressedValue, err := compressData(newBuf.Bytes())
	if err != nil {
		return err
	}

	// Update the item in the cache
	item.Value = compressedValue
	d.items[key] = item

	return nil
}

func (d *memory) GetKeysByTag(ctx context.Context, tag string) ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var activeKeys []string
	for key, item := range d.items {
		if item.ExpiryTime.Before(time.Now()) {
			continue
		}
		for _, itemTag := range item.Tags {
			if itemTag == tag {
				activeKeys = append(activeKeys, key)
				break
			}
		}
	}
	return activeKeys, nil
}

func (d *memory) Close() error {
	return nil
}

func (d *memory) Flush() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.items = make(map[string]Item)
	return nil
}

func (d *memory) Ping() error {
	return nil
}
