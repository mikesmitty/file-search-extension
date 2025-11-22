# Testing Guide for File Search Tool

## Unit Tests

Run all unit tests:
```bash
go test -v ./...
```

Run tests for a specific package:
```bash
go test -v ./internal/gemini/
go test -v .
```

### Test Coverage

#### Name Resolution Tests
- `TestResolveStoreNameFormat`: Validates store name/ID format detection
- `TestResolveFileNameFormat`: Validates file name/ID format detection
- `TestResolveDocumentNameFormat`: Validates document name/ID format detection

These tests ensure that both friendly names and resource IDs are properly recognized.

#### Upload Options Tests
- `TestUploadFileOptions`: Tests various upload option configurations
- `TestUploadFileOptionsDefaults`: Validates default values for upload options

#### Metadata Parsing Tests (in main_test.go)
- `TestParseMetadata`: Tests metadata parsing from key=value format

## Integration Tests

Integration tests require valid API credentials. Use 1Password CLI to inject credentials:

```bash
op run --env-file .env -- ./test_integration.sh
```

### Manual Integration Testing

#### 1. Test File Import with Friendly Names

```bash
# Step 1: Upload a file to Files API
op run --env-file .env -- \
  ./file-search file upload ./README.md --name "test-readme"

# Step 2: Create a test store
op run --env-file .env -- \
  ./file-search store create "TestImportStore"

# Step 3: Import file using friendly names
op run --env-file .env -- \
  ./file-search store import-file "test-readme" --store "TestImportStore"

# Step 4: Verify import
op run --env-file .env -- \
  ./file-search document list --store "TestImportStore"
```

#### 2. Test Query with Metadata Filter

```bash
# Upload a file with metadata
op run --env-file .env -- \
  ./file-search file upload ./test.pdf \
    --store "TestStore" \
    --metadata "category=research" \
    --metadata "status=reviewed"

# Query with metadata filter
op run --env-file .env -- \
  ./file-search query "summarize the research" \
    --store "TestStore" \
    --metadata-filter "category=research AND status=reviewed"
```

#### 3. Test Store Delete with Force Flag

```bash
# Delete a store that contains documents
op run --env-file .env -- \
  ./file-search store delete "TestStore" --force
```

## Testing New Features

### Feature 1: ImportFile Method

**Location**: `internal/gemini/client.go:254-275`

**Test Coverage**:
- Unit tests: Format validation for file IDs and store IDs
- Integration test: Full workflow from file upload to import

**What to test**:
- Import with file resource ID: `files/abc123`
- Import with friendly file name: `my-document.pdf`
- Import with store resource ID: `fileSearchStores/xyz789`
- Import with friendly store name: `MyStore`
- Error handling: Non-existent file, non-existent store
- Operation polling: Verify import completes successfully

### Feature 2: store import-file Command

**Location**: `main.go:153-189`

**Test Coverage**:
- CLI flag parsing: `--store` vs `--store-id`
- Name resolution: File and store friendly names
- Error messages: Missing required flags

**What to test**:
```bash
# Both flags work
./file-search store import-file "file-name" --store "StoreName"
./file-search store import-file "files/abc" --store-id "fileSearchStores/xyz"

# Mixed usage
./file-search store import-file "file-name" --store-id "fileSearchStores/xyz"
./file-search store import-file "files/abc" --store "StoreName"

# Error: neither flag provided
./file-search store import-file "file-name"  # Should error
```

### Feature 3: Metadata Filter for Query

**Location**: `main.go:422-456`

**Test Coverage**:
- Flag parsing: `--metadata-filter` value passed correctly
- Integration with Query method: Filter applied to API call

**What to test**:
```bash
# Simple filter
./file-search query "test" --store "Store" --metadata-filter "key=value"

# Complex filter (AND/OR operators)
./file-search query "test" --store "Store" --metadata-filter "key1=val1 AND key2=val2"

# Without filter (should still work)
./file-search query "test" --store "Store"
```

---

### Feature 4: JSON Output Mode

**Location**:
- `internal/gemini/client.go:14-20` (OutputFormat type)
- `main.go:34,41-47` (global --format flag)
- All List/Get methods in `client.go`

**Test Coverage**:
- OutputFormat constants (unit tests)
- All list commands support JSON output
- All get commands support JSON output
- Query command supports JSON output
- JSON is valid and parsable

**What to test**:
```bash
# Store commands
./file-search store list --format json
./file-search store get "StoreName" --format json

# File commands
./file-search file list --format json
./file-search file get "file-name" --format json

# Document commands
./file-search document list --store "StoreName" --format json
./file-search document get "doc-name" --store "StoreName" --format json

# Query command (returns full API response)
./file-search query "question" --store "Store" --format json

# Verify JSON is valid (pipe to jq)
./file-search store list --format json | jq .
./file-search file list --format json | jq '.[0].displayName'

# Default is text format
./file-search store list  # Should use text format
./file-search store list --format text  # Explicit text format
```

