# TODO – File Search Query Enhancements

## ✅ Completed (Phase 1)

### 1. ImportFile Method ✓
Added `ImportFile` function to import files from the Files API into a File Search Store.
- **Location**: `internal/gemini/client.go:254-275`
- **Features**:
  - Accepts both file and store resource IDs
  - Includes operation polling for async completion
  - Consistent with existing `UploadFile` pattern

### 2. store import-file Command ✓
Exposed CLI command for file import with full name resolution support.
- **Location**: `main.go:153-189`
- **Features**:
  - Flags: `--store` (friendly name) and `--store-id` (resource ID)
  - Resolves both file names and store names to IDs
  - Example: `file-search store import-file "doc.pdf" --store "Research"`

### 3. Metadata Filter for Queries ✓
Extended query command with metadata filtering capability.
- **Location**: `main.go:422-456`
- **Features**:
  - Flag: `--metadata-filter` for filter expressions
  - Fixed bug where Query call was missing metadataFilter parameter
  - Example: `file-search query "question" --store "Store" --metadata-filter "key=value"`

### Completed Action Items
- [x] Implement `func (c *Client) ImportFile(ctx context.Context, fileID, storeID string) error`
- [x] Add `store import-file` Cobra command with flags `--store` / `--store-id` and argument `<file-name-or-id>`
- [x] Update `query` command to accept `--metadata-filter` and forward it to `client.Query`
- [x] Write tests for the new functionality (unit tests in `client_test.go`)
- [x] Create testing documentation (`TESTING.md` and `test_integration.sh`)

### Bug Fixes Completed
- [x] Fixed `store delete` command missing `--force` flag (main.go:123-139)
- [x] Fixed `Query` method call missing metadataFilter parameter (main.go:450)

### 4. JSON Output Mode ✓
Added machine-readable JSON output for all list/get commands.
- **Location**: `internal/gemini/client.go:14-20`, all List*/Get* methods
- **Features**:
  - Global `--format` flag (values: `text`, `json`)
  - All list commands return JSON arrays
  - All get commands return JSON objects
  - Query returns full API response with grounding metadata
  - Example: `file-search store list --format json | jq .`

### 5. MCP Server Enhancements ✓
Added configurable MCP tools with built-in name resolution and short parameter names.
- **Location**: `internal/mcp/server.go`, `main.go:21,51-76,514-516`
- **Features**:
  - Configurable tools via `--mcp-tools` flag, `MCP_TOOLS` env, or `.file-search.yaml`
  - Default: Only `query` tool enabled (minimal context usage)
  - 5 tools available: `query`, `import`, `upload`, `list`, `manage`
  - Built-in name resolution (accepts both friendly names and resource IDs)
  - Short parameter names (`q`, `store`, `file`, `path`, `filter`)
  - Client reuse optimization (no per-request overhead)
- **Example**: `MCP_TOOLS=query,import,list file-search mcp`

### 6. Shell Completion ✓
Dynamic shell completion with intelligent name resolution and caching.
- **Built-in Command**: `file-search completion [bash|zsh|fish|powershell]`
- **Features**:
  - Command and flag completion (Cobra built-in)
  - **Dynamic resource name completion** (custom implementation):
    - Store names: `store get <TAB>`, `--store <TAB>`
    - File names: `file get <TAB>`
    - Document names: `document get <TAB>` (context-aware with --store)
    - Model names: `--model <TAB>`
  - TTL-based caching (5 min default) to minimize API calls
  - Configurable enable/disable via `completion_enabled` setting
  - Graceful degradation on API errors (2-second timeout)
  - 13 flag completions and 6 positional argument completions
- **Configuration**: `.file-search.yaml` or env vars (`COMPLETION_ENABLED`, `COMPLETION_CACHE_TTL`)
- **Example**: `file-search completion bash > /etc/bash_completion.d/file-search`
- **Location**:
  - Cache: `internal/completion/cache.go`
  - Completer: `internal/completion/completion.go`
  - Registration: `main.go` (ValidArgsFunction and RegisterFlagCompletionFunc)

