package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stremovskyy/cachemar"
	"github.com/stremovskyy/cachemar/drivers/memory"
)

// MockCacher is a mock implementation of the Cacher interface for testing
type MockCacher struct {
	pingError error
	getError  error
	setError  error
	data      map[string]interface{}
}

func NewMockCacher() *MockCacher {
	return &MockCacher{
		data: make(map[string]interface{}),
	}
}

func (m *MockCacher) Set(ctx context.Context, key string, value interface{}, ttl time.Duration, tags []string) error {
	if m.setError != nil {
		return m.setError
	}
	m.data[key] = value
	return nil
}

func (m *MockCacher) Get(ctx context.Context, key string, value interface{}) error {
	if m.getError != nil {
		return m.getError
	}
	if _, ok := m.data[key]; ok {
		// In a real implementation, we would unmarshal the value into the provided variable
		// For this mock, we'll just pretend it worked
		return nil
	}
	return errors.New("key not found")
}

func (m *MockCacher) Remove(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *MockCacher) RemoveByTag(ctx context.Context, tag string) error {
	return nil
}

func (m *MockCacher) RemoveByTags(ctx context.Context, tags []string) error {
	return nil
}

func (m *MockCacher) Exists(ctx context.Context, key string) (bool, error) {
	_, ok := m.data[key]
	return ok, nil
}

func (m *MockCacher) Increment(ctx context.Context, key string) error {
	return nil
}

func (m *MockCacher) Decrement(ctx context.Context, key string) error {
	return nil
}

func (m *MockCacher) GetKeysByTag(ctx context.Context, tag string) ([]string, error) {
	return []string{}, nil
}

func (m *MockCacher) Ping() error {
	return m.pingError
}

func (m *MockCacher) Close() error {
	return nil
}

func (m *MockCacher) SetPingError(err error) {
	m.pingError = err
}

func (m *MockCacher) SetGetError(err error) {
	m.getError = err
}

func (m *MockCacher) SetSetError(err error) {
	m.setError = err
}

func TestCircuitBreaker(t *testing.T) {
	// Create mock cachers
	primaryCacher := NewMockCacher()
	fallbackCacher := NewMockCacher()

	// Create a manager with circuit breaker
	manager := cachemar.NewWithOptions(
		cachemar.WithDebug(),
		cachemar.WithCircuitBreaker("primary", []string{"fallback"}, 100*time.Millisecond),
	)

	// Register the cachers
	manager.Register("primary", primaryCacher)
	manager.Register("fallback", fallbackCacher)

	// Test 1: Primary cacher is working
	ctx := context.Background()
	err := manager.Set(ctx, "key1", "value1", time.Minute, nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	var value string
	err = manager.Get(ctx, "key1", &value)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test 2: Primary cacher fails, fallback to secondary
	primaryCacher.SetPingError(errors.New("primary cacher is down"))

	// Wait for the circuit to open
	time.Sleep(10 * time.Millisecond)

	// Set a value using the fallback cacher
	err = manager.Set(ctx, "key2", "value2", time.Minute, nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Get the value from the fallback cacher
	err = manager.Get(ctx, "key2", &value)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test 3: Primary cacher comes back online
	primaryCacher.SetPingError(nil)

	// Wait for the circuit to close
	time.Sleep(200 * time.Millisecond)

	// Set a value using the primary cacher
	err = manager.Set(ctx, "key3", "value3", time.Minute, nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Get the value from the primary cacher
	err = manager.Get(ctx, "key3", &value)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCircuitBreakerWithRealCachers(t *testing.T) {
	// Create a manager with circuit breaker
	manager := cachemar.NewWithOptions(
		cachemar.WithDebug(),
		cachemar.WithCircuitBreaker(string(cachemar.RedisCacherName), []string{string(cachemar.MemoryCacherName)}, 100*time.Millisecond),
	)

	// Register the memory cacher
	memoryCacher := memory.New()
	manager.Register(string(cachemar.MemoryCacherName), memoryCacher)

	// Test: Set and get a value using the memory cacher
	ctx := context.Background()
	err := manager.Set(ctx, "key1", "value1", time.Minute, nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	var value string
	err = manager.Get(ctx, "key1", &value)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if value != "value1" {
		t.Errorf("Expected value1, got %s", value)
	}
}
