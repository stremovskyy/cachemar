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
	prev       *Item
	next       *Item
}

type Config struct {
	MaxSize int
}

type memory struct {
	mu     sync.Mutex
	items  map[string]*Item
	config Config
	head   *Item
	tail   *Item
	size   int
}

func New() cachemar.Cacher {
	return NewWithConfig(Config{MaxSize: 0})
}

func NewWithConfig(config Config) cachemar.Cacher {
	m := &memory{
		items:  make(map[string]*Item),
		config: config,
	}

	m.head = &Item{}
	m.tail = &Item{}
	m.head.next = m.tail
	m.tail.prev = m.head

	return m
}

func (d *memory) addToHead(item *Item) {
	item.prev = d.head
	item.next = d.head.next
	d.head.next.prev = item
	d.head.next = item
}

// removeItem removes an existing item from the linked list
func (d *memory) removeItem(item *Item) {
	item.prev.next = item.next
	item.next.prev = item.prev
}

// moveToHead moves an existing item to head
func (d *memory) moveToHead(item *Item) {
	d.removeItem(item)
	d.addToHead(item)
}

// removeTail removes the last item before tail
func (d *memory) removeTail() *Item {
	lastItem := d.tail.prev
	d.removeItem(lastItem)
	return lastItem
}

func (d *memory) evictLRU() {
	if d.config.MaxSize <= 0 || d.size <= d.config.MaxSize {
		return
	}

	d.cleanupExpired()

	for d.size > d.config.MaxSize {
		lru := d.removeTail()
		if lru != d.tail && lru != d.head {
			for key, item := range d.items {
				if item == lru {
					delete(d.items, key)
					d.size--
					break
				}
			}
		} else {
			break
		}
	}
}

func (d *memory) cleanupExpired() {
	now := time.Now()
	for key, item := range d.items {
		if item.ExpiryTime.Before(now) {
			d.removeItem(item)
			delete(d.items, key)
			d.size--
		}
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

	if existingItem, exists := d.items[key]; exists {
		existingItem.Value = compressedValue
		existingItem.Tags = tags
		existingItem.ExpiryTime = time.Now().Add(ttl)
		d.moveToHead(existingItem)
	} else {
		newItem := &Item{
			Value:      compressedValue,
			Tags:       tags,
			ExpiryTime: time.Now().Add(ttl),
		}

		d.items[key] = newItem
		d.addToHead(newItem)
		d.size++

		d.evictLRU()
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

	d.moveToHead(item)

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

	if item, exists := d.items[key]; exists {
		d.removeItem(item)
		delete(d.items, key)
		d.size--
	}

	return nil
}

func (d *memory) RemoveByTag(ctx context.Context, tag string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	keysToRemove := make([]string, 0)
	for key, item := range d.items {
		if item.ExpiryTime.Before(time.Now()) {
			keysToRemove = append(keysToRemove, key)
			continue
		}
		for _, itemTag := range item.Tags {
			if itemTag == tag {
				keysToRemove = append(keysToRemove, key)
				break
			}
		}
	}

	for _, key := range keysToRemove {
		if item, exists := d.items[key]; exists {
			d.removeItem(item)
			delete(d.items, key)
			d.size--
		}
	}

	return nil
}

func (d *memory) RemoveByTags(ctx context.Context, tags []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	keysToRemove := make(map[string]bool)

	for _, tag := range tags {
		for key, item := range d.items {
			for _, itemTag := range item.Tags {
				if itemTag == tag {
					keysToRemove[key] = true
					break
				}
			}
		}
	}

	for key := range keysToRemove {
		if item, exists := d.items[key]; exists {
			d.removeItem(item)
			delete(d.items, key)
			d.size--
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

	d.moveToHead(item)

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

	return nil
}

func (d *memory) Decrement(ctx context.Context, key string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	item, exists := d.items[key]
	if !exists || item.ExpiryTime.Before(time.Now()) {
		return errors.New("key not found or expired")
	}

	d.moveToHead(item)

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

	d.items = make(map[string]*Item)
	d.size = 0

	d.head.next = d.tail
	d.tail.prev = d.head

	return nil
}

func (d *memory) Ping() error {
	return nil
}