### 7. Operation Polling Command ✓
Check status of long-running file upload and import operations.
- **Commands**: `file-search operation get [operation-name]`
- **Aliases**: `op`, `operations`
- **Features**:
  - Auto-detect operation type (import vs upload)
  - Manual type specification via `--type` flag
  - Show operation status: pending, done, failed
  - Display metadata (create time, etc.)
  - Show result information (store, document)
  - Support both text and JSON output formats
  - Operation name validation with clear error messages
- **Example**: `file-search operation get "fileSearchStores/abc123/operations/op456"`
- **Location**:
  - Types: `internal/gemini/client.go:22-40` (OperationType, OperationStatus)
  - Methods: `internal/gemini/client.go:493-616` (GetOperation, FormatOperationStatus)
  - Commands: `main.go:630-686` (operation command group)
  - Tests: `internal/gemini/client_test.go:346-538` (validation, status, JSON)

### 8. Progress Indicators ✓
Show upload/indexing progress during long-running operations.
- **Features**:
  - Simple progress indicators showing iteration count and elapsed time
  - Displays during operation polling loops (every 2 seconds)
  - Single-line updates using `\r` for clean output
  - Global `--quiet` flag (`-q`) to suppress all progress output
  - Works well in both TTY and non-TTY environments
- **Example**: `file-search file upload file.pdf --store "Store" -q`
- **Location**:
  - Types: `internal/gemini/client.go:263-265` (ImportFileOptions)
  - UploadFile: `internal/gemini/client.go:306-335`
  - ImportFile: `internal/gemini/client.go:375-399`
  - Global flag: `main.go:24,39`
- **Implementation**: Simple progress (iteration + time) without external dependencies

### 9. Batch Operations ✓
Support glob patterns and multiple files for upload and import.
- **Features**:
  - Accepts multiple file arguments
  - Expands shell glob patterns (e.g., `*.pdf`, `docs/*.md`)
  - Processes files in parallel with configurable concurrency (default: 5)
  - Shows progress for batch operations
  - Continues on individual failures without stopping entire batch
  - Summary report: "X succeeded, Y failed"
  - Exit code 1 if any file fails
- **Commands**:
  - `file-search file upload *.pdf --store "Research"`
  - `file-search file upload docs/*.md notes/*.txt --store "Documentation"`
  - `file-search store import-file file1.pdf file2.pdf file3.pdf --store "Research"`
- **Implementation**:
  - Parallel processing with worker pool
  - Comprehensive error collection and reporting
  - Integration with progress indicators

---

## 🛠 Code Review Improvements (High Priority)

### 11. Refactor Magic Strings & Constants ✓
**Priority**: High
**Status**: Completed
**Goal**: Improve maintainability and reduce risk of typos.

- [x] Define default model constant (e.g. `DefaultModel = "gemini-2.5-flash"`) in one place and propagate to CLI and MCP.
- [x] Define resource prefixes (`fileSearchStores/`, `files/`) as constants.
- [x] Replace all hardcoded occurrences with these constants.
- [x] Remove hardcoded model list in `internal/completion/completion.go` and fetch dynamically or use shared constant.

### 12. Improve MCP Tool Documentation ✓
**Priority**: High
**Status**: Completed
**Goal**: Help the AI model use tools more effectively.

- [x] Add examples to `metadata_filter` description in `internal/mcp/server.go`.
- [x] Add examples to other complex tool parameters (e.g., `metadata` JSON format).

### 13. Enhance Testing Infrastructure ✓
**Priority**: Medium
**Status**: Completed
**Goal**: Enable deterministic and end-to-end testing.

- [ ] Implement Record/Replay tests using `go-vcr` to mock API calls.
- [ ] Create End-to-End (E2E) test suite for full flow verification (provision -> upload -> query -> delete).
- [x] Update CI workflow to test against multiple Go versions (matrix: stable, oldstable, go.mod?).

