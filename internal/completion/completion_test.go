package completion

import (
	"testing"
	"time"
)

func TestNewCompleter(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		enabled  bool
		cacheTTL time.Duration
	}{
		{
			name:     "with API key and enabled",
			apiKey:   "test-api-key",
			enabled:  true,
			cacheTTL: 5 * time.Minute,
		},
		{
			name:     "with API key but disabled",
			apiKey:   "test-api-key",
			enabled:  false,
			cacheTTL: 5 * time.Minute,
		},
		{
			name:     "without API key",
			apiKey:   "",
			enabled:  true,
			cacheTTL: 5 * time.Minute,
		},
		{
			name:     "custom cache TTL",
			apiKey:   "test-api-key",
			enabled:  true,
			cacheTTL: 10 * time.Minute,
		},
		{
			name:     "zero TTL (should use default)",
			apiKey:   "test-api-key",
			enabled:  true,
			cacheTTL: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completer := NewCompleter(tt.apiKey, tt.enabled, tt.cacheTTL)

			if completer == nil {
				t.Fatal("NewCompleter returned nil")
			}

			if completer.apiKey != tt.apiKey {
				t.Errorf("Expected apiKey %q, got %q", tt.apiKey, completer.apiKey)
			}

			if completer.enabled != tt.enabled {
				t.Errorf("Expected enabled %v, got %v", tt.enabled, completer.enabled)
			}

			if completer.cache == nil {
				t.Error("Cache not initialized")
			}

			expectedTTL := tt.cacheTTL
			if expectedTTL == 0 {
				expectedTTL = 5 * time.Minute
			}
			if completer.cache.ttl != expectedTTL {
				t.Errorf("Expected cache TTL %v, got %v", expectedTTL, completer.cache.ttl)
			}

			if completer.clientInit {
				t.Error("Client should not be initialized yet")
			}

			if completer.client != nil {
				t.Error("Client should be nil on creation")
			}
		})
	}
}

