package gemini

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mikesmitty/file-search/internal/constants"
	"google.golang.org/genai"
)

// OperationType represents the type of long-running operation
type OperationType string

const (
	OperationTypeImport OperationType = "import"
	OperationTypeUpload OperationType = "upload"
)

// OperationStatus represents the detailed status of an operation
type OperationStatus struct {
	Name         string         `json:"name"`
	Type         OperationType  `json:"type"`
	Done         bool           `json:"done"`
	Failed       bool           `json:"failed"`
	ErrorMessage string         `json:"errorMessage,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	Parent       string         `json:"parent,omitempty"`
	DocumentName string         `json:"documentName,omitempty"`
}

type Client struct {
	client *genai.Client
}

func NewClient(ctx context.Context, apiKey string) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY not set")
	}

	cfg := &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	}

	client, err := genai.NewClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &Client{client: client}, nil
}

func (c *Client) Close() {
	// No-op for this SDK as it doesn't expose Close
}

func (c *Client) ListStores(ctx context.Context) ([]*genai.FileSearchStore, error) {
	resp, err := c.client.FileSearchStores.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	var stores []*genai.FileSearchStore
	stores = append(stores, resp.Items...)

	for resp.NextPageToken != "" {
		resp, err = resp.Next(ctx)
		if err != nil {
			return nil, err
		}
		stores = append(stores, resp.Items...)
	}
	return stores, nil
}

func (c *Client) ListModels(ctx context.Context) ([]*genai.Model, error) {
	resp, err := c.client.Models.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	var models []*genai.Model
	models = append(models, resp.Items...)

	for resp.NextPageToken != "" {
		resp, err = resp.Next(ctx)
		if err != nil {
			return nil, err
		}
		models = append(models, resp.Items...)
	}
	return models, nil
}

// ResolveStoreName resolves a display name or partial name to a full store resource name.
// If the input is already a resource name (starts with "fileSearchStores/"), returns it as-is.
func (c *Client) ResolveStoreName(ctx context.Context, nameOrID string) (string, error) {
	// If already a resource name, return as-is
	if strings.HasPrefix(nameOrID, constants.StoreResourcePrefix) {
		return nameOrID, nil
	}

	// Otherwise, search for display name
	stores, err := c.ListStores(ctx)
	if err != nil {
		return "", err
	}

	for _, s := range stores {
		if s.DisplayName == nameOrID {
			return s.Name, nil
		}
	}

	return "", fmt.Errorf("store not found: %s", nameOrID)
}

// ResolveFileName resolves a file display name to a full file resource name.
// If the input is already a resource name (starts with "files/"), returns it as-is.
func (c *Client) ResolveFileName(ctx context.Context, nameOrID string) (string, error) {
	// If already a resource name, return as-is
	if strings.HasPrefix(nameOrID, constants.FileResourcePrefix) {
		return nameOrID, nil
	}

	// Otherwise, search for display name
	files, err := c.ListFiles(ctx)
	if err != nil {
		return "", err
	}

	for _, f := range files {
		if f.DisplayName == nameOrID {
			return f.Name, nil
		}
	}

	return "", fmt.Errorf("file not found: %s", nameOrID)
}

// ResolveDocumentName resolves a document display name to a full document resource name.
// If the input is already a resource name (contains "documents/"), returns it as-is.
// Requires the store name/ID to scope the search.
func (c *Client) ResolveDocumentName(ctx context.Context, storeNameOrID, docNameOrID string) (string, error) {
	// If already a resource name, return as-is
	if strings.Contains(docNameOrID, constants.DocumentResourcePrefix) {
		return docNameOrID, nil
	}

	// Resolve store name first
	storeID, err := c.ResolveStoreName(ctx, storeNameOrID)
	if err != nil {
		return "", err
	}

	// Search for document by display name
	docs, err := c.ListDocuments(ctx, storeID)
	if err != nil {
		return "", err
	}

	for _, doc := range docs {
		if doc.DisplayName == docNameOrID {
			return doc.Name, nil
		}
	}

	return "", fmt.Errorf("document not found in store %s: %s", storeID, docNameOrID)
}

// GetStoreNames returns a list of all store display names for completion.
func (c *Client) GetStoreNames(ctx context.Context) ([]string, error) {
	stores, err := c.ListStores(ctx)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(stores))
	for _, s := range stores {
		names = append(names, s.DisplayName)
	}
	return names, nil
}

// GetFileNames returns a list of all file display names for completion.
func (c *Client) GetFileNames(ctx context.Context) ([]string, error) {
	files, err := c.ListFiles(ctx)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(files))
	for _, f := range files {
		names = append(names, f.DisplayName)
	}
	return names, nil
}

// GetDocumentNames returns a list of all document display names in a store for completion.
func (c *Client) GetDocumentNames(ctx context.Context, storeID string) ([]string, error) {
	docs, err := c.ListDocuments(ctx, storeID)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(docs))
	for _, doc := range docs {
		names = append(names, doc.DisplayName)
	}
	return names, nil
}

func (c *Client) GetStore(ctx context.Context, name string) (*genai.FileSearchStore, error) {
	return c.client.FileSearchStores.Get(ctx, name, nil)
}

func (c *Client) DeleteStore(ctx context.Context, name string, force bool) error {
	// Build optional delete config
	cfg := &genai.DeleteFileSearchStoreConfig{}
	if force {
		cfg.Force = new(bool)
		*cfg.Force = true
	}
	return c.client.FileSearchStores.Delete(ctx, name, cfg)
}

func (c *Client) CreateStore(ctx context.Context, displayName string) (*genai.FileSearchStore, error) {
	return c.client.FileSearchStores.Create(ctx, &genai.CreateFileSearchStoreConfig{
		DisplayName: displayName,
	})
}

type UploadFileOptions struct {
	StoreName      string
	DisplayName    string
	MIMEType       string
	MaxChunkTokens int
	ChunkOverlap   int
	Metadata       map[string]string
	Quiet          bool
}

type ImportFileOptions struct {
	Quiet bool
}

// UploadFile uploads a file and optionally indexes it in a store.
// It returns the created File (if no store) or nil (if store upload, as operation handles it).
// For store uploads, it polls until completion.
func (c *Client) UploadFile(ctx context.Context, path string, opts *UploadFileOptions) (*genai.File, error) {
	if opts == nil {
		opts = &UploadFileOptions{}
	}

	// If storeName is provided, use UploadToFileSearchStoreFromPath (direct)
	// If not, just UploadFromPath (Files API only)

	if opts.StoreName != "" {
		if !opts.Quiet {
			fmt.Printf("Uploading %s to store %s...\n", path, opts.StoreName)
		}

		config := &genai.UploadToFileSearchStoreConfig{
			DisplayName: opts.DisplayName,
			MIMEType:    opts.MIMEType,
		}

		// Add chunking config if specified
		if opts.MaxChunkTokens > 0 || opts.ChunkOverlap > 0 {
			config.ChunkingConfig = &genai.ChunkingConfig{
				WhiteSpaceConfig: &genai.WhiteSpaceConfig{},
			}
			if opts.MaxChunkTokens > 0 {
				maxTokens := int32(opts.MaxChunkTokens)
				config.ChunkingConfig.WhiteSpaceConfig.MaxTokensPerChunk = &maxTokens
			}
			if opts.ChunkOverlap > 0 {
				overlapTokens := int32(opts.ChunkOverlap)
				config.ChunkingConfig.WhiteSpaceConfig.MaxOverlapTokens = &overlapTokens
			}
		}

		// Add metadata if specified
		if len(opts.Metadata) > 0 {
			config.CustomMetadata = make([]*genai.CustomMetadata, 0, len(opts.Metadata))
			for key, value := range opts.Metadata {
				config.CustomMetadata = append(config.CustomMetadata, &genai.CustomMetadata{
					Key:         key,
					StringValue: value,
				})
			}
		}

		op, err := c.client.FileSearchStores.UploadToFileSearchStoreFromPath(ctx, path, opts.StoreName, config)
		if err != nil {
			return nil, err
		}

		// Poll with optional progress indicator
		startTime := time.Now()
		if !opts.Quiet {
			fmt.Print("Indexing...")
		}
		for !op.Done {
			if !opts.Quiet {
				elapsed := time.Since(startTime)
				fmt.Printf("\rIndexing... (%s elapsed)", elapsed.Round(time.Second))
			}

			time.Sleep(2 * time.Second)
			op, err = c.client.Operations.GetUploadToFileSearchStoreOperation(ctx, op, nil)
			if err != nil {
				if !opts.Quiet {
					fmt.Println() // New line before error
				}
				return nil, err
			}
		}
		if !opts.Quiet {
			fmt.Println("\n✓ Upload and index complete.")
		}
		return nil, nil
	}

	// Just upload to Files API
	config := &genai.UploadFileConfig{
		DisplayName: opts.DisplayName,
		MIMEType:    opts.MIMEType,
	}
	// Note: metadata might not be supported for Files API uploads
	// Only chunking config is for store uploads

	res, err := c.client.Files.UploadFromPath(ctx, path, config)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ImportFile imports an existing file from the Files API into a File Search Store.
// fileID should be a file resource name (e.g., "files/abc123").
// storeID should be a store resource name (e.g., "fileSearchStores/xyz789").
func (c *Client) ImportFile(ctx context.Context, fileID, storeID string, opts *ImportFileOptions) error {
	if opts == nil {
		opts = &ImportFileOptions{}
	}

	if !opts.Quiet {
		fmt.Printf("Importing file %s into store %s...\n", fileID, storeID)
	}

	op, err := c.client.FileSearchStores.ImportFile(ctx, storeID, fileID, &genai.ImportFileConfig{})
	if err != nil {
		return err
	}

	// Poll operation until complete with optional progress indicator
	startTime := time.Now()
	if !opts.Quiet {
		fmt.Printf("Operation ID: %s\n", op.Name)
		fmt.Print("Importing...")
	}
	for !op.Done {
		if !opts.Quiet {
			elapsed := time.Since(startTime)
			fmt.Printf("\rImporting... (%s elapsed)", elapsed.Round(time.Second))
		}

		time.Sleep(2 * time.Second)
		op, err = c.client.Operations.GetImportFileOperation(ctx, op, nil)
		if err != nil {
			if !opts.Quiet {
				fmt.Println() // New line before error
			}
			return err
		}
	}
	if !opts.Quiet {
		fmt.Println("\n✓ Import complete.")
	}
	return nil
}

func (c *Client) ListFiles(ctx context.Context) ([]*genai.File, error) {
	resp, err := c.client.Files.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	var files []*genai.File
	files = append(files, resp.Items...)

	for resp.NextPageToken != "" {
		resp, err = resp.Next(ctx)
		if err != nil {
			return nil, err
		}
		files = append(files, resp.Items...)
	}
	return files, nil
}

func (c *Client) GetFile(ctx context.Context, name string) (*genai.File, error) {
	return c.client.Files.Get(ctx, name, nil)
}

func (c *Client) ListDocuments(ctx context.Context, storeName string) ([]*genai.Document, error) {
	resp, err := c.client.FileSearchStores.Documents.List(ctx, storeName, nil)
	if err != nil {
		return nil, err
	}

	var docs []*genai.Document
	docs = append(docs, resp.Items...)

	for resp.NextPageToken != "" {
		resp, err = resp.Next(ctx)
		if err != nil {
			return nil, err
		}
		docs = append(docs, resp.Items...)
	}
	return docs, nil
}

func (c *Client) GetDocument(ctx context.Context, name string) (*genai.Document, error) {
	return c.client.FileSearchStores.Documents.Get(ctx, name, nil)
}

func (c *Client) DeleteDocument(ctx context.Context, name string, force bool) error {
	cfg := &genai.DeleteDocumentConfig{}
	if force {
		cfg.Force = new(bool)
		*cfg.Force = true
	}
	return c.client.FileSearchStores.Documents.Delete(ctx, name, cfg)
}

func (c *Client) DeleteFile(ctx context.Context, name string) error {
	_, err := c.client.Files.Delete(ctx, name, nil)
	return err
}

func (c *Client) Query(ctx context.Context, text string, storeName string, modelName string, metadataFilter string) (*genai.GenerateContentResponse, error) {
	var config *genai.GenerateContentConfig

	if storeName != "" {
		fs := &genai.FileSearch{FileSearchStoreNames: []string{storeName}}
		if metadataFilter != "" {
			fs.MetadataFilter = metadataFilter
		}
		config = &genai.GenerateContentConfig{Tools: []*genai.Tool{{FileSearch: fs}}}
	}

	return c.client.Models.GenerateContent(ctx, modelName, genai.Text(text), config)
}

// GetOperation retrieves the status of a long-running operation.
// If operationType is empty, it will try both import and upload types.
func (c *Client) GetOperation(ctx context.Context, operationName string, operationType OperationType) (*OperationStatus, error) {
	// Validate operation name format
	if !strings.HasPrefix(operationName, constants.StoreResourcePrefix) {
		return nil, fmt.Errorf("invalid operation name: must start with '%s'", constants.StoreResourcePrefix)
	}
	if !strings.Contains(operationName, constants.OperationResourcePrefix) {
		return nil, fmt.Errorf("invalid operation name: must contain '%s'", constants.OperationResourcePrefix)
	}

	// If type specified, use it directly
	if operationType == OperationTypeImport {
		return c.getImportOperation(ctx, operationName)
	}
	if operationType == OperationTypeUpload {
		return c.getUploadOperation(ctx, operationName)
	}

	// Try both types if not specified
	status, err := c.getImportOperation(ctx, operationName)
	if err == nil {
		return status, nil
	}

	return c.getUploadOperation(ctx, operationName)
}

func (c *Client) getImportOperation(ctx context.Context, operationName string) (*OperationStatus, error) {
	op := &genai.ImportFileOperation{Name: operationName}
	result, err := c.client.Operations.GetImportFileOperation(ctx, op, nil)
	if err != nil {
		return nil, err
	}

	status := &OperationStatus{
		Name:     result.Name,
		Type:     OperationTypeImport,
		Done:     result.Done,
		Metadata: result.Metadata,
	}

	if result.Error != nil {
		status.Failed = true
		if msg, ok := result.Error["message"].(string); ok {
			status.ErrorMessage = msg
		} else {
			status.ErrorMessage = fmt.Sprintf("%v", result.Error)
		}
	}

	if result.Response != nil {
		status.Parent = result.Response.Parent
		status.DocumentName = result.Response.DocumentName
	}

	return status, nil
}

func (c *Client) getUploadOperation(ctx context.Context, operationName string) (*OperationStatus, error) {
	op := &genai.UploadToFileSearchStoreOperation{Name: operationName}
	result, err := c.client.Operations.GetUploadToFileSearchStoreOperation(ctx, op, nil)
	if err != nil {
		return nil, err
	}

	status := &OperationStatus{
		Name:     result.Name,
		Type:     OperationTypeUpload,
		Done:     result.Done,
		Metadata: result.Metadata,
	}

	if result.Error != nil {
		status.Failed = true
		if msg, ok := result.Error["message"].(string); ok {
			status.ErrorMessage = msg
		} else {
			status.ErrorMessage = fmt.Sprintf("%v", result.Error)
		}
	}

	if result.Response != nil {
		status.Parent = result.Response.Parent
		status.DocumentName = result.Response.DocumentName
	}

	return status, nil
}
