package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mikesmitty/file-search/internal/constants"
	"github.com/mikesmitty/file-search/internal/gemini"
	"google.golang.org/genai"
)

// GeminiClient defines the interface required by the MCP server
type GeminiClient interface {
	ListStores(ctx context.Context) ([]*genai.FileSearchStore, error)
	ListFiles(ctx context.Context) ([]*genai.File, error)
	ResolveStoreName(ctx context.Context, nameOrID string) (string, error)
	ListDocuments(ctx context.Context, storeID string) ([]*genai.Document, error)
	CreateStore(ctx context.Context, displayName string) (*genai.FileSearchStore, error)
	DeleteStore(ctx context.Context, name string, force bool) error
	ResolveFileName(ctx context.Context, nameOrID string) (string, error)
	ImportFile(ctx context.Context, fileID, storeID string, opts *gemini.ImportFileOptions) error
	Query(ctx context.Context, text string, storeName string, modelName string, metadataFilter string) (*genai.GenerateContentResponse, error)
	UploadFile(ctx context.Context, path string, opts *gemini.UploadFileOptions) (*genai.File, error)
	DeleteFile(ctx context.Context, name string) error
	ResolveDocumentName(ctx context.Context, storeNameOrID, docNameOrID string) (string, error)
	DeleteDocument(ctx context.Context, name string) error
	Close()
}

func RunServer(ctx context.Context, client GeminiClient, enabledTools []string) error {
	s := NewServer(client, enabledTools)
	return server.ServeStdio(s)
}

