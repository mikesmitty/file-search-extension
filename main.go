package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mikesmitty/file-search-extension/internal/completion"
	"github.com/mikesmitty/file-search-extension/internal/constants"
	"github.com/mikesmitty/file-search-extension/internal/gemini"
	"github.com/mikesmitty/file-search-extension/internal/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/genai"
)

var (
	cfgFile      string
	apiKey       string
	apiKeyEnv    string
	outputFormat string
	mcpTools     string
	quiet        bool

	// Build info
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "file-search",
	Short: "Gemini File Search & MCP Tool",
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.file-search.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Gemini API Key")
	rootCmd.PersistentFlags().StringVar(&apiKeyEnv, "api-key-env", "", "Environment variable to read API Key from")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "format", "text", "Output format: text or json")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress progress indicators")

	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("api_key_env", rootCmd.PersistentFlags().Lookup("api-key-env"))
}

// getMCPTools returns the list of enabled MCP tools
// Supports comma-separated string from flag/env/config
// Default: ["query"]
func getMCPTools() []string {
	// Check if set via flag/env/config
	toolsStr := viper.GetString("mcp_tools")
	if toolsStr == "" {
		// Default to query only
		return []string{"query"}
	}

	// Parse comma-separated list
	tools := strings.Split(toolsStr, ",")
	result := make([]string, 0, len(tools))
	for _, tool := range tools {
		trimmed := strings.TrimSpace(tool)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	if len(result) == 0 {
		return []string{"query"}
	}
	return result
}

var globalCompleter *completion.Completer

// getCompleter returns or initializes the global completer instance
func getCompleter() *completion.Completer {
	if globalCompleter != nil {
		return globalCompleter
	}

	// Get configuration
	enabled := viper.GetBool("completion_enabled")
	cacheTTL := viper.GetDuration("completion_cache_ttl")
	if cacheTTL == 0 {
		cacheTTL = 300 * time.Second // 5 minutes default
	}

	// Get API key
	key, err := getAPIKey()
	if err != nil || key == "" {
		// If no API key, create disabled completer
		globalCompleter = completion.NewCompleter("", false, cacheTTL)
		return globalCompleter
	}

	// Create completer with configuration
	globalCompleter = completion.NewCompleter(key, enabled, cacheTTL)
	return globalCompleter
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".file-search")
	}

	// Set defaults
	viper.SetDefault("completion_enabled", true)
	viper.SetDefault("completion_cache_ttl", "300s")
	viper.SetDefault("mcp_tools", "query")

	// Bind environment variables
	viper.BindEnv("api_key", "GOOGLE_API_KEY", "GEMINI_API_KEY")
	viper.BindEnv("mcp_tools", "MCP_TOOLS")
	viper.BindEnv("completion_enabled", "COMPLETION_ENABLED")
	viper.BindEnv("completion_cache_ttl", "COMPLETION_CACHE_TTL")

	if err := viper.ReadInConfig(); err == nil {
		// fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func getAPIKey() (string, error) {
	// 1. Check if a custom env var is specified
	if envVar := viper.GetString("api_key_env"); envVar != "" {
		if key := os.Getenv(envVar); key != "" {
			return key, nil
		}
	}

	// 2. Check standard config/env
	key := viper.GetString("api_key")
	if key == "" {
		return "", fmt.Errorf("API key not set. Use --api-key, --api-key-env, config file, or GOOGLE_API_KEY/GEMINI_API_KEY")
	}
	return key, nil
}

func getClient(ctx context.Context) (*gemini.Client, error) {
	key, err := getAPIKey()
	if err != nil {
		return nil, err
	}
	return gemini.NewClient(ctx, key)
}

