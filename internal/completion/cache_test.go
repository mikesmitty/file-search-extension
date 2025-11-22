package completion

import (
	"sync"
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	tests := []struct {
		name        string
		ttl         time.Duration
		expectedTTL time.Duration
	}{
		{
			name:        "default TTL when zero",
			ttl:         0,
			expectedTTL: 5 * time.Minute,
		},
		{
			name:        "default TTL when negative",
			ttl:         -1 * time.Second,
			expectedTTL: 5 * time.Minute,
		},
		{
			name:        "custom TTL",
			ttl:         10 * time.Second,
			expectedTTL: 10 * time.Second,
		},
		{
			name:        "large custom TTL",
			ttl:         1 * time.Hour,
			expectedTTL: 1 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewCache(tt.ttl)
			if cache == nil {
				t.Fatal("NewCache returned nil")
			}
			if cache.ttl != tt.expectedTTL {
				t.Errorf("Expected TTL %v, got %v", tt.expectedTTL, cache.ttl)
			}
			if cache.entries == nil {
				t.Error("entries map not initialized")
			}
			if len(cache.entries) != 0 {
				t.Error("entries map should be empty initially")
			}
		})
	}
}

func TestCacheGet(t *testing.T) {
	t.Run("miss on non-existent key", func(t *testing.T) {
		cache := NewCache(5 * time.Minute)
		values, ok := cache.Get("non-existent")
		if ok {
			t.Error("Expected cache miss, got hit")
		}
		if values != nil {
			t.Errorf("Expected nil values on miss, got %v", values)
		}
	})

	t.Run("hit on existing non-expired entry", func(t *testing.T) {
		cache := NewCache(5 * time.Minute)
		expected := []string{"value1", "value2"}
		cache.Set("key1", expected)

		values, ok := cache.Get("key1")
		if !ok {
			t.Error("Expected cache hit, got miss")
		}
		if len(values) != len(expected) {
			t.Errorf("Expected %d values, got %d", len(expected), len(values))
		}
		for i, v := range values {
			if v != expected[i] {
				t.Errorf("Expected value[%d] = %s, got %s", i, expected[i], v)
			}
		}
	})

	t.Run("miss on expired entry", func(t *testing.T) {
		cache := NewCache(100 * time.Millisecond)
		cache.Set("key1", []string{"value1"})

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		values, ok := cache.Get("key1")
		if ok {
			t.Error("Expected cache miss for expired entry, got hit")
		}
		if values != nil {
			t.Errorf("Expected nil values for expired entry, got %v", values)
		}
	})

	t.Run("entry at exact expiration time", func(t *testing.T) {
		cache := NewCache(50 * time.Millisecond)
		cache.Set("key1", []string{"value1"})

		// Wait slightly less than TTL
		time.Sleep(40 * time.Millisecond)
		values, ok := cache.Get("key1")
		if !ok {
			t.Error("Expected cache hit before expiration, got miss")
		}
		if values == nil {
			t.Error("Expected non-nil values before expiration")
		}
	})
}

func TestCacheSet(t *testing.T) {
	t.Run("store values with TTL", func(t *testing.T) {
		cache := NewCache(5 * time.Minute)
		values := []string{"a", "b", "c"}
		cache.Set("key1", values)

		entry, exists := cache.entries["key1"]
		if !exists {
			t.Fatal("Entry not stored in cache")
		}
		if len(entry.Values) != len(values) {
			t.Errorf("Expected %d values stored, got %d", len(values), len(entry.Values))
		}
		if entry.ExpiresAt.Before(time.Now()) {
			t.Error("Entry already expired at creation")
		}
	})

	t.Run("overwrite existing key", func(t *testing.T) {
		cache := NewCache(5 * time.Minute)
		cache.Set("key1", []string{"old1", "old2"})
		cache.Set("key1", []string{"new1", "new2", "new3"})

		values, ok := cache.Get("key1")
		if !ok {
			t.Fatal("Expected cache hit after overwrite")
		}
		if len(values) != 3 {
			t.Errorf("Expected 3 values after overwrite, got %d", len(values))
		}
		if values[0] != "new1" {
			t.Errorf("Expected first value 'new1', got '%s'", values[0])
		}
	})

	t.Run("expiration calculation", func(t *testing.T) {
		ttl := 10 * time.Second
		cache := NewCache(ttl)
		before := time.Now()
		cache.Set("key1", []string{"value1"})
		after := time.Now()

		entry := cache.entries["key1"]
		expectedMin := before.Add(ttl)
		expectedMax := after.Add(ttl)

		if entry.ExpiresAt.Before(expectedMin) || entry.ExpiresAt.After(expectedMax) {
			t.Errorf("Expiration time %v not in expected range %v - %v",
				entry.ExpiresAt, expectedMin, expectedMax)
		}
	})

	t.Run("empty values slice", func(t *testing.T) {
		cache := NewCache(5 * time.Minute)
		cache.Set("key1", []string{})

		values, ok := cache.Get("key1")
		if !ok {
			t.Error("Expected cache hit for empty slice")
		}
		if len(values) != 0 {
			t.Errorf("Expected empty slice, got %d values", len(values))
		}
	})
}

