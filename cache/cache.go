// Package cache provides a simple, generic, in-memory key-value store for the GoKit-Lite toolkit.
package cache

// Cache represents a generic, lightweight, in-memory cache.
// It is not currently thread-safe and does not support TTLs.
// K must be a comparable type so it can be used as a map key.
// V can be any type.
type Cache[K comparable, V any] struct {
	store map[K]V
}

// New creates and returns a new initialized Cache instance.
func New[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		store: make(map[K]V),
	}
}

// Set adds a key-value pair to the cache.
// If the key already exists, its value is overwritten.
func (c *Cache[K, V]) Set(key K, value V) {
	c.store[key] = value
}

// Get retrieves a value from the cache by its key.
// It returns the value and a boolean indicating whether the key was found.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	val, ok := c.store[key]
	return val, ok
}

// Delete removes a key-value pair from the cache.
// If the key does not exist, it performs a no-op.
func (c *Cache[K, V]) Delete(key K) {
	delete(c.store, key)
}

// Size returns the current number of items in the cache.
func (c *Cache[K, V]) Size() int {
	return len(c.store)
}

// Clear removes all items from the cache, resetting it to an empty state.
func (c *Cache[K, V]) Clear() {
	clear(c.store)
}
