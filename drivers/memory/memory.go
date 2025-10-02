package memory

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"sync"
	"time"

	"github.com/stremovskyy/cachemar"
)

type Item struct {
	Key        string
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

// removeEntry disconnects an item from the list and map bookkeeping
func (d *memory) removeEntry(item *Item) {
	if item == nil || item == d.head || item == d.tail {
		return
	}
	if item.prev == nil || item.next == nil {
		return
	}

	d.removeItem(item)
	delete(d.items, item.Key)
	if d.size > 0 {
		d.size--
	}
	item.prev = nil
	item.next = nil
}

// moveToHead moves an existing item to head
func (d *memory) moveToHead(item *Item) {
	d.removeItem(item)
	d.addToHead(item)
}

func (d *memory) evictLRU() {
	if d.config.MaxSize <= 0 || d.size <= d.config.MaxSize {
		return
	}

	now := time.Now()
	for d.size > d.config.MaxSize {
		candidate := d.tail.prev
		if candidate == d.head {
			break
		}

		if d.isExpired(candidate, now) {
			d.removeEntry(candidate)
			continue
		}

		d.removeEntry(candidate)
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

	encodedValue := buf.Bytes()

	expiry := time.Time{}
	if ttl > 0 {
		expiry = time.Now().Add(ttl)
	}

	if existingItem, exists := d.items[key]; exists {
		existingItem.Value = encodedValue
		existingItem.Tags = tags
		existingItem.ExpiryTime = expiry
		existingItem.Key = key
		d.moveToHead(existingItem)
	} else {
		newItem := &Item{
			Key:        key,
			Value:      encodedValue,
			Tags:       tags,
			ExpiryTime: expiry,
		}

		d.items[key] = newItem
		d.addToHead(newItem)
		d.size++

		d.evictLRU()
	}

	return nil
}

func (d *memory) Get(ctx context.Context, key string, value interface{}) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	item, exists := d.items[key]
	if !exists {
		return cachemar.ErrNotFound
	}

	now := time.Now()
	if d.isExpired(item, now) {
		d.removeEntry(item)
		return cachemar.ErrNotFound
	}

	d.moveToHead(item)

	buf := bytes.NewBuffer(item.Value)
	dec := gob.NewDecoder(buf)

	if err := dec.Decode(value); err != nil {
		return err
	}

	return nil
}

func (d *memory) Remove(ctx context.Context, key string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if item, exists := d.items[key]; exists {
		d.removeEntry(item)
	}

	return nil
}

func (d *memory) Exists(ctx context.Context, key string) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	item, exists := d.items[key]
	if !exists {
		return false, nil
	}

	now := time.Now()
	if d.isExpired(item, now) {
		d.removeEntry(item)
		return false, nil
	}
	return true, nil
}

func (d *memory) Increment(ctx context.Context, key string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	item, exists := d.items[key]
	if !exists {
		return errors.New("key not found or expired")
	}

	now := time.Now()
	if d.isExpired(item, now) {
		d.removeEntry(item)
		return errors.New("key not found or expired")
	}

	d.moveToHead(item)

	var intValue int
	buf := bytes.NewBuffer(item.Value)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&intValue); err != nil {
		return errors.New("value is not an integer")
	}

	// Increment the value
	intValue++

	// Re-encode the value
	var newBuf bytes.Buffer
	enc := gob.NewEncoder(&newBuf)
	if err := enc.Encode(intValue); err != nil {
		return err
	}

	// Update the item in the cache
	item.Value = newBuf.Bytes()

	return nil
}

func (d *memory) Decrement(ctx context.Context, key string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	item, exists := d.items[key]
	if !exists {
		return errors.New("key not found or expired")
	}

	now := time.Now()
	if d.isExpired(item, now) {
		d.removeEntry(item)
		return errors.New("key not found or expired")
	}

	d.moveToHead(item)

	var intValue int
	buf := bytes.NewBuffer(item.Value)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&intValue); err != nil {
		return errors.New("value is not an integer")
	}

	// Decrement the value
	intValue--

	// Re-encode the value
	var newBuf bytes.Buffer
	enc := gob.NewEncoder(&newBuf)
	if err := enc.Encode(intValue); err != nil {
		return err
	}

	// Update the item in the cache
	item.Value = newBuf.Bytes()

	return nil
}

func (d *memory) GetKeysByTag(ctx context.Context, tag string) ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var activeKeys []string
	now := time.Now()
	for key, item := range d.items {
		if d.isExpired(item, now) {
			d.removeEntry(item)
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

func (d *memory) RemoveByTag(ctx context.Context, tag string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	for _, item := range d.items {
		if d.isExpired(item, now) {
			d.removeEntry(item)
			continue
		}
		for _, itemTag := range item.Tags {
			if itemTag == tag {
				d.removeEntry(item)
				break
			}
		}
	}

	return nil
}

func (d *memory) RemoveByTags(ctx context.Context, tags []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	for _, item := range d.items {
		if d.isExpired(item, now) {
			d.removeEntry(item)
			continue
		}
		removed := false
		for _, tag := range tags {
			if removed {
				break
			}
			for _, itemTag := range item.Tags {
				if itemTag == tag {
					d.removeEntry(item)
					removed = true
					break
				}
			}
		}
	}

	return nil
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

func (d *memory) isExpired(item *Item, now time.Time) bool {
	if item == nil {
		return false
	}
	return !item.ExpiryTime.IsZero() && item.ExpiryTime.Before(now)
}