// NewServer creates a new MCP server instance with the configured tools.
// It is exported to allow testing of the server configuration and tool registration.
func NewServer(client GeminiClient, enabledTools []string) *server.MCPServer {
	s := server.NewMCPServer(
		"Gemini File Search",
		"1.0.0",
	)

	// Helper to check if a tool is enabled
	isToolEnabled := func(name string) bool {
		for _, t := range enabledTools {
			if t == name {
				return true
			}
		}
		return false
	}

	// Helper to get string argument
	getStringArg := func(args map[string]interface{}, key string) (string, bool) {
		val, ok := args[key]
		if !ok {
			return "", false
		}
		str, ok := val.(string)
		return str, ok
	}

	// Helper to get bool argument
	getBoolArg := func(args map[string]interface{}, key string) bool {
		val, ok := args[key]
		if !ok {
			return false
		}
		b, ok := val.(bool)
		return b && ok
	}

	// Tool: list_stores
	if isToolEnabled("list_stores") || isToolEnabled("all") {
		s.AddTool(mcp.NewTool("list_stores",
			mcp.WithDescription("List all File Search Stores. Returns a JSON array of store objects containing name, displayName, and other metadata."),
		), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			stores, err := client.ListStores(ctx)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			res, err := mcp.NewToolResultJSON(stores)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return res, nil
		})
	}

	// Tool: list_files
	if isToolEnabled("list_files") || isToolEnabled("all") {
		s.AddTool(mcp.NewTool("list_files",
			mcp.WithDescription("List all files in the Gemini Files API. Returns a JSON array of file objects."),
		), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			files, err := client.ListFiles(ctx)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			res, err := mcp.NewToolResultJSON(files)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return res, nil
		})
	}

	// Tool: list_documents
	if isToolEnabled("list_documents") || isToolEnabled("all") {
		s.AddTool(mcp.NewTool("list_documents",
			mcp.WithDescription("List all documents within a specified File Search Store. Returns a JSON array of document objects."),
			mcp.WithString("store_name", mcp.Required(), mcp.Description("The resource name or display name of the store to list documents from.")),
		), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, ok := request.Params.Arguments.(map[string]interface{})
			if !ok {
				return mcp.NewToolResultError("arguments must be a map"), nil
			}
			storeName, ok := getStringArg(args, "store_name")
			if !ok {
				return mcp.NewToolResultError("store_name must be a string"), nil
			}

			// Resolve store name
			storeID, err := client.ResolveStoreName(ctx, storeName)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to resolve store name: %v", err)), nil
			}

			docs, err := client.ListDocuments(ctx, storeID)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			res, err := mcp.NewToolResultJSON(docs)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return res, nil
		})
	}

	// Tool: create_store
	if isToolEnabled("create_store") || isToolEnabled("all") {
		s.AddTool(mcp.NewTool("create_store",
			mcp.WithDescription("Create a new File Search Store."),
			mcp.WithString("display_name", mcp.Required(), mcp.Description("The human-readable name for the new store.")),
		), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, ok := request.Params.Arguments.(map[string]interface{})
			if !ok {
				return mcp.NewToolResultError("arguments must be a map"), nil
			}
			displayName, ok := getStringArg(args, "display_name")
			if !ok {
				return mcp.NewToolResultError("display_name must be a string"), nil
			}

			store, err := client.CreateStore(ctx, displayName)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			res, err := mcp.NewToolResultJSON(store)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return res, nil
		})
	}

	// Tool: delete_store
	if isToolEnabled("delete_store") || isToolEnabled("all") {
		s.AddTool(mcp.NewTool("delete_store",
			mcp.WithDescription("Delete a File Search Store."),
			mcp.WithString("store_name", mcp.Required(), mcp.Description("The resource name or display name of the store to delete.")),
			mcp.WithBoolean("force", mcp.Description("Force delete even if the store contains documents.")),
		), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, ok := request.Params.Arguments.(map[string]interface{})
			if !ok {
				return mcp.NewToolResultError("arguments must be a map"), nil
			}
			storeName, ok := getStringArg(args, "store_name")
			if !ok {
				return mcp.NewToolResultError("store_name must be a string"), nil
			}
			force := getBoolArg(args, "force")

			// Resolve store name
			storeID, err := client.ResolveStoreName(ctx, storeName)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to resolve store name: %v", err)), nil
			}

			err = client.DeleteStore(ctx, storeID, force)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Deleted store: %s", storeID)), nil
		})
	}

	// Tool: import_file_to_store
	if isToolEnabled("import_file_to_store") || isToolEnabled("all") {
		s.AddTool(mcp.NewTool("import_file_to_store",
			mcp.WithDescription("Import a file from the Files API into a File Search Store."),
			mcp.WithString("file_name", mcp.Required(), mcp.Description("The resource name or display name of the file to import.")),
			mcp.WithString("store_name", mcp.Required(), mcp.Description("The resource name or display name of the store to import into.")),
		), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, ok := request.Params.Arguments.(map[string]interface{})
			if !ok {
				return mcp.NewToolResultError("arguments must be a map"), nil
			}
			fileName, ok := getStringArg(args, "file_name")
			if !ok {
				return mcp.NewToolResultError("file_name must be a string"), nil
			}
			storeName, ok := getStringArg(args, "store_name")
			if !ok {
				return mcp.NewToolResultError("store_name must be a string"), nil
			}

			// Resolve file name
			fileID, err := client.ResolveFileName(ctx, fileName)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to resolve file name: %v", err)), nil
			}

			// Resolve store name
			storeID, err := client.ResolveStoreName(ctx, storeName)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to resolve store name: %v", err)), nil
			}

			// Note: ImportFile now returns error only, but prints progress to stdout if not quiet.
			// Since we are in MCP, we can't easily stream progress.
			// We'll use Quiet=true to avoid stdout noise and just wait for completion.
			err = client.ImportFile(ctx, fileID, storeID, &gemini.ImportFileOptions{Quiet: true})
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Imported file %s into store %s", fileID, storeID)), nil
		})
	}

	// Tool: query_knowledge_base
	if isToolEnabled("query_knowledge_base") || isToolEnabled("query") || isToolEnabled("all") {
		s.AddTool(mcp.NewTool("query_knowledge_base",
			mcp.WithDescription("Query the knowledge base using Gemini File Search. Use this to answer questions based on uploaded documents."),
			mcp.WithString("query", mcp.Required(), mcp.Description("The question or query to ask.")),
			mcp.WithString("store_name", mcp.Description("The resource name or display name of the store to search. If omitted, searches all stores (if supported) or requires specific configuration.")),
			mcp.WithString("model", mcp.Description("The model to use (default: "+constants.DefaultModel+").")),
			mcp.WithString("metadata_filter", mcp.Description("Optional metadata filter expression to narrow search results. Examples: 'category = \"research\"' for exact match, 'status = \"reviewed\" AND priority = \"high\"' for multiple conditions, 'author = \"Smith\"' for filtering by author metadata.")),
		), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, ok := request.Params.Arguments.(map[string]interface{})
			if !ok {
				return mcp.NewToolResultError("arguments must be a map"), nil
			}
			query, ok := getStringArg(args, "query")
			if !ok {
				return mcp.NewToolResultError("query must be a string"), nil
			}
			storeName, _ := getStringArg(args, "store_name")
			model, _ := getStringArg(args, "model")
			if model == "" {
				model = constants.DefaultModel
			}
			metadataFilter, _ := getStringArg(args, "metadata_filter")

			var storeID string
			var err error
			if storeName != "" {
				storeID, err = client.ResolveStoreName(ctx, storeName)
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("Failed to resolve store name: %v", err)), nil
				}
			}

			resp, err := client.Query(ctx, query, storeID, model, metadataFilter)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			res, err := mcp.NewToolResultJSON(resp)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return res, nil
		})
	}

	// Tool: upload_file
	if isToolEnabled("upload_file") || isToolEnabled("upload") || isToolEnabled("all") {
		s.AddTool(mcp.NewTool("upload_file",
			mcp.WithDescription("Upload a local file to Gemini Files API and optionally add it to a store."),
			mcp.WithString("path", mcp.Required(), mcp.Description("Absolute path to the local file.")),
			mcp.WithString("store_name", mcp.Description("The resource name or display name of the store to add the file to.")),
			mcp.WithString("mime_type", mcp.Description("The MIME type of the file (optional).")),
			mcp.WithString("metadata", mcp.Description("Optional metadata as a JSON string. Examples: '{\"category\": \"research\", \"author\": \"Smith\"}' for multiple fields, '{\"status\": \"draft\"}' for single field, '{\"project\": \"Q4-2024\", \"priority\": \"high\"}' for project tracking. Only used if store_name is provided.")),
		), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, ok := request.Params.Arguments.(map[string]interface{})
			if !ok {
				return mcp.NewToolResultError("arguments must be a map"), nil
			}
			path, ok := getStringArg(args, "path")
			if !ok {
				return mcp.NewToolResultError("path must be a string"), nil
			}
			storeName, _ := getStringArg(args, "store_name")
			mimeType, _ := getStringArg(args, "mime_type")
			metadataJSON, _ := getStringArg(args, "metadata")

			var metadata map[string]string
			if metadataJSON != "" {
				// Try to parse as JSON map[string]string
				if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("Failed to parse metadata JSON: %v", err)), nil
				}
			}

			var storeID string
			var err error
			if storeName != "" {
				storeID, err = client.ResolveStoreName(ctx, storeName)
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("Failed to resolve store name: %v", err)), nil
				}
			}

			opts := &gemini.UploadFileOptions{
				StoreName: storeID,
				MIMEType:  mimeType,
				Metadata:  metadata,
				Quiet:     true, // Suppress stdout progress
			}

			file, err := client.UploadFile(ctx, path, opts)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// If file is nil, it means it was uploaded to a store (UploadFile returns nil for store uploads as it handles the operation)
			if file == nil {
				return mcp.NewToolResultText(fmt.Sprintf("Uploaded %s to store %s", path, storeName)), nil
			}

			res, err := mcp.NewToolResultJSON(file)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return res, nil
		})
	}

	// Tool: delete_file
	if isToolEnabled("delete_file") || isToolEnabled("delete") || isToolEnabled("all") {
		s.AddTool(mcp.NewTool("delete_file",
			mcp.WithDescription("Delete a file from the Gemini Files API."),
			mcp.WithString("file_name", mcp.Required(), mcp.Description("The resource name or display name of the file to delete.")),
		), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, ok := request.Params.Arguments.(map[string]interface{})
			if !ok {
				return mcp.NewToolResultError("arguments must be a map"), nil
			}
			fileName, ok := getStringArg(args, "file_name")
			if !ok {
				return mcp.NewToolResultError("file_name must be a string"), nil
			}

			fileID, err := client.ResolveFileName(ctx, fileName)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to resolve file name: %v", err)), nil
			}

			err = client.DeleteFile(ctx, fileID)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Deleted file: %s", fileID)), nil
		})
	}

	// Tool: delete_document
	if isToolEnabled("delete_document") || isToolEnabled("delete") || isToolEnabled("all") {
		s.AddTool(mcp.NewTool("delete_document",
			mcp.WithDescription("Delete a document from a File Search Store."),
			mcp.WithString("store_name", mcp.Required(), mcp.Description("The resource name or display name of the store.")),
			mcp.WithString("document_name", mcp.Required(), mcp.Description("The resource name or display name of the document.")),
		), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, ok := request.Params.Arguments.(map[string]interface{})
			if !ok {
				return mcp.NewToolResultError("arguments must be a map"), nil
			}
			storeName, ok := getStringArg(args, "store_name")
			if !ok {
				return mcp.NewToolResultError("store_name must be a string"), nil
			}
			docName, ok := getStringArg(args, "document_name")
			if !ok {
				return mcp.NewToolResultError("document_name must be a string"), nil
			}

			// Resolve store
			storeID, err := client.ResolveStoreName(ctx, storeName)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to resolve store name: %v", err)), nil
			}

			// Resolve document
			docID, err := client.ResolveDocumentName(ctx, storeID, docName)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to resolve document name: %v", err)), nil
			}

			err = client.DeleteDocument(ctx, docID)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Deleted document: %s from store %s", docID, storeID)), nil
		})
	}

	return s
}
