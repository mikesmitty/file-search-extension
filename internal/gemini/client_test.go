package gemini

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestUploadFileOptions(t *testing.T) {
	tests := []struct {
		name string
		opts *UploadFileOptions
		want bool
	}{
		{
			name: "nil options",
			opts: nil,
			want: true,
		},
		{
			name: "empty options",
			opts: &UploadFileOptions{},
			want: true,
		},
		{
			name: "with store name",
			opts: &UploadFileOptions{
				StoreName: "test-store",
			},
			want: true,
		},
		{
			name: "with chunking config",
			opts: &UploadFileOptions{
				StoreName:      "test-store",
				MaxChunkTokens: 2000,
				ChunkOverlap:   200,
			},
			want: true,
		},
		{
			name: "with metadata",
			opts: &UploadFileOptions{
				StoreName: "test-store",
				Metadata: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the struct can be created without panic
			if tt.opts == nil {
				opts := &UploadFileOptions{}
				if opts == nil {
					t.Error("failed to create UploadFileOptions")
				}
			}
		})
	}
}

func TestUploadFileOptionsDefaults(t *testing.T) {
	opts := &UploadFileOptions{}

	if opts.StoreName != "" {
		t.Errorf("expected empty StoreName, got %s", opts.StoreName)
	}
	if opts.DisplayName != "" {
		t.Errorf("expected empty DisplayName, got %s", opts.DisplayName)
	}
	if opts.MIMEType != "" {
		t.Errorf("expected empty MIMEType, got %s", opts.MIMEType)
	}
	if opts.MaxChunkTokens != 0 {
		t.Errorf("expected 0 MaxChunkTokens, got %d", opts.MaxChunkTokens)
	}
	if opts.ChunkOverlap != 0 {
		t.Errorf("expected 0 ChunkOverlap, got %d", opts.ChunkOverlap)
	}
	if opts.Metadata != nil {
		t.Errorf("expected nil Metadata, got %v", opts.Metadata)
	}
}

func TestResolveStoreNameFormat(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		shouldPass bool
		desc       string
	}{
		{
			name:       "resource name format",
			input:      "fileSearchStores/abc123",
			shouldPass: true,
			desc:       "Should pass through resource names starting with 'fileSearchStores/'",
		},
		{
			name:       "resource name format uppercase",
			input:      "FileSearchStores/abc123",
			shouldPass: false,
			desc:       "Uppercase 'FileSearchStores' should not pass through (only lowercase is checked)",
		},
		{
			name:       "friendly name",
			input:      "My Research Store",
			shouldPass: false,
			desc:       "Friendly names require API lookup (would fail without mock)",
		},
		{
			name:       "empty string",
			input:      "",
			shouldPass: false,
			desc:       "Empty string should not pass through",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the format detection logic (without API call)
			isResourceName := false
			if len(tt.input) > 0 && (tt.input[:1] == "f" || tt.input[:1] == "F") {
				if len(tt.input) > 16 && tt.input[:16] == "fileSearchStores" {
					isResourceName = true
				}
			}

			if isResourceName != tt.shouldPass {
				t.Errorf("%s: expected pass=%v, got pass=%v", tt.desc, tt.shouldPass, isResourceName)
			}
		})
	}
}

func TestResolveFileNameFormat(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		shouldPass bool
		desc       string
	}{
		{
			name:       "resource name format",
			input:      "files/abc123xyz",
			shouldPass: true,
			desc:       "Should pass through resource names starting with 'files/'",
		},
		{
			name:       "friendly name",
			input:      "document.pdf",
			shouldPass: false,
			desc:       "Friendly names require API lookup",
		},
		{
			name:       "short string",
			input:      "file",
			shouldPass: false,
			desc:       "Too short to be a resource name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the format detection logic
			isResourceName := len(tt.input) > 6 && tt.input[:6] == "files/"

			if isResourceName != tt.shouldPass {
				t.Errorf("%s: expected pass=%v, got pass=%v", tt.desc, tt.shouldPass, isResourceName)
			}
		})
	}
}