// printOutput handles formatting and printing of results
func printOutput(data interface{}, format string) error {
	if format == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	// Text formatting based on type
	switch v := data.(type) {
	case []*genai.FileSearchStore:
		for _, s := range v {
			fmt.Printf("%s (%s)\n", s.DisplayName, s.Name)
		}
	case *genai.FileSearchStore:
		fmt.Printf("Name: %s\n", v.Name)
		fmt.Printf("Display Name: %s\n", v.DisplayName)
		fmt.Printf("Create Time: %s\n", v.CreateTime)
		fmt.Printf("Update Time: %s\n", v.UpdateTime)
		fmt.Printf("Active Documents: %d\n", v.ActiveDocumentsCount)
		fmt.Printf("Pending Documents: %d\n", v.PendingDocumentsCount)
		fmt.Printf("Failed Documents: %d\n", v.FailedDocumentsCount)
		fmt.Printf("Total Size: %d bytes\n", v.SizeBytes)
	case []*genai.File:
		for _, f := range v {
			fmt.Printf("%s (%s) - %s\n", f.DisplayName, f.Name, f.URI)
		}
	case *genai.File:
		fmt.Printf("Name: %s\n", v.Name)
		fmt.Printf("Display Name: %s\n", v.DisplayName)
		fmt.Printf("URI: %s\n", v.URI)
		fmt.Printf("MIME Type: %s\n", v.MIMEType)
		fmt.Printf("Size: %d bytes\n", v.SizeBytes)
		fmt.Printf("Create Time: %s\n", v.CreateTime)
		fmt.Printf("Update Time: %s\n", v.UpdateTime)
		fmt.Printf("State: %s\n", v.State)
	case []*genai.Document:
		for _, doc := range v {
			fmt.Printf("%s (%s) - %s - %d bytes\n", doc.DisplayName, doc.Name, doc.State, doc.SizeBytes)
		}
	case *genai.Document:
		fmt.Printf("Name: %s\n", v.Name)
		fmt.Printf("Display Name: %s\n", v.DisplayName)
		fmt.Printf("State: %s\n", v.State)
		fmt.Printf("Size: %d bytes\n", v.SizeBytes)
		fmt.Printf("MIME Type: %s\n", v.MIMEType)
		fmt.Printf("Create Time: %s\n", v.CreateTime)
		fmt.Printf("Update Time: %s\n", v.UpdateTime)
		if len(v.CustomMetadata) > 0 {
			fmt.Println("Custom Metadata:")
			for _, meta := range v.CustomMetadata {
				fmt.Printf("  %s: %s\n", meta.Key, meta.StringValue)
			}
		}
	case *genai.GenerateContentResponse:
		for _, cand := range v.Candidates {
			for _, part := range cand.Content.Parts {
				fmt.Printf("%v\n", part.Text)
			}
			if cand.GroundingMetadata != nil {
				fmt.Printf("\n[Grounding Metadata Found]\n")
			}
		}
	case *gemini.OperationStatus:
		fmt.Printf("Operation: %s\n", v.Name)
		fmt.Printf("Type: %s\n", v.Type)

		if v.Failed {
			fmt.Printf("Status: FAILED\n")
			fmt.Printf("Error: %s\n", v.ErrorMessage)
		} else if v.Done {
			fmt.Printf("Status: DONE\n")
			if v.Parent != "" {
				fmt.Printf("Store: %s\n", v.Parent)
			}
			if v.DocumentName != "" {
				fmt.Printf("Document: %s\n", v.DocumentName)
			}
		} else {
			fmt.Printf("Status: PENDING\n")
		}

		if len(v.Metadata) > 0 {
			fmt.Println("\nMetadata:")
			for k, val := range v.Metadata {
				fmt.Printf("  %s: %v\n", k, val)
			}
		}
	default:
		// Fallback for simple strings or unknown types
		fmt.Printf("%v\n", v)
	}
	return nil
}

