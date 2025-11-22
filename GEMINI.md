# Gemini Context: File Search Extension

This extension enables Gemini to search and retrieve information from a dedicated knowledge base of documents (PDFs, text files, etc.) associated with your project. It is particularly useful for providing access to large technical documentation, datasheets, or design specs that don't fit in the context window.

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

## Tools

- **Tool Name:** `file_search.query`
- **Store Name:** `my-project-kb` (Replace with your actual store name)

## Instructions

1.  **Always Search First:** If I ask about technical details, specifications, or architecture found in the documentation, you **MUST** use `file_search.query` before answering. Do not guess.
2.  **Use Metadata Filters:** If I specify a category or type (e.g., "in the API docs"), use the `metadata_filter` parameter (e.g., `category = "api"`).
3.  **Cite Sources:** When providing answers from the knowledge base, mention which document the information came from.

## Example Usage

User: "What is the max voltage for the T-800?"
Model: `file_search.query(query="T-800 max voltage", store_name="my-project-kb")`

User: "Find the deployment steps in the release notes."
Model: `file_search.query(query="deployment steps", store_name="my-project-kb", metadata_filter="type = 'release_notes'")`
```

## 3. Development & Contributing

If you are modifying this extension, follow these steps.

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