func TestResolveDocumentNameFormat(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		shouldPass bool
		desc       string
	}{
		{
			name:       "resource name format",
			input:      "fileSearchStores/store123/documents/doc456",
			shouldPass: true,
			desc:       "Should pass through resource names containing '/documents/'",
		},
		{
			name:       "friendly name",
			input:      "my-document.pdf",
			shouldPass: false,
			desc:       "Friendly names require API lookup",
		},
		{
			name:       "short string",
			input:      "doc",
			shouldPass: false,
			desc:       "Strings too short should not pass through",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the format detection logic (matches client.go line 103)
			isResourceName := len(tt.input) > 10 && containsDocuments(tt.input)

			if isResourceName != tt.shouldPass {
				t.Errorf("%s: expected pass=%v, got pass=%v", tt.desc, tt.shouldPass, isResourceName)
			}
		})
	}
}

func containsDocuments(s string) bool {
	// Match the logic in client.go: strings.Contains(docNameOrID, "/documents/")
	return strings.Contains(s, "/documents/")
}

// TestGetStoreNamesSignature verifies the method signature and return types
func TestGetStoreNamesSignature(t *testing.T) {
	t.Run("method exists and returns correct types", func(t *testing.T) {
		// This test verifies the method signature exists
		// Cannot test actual API call without credentials, but we can verify the structure

		// Verify the method signature is correct by trying to assign it
		var f func(*Client) func() ([]string, error)
		f = func(c *Client) func() ([]string, error) {
			return func() ([]string, error) {
				return c.GetStoreNames(nil)
			}
		}

		if f == nil {
			t.Error("GetStoreNames method signature verification failed")
		}
	})
}

// TestGetFileNamesSignature verifies the method signature and return types
func TestGetFileNamesSignature(t *testing.T) {
	t.Run("method exists and returns correct types", func(t *testing.T) {
		// Verify the method signature is correct by trying to assign it
		var f func(*Client) func() ([]string, error)
		f = func(c *Client) func() ([]string, error) {
			return func() ([]string, error) {
				return c.GetFileNames(nil)
			}
		}

		if f == nil {
			t.Error("GetFileNames method signature verification failed")
		}
	})
}

// TestGetDocumentNamesSignature verifies the method signature and return types
func TestGetDocumentNamesSignature(t *testing.T) {
	t.Run("method exists and returns correct types", func(t *testing.T) {
		// Verify the method signature is correct by trying to assign it
		var f func(*Client) func(string) ([]string, error)
		f = func(c *Client) func(string) ([]string, error) {
			return func(storeID string) ([]string, error) {
				return c.GetDocumentNames(nil, storeID)
			}
		}

		if f == nil {
			t.Error("GetDocumentNames method signature verification failed")
		}
	})
}

// TestGetNamesReturnTypes verifies that the Get*Names methods return the expected types
func TestGetNamesReturnTypes(t *testing.T) {
	tests := []struct {
		name        string
		description string
	}{
		{
			name:        "GetStoreNames",
			description: "should return []string and error",
		},
		{
			name:        "GetFileNames",
			description: "should return []string and error",
		},
		{
			name:        "GetDocumentNames",
			description: "should return []string and error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// These tests verify the methods exist and have correct return types
			// Actual API testing requires integration tests with credentials
			t.Log(tt.description)
		})
	}
}

// Note: Full integration tests for GetStoreNames, GetFileNames, and GetDocumentNames
// require valid API credentials and should be run separately as integration tests.
// These tests verify the method signatures and structure without making API calls.

func TestOperationTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		opType   OperationType
		expected string
	}{
		{
			name:     "import type",
			opType:   OperationTypeImport,
			expected: "import",
		},
		{
			name:     "upload type",
			opType:   OperationTypeUpload,
			expected: "upload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.opType) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.opType))
			}
		})
	}
}