func main() {
	ctx := context.Background()

	// Store Commands
	var storeCmd = &cobra.Command{
		Use:   "store",
		Short: "Manage File Search Stores",
	}

	storeCmd.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all File Search Stores",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(ctx)
			if err != nil {
				return err
			}
			defer client.Close()
			stores, err := client.ListStores(ctx)
			if err != nil {
				return err
			}
			return printOutput(stores, outputFormat)
		},
	})

	storeCmd.AddCommand(&cobra.Command{
		Use:   "get [name]",
		Short: "Get details of a File Search Store",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return getCompleter().GetStoreNames(), cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			// Resolve store name to ID
			storeID, err := client.ResolveStoreName(ctx, args[0])
			if err != nil {
				return err
			}

			store, err := client.GetStore(ctx, storeID)
			if err != nil {
				return err
			}
			return printOutput(store, outputFormat)
		},
	})

	var deleteStoreForce bool
	deleteStoreCmd := &cobra.Command{
		Use:     "delete [name]",
		Aliases: []string{"rm", "del"},
		Short:   "Delete a File Search Store",
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return getCompleter().GetStoreNames(), cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			// Resolve store name to ID
			storeID, err := client.ResolveStoreName(ctx, args[0])
			if err != nil {
				return err
			}

			err = client.DeleteStore(ctx, storeID, deleteStoreForce)
			if err != nil {
				return err
			}
			if outputFormat == "json" {
				return printOutput(map[string]string{"status": "deleted", "name": args[0]}, "json")
			}
			fmt.Printf("Deleted store: %s\n", args[0])
			return nil
		},
	}
	deleteStoreCmd.Flags().BoolVar(&deleteStoreForce, "force", false, "Force delete even if store contains documents")
	storeCmd.AddCommand(deleteStoreCmd)

	storeCmd.AddCommand(&cobra.Command{
		Use:     "create [display_name]",
		Aliases: []string{"new"},
		Short:   "Create a new File Search Store",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(ctx)
			if err != nil {
				return err
			}
			defer client.Close()
			store, err := client.CreateStore(ctx, args[0])
			if err != nil {
				return err
			}
			if outputFormat == "json" {
				return printOutput(store, "json")
			}
			fmt.Printf("Created store: %s (%s)\n", store.DisplayName, store.Name)
			return nil
		},
	})

	var importFileStore string
	var importFileStoreID string
	importFileCmd := &cobra.Command{
		Use:   "import-file [file-name-or-id]",
		Short: "Import a file from Files API into a Store",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if importFileStore == "" && importFileStoreID == "" {
				return fmt.Errorf("either --store or --store-id is required")
			}
			client, err := getClient(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			// Resolve file name to ID
			fileID, err := client.ResolveFileName(ctx, args[0])
			if err != nil {
				return err
			}

			// Resolve store name to ID if --store was used
			storeID := importFileStoreID
			if importFileStore != "" {
				storeID, err = client.ResolveStoreName(ctx, importFileStore)
				if err != nil {
					return err
				}
			}

			err = client.ImportFile(ctx, fileID, storeID, &gemini.ImportFileOptions{
				Quiet: quiet,
			})
			if err != nil {
				return err
			}
			if outputFormat == "json" {
				return printOutput(map[string]string{"status": "imported", "file": fileID, "store": storeID}, "json")
			}
			// ImportFile already prints progress if not quiet, but we can add a final success message if needed
			// The client.ImportFile method prints "Import complete" so we are good.
			return nil
		},
	}
	importFileCmd.Flags().StringVar(&importFileStore, "store", "", "Store display name")
	importFileCmd.Flags().StringVar(&importFileStoreID, "store-id", "", "Store resource ID ("+constants.StoreResourcePrefix+"xxx)")
	importFileCmd.RegisterFlagCompletionFunc("store", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getCompleter().GetStoreNames(), cobra.ShellCompDirectiveNoFileComp
	})
	importFileCmd.RegisterFlagCompletionFunc("store-id", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getCompleter().GetStoreNames(), cobra.ShellCompDirectiveNoFileComp
	})
	storeCmd.AddCommand(importFileCmd)

	// File Commands
	var fileCmd = &cobra.Command{
		Use:   "file",
		Short: "Manage Files",
	}

	fileCmd.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List uploaded files",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(ctx)
			if err != nil {
				return err
			}
			defer client.Close()
			files, err := client.ListFiles(ctx)
			if err != nil {
				return err
			}
			return printOutput(files, outputFormat)
		},
	})

	fileCmd.AddCommand(&cobra.Command{
		Use:   "get [name]",
		Short: "Get details of a file",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return getCompleter().GetFileNames(), cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			// Resolve file name to ID
			fileID, err := client.ResolveFileName(ctx, args[0])
			if err != nil {
				return err
			}

			file, err := client.GetFile(ctx, fileID)
			if err != nil {
				return err
			}
			return printOutput(file, outputFormat)
		},
	})

	fileCmd.AddCommand(&cobra.Command{
		Use:     "delete [name]",
		Aliases: []string{"rm", "del"},
		Short:   "Delete a file",
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return getCompleter().GetFileNames(), cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			// Resolve file name to ID
			fileID, err := client.ResolveFileName(ctx, args[0])
			if err != nil {
				return err
			}

			err = client.DeleteFile(ctx, fileID)
			if err != nil {
				return err
			}
			if outputFormat == "json" {
				return printOutput(map[string]string{"status": "deleted", "file": fileID}, "json")
			}
			fmt.Printf("Deleted file: %s\n", args[0])
			return nil
		},
	})

	var uploadStoreName string
	var uploadStoreID string
	var uploadDisplayName string
	var uploadMimeType string
	var uploadChunkSize int
	var uploadChunkOverlap int
	var uploadMetadata []string
	uploadCmd := &cobra.Command{
		Use:   "upload [path]",
		Short: "Upload and import a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			// Parse metadata from key=value strings
			metadataMap := make(map[string]string)
			for _, meta := range uploadMetadata {
				parts := strings.SplitN(meta, "=", 2)
				if len(parts) == 2 {
					metadataMap[parts[0]] = parts[1]
				}
			}

			// Auto-set display name from filename if not provided
			displayName := uploadDisplayName
			if displayName == "" {
				displayName = filepath.Base(args[0])
			}

			// Resolve store name to ID if --store was used
			storeID := uploadStoreID
			if uploadStoreName != "" {
				storeID, err = client.ResolveStoreName(ctx, uploadStoreName)
				if err != nil {
					return err
				}
			}

			opts := &gemini.UploadFileOptions{
				StoreName:      storeID,
				DisplayName:    displayName,
				MIMEType:       uploadMimeType,
				MaxChunkTokens: uploadChunkSize,
				ChunkOverlap:   uploadChunkOverlap,
				Metadata:       metadataMap,
				Quiet:          quiet,
			}

			file, err := client.UploadFile(ctx, args[0], opts)
			if err != nil {
				return err
			}

			if outputFormat == "json" {
				if file != nil {
					return printOutput(file, "json")
				}
				return printOutput(map[string]string{"status": "uploaded_and_indexed"}, "json")
			}

			if file != nil {
				fmt.Printf("Uploaded file: %s (URI: %s)\n", file.DisplayName, file.URI)
			}
			return nil
		},
	}
	uploadCmd.Flags().StringVar(&uploadStoreName, "store", "", "Store display name (optional)")
	uploadCmd.Flags().StringVar(&uploadStoreID, "store-id", "", "Store resource ID (optional, "+constants.StoreResourcePrefix+"xxx)")
	uploadCmd.Flags().StringVar(&uploadDisplayName, "name", "", "Display name (optional)")
	uploadCmd.Flags().StringVar(&uploadMimeType, "mime-type", "", "MIME type (optional, e.g. text/plain, application/pdf)")
	uploadCmd.Flags().IntVar(&uploadChunkSize, "chunk-size", 0, "Max tokens per chunk (for store uploads)")
	uploadCmd.Flags().IntVar(&uploadChunkOverlap, "chunk-overlap", 0, "Overlap tokens between chunks (for store uploads)")
	uploadCmd.Flags().StringArrayVar(&uploadMetadata, "metadata", []string{}, "Custom metadata as key=value (repeatable, for store uploads)")
	uploadCmd.RegisterFlagCompletionFunc("store", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getCompleter().GetStoreNames(), cobra.ShellCompDirectiveNoFileComp
	})
	uploadCmd.RegisterFlagCompletionFunc("store-id", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getCompleter().GetStoreNames(), cobra.ShellCompDirectiveNoFileComp
	})
	fileCmd.AddCommand(uploadCmd)

	// Document Commands
	var documentCmd = &cobra.Command{
		Use:     "document",
		Aliases: []string{"doc"},
		Short:   "Manage Documents in Stores",
	}

	var docListStore string
	var docListStoreID string
	docListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List documents in a store",
		RunE: func(cmd *cobra.Command, args []string) error {
			if docListStore == "" && docListStoreID == "" {
				return fmt.Errorf("either --store or --store-id is required")
			}
			client, err := getClient(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			// Resolve store name to ID if --store was used
			storeID := docListStoreID
			if docListStore != "" {
				storeID, err = client.ResolveStoreName(ctx, docListStore)
				if err != nil {
					return err
				}
			}

			docs, err := client.ListDocuments(ctx, storeID)
			if err != nil {
				return err
			}
			return printOutput(docs, outputFormat)
		},
	}
	docListCmd.Flags().StringVar(&docListStore, "store", "", "Store display name")
	docListCmd.Flags().StringVar(&docListStoreID, "store-id", "", "Store resource ID ("+constants.StoreResourcePrefix+"xxx)")
	docListCmd.RegisterFlagCompletionFunc("store", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getCompleter().GetStoreNames(), cobra.ShellCompDirectiveNoFileComp
	})
	docListCmd.RegisterFlagCompletionFunc("store-id", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getCompleter().GetStoreNames(), cobra.ShellCompDirectiveNoFileComp
	})
	documentCmd.AddCommand(docListCmd)

	var docGetStore string
	var docGetStoreID string
	docGetCmd := &cobra.Command{
		Use:   "get [name]",
		Short: "Get document details",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			storeFlag, _ := cmd.Flags().GetString("store")
			if storeFlag != "" {
				return getCompleter().GetDocumentNames(storeFlag), cobra.ShellCompDirectiveNoFileComp
			}
			storeIDFlag, _ := cmd.Flags().GetString("store-id")
			if storeIDFlag != "" {
				return getCompleter().GetDocumentNames(storeIDFlag), cobra.ShellCompDirectiveNoFileComp
			}
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			// If store is provided, resolve document name within that store
			docID := args[0]
			if docGetStore != "" || docGetStoreID != "" {
				storeRef := docGetStoreID
				if docGetStore != "" {
					storeRef = docGetStore
				}
				docID, err = client.ResolveDocumentName(ctx, storeRef, args[0])
				if err != nil {
					return err
				}
			}

			doc, err := client.GetDocument(ctx, docID)
			if err != nil {
				return err
			}
			return printOutput(doc, outputFormat)
		},
	}
	docGetCmd.Flags().StringVar(&docGetStore, "store", "", "Store display name (optional, for name resolution)")
	docGetCmd.Flags().StringVar(&docGetStoreID, "store-id", "", "Store resource ID (optional, for name resolution)")
	docGetCmd.RegisterFlagCompletionFunc("store", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getCompleter().GetStoreNames(), cobra.ShellCompDirectiveNoFileComp
	})
	docGetCmd.RegisterFlagCompletionFunc("store-id", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getCompleter().GetStoreNames(), cobra.ShellCompDirectiveNoFileComp
	})
	documentCmd.AddCommand(docGetCmd)

	var docDelStore string
	var docDelStoreID string
	docDelCmd := &cobra.Command{
		Use:     "delete [name]",
		Aliases: []string{"rm", "del"},
		Short:   "Delete a document",
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			storeFlag, _ := cmd.Flags().GetString("store")
			if storeFlag != "" {
				return getCompleter().GetDocumentNames(storeFlag), cobra.ShellCompDirectiveNoFileComp
			}
			storeIDFlag, _ := cmd.Flags().GetString("store-id")
			if storeIDFlag != "" {
				return getCompleter().GetDocumentNames(storeIDFlag), cobra.ShellCompDirectiveNoFileComp
			}
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			// If store is provided, resolve document name within that store
			docID := args[0]
			if docDelStore != "" || docDelStoreID != "" {
				storeRef := docDelStoreID
				if docDelStore != "" {
					storeRef = docDelStore
				}
				docID, err = client.ResolveDocumentName(ctx, storeRef, args[0])
				if err != nil {
					return err
				}
			}

			err = client.DeleteDocument(ctx, docID)
			if err != nil {
				return err
			}
			if outputFormat == "json" {
				return printOutput(map[string]string{"status": "deleted", "document": docID}, "json")
			}
			fmt.Printf("Deleted document: %s\n", args[0])
			return nil
		},
	}
	docDelCmd.Flags().StringVar(&docDelStore, "store", "", "Store display name (optional, for name resolution)")
	docDelCmd.Flags().StringVar(&docDelStoreID, "store-id", "", "Store resource ID (optional, for name resolution)")
	docDelCmd.RegisterFlagCompletionFunc("store", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getCompleter().GetStoreNames(), cobra.ShellCompDirectiveNoFileComp
	})
	docDelCmd.RegisterFlagCompletionFunc("store-id", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getCompleter().GetStoreNames(), cobra.ShellCompDirectiveNoFileComp
	})
	documentCmd.AddCommand(docDelCmd)

	// Query Command
	var queryStoreName string
	var queryStoreID string
	var queryModel string
	var queryMetadataFilter string
	var queryCmd = &cobra.Command{
		Use:   "query [text]",
		Short: "Query with optional file search",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			// Resolve store name to ID if --store was used
			storeID := queryStoreID
			if queryStoreName != "" {
				storeID, err = client.ResolveStoreName(ctx, queryStoreName)
				if err != nil {
					return err
				}
			}

			if queryModel == "" {
				queryModel = constants.DefaultModel
			}
			resp, err := client.Query(ctx, args[0], storeID, queryModel, queryMetadataFilter)
			if err != nil {
				return err
			}
			return printOutput(resp, outputFormat)
		},
	}
	queryCmd.Flags().StringVar(&queryStoreName, "store", "", "Store display name (optional)")
	queryCmd.Flags().StringVar(&queryStoreID, "store-id", "", "Store resource ID (optional, "+constants.StoreResourcePrefix+"xxx)")
	queryCmd.Flags().StringVar(&queryModel, "model", constants.DefaultModel, "Model name")
	queryCmd.Flags().StringVar(&queryMetadataFilter, "metadata-filter", "", "Metadata filter expression (optional)")
	queryCmd.RegisterFlagCompletionFunc("store", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getCompleter().GetStoreNames(), cobra.ShellCompDirectiveNoFileComp
	})
	queryCmd.RegisterFlagCompletionFunc("store-id", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getCompleter().GetStoreNames(), cobra.ShellCompDirectiveNoFileComp
	})
	queryCmd.RegisterFlagCompletionFunc("model", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getCompleter().GetModelNames(), cobra.ShellCompDirectiveNoFileComp
	})

	// MCP Command
	mcpCmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start MCP Server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// For MCP, we start the server even without API key configured.
			// Tools will fail gracefully when invoked if auth is missing.
			key, _ := getAPIKey()

			var client *gemini.Client
			var err error
			if key != "" {
				client, err = gemini.NewClient(ctx, key)
				if err != nil {
					return err
				}
				defer client.Close()
			}

			tools := getMCPTools()
			return mcp.RunServer(ctx, client, tools)
		},
	}
	mcpCmd.Flags().StringVar(&mcpTools, "mcp-tools", "", "Comma-separated list of MCP tools to enable (default: query)")
	viper.BindPFlag("mcp_tools", mcpCmd.Flags().Lookup("mcp-tools"))
	viper.BindEnv("mcp_tools", "MCP_TOOLS")

	// Operation Commands
	var operationCmd = &cobra.Command{
		Use:     "operation",
		Aliases: []string{"op", "operations"},
		Short:   "Manage long-running operations",
	}

	var operationType string
	operationGetCmd := &cobra.Command{
		Use:   "get [operation-name]",
		Short: "Get the status of a long-running operation",
		Long: `Get the status of a long-running file upload or import operation.

Operation names follow the format: fileSearchStores/{store-id}/operations/{operation-id}

Examples:
  # Get operation status (auto-detect type)
  file-search operation get "fileSearchStores/abc123/operations/op456"

  # Get operation status with specific type
  file-search operation get "fileSearchStores/abc123/operations/op456" --type import

  # Get operation status in JSON format
  file-search operation get "fileSearchStores/abc123/operations/op456" --format json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			var opType gemini.OperationType
			switch operationType {
			case "import":
				opType = gemini.OperationTypeImport
			case "upload":
				opType = gemini.OperationTypeUpload
			case "":
				// Auto-detect (empty string is valid)
				opType = ""
			default:
				return fmt.Errorf("invalid operation type: %s (must be 'import' or 'upload')", operationType)
			}

			status, err := client.GetOperation(ctx, args[0], opType)
			if err != nil {
				return err
			}

			return printOutput(status, outputFormat)
		},
	}
	operationGetCmd.Flags().StringVar(&operationType, "type", "", "Operation type: import or upload (auto-detect if not specified)")
	operationCmd.AddCommand(operationGetCmd)

	rootCmd.AddCommand(storeCmd, fileCmd, documentCmd, queryCmd, operationCmd, mcpCmd)

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("file-search %s (%s) built at %s\n", version, commit, date)
		},
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
