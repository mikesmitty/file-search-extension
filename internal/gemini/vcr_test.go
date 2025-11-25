package gemini

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

// newVCRClient creates a client that records/replays interactions using go-vcr.
// It returns the client and a cleanup function that must be called.
func newVCRClient(t *testing.T, cassetteName string) (*Client, func()) {
	t.Helper()

	// Ensure testdata directory exists
	fixtureDir := filepath.Join("testdata", "fixtures")
	if err := os.MkdirAll(fixtureDir, 0755); err != nil {
		t.Fatalf("Failed to create fixture directory: %v", err)
	}

	cassettePath := filepath.Join(fixtureDir, cassetteName)

	// Add a hook to remove the API key from recorded interactions
	// The Gemini API uses "x-goog-api-key" header or "key" query param
	hook := func(i *cassette.Interaction) error {
		// Scrub header
		// go-vcr v3 stores headers as map[string][]string
		// We need to remove the key regardless of casing
		for k := range i.Request.Headers {
			if k == "X-Goog-Api-Key" || k == "x-goog-api-key" {
				delete(i.Request.Headers, k)
			}
		}

		// Scrub query param (if present)
		u, err := url.Parse(i.Request.URL)
		if err == nil {
			q := u.Query()
			if q.Has("key") {
				q.Del("key")
				u.RawQuery = q.Encode()
				i.Request.URL = u.String()
			}
		}

		return nil
	}

	// Configure recorder
	r, err := recorder.New(cassettePath, recorder.WithHook(hook, recorder.BeforeSaveHook))
	if err != nil {
		t.Fatalf("Failed to create recorder: %v", err)
	}

	// Use real API key if recording, dummy if replaying
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		if r.Mode() == recorder.ModeRecordOnce {
			t.Skip("GEMINI_API_KEY not set, skipping recording test")
		}
		apiKey = "dummy-api-key-for-replay"
	}

	// Create client with the recorder's HTTP client
	client, err := NewClient(context.Background(), apiKey, r.GetDefaultClient())
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	cleanup := func() {
		if err := r.Stop(); err != nil {
			t.Errorf("Failed to stop recorder: %v", err)
		}
	}

	return client, cleanup
}