**JSON Output Structure**:
- List commands: Returns array of objects
- Get commands: Returns single object
- Query command: Returns full GenerateContentResponse with candidates, grounding metadata, etc.
- All fields from API are included (not just display subset)

---

### Feature 5: MCP Server Enhancements

**Location**:
- `internal/mcp/server.go` (all tools)
- `main.go:21,51-76,514-516` (configuration)

**Test Coverage**:
- Tool configuration (flag, env, config file)
- Name resolution in all tools
- Client reuse pattern
- All 5 tools functional

**What to test**:

#### Configuration Testing
```bash
# Default configuration (query only)
file-search mcp

# Via environment variable
MCP_TOOLS=query,import,list file-search mcp

# Via command flag
file-search mcp --mcp-tools query,import,list

# Via config file (.file-search.yaml)
# See .file-search.yaml.example for format
```

#### MCP Client Configuration (Claude Desktop)
Add to Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "file-search": {
      "command": "/path/to/file-search",
      "args": ["mcp"],
      "env": {
        "GOOGLE_API_KEY": "your-api-key",
        "MCP_TOOLS": "query,import,list"
      }
    }
  }
}
```

#### Tool Testing

**Query Tool** (always enabled by default):
```
Tool: query
Parameters:
  - q (required): "What is in my research store?"
  - store (optional): "Research" or "fileSearchStores/abc123"
  - filter (optional): "category=important"
  - model (optional): "gemini-2.5-flash"
```

**Import Tool** (enable with `MCP_TOOLS=query,import`):
```
Tool: import
Parameters:
  - file (required): "document.pdf" or "files/abc123"
  - store (required): "Research" or "fileSearchStores/abc123"
```

**Upload Tool** (enable with `MCP_TOOLS=query,upload`):
```
Tool: upload
Parameters:
  - path (required): "/path/to/file.pdf"
  - store (required): "Research"
  - name (optional): "My Document"
  - metadata (optional): {"category": "research", "status": "draft"}
```

**List Tool** (enable with `MCP_TOOLS=query,list`):
```
Tool: list
Parameters:
  - type (required): "stores" | "files" | "docs"
  - store (optional, required for docs): "Research"
```

**Manage Tool** (enable with `MCP_TOOLS=query,manage`):
```
Tool: manage
Parameters:
  - action (required): "create" | "delete"
  - type (required): "store" | "doc"
  - name (required): store/doc name
  - store (optional, required for docs): "Research"
  - force (optional, for store deletion): true
```

#### Name Resolution Testing
All tools support both friendly names and resource IDs:
- Store: "Research" → "fileSearchStores/abc123"
- File: "document.pdf" → "files/xyz789"
- Document: "my-doc" → "fileSearchStores/abc/documents/xyz"

Test by using friendly names in all tool parameters - they should resolve automatically.

#### Context Management Testing
1. **Minimal config** (query only): Should have minimal context usage
2. **Full config** (all tools): Verify all tools available but context acceptable
3. **Custom config**: Test various combinations

---

### Feature 6: Shell Completion

**Location**:
- `internal/completion/cache.go` (TTL cache)
- `internal/completion/completion.go` (Completer with API integration)
- `main.go` (ValidArgsFunction and RegisterFlagCompletionFunc on all commands)

**Test Coverage**:
- Completion script generation for all shells
- Dynamic resource name completion
- Cache behavior and TTL
- Graceful degradation on API errors
- Configuration (enabled/disabled, cache TTL)

**What to test**:

#### Setup Shell Completion
```bash
# Bash
./file-search completion bash > /tmp/file-search-completion.bash
source /tmp/file-search-completion.bash

# Zsh
./file-search completion zsh > /tmp/_file-search
# Add to fpath and reload
```

#### Test Positional Argument Completion
```bash
# Store names (should show list of stores)
./file-search store get <TAB>
./file-search store delete <TAB>

# File names (should show list of files)
./file-search file get <TAB>
./file-search file delete <TAB>

# Document names (context-aware - needs --store flag first)
./file-search document get --store "MyStore" <TAB>
./file-search document delete --store "MyStore" <TAB>
```

#### Test Flag Completion
```bash
# Store flags (should show list of stores)
./file-search query --store <TAB>
./file-search file upload test.pdf --store <TAB>
./file-search document list --store <TAB>

# Model flag (should show static list of models)
./file-search query "test" --model <TAB>
```

#### Test Configuration
```bash
# Disable completion
export COMPLETION_ENABLED=false
./file-search store get <TAB>  # Should not show completions

# Re-enable
export COMPLETION_ENABLED=true

# Change cache TTL
export COMPLETION_CACHE_TTL=60s  # 1 minute cache
```

#### Test Caching Behavior
1. First completion call: May take up to 2 seconds (API call)
2. Subsequent calls within TTL: Should be instant (cached)
3. After TTL expires: New API call made
4. Cache key isolation: Stores, files, and documents cached separately

#### Test Error Handling
```bash
# No API key set (completion should fail gracefully)
unset GOOGLE_API_KEY
unset GEMINI_API_KEY
./file-search store get <TAB>  # Should not error, just no suggestions

