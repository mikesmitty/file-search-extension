# Gemini Context: File Search Query

This tool enables Gemini to search and retrieve information from a dedicated knowledge base of documents (PDFs, text files, etc.) associated with your project. It is particularly useful for providing access to large technical documentation, datasheets, or design specs that don't fit in the context window.

## 1. Setup (One-time per project)

Before the model can use your documents, you need to index them into a "File Search Store".

### Create a Store
Run this command in your terminal to create a store. Replace `my-project-kb` with a unique name.

```sh
file-search store create my-project-kb
```

### Upload Documents
Upload your documents to the store. You can upload individual files or entire directories.

```sh
# Upload a single file
file-search file upload --store my-project-kb ./datasheets/spec-v1.pdf

# Upload a directory of documents
file-search file upload --store my-project-kb ./docs/
```

## 2. Model Instructions (GEMINI.md)

To enable Gemini to use this knowledge base, add a `GEMINI.md` file to the root of your project with the following content.

**Copy & Paste Template:**

```markdown
# Gemini Context: Project Knowledge Base

This project uses a searchable knowledge base for supplementary documentation.

## Available Tools

This MCP server provides 10 tools for interacting with Gemini File Search. By default, only `query_knowledge_base` is enabled to minimize context usage.

### Core Tools (Recommended)
- **`query_knowledge_base`** - Query the knowledge base with optional metadata filtering
- **`import_file_to_store`** - Import files from Files API into a store
- **`upload_file`** - Upload and index files directly to a store

### List/Discovery Tools
- **`list_stores`** - List all File Search Stores
- **`list_files`** - List all files in the Gemini Files API
- **`list_documents`** - List documents within a specific store

### Management Tools
- **`create_store`** - Create a new File Search Store
- **`delete_store`** - Delete a File Search Store (with optional force flag)
- **`delete_file`** - Delete a file from the Files API
- **`delete_document`** - Delete a document from a store

### Tool Configuration

Configure which tools are available via the `MCP_TOOLS` environment variable or `.file-search.yaml`:

```bash
# Minimal (default) - only query
MCP_TOOLS=query_knowledge_base file-search mcp

# Balanced - query + discovery
MCP_TOOLS=query_knowledge_base,list_stores,list_files,list_documents file-search mcp

# Full access - all tools
MCP_TOOLS=all file-search mcp
```

## Store Configuration

- **Store Name:** `my-project-kb` (Replace with your actual store name)

## Instructions

1.  **Always Search First:** If I ask about technical details, specifications, or architecture found in the documentation, you **MUST** use `query_knowledge_base` before answering. Do not guess.
2.  **Use Metadata Filters:** If I specify a category or type (e.g., "in the API docs"), use the `metadata_filter` parameter (e.g., `category = "api"`).
3.  **Cite Sources:** When providing answers from the knowledge base, mention which document the information came from.

## Example Usage

User: "What is the max voltage for the T-800?"
Model: `query_knowledge_base(query="T-800 max voltage", store_name="my-project-kb")`

User: "Find the deployment steps in the release notes."
Model: `query_knowledge_base(query="deployment steps", store_name="my-project-kb", metadata_filter="type = 'release_notes'")`
```

## 3. Development & Contributing

If you are modifying this tool, follow these steps.

### Prerequisites
- Go 1.23+
- `npm` (for MCP Inspector)

### Build
```sh
go build -o file-search .
```

### Test
Run unit and integration tests:
```sh
go test -v ./...
```

### Verifying the MCP Server

You can verify the MCP server functionality using the `mcp-inspector`. A helper script is provided:

```sh
./scripts/verify_mcp.sh
```
This opens the inspector at `http://localhost:5173`.

### Project Structure
- `main.go`: CLI entry point.
- `internal/mcp/`: MCP server implementation and tools.
- `internal/gemini/`: Client library for Gemini API.
