package cmd

import (
	"context"
	"fmt"

	"github.com/mikesmitty/file-search/internal/constants"
	"github.com/mikesmitty/file-search/internal/gemini"
	"github.com/spf13/cobra"
)

var storeCmd = &cobra.Command{
	Use:   "store",
	Short: "Manage File Search Stores",
}

func init() {
	rootCmd.AddCommand(storeCmd)

	// Store list
	storeCmd.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all File Search Stores",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
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

	// Store get
	storeCmd.AddCommand(&cobra.Command{
		Use:   "get [name]",
		Short: "Get details of a File Search Store",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return getCompleter().GetStoreNames(), cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
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

	// Store delete
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
			ctx := context.Background()
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

	// Store create
	storeCmd.AddCommand(&cobra.Command{
		Use:     "create [display_name]",
		Aliases: []string{"new", "add"},
		Short:   "Create a new File Search Store",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
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

	// Store import-file
	var importFileStore string
	var importFileStoreID string
	var importConcurrency int
	importFileCmd := &cobra.Command{
		Use:   "import-file [file-name-or-id]...",
		Short: "Import files from Files API into a Store",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if importFileStore == "" && importFileStoreID == "" {
				return fmt.Errorf("either --store or --store-id is required")
			}
			ctx := context.Background()
			client, err := getClient(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			// Resolve store name to ID if --store was used
			storeID := importFileStoreID
			if importFileStore != "" {
				storeID, err = client.ResolveStoreName(ctx, importFileStore)
				if err != nil {
					return err
				}
			}

			// Define the processor function for a single file ID/name
			processor := func(ctx context.Context, fileIDOrName string) error {
				// Resolve file name to ID
				fileID, err := client.ResolveFileName(ctx, fileIDOrName)
				if err != nil {
					return err
				}
				
				if !quiet {
					fmt.Printf("[+] Starting import: %s\n", fileIDOrName)
				}

				err = client.ImportFile(ctx, fileID, storeID, &gemini.ImportFileOptions{
					Quiet: true, // Force quiet for inner operation
				})
				return err
			}

			// Define the progress callback
			onProgress := func(current, total int, file string, err error) {
				if err != nil {
					fmt.Printf("[%d/%d] ✗ Failed: %s (%v)\n", current, total, file, err)
				} else {
					fmt.Printf("[%d/%d] ✓ Finished: %s\n", current, total, file)
				}
			}

			// Process files using the batch processor
			batchResult := processBatch(ctx, args, processor, &BatchOptions{
				Concurrency: importConcurrency,
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
					filesSummary = append(filesSummary, map[string]interface{}{"file": f, "status": "success", "store": storeID})
				}
				for f, err := range batchResult.Failed {
					filesSummary = append(filesSummary, map[string]interface{}{"file": f, "status": "failed", "error": err.Error(), "store": storeID})
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
					return fmt.Errorf("some files failed to import")
				}
				if !quiet && len(args) == 1 && len(batchResult.Succeeded) == 1 {
					// If single file and succeeded, print success message
					fmt.Printf("Imported file: %s to store: %s\n", batchResult.Succeeded[0], storeID)
				}
			}
			return nil
		},
	}
	importFileCmd.Flags().StringVar(&importFileStore, "store", "", "Store display name")
	importFileCmd.Flags().StringVar(&importFileStoreID, "store-id", "", "Store resource ID ("+constants.StoreResourcePrefix+"xxx)")
	importFileCmd.Flags().IntVar(&importConcurrency, "concurrency", 5, "Number of parallel imports")
	importFileCmd.RegisterFlagCompletionFunc("store", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getCompleter().GetStoreNames(), cobra.ShellCompDirectiveNoFileComp
	})
	importFileCmd.RegisterFlagCompletionFunc("store-id", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getCompleter().GetStoreNames(), cobra.ShellCompDirectiveNoFileComp
	})
	storeCmd.AddCommand(importFileCmd)
}