func TestCompleterGetModelNames(t *testing.T) {
	t.Run("returns static list", func(t *testing.T) {
		completer := NewCompleter("test-key", true, 5*time.Minute)
		models := completer.GetModelNames()

		if len(models) == 0 {
			t.Error("Expected non-empty model list")
		}

		// Check for expected models
		expectedModels := map[string]bool{
			"gemini-2.5-flash": true,
			"gemini-2.5-pro":   true,
		}

		for _, model := range models {
			if !expectedModels[model] {
				t.Errorf("Unexpected model in list: %s", model)
			}
		}

		for expectedModel := range expectedModels {
			found := false
			for _, model := range models {
				if model == expectedModel {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected model %s not found in list", expectedModel)
			}
		}
	})

	t.Run("returns same list when disabled", func(t *testing.T) {
		completer := NewCompleter("test-key", false, 5*time.Minute)
		models := completer.GetModelNames()

		if len(models) == 0 {
			t.Error("Expected non-empty model list even when disabled")
		}
	})

	t.Run("does not make API calls", func(t *testing.T) {
		completer := NewCompleter("invalid-key", true, 5*time.Minute)
		models := completer.GetModelNames()

		// Should succeed even with invalid key since it's static
		if len(models) == 0 {
			t.Error("Expected static model list to work with invalid key")
		}

		// Client should not have been initialized
		if completer.clientInit {
			t.Error("Client should not be initialized for static method")
		}
	})
}

func TestCompleterDisabledBehavior(t *testing.T) {
	tests := []struct {
		name   string
		method func(*Completer) []string
	}{
		{
			name: "GetStoreNames",
			method: func(c *Completer) []string {
				return c.GetStoreNames()
			},
		},
		{
			name: "GetFileNames",
			method: func(c *Completer) []string {
				return c.GetFileNames()
			},
		},
		{
			name: "GetDocumentNames",
			method: func(c *Completer) []string {
				return c.GetDocumentNames("test-store")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+" when disabled", func(t *testing.T) {
			completer := NewCompleter("test-key", false, 5*time.Minute)
			result := tt.method(completer)

			if len(result) != 0 {
				t.Errorf("Expected empty slice when disabled, got %v", result)
			}

			// Client should not have been initialized
			if completer.clientInit {
				t.Error("Client should not be initialized when disabled")
			}
		})
	}
}

func TestCompleterGetDocumentNamesEmptyStore(t *testing.T) {
	t.Run("returns empty when storeRef is empty", func(t *testing.T) {
		completer := NewCompleter("test-key", true, 5*time.Minute)
		result := completer.GetDocumentNames("")

		if len(result) != 0 {
			t.Errorf("Expected empty slice for empty storeRef, got %v", result)
		}
	})

	t.Run("returns empty when disabled and empty storeRef", func(t *testing.T) {
		completer := NewCompleter("test-key", false, 5*time.Minute)
		result := completer.GetDocumentNames("")

		if len(result) != 0 {
			t.Errorf("Expected empty slice when disabled with empty storeRef, got %v", result)
		}
	})
}

func TestCompleterCacheKeyIsolation(t *testing.T) {
	t.Run("different resource types use different cache keys", func(t *testing.T) {
		completer := NewCompleter("test-key", true, 5*time.Minute)

		// Manually populate cache to test isolation
		completer.cache.Set("stores", []string{"store1", "store2"})
		completer.cache.Set("files", []string{"file1", "file2"})
		completer.cache.Set("docs:teststore", []string{"doc1", "doc2"})

		// Verify cache isolation
		stores, ok := completer.cache.Get("stores")
		if !ok || len(stores) != 2 {
			t.Error("Store cache key not isolated properly")
		}

		files, ok := completer.cache.Get("files")
		if !ok || len(files) != 2 {
			t.Error("File cache key not isolated properly")
		}

		docs, ok := completer.cache.Get("docs:teststore")
		if !ok || len(docs) != 2 {
			t.Error("Document cache key not isolated properly")
		}
	})

	t.Run("document names cached per store", func(t *testing.T) {
		completer := NewCompleter("test-key", true, 5*time.Minute)

		// Manually populate cache with different stores
		completer.cache.Set("docs:store1", []string{"doc1", "doc2"})
		completer.cache.Set("docs:store2", []string{"doc3", "doc4"})

		// Verify different stores have different cache entries
		docs1, ok1 := completer.cache.Get("docs:store1")
		docs2, ok2 := completer.cache.Get("docs:store2")

		if !ok1 || !ok2 {
			t.Error("Document cache entries not created properly")
		}

		if len(docs1) != 2 || docs1[0] != "doc1" {
			t.Error("Store1 documents not cached correctly")
		}

		if len(docs2) != 2 || docs2[0] != "doc3" {
			t.Error("Store2 documents not cached correctly")
		}
	})
}

func TestCompleterClose(t *testing.T) {
	t.Run("close when client not initialized", func(t *testing.T) {
		completer := NewCompleter("test-key", true, 5*time.Minute)
		completer.Close() // Should not panic
	})

	t.Run("close when client is nil", func(t *testing.T) {
		completer := NewCompleter("test-key", true, 5*time.Minute)
		completer.client = nil
		completer.Close() // Should not panic
	})
}

func TestCompleterCacheBehavior(t *testing.T) {
	t.Run("cache respects TTL", func(t *testing.T) {
		shortTTL := 100 * time.Millisecond
		completer := NewCompleter("test-key", true, shortTTL)

		// Manually set cache entry
		completer.cache.Set("stores", []string{"store1"})

		// Verify it's in cache
		if cached, ok := completer.cache.Get("stores"); !ok || len(cached) != 1 {
			t.Error("Cache entry not set properly")
		}

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		// Verify it's expired
		if _, ok := completer.cache.Get("stores"); ok {
			t.Error("Cache entry should have expired")
		}
	})

	t.Run("cache stores values correctly", func(t *testing.T) {
		completer := NewCompleter("test-key", true, 5*time.Minute)

		testData := []string{"value1", "value2", "value3"}
		completer.cache.Set("test-key", testData)

		retrieved, ok := completer.cache.Get("test-key")
		if !ok {
			t.Error("Failed to retrieve cached values")
		}

		if len(retrieved) != len(testData) {
			t.Errorf("Expected %d values, got %d", len(testData), len(retrieved))
		}

		for i, v := range retrieved {
			if v != testData[i] {
				t.Errorf("Value mismatch at index %d: expected %s, got %s", i, testData[i], v)
			}
		}
	})
}

// Note: Tests that require actual API calls (GetStoreNames, GetFileNames, GetDocumentNames
// with real API) should be in integration tests with build tags and require API credentials.
// These tests focus on the logic we can test without API access:
// - Initialization
// - Disabled behavior
// - Cache interaction
// - Static methods (GetModelNames)
// - Error conditions we can trigger