func TestCacheClear(t *testing.T) {
	t.Run("clears all entries", func(t *testing.T) {
		cache := NewCache(5 * time.Minute)
		cache.Set("key1", []string{"value1"})
		cache.Set("key2", []string{"value2"})
		cache.Set("key3", []string{"value3"})

		if len(cache.entries) != 3 {
			t.Fatalf("Expected 3 entries before clear, got %d", len(cache.entries))
		}

		cache.Clear()

		if len(cache.entries) != 0 {
			t.Errorf("Expected 0 entries after clear, got %d", len(cache.entries))
		}
	})

	t.Run("can add new entries after clear", func(t *testing.T) {
		cache := NewCache(5 * time.Minute)
		cache.Set("key1", []string{"value1"})
		cache.Clear()
		cache.Set("key2", []string{"value2"})

		values, ok := cache.Get("key2")
		if !ok {
			t.Error("Expected cache hit after clear and re-add")
		}
		if len(values) != 1 || values[0] != "value2" {
			t.Errorf("Expected [value2], got %v", values)
		}
	})

	t.Run("clear empty cache", func(t *testing.T) {
		cache := NewCache(5 * time.Minute)
		cache.Clear() // Should not panic
		if len(cache.entries) != 0 {
			t.Error("Expected empty cache after clearing empty cache")
		}
	})
}

func TestCacheThreadSafety(t *testing.T) {
	t.Run("concurrent reads", func(t *testing.T) {
		cache := NewCache(5 * time.Minute)
		cache.Set("key1", []string{"value1", "value2"})

		var wg sync.WaitGroup
		readers := 100

		for i := 0; i < readers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				values, ok := cache.Get("key1")
				if !ok {
					t.Error("Expected cache hit")
				}
				if len(values) != 2 {
					t.Errorf("Expected 2 values, got %d", len(values))
				}
			}()
		}

		wg.Wait()
	})

	t.Run("concurrent writes", func(t *testing.T) {
		cache := NewCache(5 * time.Minute)
		var wg sync.WaitGroup
		writers := 100

		for i := 0; i < writers; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				key := "key1"
				values := []string{string(rune('a' + id))}
				cache.Set(key, values)
			}(i)
		}

		wg.Wait()

		// Should have exactly one entry (last writer wins)
		if len(cache.entries) != 1 {
			t.Errorf("Expected 1 entry after concurrent writes, got %d", len(cache.entries))
		}
	})

	t.Run("concurrent reads and writes", func(t *testing.T) {
		cache := NewCache(5 * time.Minute)
		cache.Set("key1", []string{"initial"})

		var wg sync.WaitGroup
		operations := 200

		for i := 0; i < operations; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				if id%2 == 0 {
					// Reader
					cache.Get("key1")
				} else {
					// Writer
					cache.Set("key1", []string{string(rune('a' + id))})
				}
			}(i)
		}

		wg.Wait() // Should not deadlock or panic
	})

	t.Run("concurrent clear and operations", func(t *testing.T) {
		cache := NewCache(5 * time.Minute)
		var wg sync.WaitGroup
		operations := 100

		for i := 0; i < operations; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				switch id % 3 {
				case 0:
					cache.Get("key1")
				case 1:
					cache.Set("key1", []string{"value"})
				case 2:
					cache.Clear()
				}
			}(i)
		}

		wg.Wait() // Should not deadlock or panic
	})
}

func TestCacheExpiration(t *testing.T) {
	t.Run("multiple entries with different expiration times", func(t *testing.T) {
		cache := NewCache(100 * time.Millisecond)

		// Add first entry
		cache.Set("key1", []string{"value1"})
		time.Sleep(50 * time.Millisecond)

		// Add second entry (will expire later)
		cache.Set("key2", []string{"value2"})

		// Wait for first entry to expire
		time.Sleep(60 * time.Millisecond)

		// key1 should be expired, key2 should still be valid
		_, ok1 := cache.Get("key1")
		if ok1 {
			t.Error("Expected key1 to be expired")
		}

		values2, ok2 := cache.Get("key2")
		if !ok2 {
			t.Error("Expected key2 to still be valid")
		}
		if len(values2) != 1 || values2[0] != "value2" {
			t.Errorf("Expected [value2], got %v", values2)
		}
	})

	t.Run("entry becomes valid again after re-set", func(t *testing.T) {
		cache := NewCache(100 * time.Millisecond)
		cache.Set("key1", []string{"value1"})

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		// Verify expired
		_, ok := cache.Get("key1")
		if ok {
			t.Error("Expected key1 to be expired")
		}

		// Re-set the key
		cache.Set("key1", []string{"value2"})

		// Should be valid again
		values, ok := cache.Get("key1")
		if !ok {
			t.Error("Expected key1 to be valid after re-set")
		}
		if len(values) != 1 || values[0] != "value2" {
			t.Errorf("Expected [value2], got %v", values)
		}
	})
}
