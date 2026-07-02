package cache

import (
	"sync"
	"time"
)

// cacheItem represents a single cached value with its expiration time.
type cacheItem[V any] struct {
	Value     V
	ExpiresAt time.Time
}

// Cache represents a generic, thread-safe, in-memory cache with TTL support.
// K must be a comparable type so it can be used as a map key.
// V can be any type.
type Cache[K comparable, V any] struct {
	store       map[K]cacheItem[V]
	mu          sync.RWMutex
	stopCleanup chan struct{}
}

// New creates and returns a new initialized Cache instance.
func New[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		store: make(map[K]cacheItem[V]),
	}
}

// Set adds a key-value pair to the cache permanently (no expiration).
// If the key already exists, its value is overwritten.
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store[key] = cacheItem[V]{
		Value: value,
		// Zero time means it never expires.
	}
}

// SetWithTTL adds a key-value pair to the cache with a specified time-to-live.
func (c *Cache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store[key] = cacheItem[V]{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Get retrieves a value from the cache by its key.
// It returns the value and a boolean indicating whether the key was found.
// If the key exists but has expired, it is automatically removed and false is returned.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	item, ok := c.store[key]
	c.mu.RUnlock()

	if !ok {
		var zero V
		return zero, false
	}

	// Check expiration
	if !item.ExpiresAt.IsZero() && time.Now().After(item.ExpiresAt) {
		// Lock to remove expired item
		c.mu.Lock()
		defer c.mu.Unlock()

		// Double check to ensure it wasn't updated by another goroutine during lock escalation
		item, ok = c.store[key]
		if ok && !item.ExpiresAt.IsZero() && time.Now().After(item.ExpiresAt) {
			delete(c.store, key)
		}

		var zero V
		return zero, false
	}

	return item.Value, true
}

// Has checks if a key exists in the cache and has not expired.
func (c *Cache[K, V]) Has(key K) bool {
	_, ok := c.Get(key)
	return ok
}

// Delete removes a key-value pair from the cache.
// If the key does not exist, it performs a no-op.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, key)
}

// Size returns the current number of items in the underlying map.
// Note: This includes expired items that have not yet been swept by the cleanup worker.
func (c *Cache[K, V]) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.store)
}

// Clear removes all items from the cache, resetting it to an empty state.
func (c *Cache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	clear(c.store)
}

// StartCleanup starts a background goroutine that removes expired items at the specified interval.
// If a cleanup worker is already running, this function does nothing.
func (c *Cache[K, V]) StartCleanup(interval time.Duration) {
	c.mu.Lock()
	if c.stopCleanup != nil {
		c.mu.Unlock()
		return
	}
	stopChan := make(chan struct{})
	c.stopCleanup = stopChan
	c.mu.Unlock()

	go func(stop <-chan struct{}) {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.DeleteExpired()
			case <-stop:
				return
			}
		}
	}(stopChan)
}

// StopCleanup gracefully signals the background cleanup worker to stop.
func (c *Cache[K, V]) StopCleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.stopCleanup != nil {
		close(c.stopCleanup)
		c.stopCleanup = nil
	}
}

// DeleteExpired performs a sweep of the cache to remove all expired items.
func (c *Cache[K, V]) DeleteExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for k, v := range c.store {
		if !v.ExpiresAt.IsZero() && now.After(v.ExpiresAt) {
			delete(c.store, k)
		}
	}
}
