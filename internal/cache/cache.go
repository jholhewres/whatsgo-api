package cache

import (
	"sync"
	"time"
)

// Entry holds a cached value with an expiration time.
type Entry[T any] struct {
	Value     T
	ExpiresAt time.Time
}

// Cache is a TTL-based in-memory cache using sync.RWMutex.
type Cache[K comparable, T any] struct {
	mu      sync.RWMutex
	entries map[K]*Entry[T]
	ttl     time.Duration
}

// New creates a cache with the given default TTL.
func New[K comparable, T any](ttl time.Duration) *Cache[K, T] {
	return &Cache[K, T]{
		entries: make(map[K]*Entry[T]),
		ttl:     ttl,
	}
}

// Get returns the cached value and true if it exists and has not expired.
func (c *Cache[K, T]) Get(key K) (T, bool) {
	c.mu.RLock()
	e, ok := c.entries[key]
	if !ok || time.Now().After(e.ExpiresAt) {
		c.mu.RUnlock()
		var zero T
		return zero, false
	}
	val := e.Value
	c.mu.RUnlock()
	return val, true
}

// Set stores a value with the cache's default TTL.
func (c *Cache[K, T]) Set(key K, value T) {
	c.mu.Lock()
	c.entries[key] = &Entry[T]{Value: value, ExpiresAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}

// SetWithTTL stores a value with a custom TTL.
func (c *Cache[K, T]) SetWithTTL(key K, value T, ttl time.Duration) {
	c.mu.Lock()
	c.entries[key] = &Entry[T]{Value: value, ExpiresAt: time.Now().Add(ttl)}
	c.mu.Unlock()
}

// Flush removes all entries from the cache.
func (c *Cache[K, T]) Flush() {
	c.mu.Lock()
	c.entries = make(map[K]*Entry[T])
	c.mu.Unlock()
}

// Invalidate removes a single entry.
func (c *Cache[K, T]) Invalidate(key K) {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}

// Cleanup removes all expired entries.
func (c *Cache[K, T]) Cleanup() {
	c.mu.Lock()
	now := time.Now()
	for k, e := range c.entries {
		if now.After(e.ExpiresAt) {
			delete(c.entries, k)
		}
	}
	c.mu.Unlock()
}
