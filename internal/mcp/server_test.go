package mcp

import (
	"context"
	"reflect"
	"testing"

	"github.com/mikesmitty/file-search-extension/internal/gemini"
	"google.golang.org/genai"
)

// MockGeminiClient implements GeminiClient for testing
type MockGeminiClient struct {
	ListStoresFunc          func(ctx context.Context) ([]*genai.FileSearchStore, error)
	ListFilesFunc           func(ctx context.Context) ([]*genai.File, error)
	ResolveStoreNameFunc    func(ctx context.Context, nameOrID string) (string, error)
	ListDocumentsFunc       func(ctx context.Context, storeID string) ([]*genai.Document, error)
	CreateStoreFunc         func(ctx context.Context, displayName string) (*genai.FileSearchStore, error)
	DeleteStoreFunc         func(ctx context.Context, name string, force bool) error
	ResolveFileNameFunc     func(ctx context.Context, nameOrID string) (string, error)
	ImportFileFunc          func(ctx context.Context, fileID, storeID string, opts *gemini.ImportFileOptions) error
	QueryFunc               func(ctx context.Context, text string, storeName string, modelName string, metadataFilter string) (*genai.GenerateContentResponse, error)
	UploadFileFunc          func(ctx context.Context, path string, opts *gemini.UploadFileOptions) (*genai.File, error)
	DeleteFileFunc          func(ctx context.Context, name string) error
	ResolveDocumentNameFunc func(ctx context.Context, storeNameOrID, docNameOrID string) (string, error)
	DeleteDocumentFunc      func(ctx context.Context, name string) error
	CloseFunc               func()
}

func (m *MockGeminiClient) ListStores(ctx context.Context) ([]*genai.FileSearchStore, error) {
	return m.ListStoresFunc(ctx)
}
func (m *MockGeminiClient) ListFiles(ctx context.Context) ([]*genai.File, error) {
	return m.ListFilesFunc(ctx)
}
func (m *MockGeminiClient) ResolveStoreName(ctx context.Context, nameOrID string) (string, error) {
	return m.ResolveStoreNameFunc(ctx, nameOrID)
}
func (m *MockGeminiClient) ListDocuments(ctx context.Context, storeID string) ([]*genai.Document, error) {
	return m.ListDocumentsFunc(ctx, storeID)
}
func (m *MockGeminiClient) CreateStore(ctx context.Context, displayName string) (*genai.FileSearchStore, error) {
	return m.CreateStoreFunc(ctx, displayName)
}
func (m *MockGeminiClient) DeleteStore(ctx context.Context, name string, force bool) error {
	return m.DeleteStoreFunc(ctx, name, force)
}
func (m *MockGeminiClient) ResolveFileName(ctx context.Context, nameOrID string) (string, error) {
	return m.ResolveFileNameFunc(ctx, nameOrID)
}
func (m *MockGeminiClient) ImportFile(ctx context.Context, fileID, storeID string, opts *gemini.ImportFileOptions) error {
	return m.ImportFileFunc(ctx, fileID, storeID, opts)
}
func (m *MockGeminiClient) Query(ctx context.Context, text string, storeName string, modelName string, metadataFilter string) (*genai.GenerateContentResponse, error) {
	return m.QueryFunc(ctx, text, storeName, modelName, metadataFilter)
}
func (m *MockGeminiClient) UploadFile(ctx context.Context, path string, opts *gemini.UploadFileOptions) (*genai.File, error) {
	return m.UploadFileFunc(ctx, path, opts)
}
func (m *MockGeminiClient) DeleteFile(ctx context.Context, name string) error {
	return m.DeleteFileFunc(ctx, name)
}
func (m *MockGeminiClient) ResolveDocumentName(ctx context.Context, storeNameOrID, docNameOrID string) (string, error) {
	return m.ResolveDocumentNameFunc(ctx, storeNameOrID, docNameOrID)
}
func (m *MockGeminiClient) DeleteDocument(ctx context.Context, name string) error {
	return m.DeleteDocumentFunc(ctx, name)
}
func (m *MockGeminiClient) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}

// getReflectedToolsMap is a helper to access the internal tools map via reflection
func getReflectedToolsMap(server interface{}) reflect.Value {
	return reflect.ValueOf(server).Elem().FieldByName("tools")
}

// TestServerTools verifies that tools are registered correctly.
func TestNewServer_ToolRegistration(t *testing.T) {
	mockClient := &MockGeminiClient{}
	enabledTools := []string{"all"}

	s := NewServer(mockClient, enabledTools)

	val := getReflectedToolsMap(s)
	if !val.IsValid() {
		t.Fatalf("Could not find 'tools' field in MCPServer")
	}

	if val.Kind() != reflect.Map {
		t.Fatalf("'tools' field is not a map")
	}

	registeredTools := val.MapKeys()
	if len(registeredTools) == 0 {
		t.Errorf("Expected tools to be registered, got 0")
	}

	expectedTools := []string{
		"list_stores",
		"list_files",
		"list_documents",
		"create_store",
		"delete_store",
		"import_file_to_store",
		"query_knowledge_base",
		"upload_file",
		"delete_file",
		"delete_document",
	}

	for _, expected := range expectedTools {
		found := false
		for _, key := range registeredTools {
			if key.String() == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected tool %s to be registered", expected)
		}
	}
}

func TestNewServer_SelectiveToolRegistration(t *testing.T) {
	mockClient := &MockGeminiClient{}
	enabledTools := []string{"query", "upload"}

	s := NewServer(mockClient, enabledTools)

	val := getReflectedToolsMap(s)
	if !val.IsValid() {
		t.Fatalf("Could not find 'tools' field in MCPServer")
	}

	registeredTools := val.MapKeys()

	// query -> query_knowledge_base
	// upload -> upload_file

	expectedTools := []string{
		"query_knowledge_base",
		"upload_file",
	}

	for _, expected := range expectedTools {
		found := false
		for _, key := range registeredTools {
			if key.String() == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected tool %s to be registered", expected)
		}
	}

	// Ensure others are NOT registered
	unexpectedTools := []string{
		"delete_file",
		"create_store",
	}

	for _, unexpected := range unexpectedTools {
		found := false
		for _, key := range registeredTools {
			if key.String() == unexpected {
				found = true
				break
			}
		}
		if found {
			t.Errorf("Did not expect tool %s to be registered", unexpected)
		}
	}
}

// To make this testable, let's assume we refactor RunServer to return the server instance in a future step.
// For now, I will add a test that ensures the MockClient satisfies the interface.
func TestMockClientSatisfiesInterface(t *testing.T) {
	var _ GeminiClient = &MockGeminiClient{}
}
