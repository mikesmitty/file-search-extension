# Gemini File Search

A CLI tool and Model Context Protocol (MCP) server for the Google Gemini File Search API. This tool allows you to manage file stores, upload documents, and perform semantic searches using Gemini's advanced retrieval capabilities.

## Installation

### Releases (Recommended)
Download the latest release for your platform from the [Releases page](https://github.com/mikesmitty/file-search/releases).

### Verification
All release artifacts are signed using [Cosign](https://github.com/sigstore/cosign) (keyless) and include build provenance attestations.
You can verify the integrity of the downloaded artifacts using the following commands:

First, verify your downloaded archive matches the checksum:

```bash
sha256sum -c checksums.txt --ignore-missing
```

Then verify that the checksums file itself is authentic using [GitHub CLI](https://cli.github.com/):

```bash
gh attestation verify checksums.txt -R mikesmitty/file-search
```
<!-- x-release-please-start-version -->
```bash
$ gh attestation verify checksums.txt -R mikesmitty/file-search
Loaded digest sha256:f45aa9456b79bfeb56dc82ee85dd0730669a5716f8930250485c57a25d271000 for file://checksums.txt
Loaded 1 attestation from GitHub API

The following policy criteria will be enforced:
- Predicate type must match:................ https://slsa.dev/provenance/v1
- Source Repository Owner URI must match:... https://github.com/mikesmitty
- Source Repository URI must match:......... https://github.com/mikesmitty/file-search
- Subject Alternative Name must match regex: (?i)^https://github.com/mikesmitty/file-search/
- OIDC Issuer must match:................... https://token.actions.githubusercontent.com

✓ Verification succeeded!

The following 1 attestation matched the policy criteria

- Attestation #1
  - Build repo:..... mikesmitty/file-search
  - Build workflow:. .github/workflows/goreleaser.yml@refs/tags/v0.6.2
  - Signer repo:.... mikesmitty/file-search
  - Signer workflow: .github/workflows/goreleaser.yml@refs/tags/v0.6.2
```
<!-- x-release-please-end -->

Alternatively, you can verify using Cosign:

<!-- x-release-please-start-version -->
```bash
cosign verify-blob \
  --certificate-identity "https://github.com/mikesmitty/file-search/.github/workflows/release.yml@refs/tags/v0.6.2" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  --bundle checksums.txt.sigstore.json \
  checksums.txt
```
<!-- x-release-please-end -->

### Go Install (Alternative)
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
## Quick Start Guide

Here is a typical workflow to get you started with the CLI:

1.  **Create a Store**:
    ```bash
    file-search store create "My Knowledge Base"
    ```

2.  **Upload a Document**:
    This uploads a local file and adds it to your store in one step.
    ```bash
    file-search file upload ./path/to/my-doc.pdf --store "My Knowledge Base"
    ```

3.  **Query**:
    Ask a question about your documents.
    ```bash
    file-search query "What are the key points in the document?" --store "My Knowledge Base"
    ```

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

### Operations
Manage long-running operations.

```bash
# Get operation status
file-search operation get <operation-name>
```

## MCP Server Integration

This tool functions as a Model Context Protocol (MCP) server, allowing AI assistants to access your documents.

### Antigravity
To use with Antigravity, ensure the `file-search` binary is in your path or referenced correctly in your configuration.

See [Antigravity MCP docs](https://antigravity.google/docs/mcp#connecting-custom-mcp-servers)

### Gemini

#### Option 1: As an MCP Server
To install `file-search` as an MCP server in Gemini for the current project:

```bash
gemini mcp add -e GEMINI_API_KEY=your-api-key file-search file-search mcp
```

To install it at the user-level scope (available across all projects):

```bash
gemini mcp add -s user -e GEMINI_API_KEY=your-api-key file-search file-search mcp
```

#### Option 2: As an Extension (Project-based)
To install the extension for a specific project:

```bash
gemini extensions install https://github.com/mikesmitty/file-search
```

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

### Usage

Once installed, the tools provided by `file-search` are automatically available to Gemini. You can interact with your knowledge base using natural language.

**Examples:**

*   **Creating a Store:** "Create a new file search store called Skynet."
*   **Listing Stores:** "List all my file search stores."
*   **Uploading:** "Upload the file `spec-v1.pdf` to the Skynet store."
*   **Listing Documents:** "Show me all documents in the Skynet store."
*   **Querying:** "Search the Skynet knowledge base for information about the T-800's power source."

Gemini will intelligently select the appropriate tool (`query_knowledge_base`, `list_stores`, `upload_file`, etc.) based on your request.
