package completion

import (
	"context"
	"time"

	"github.com/mikesmitty/file-search-extension/internal/gemini"
)

// Completer provides completion suggestions for CLI arguments
type Completer struct {
	cache      *Cache
	apiKey     string
	enabled    bool
	client     *gemini.Client
	clientInit bool
}

// NewCompleter creates a new Completer with the specified configuration
func NewCompleter(apiKey string, enabled bool, cacheTTL time.Duration) *Completer {
	return &Completer{
		cache:   NewCache(cacheTTL),
		apiKey:  apiKey,
		enabled: enabled,
	}
}

// ensureClient lazily initializes the gemini client
func (c *Completer) ensureClient(ctx context.Context) (*gemini.Client, error) {
	if c.clientInit {
		return c.client, nil
	}

	client, err := gemini.NewClient(ctx, c.apiKey)
	if err != nil {
		return nil, err
	}

	c.client = client
	c.clientInit = true
	return c.client, nil
}

// Close closes the underlying gemini client if initialized
func (c *Completer) Close() {
	if c.client != nil {
		c.client.Close()
	}
}

// GetStoreNames returns a list of store names for completion.
// Returns empty slice if disabled or on error (graceful degradation).
func (c *Completer) GetStoreNames() []string {
	if !c.enabled {
		return []string{}
	}

	// Check cache first
	if cached, ok := c.cache.Get("stores"); ok {
		return cached
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Ensure client is initialized
	client, err := c.ensureClient(ctx)
	if err != nil {
		return []string{} // Silent failure
	}

	// Get store names from API
	names, err := client.GetStoreNames(ctx)
	if err != nil {
		return []string{} // Silent failure
	}

	// Cache the results
	c.cache.Set("stores", names)

	return names
}

// GetFileNames returns a list of file names for completion.
// Returns empty slice if disabled or on error (graceful degradation).
func (c *Completer) GetFileNames() []string {
	if !c.enabled {
		return []string{}
	}

	// Check cache first
	if cached, ok := c.cache.Get("files"); ok {
		return cached
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Ensure client is initialized
	client, err := c.ensureClient(ctx)
	if err != nil {
		return []string{} // Silent failure
	}

	// Get file names from API
	names, err := client.GetFileNames(ctx)
	if err != nil {
		return []string{} // Silent failure
	}

	// Cache the results
	c.cache.Set("files", names)

	return names
}

// GetDocumentNames returns a list of document names for completion within a store.
// Returns empty slice if disabled or on error (graceful degradation).
func (c *Completer) GetDocumentNames(storeRef string) []string {
	if !c.enabled || storeRef == "" {
		return []string{}
	}

	// Cache key includes store reference
	cacheKey := "docs:" + storeRef

	// Check cache first
	if cached, ok := c.cache.Get(cacheKey); ok {
		return cached
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Ensure client is initialized
	client, err := c.ensureClient(ctx)
	if err != nil {
		return []string{} // Silent failure
	}

	// Resolve store name to ID
	storeID, err := client.ResolveStoreName(ctx, storeRef)
	if err != nil {
		return []string{} // Silent failure
	}

	// Get document names from API
	names, err := client.GetDocumentNames(ctx, storeID)
	if err != nil {
		return []string{} // Silent failure
	}

	// Cache the results
	c.cache.Set(cacheKey, names)

	return names
}

// GetModelNames returns a static list of available model names
func (c *Completer) GetModelNames() []string {
	return []string{
		"gemini-2.5-flash",
		"gemini-2.5-flash-lite",
		"gemini-2.5-pro",
		"gemini-2.0-flash",
		"gemini-2.0-flash-lite",
	}
}
