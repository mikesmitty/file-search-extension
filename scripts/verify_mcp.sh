#!/bin/bash

# MCP Server Verification Script
# This script verifies that the MCP server exposes the correct tools

set -e

# Ensure we are running from the project root
cd "$(dirname "$0")/.."

echo "Building file-search binary..."
go build -o file-search .

echo ""
echo "Starting MCP Inspector..."
echo "The inspector will open at http://localhost:5173"
echo ""
echo "Expected tools (with --mcp-tools=all):"
echo "  1. list_stores"
echo "  2. list_files"  
echo "  3. list_documents"
echo "  4. create_store"
echo "  5. delete_store"
echo "  6. import_file_to_store"
echo "  7. query_knowledge_base (with metadata_filter parameter)"
echo "  8. upload_file (with metadata parameter)"
echo "  9. delete_file"
echo "  10. delete_document"
echo ""
echo "Press Ctrl+C to stop the inspector"
echo ""

# Run the inspector with all tools enabled
MCP_TOOLS=all npx @modelcontextprotocol/inspector ./file-search mcp
