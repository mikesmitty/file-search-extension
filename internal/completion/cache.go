package completion

import (
	"sync"
	"time"
)

// CacheEntry represents a cached list of completion values with expiration
type CacheEntry struct {
	Values    []string
	ExpiresAt time.Time
}

// Cache provides thread-safe TTL-based caching for completion values
type Cache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	ttl     time.Duration
}

// NewCache creates a new Cache with the specified TTL.
// If ttl is <= 0, defaults to 5 minutes.
func NewCache(ttl time.Duration) *Cache {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &Cache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
	}
}

// Get retrieves cached values for the given key.
// Returns (values, true) if found and not expired, (nil, false) otherwise.
func (c *Cache) Get(key string) ([]string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists || time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Values, true
}

// Set stores values in the cache with the configured TTL
func (c *Cache) Set(key string, values []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &CacheEntry{
		Values:    values,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
}
