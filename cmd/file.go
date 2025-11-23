package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mikesmitty/file-search/internal/constants"
	"github.com/mikesmitty/file-search/internal/gemini"
	"github.com/spf13/cobra"
)

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "Manage Files",
}

func init() {
	rootCmd.AddCommand(fileCmd)

	// File list
	fileCmd.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List uploaded files",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
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

	// File get
	fileCmd.AddCommand(&cobra.Command{
		Use:   "get [name]",
		Short: "Get details of a file",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return getCompleter().GetFileNames(), cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
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

	// File delete
	fileCmd.AddCommand(&cobra.Command{
		Use:     "delete [name]",
		Aliases: []string{"rm", "del"},
		Short:   "Delete a file",
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return getCompleter().GetFileNames(), cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
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

	// File upload
	var uploadStoreName string
	var uploadStoreID string
	var uploadDisplayName string
	var uploadMimeType string
	var uploadChunkSize int
	var uploadChunkOverlap int
	var uploadMetadata []string
	var uploadConcurrency int
	uploadCmd := &cobra.Command{
		Use:   "upload [path]...",
		Short: "Upload and import files",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := getClient(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			if len(args) > 1 && uploadDisplayName != "" {
				return fmt.Errorf("cannot use --name with multiple files")
			}

			// Parse metadata from key=value strings
			metadataMap := make(map[string]string)
			for _, meta := range uploadMetadata {
				parts := strings.SplitN(meta, "=", 2)
				if len(parts) == 2 {
					metadataMap[parts[0]] = parts[1]
				}
			}

			// Resolve store name to ID if --store was used
			storeID := uploadStoreID
			if uploadStoreName != "" {
				storeID, err = client.ResolveStoreName(ctx, uploadStoreName)
				if err != nil {
					return err
				}
			}

			// Define the processor function for a single file
			processor := func(ctx context.Context, path string) error {
				displayName := uploadDisplayName
				if displayName == "" {
					displayName = filepath.Base(path)
				}

				if !quiet {
					fmt.Printf("[+] Starting upload: %s\n", displayName)
				}

				opts := &gemini.UploadFileOptions{
					StoreName:      storeID,
					DisplayName:    displayName,
					MIMEType:       uploadMimeType,
					MaxChunkTokens: uploadChunkSize,
					ChunkOverlap:   uploadChunkOverlap,
					Metadata:       metadataMap,
					Quiet:          true, // Force quiet for inner operation to prevent output interleaving
				}
				_, err := client.UploadFile(ctx, path, opts)
				return err
			}

			// Define the progress callback
			onProgress := func(current, total int, file string, err error) {
				if err != nil {
					fmt.Printf("[%d/%d] ✗ Failed: %s (%v)\n", current, total, filepath.Base(file), err)
				} else {
					fmt.Printf("[%d/%d] ✓ Finished: %s\n", current, total, filepath.Base(file))
				}
			}

			// Process files using the batch processor
			batchResult := processBatch(ctx, args, processor, &BatchOptions{
				Concurrency: uploadConcurrency,
				Quiet:       quiet,
				OnProgress:  onProgress,
			})

			// Print summary
			if !quiet {
				if len(args) > 1 { // Only print summary if multiple files were processed
					fmt.Printf("\n\nSummary:\n")
					fmt.Printf("  ✓ Succeeded: %d\n", len(batchResult.Succeeded))
					fmt.Printf("  ✗ Failed: %d\n", len(batchResult.Failed))
				}
			}

			if outputFormat == "json" {
				// For JSON, aggregate results
				jsonResult := make(map[string]interface{})
				jsonResult["total"] = batchResult.Total
				jsonResult["succeeded"] = len(batchResult.Succeeded)
				jsonResult["failed"] = len(batchResult.Failed)

				filesSummary := make([]map[string]interface{}, 0, batchResult.Total)
				for _, f := range batchResult.Succeeded {
					filesSummary = append(filesSummary, map[string]interface{}{"file": f, "status": "success"})
				}
				for f, err := range batchResult.Failed {
					filesSummary = append(filesSummary, map[string]interface{}{"file": f, "status": "failed", "error": err.Error()})
				}
				jsonResult["files"] = filesSummary
				return printOutput(jsonResult, "json")

			} else { // Text output
				if len(batchResult.Failed) > 0 {
					if !quiet {
						fmt.Printf("\nFailed files:\n")
						for f, err := range batchResult.Failed {
							fmt.Printf("  - %s: %v\n", f, err)
						}
					}
					return fmt.Errorf("some files failed to upload")
				}
				if !quiet && len(args) == 1 && len(batchResult.Succeeded) == 1 {
					// If single file and succeeded, print success message
					fmt.Printf("Uploaded file: %s\n", batchResult.Succeeded[0])
				}
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
	uploadCmd.Flags().IntVar(&uploadConcurrency, "concurrency", 5, "Number of parallel uploads")
	uploadCmd.RegisterFlagCompletionFunc("store", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getCompleter().GetStoreNames(), cobra.ShellCompDirectiveNoFileComp
	})
	uploadCmd.RegisterFlagCompletionFunc("store-id", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getCompleter().GetStoreNames(), cobra.ShellCompDirectiveNoFileComp
	})
	fileCmd.AddCommand(uploadCmd)
}