### 14. Binary Name Consistency ✓
**Priority**: High
**Status**: Completed
**Goal**: Ensure the MCP server binary name is handled consistently across configuration and deployment.

- [x] Ensure `gemini-extension.json` uses the correct binary name (`file-search` vs `gemini-file-search`).
- [x] Verify `.goreleaser.yml` produces the expected binary name.
- [x] Check `scripts/` for hardcoded binary names.
- [x] Handle Windows `.exe` extension requirement (e.g., separate config or platform-aware runner).

### 15. Documentation & Scripts Updates ✓
**Priority**: Medium
**Status**: Completed
**Goal**: Ensure documentation and examples match the actual code behavior.

- [x] Update `GEMINI.md` to list all available MCP tools (currently only lists `query`).
- [x] Populate `README.md` with project description, installation, and usage instructions.
- [x] Fix `.file-search.yaml.example`: The `list` and `manage` tool groups are mentioned but not implemented in `server.go`. Either implement them or update the example to use specific tool names.
- [x] Verify `scripts/test_integration.sh` works with the final binary name.

### 20. Alternative Installation Workflow ✓
**Priority**: Low
**Status**: Completed
**Goal**: Simplify installation and MCP server registration for end users.

- [x] Investigate using Gemini CLI's `gemini mcp add` workflow instead of manual extension installation.
- [x] Documented in README.md under "Gemini > Option 1: As an MCP Server".

### 21. Homebrew Installation Support
**Priority**: Low
**Status**: Not Started
**Goal**: Allow users to install via `brew install file-search`.

- [ ] Create Homebrew tap repository (e.g., `mikesmitty/homebrew-tap`).
- [ ] Configure goreleaser to publish to the tap.
- [ ] Verify installation with `brew install mikesmitty/tap/file-search`.
- [ ] Update documentation with Homebrew installation instructions.

### 18. Display Grounding Details ✓
**Priority**: Medium
**Status**: Completed
**Goal**: Show detailed grounding attribution in query results.

- [x] Display grounding metadata details in text output format (currently only shows `[Grounding Metadata Found]`).
- [x] Show grounding chunks with source documents, snippets, and relevance scores.
- [x] Format grounding attribution in a user-friendly way (e.g., "From: documentName.pdf, Page X").
- [x] Ensure JSON output already includes full grounding details (verify current implementation).

### 19. Pagination Support ✓
**Priority**: High
**Status**: Completed
**Goal**: Ensure all resources are listed, not just the first page.

- [x] Implement pagination for `ListStores`
- [x] Implement pagination for `ListFiles`
- [x] Implement pagination for `ListDocuments`
- [x] Implement pagination for `ListModels`
- [x] Verify `Resolve*` and `Get*Names` helpers use paginated lists

---

## 🎯 Recommended Implementation Order

### Phase 2 (High Value) ✅ COMPLETED
1. ✅ **JSON Output Mode** (#4) - Enables automation and scripting
2. ✅ **MCP Server Enhancements** (#5) - Exposes new features to AI agents
3. ✅ **Shell Completion** (#6) - Dynamic completion with caching

### Phase 3 (Medium Priority) ✅ COMPLETED
4. ✅ **Operation Polling** (#7) - Check status of long-running operations

### Phase 4 (Nice to Have) ✅ COMPLETED
5. ✅ **Progress Indicators** (#8) - Simple progress with --quiet flag

### Phase 5 (Lower Priority) ✅ COMPLETED
6. ✅ **Batch Operations** (#9) - Multiple files with glob patterns and parallel processing

---

## 📝 Notes

### Testing Strategy
- Continue adding unit tests for new features
- Update `TESTING.md` with new test procedures
- Expand `test_integration.sh` as new commands are added
- Consider adding golden file tests for JSON output

### Documentation Updates Needed
- Update main README.md with new features as they're implemented
- Keep TESTING.md in sync with new functionality
- Add examples for each new command to docs