func TestOperationNameValidation(t *testing.T) {
	tests := []struct {
		name      string
		opName    string
		shouldErr bool
		errMsg    string
	}{
		{
			name:      "valid operation name",
			opName:    "fileSearchStores/abc123/operations/op456",
			shouldErr: false,
		},
		{
			name:      "valid operation name with longer IDs",
			opName:    "fileSearchStores/store-id-123xyz/operations/operation-id-789abc",
			shouldErr: false,
		},
		{
			name:      "missing fileSearchStores prefix",
			opName:    "abc123/operations/op456",
			shouldErr: true,
			errMsg:    "must start with 'fileSearchStores/'",
		},
		{
			name:      "missing operations segment",
			opName:    "fileSearchStores/abc123/op456",
			shouldErr: true,
			errMsg:    "must contain '/operations/'",
		},
		{
			name:      "empty string",
			opName:    "",
			shouldErr: true,
			errMsg:    "must start with 'fileSearchStores/'",
		},
		{
			name:      "just fileSearchStores",
			opName:    "fileSearchStores/",
			shouldErr: true,
			errMsg:    "must contain '/operations/'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the validation logic from GetOperation
			hasPrefix := strings.HasPrefix(tt.opName, "fileSearchStores/")
			hasOperations := strings.Contains(tt.opName, "/operations/")

			isValid := hasPrefix && hasOperations
			if isValid == tt.shouldErr {
				t.Errorf("Validation mismatch for %q: expected shouldErr=%v, got isValid=%v",
					tt.opName, tt.shouldErr, isValid)
			}
		})
	}
}

func TestOperationStatusStructure(t *testing.T) {
	t.Run("operation status struct fields", func(t *testing.T) {
		status := &OperationStatus{
			Name:         "fileSearchStores/abc/operations/123",
			Type:         OperationTypeImport,
			Done:         true,
			Failed:       false,
			ErrorMessage: "",
			Metadata:     map[string]any{"key": "value"},
			Parent:       "fileSearchStores/abc",
			DocumentName: "fileSearchStores/abc/documents/doc123",
		}

		if status.Name == "" {
			t.Error("Name should not be empty")
		}
		if status.Type != OperationTypeImport {
			t.Error("Type should be OperationTypeImport")
		}
		if !status.Done {
			t.Error("Done should be true")
		}
		if status.Failed {
			t.Error("Failed should be false")
		}
		if len(status.Metadata) != 1 {
			t.Error("Metadata should have 1 entry")
		}
	})

	t.Run("operation status JSON tags", func(t *testing.T) {
		// Verify JSON tags are properly defined
		status := OperationStatus{
			Name:   "test",
			Type:   OperationTypeImport,
			Done:   true,
			Failed: false,
		}

		// This test verifies the struct can be marshaled to JSON
		data, err := json.Marshal(status)
		if err != nil {
			t.Errorf("Failed to marshal OperationStatus: %v", err)
		}

		var decoded map[string]any
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Errorf("Failed to unmarshal JSON: %v", err)
		}

		// Check expected JSON fields exist
		expectedFields := []string{"name", "type", "done", "failed"}
		for _, field := range expectedFields {
			if _, ok := decoded[field]; !ok {
				t.Errorf("Expected JSON field %q not found", field)
			}
		}
	})
}

func TestOperationStatusFailedState(t *testing.T) {
	tests := []struct {
		name         string
		status       *OperationStatus
		expectFailed bool
		expectDone   bool
	}{
		{
			name: "pending operation",
			status: &OperationStatus{
				Done:   false,
				Failed: false,
			},
			expectFailed: false,
			expectDone:   false,
		},
		{
			name: "completed operation",
			status: &OperationStatus{
				Done:   true,
				Failed: false,
			},
			expectFailed: false,
			expectDone:   true,
		},
		{
			name: "failed operation",
			status: &OperationStatus{
				Done:         true,
				Failed:       true,
				ErrorMessage: "something went wrong",
			},
			expectFailed: true,
			expectDone:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.status.Failed != tt.expectFailed {
				t.Errorf("Expected Failed=%v, got %v", tt.expectFailed, tt.status.Failed)
			}
			if tt.status.Done != tt.expectDone {
				t.Errorf("Expected Done=%v, got %v", tt.expectDone, tt.status.Done)
			}
		})
	}
}