# Invalid API key (completion should timeout after 2s)
export GOOGLE_API_KEY=invalid
./file-search store get <TAB>  # Should timeout gracefully
```

#### Completion Coverage
- **Positional arguments**: 6 commands (store get/delete, file get/delete, document get/delete)
- **Flag completions**: 13 flags (--store, --store-id on 6 commands + --model on query)
- **Context-aware**: Document completions use --store flag value

#### Automated Integration Testing

The `test_integration.sh` script now includes completion tests (Test 8):

```bash
# Run with 1Password CLI for credentials
op run --env-file .env -- ./test_integration.sh
```

**What it tests:**
- ✅ Completion script generation
- ✅ Store name completion with real API
- ✅ File name completion with real API
- ✅ Model name completion (static list)
- ✅ Cache behavior (second call faster than first)
- ✅ Disabled completion configuration
- ✅ Graceful error handling

**Test Coverage:**
- Unit tests: 32.9% (cache: 100%, logic: all testable paths)
- Integration tests: Real API calls with credentials
- Manual tests: Interactive shell completion

---

### Feature 7: Operation Polling

**Location**:
- `internal/gemini/client.go:493-616` (GetOperation, FormatOperationStatus)
- `main.go:630-686` (operation command group)

**Test Coverage**:
- OperationType constants validation
- Operation name format validation
- OperationStatus struct fields and JSON marshaling
- Operation status states (pending/done/failed)

**What to test**:

#### Basic Operation Polling
```bash
# After uploading or importing a file, you'll get an operation ID
# Poll the operation status
./file-search operation get "fileSearchStores/abc123/operations/op456"

# Specify operation type explicitly
./file-search operation get "fileSearchStores/abc123/operations/op456" --type import

# JSON output
./file-search operation get "fileSearchStores/abc123/operations/op456" --format json
```

#### Operation Name Validation
```bash
# Test invalid operation names
./file-search operation get "invalid"  # Should error: must start with 'fileSearchStores/'
./file-search operation get "fileSearchStores/abc"  # Should error: must contain '/operations/'
```

#### Auto-Detection of Operation Type
```bash
# Without --type flag, it tries both import and upload types
./file-search operation get "fileSearchStores/abc123/operations/op456"
```

#### Poll Until Complete (Bash Script)
```bash
#!/bin/bash
OP_NAME="fileSearchStores/abc123/operations/op456"

while ! ./file-search operation get "$OP_NAME" --format json | jq -e '.done == true' >/dev/null; do
  echo "Waiting for operation to complete..."
  sleep 2
done

echo "Operation complete!"
./file-search operation get "$OP_NAME"
```

#### Integration Test
The `test_integration.sh` script includes operation polling tests (Test 9):
- Verifies operation command exists
- Tests operation name validation
- Provides usage examples

**Operation Status Output:**
```
Operation: fileSearchStores/abc123/operations/op456
Type: import
Status: DONE
Store: fileSearchStores/abc123
Document: fileSearchStores/abc123/documents/doc789

Metadata:
  createTime: 2025-11-21T10:30:00Z
```

**JSON Output:**
```json
{
  "name": "fileSearchStores/abc123/operations/op456",
  "type": "import",
  "done": true,
  "failed": false,
  "parent": "fileSearchStores/abc123",
  "documentName": "fileSearchStores/abc123/documents/doc789",
  "metadata": {
    "createTime": "2025-11-21T10:30:00Z"
  }
}
```

## Bug Fixes Tested

### Bug Fix: DeleteStore Missing Force Parameter

**Location**: `main.go:123-139`

**What was fixed**: The `store delete` command was calling `client.DeleteStore()` with only 2 arguments when 3 were required.

**Test**:
```bash
# Should now work with --force flag
./file-search store delete "StoreName" --force

# Should work without --force (defaults to false)
./file-search store delete "EmptyStore"
```

### Bug Fix: Query Missing metadataFilter Parameter

**Location**: `main.go:450`

**What was fixed**: The Query method call was missing the 5th parameter (metadataFilter).

**Test**: Verified through unit tests and integration testing with `--metadata-filter` flag.

## Continuous Integration

To add to CI pipeline:

```yaml
# .github/workflows/test.yml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run tests
        run: go test -v ./...
      - name: Build
        run: go build -o file-search .
      - name: Verify commands
        run: |
          ./file-search store --help | grep import-file
          ./file-search query --help | grep metadata-filter
```

## Test Checklist

Before releasing:

- [ ] All unit tests pass: `go test -v ./...`
- [ ] Build succeeds: `go build -o file-search .`
- [ ] New commands appear in help: `./file-search store --help`
- [ ] Integration test with file import works end-to-end
- [ ] Integration test with metadata filter works
- [ ] Store delete with `--force` flag works
- [ ] Both friendly names and resource IDs work for all commands
- [ ] Error messages are helpful when invalid input is provided
