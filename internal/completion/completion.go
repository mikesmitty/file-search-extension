package completion

import (
	"context"
	"strings"
	"time"

	"github.com/mikesmitty/file-search-extension/internal/constants"
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

// GetModelNames returns a list of available model names
func (c *Completer) GetModelNames() []string {
	// Always include the default/static list as fallback or base
	defaults := constants.GetModelList()

	if !c.enabled {
		return defaults
	}

	// Check cache first
	if cached, ok := c.cache.Get("models"); ok {
		return cached
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Ensure client is initialized
	client, err := c.ensureClient(ctx)
	if err != nil {
		return defaults
	}

	// Get models from API
	models, err := client.ListModels(ctx)
	if err != nil {
		return defaults
	}

	// Extract names
	var names []string
	for _, m := range models {
		// Filter for Gemini models if needed, but for now just take them all
		// The name usually comes as "models/gemini-pro", we might want to strip "models/"
		// or keep it depending on what the API expects.
		// The CLI usually expects "gemini-pro".
		// Let's check what the API returns. Usually "models/gemini-pro".
		// But the user input usually doesn't have "models/".
		// The `genai` SDK `GenerateContent` usually takes "gemini-pro".
		name := strings.TrimPrefix(m.Name, "models/")
		names = append(names, name)
	}

	// If we got no models, return defaults
	if len(names) == 0 {
		return defaults
	}

	// Cache the results
	c.cache.Set("models", names)

	return names
}
