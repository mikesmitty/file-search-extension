# Gemini File Search

A CLI tool and Model Context Protocol (MCP) server for the Google Gemini File Search API. This tool allows you to manage file stores, upload documents, and perform semantic searches using Gemini's advanced retrieval capabilities.

## Installation

### Primary: Releases
Download the latest release for your platform from the [Releases page](https://github.com/mikesmitty/file-search/releases).

### Verification
All release artifacts are signed using [Cosign](https://github.com/sigstore/cosign) (keyless). You can verify the integrity of the downloaded artifacts using the following command:

<!-- x-release-please-start-version -->
```bash
cosign verify-blob \
  --certificate-identity "https://github.com/mikesmitty/file-search/.github/workflows/release.yml@refs/tags/v0.2.0" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  --bundle <ARTIFACT>.sigstore.json \
  <ARTIFACT>
```
<!-- x-release-please-end -->
Replacing `<ARTIFACT>` with the filename (e.g., `darwin.arm64.file-search.tar.gz`).

### Fallback: Go Install
If you have Go installed, you can install the tool directly:

```bash
go install github.com/mikesmitty/file-search@latest
```

## Configuration

### API Key
You must provide a Gemini API key to use this tool. You can set it via an environment variable:

```bash
export GEMINI_API_KEY="your-api-key"
# OR
export GOOGLE_API_KEY="your-api-key"
```

Alternatively, you can pass it as a flag `--api-key` or configure it in `$HOME/.file-search.yaml`.

> [!IMPORTANT]
> **API Usage Fees**: Using the Gemini and the Gemini File Search APIs can involve costs for embeddings with paid tier API keys. The FileSearch API is free for free tier users, but note that Gemini queries may be subject to use for product improvement. I'm not a lawyer, so be sure to review the [Gemini API Pricing](https://ai.google.dev/gemini-api/docs/pricing) page better to understand the potential associated fees.

## CLI Usage

The `file-search` CLI provides several commands to manage your knowledge base.

### Stores
Manage File Search Stores (collections of documents).

```bash
# List all stores
file-search store list

# Create a new store
file-search store create "My Knowledge Base"

# Get store details
file-search store get "My Knowledge Base"

# Delete a store
file-search store delete "My Knowledge Base"
```

### Files
Manage files uploaded to the Gemini Files API. These are raw files that can be used for various purposes, including adding to stores.

> [!NOTE]
> **Data Retention**: Files uploaded to the Gemini Files API are stored for 48 hours and cannot be downloaded from the API. However, documents that are processed and added to a File Search Store will remain available for search until deleted.

```bash
# Upload a file (raw upload)
file-search file upload ./path/to/doc.pdf

# List uploaded files
file-search file list

# Delete a file
file-search file delete "doc.pdf"
```

### Documents
Manage documents within a Store. These are files that have been indexed and are ready for search.

```bash
# List documents in a store
file-search document list --store "My Knowledge Base"

# Get document details
file-search document get "doc.pdf" --store "My Knowledge Base"

# Delete a document from a store
file-search document delete "doc.pdf" --store "My Knowledge Base"
```

### Query
Perform a semantic search against your knowledge base.

```bash
file-search query "What is the max voltage?" --store "My Knowledge Base"
```

## MCP Server Integration

This tool functions as a Model Context Protocol (MCP) server, allowing AI assistants to access your documents.

### Antigravity
To use with Antigravity, ensure the `file-search` binary is in your path or referenced correctly in your configuration.

*(Installation details for Antigravity to be added)*

### Gemini
To enable the extension in Gemini:

1. Ensure `gemini-extension.json` is present in your project root (or the directory where you run Gemini).
2. The manifest should point to the `file-search` binary.
3. When you start a conversation, Gemini will detect the extension and allow you to use the `@file-search` tool (or whatever name is configured).

### Claude Code
To use with Claude Desktop or Claude Code, add the server configuration to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "file-search": {
      "command": "/path/to/file-search",
      "args": ["mcp"],
      "env": {
        "GEMINI_API_KEY": "your-api-key"
      }
    }
  }
}
```