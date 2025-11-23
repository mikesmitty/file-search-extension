package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/mikesmitty/file-search/internal/completion"
	"github.com/mikesmitty/file-search/internal/gemini"
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
	verbose      bool
	debug        bool

	// Build info - set by main package
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "file-search",
	Short: "File Search Query & MCP Tool",
	Long: `File Search Query is a CLI tool and Model Context Protocol (MCP) server that 
enables interaction with the Google Gemini File Search API.

It allows you to manage file stores, upload documents, and perform semantic searches.`,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.file-search.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Gemini API Key")
	rootCmd.PersistentFlags().StringVar(&apiKeyEnv, "api-key-env", "", "Environment variable to read API Key from")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "format", "text", "Output format: text or json")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress progress indicators")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug output (JSON)")

	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("api_key_env", rootCmd.PersistentFlags().Lookup("api-key-env"))
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
	viper.SetDefault("mcp_tools", "all")

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
				fmt.Printf("\n[Grounding Metadata]\n")

				// Debug output: Print full metadata as JSON if --debug is set
				if debug {
					debugJSON, err := json.MarshalIndent(cand.GroundingMetadata, "", "  ")
					if err == nil {
						fmt.Println(string(debugJSON))
					}
				}

				if len(cand.GroundingMetadata.GroundingChunks) > 0 {
					fmt.Println("\nSources:")
					for i, chunk := range cand.GroundingMetadata.GroundingChunks {
						if chunk.Web != nil {
							fmt.Printf("  %d. [Web] %s (%s)\n", i+1, chunk.Web.Title, chunk.Web.URI)
						} else if chunk.RetrievedContext != nil {
							title := chunk.RetrievedContext.Title
							if title == "" {
								title = "Unknown Document"
							}

							// Build location string (URI and/or Page)
							var locParts []string
							if chunk.RetrievedContext.URI != "" {
								locParts = append(locParts, fmt.Sprintf("URI: %s", chunk.RetrievedContext.URI))
							}

							// Check for RAGChunk page numbers
							if chunk.RetrievedContext.RAGChunk != nil && chunk.RetrievedContext.RAGChunk.PageSpan != nil {
								span := chunk.RetrievedContext.RAGChunk.PageSpan
								if span.FirstPage > 0 {
									if span.FirstPage == span.LastPage || span.LastPage == 0 {
										locParts = append(locParts, fmt.Sprintf("Page %d", span.FirstPage))
									} else {
										locParts = append(locParts, fmt.Sprintf("Pages %d-%d", span.FirstPage, span.LastPage))
									}
								}
							}

							// Fallback: Extract page number from text using regex
							// Look for pattern like "--- PAGE 17 ---"
							if chunk.RetrievedContext.Text != "" {
								re := regexp.MustCompile(`--- PAGE (\d+) ---`)
								matches := re.FindStringSubmatch(chunk.RetrievedContext.Text)
								if len(matches) > 1 {
									// Only add if we haven't already added a page number from RAGChunk
									alreadyHasPage := false
									for _, part := range locParts {
										if strings.Contains(part, "Page") {
											alreadyHasPage = true
											break
										}
									}
									if !alreadyHasPage {
										locParts = append(locParts, fmt.Sprintf("Page %s", matches[1]))
									}
								}
							}

							locStr := ""
							if len(locParts) > 0 {
								locStr = fmt.Sprintf(" (%s)", strings.Join(locParts, ", "))
							}

							fmt.Printf("  %d. [Doc] %s%s\n", i+1, title, locStr)

							if chunk.RetrievedContext.Text != "" {
								text := chunk.RetrievedContext.Text

								if verbose {
									// Verbose mode: Print full text but collapse excessive newlines
									// Replace 3+ newlines with 2
									re := regexp.MustCompile(`\n{3,}`)
									text = re.ReplaceAllString(text, "\n\n")
									fmt.Printf("     Full Text:\n%s\n", text)
								} else {
									// Default mode: Clean up snippet (single line)
									text = strings.ReplaceAll(text, "\n", " ")
									text = strings.ReplaceAll(text, "\r", " ")
									text = strings.Join(strings.Fields(text), " ") // Collapse multiple spaces

									// Truncate text if too long
									if len(text) > 200 {
										text = text[:197] + "..."
									}
									// Indent the snippet
									fmt.Printf("     Snippet: %s\n", text)
								}
							}
						}
					}
				}
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

// Execute runs the root command
func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}
