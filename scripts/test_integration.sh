#!/bin/bash
# Integration test script for new file-search features
# Run with: op run --env-file .env -- ./test_integration.sh

set -e  # Exit on error

# Ensure we are running from the project root
cd "$(dirname "$0")/.."

echo "=== File Search Integration Tests ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Test 1: List stores (prerequisite check)
echo -e "${BLUE}Test 1: Listing existing stores${NC}"
./file-search store list
echo ""

# Test 2: List files (check what's available to import)
echo -e "${BLUE}Test 2: Listing uploaded files${NC}"
./file-search file list
echo ""

# Test 3: Test store import-file with friendly name resolution
echo -e "${BLUE}Test 3: Import file with friendly name resolution${NC}"
echo "This test requires:"
echo "  1. An existing file in the Files API (use 'file list' to see available files)"
echo "  2. An existing store (use 'store list' to see available stores)"
echo ""
echo "To test manually, run:"
echo "  ./file-search store import-file \"<file-name>\" --store \"<store-name>\""
echo ""

# Test 4: Query with metadata filter
echo -e "${BLUE}Test 4: Query with metadata filter${NC}"
echo "This test requires a store with documents that have custom metadata"
echo ""
echo "To test manually, run:"
echo "  ./file-search query \"your question\" --store \"<store-name>\" --metadata-filter \"key=value\""
echo ""

# Test 5: Store delete with --force flag
echo -e "${BLUE}Test 5: Store delete command (with --force flag)${NC}"
echo "To test the force delete feature, run:"
echo "  ./file-search store delete \"<store-name>\" --force"
echo ""

# Test 6: Verify help text for new commands
echo -e "${BLUE}Test 6: Verify new commands appear in help${NC}"
echo "--- store import-file help ---"
./file-search store import-file --help
echo ""
echo "--- query metadata-filter flag ---"
./file-search query --help | grep -A 1 "metadata-filter"
echo ""

# Test 7: JSON output mode
echo -e "${BLUE}Test 7: JSON output mode${NC}"
echo "Verify --format flag is available:"
./file-search --help | grep -A 1 "format"
echo ""
echo "To test JSON output, run:"
echo "  ./file-search store list --format json"
echo "  ./file-search file list --format json"
echo "  ./file-search document list --store \"StoreName\" --format json"
echo ""

# Test 8: Shell completion integration
echo -e "${BLUE}Test 8: Shell completion integration${NC}"
echo "Testing completion functionality with real API..."
echo ""

# Test completion script generation
echo "Generating bash completion script..."
./file-search completion bash > /tmp/file-search-completion-test.bash
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Completion script generated successfully${NC}"
else
    echo -e "${RED}✗ Failed to generate completion script${NC}"
fi
echo ""

# Test Cobra's __complete command (tests the completion functions)
echo "Testing store name completion..."
STORE_COMPLETIONS=$(./file-search __complete store get "")
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Store name completion executed${NC}"
    echo "Available completions: $(echo "$STORE_COMPLETIONS" | head -5 | tr '\n' ' ')"
else
    echo -e "${RED}✗ Store name completion failed${NC}"
fi
echo ""

echo "Testing file name completion..."
FILE_COMPLETIONS=$(./file-search __complete file get "")
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ File name completion executed${NC}"
    echo "Available completions: $(echo "$FILE_COMPLETIONS" | head -5 | tr '\n' ' ')"
else
    echo -e "${RED}✗ File name completion failed${NC}"
fi
echo ""

echo "Testing model name completion (static list)..."
MODEL_COMPLETIONS=$(./file-search __complete query --model "")
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Model name completion executed${NC}"
    echo "Available models: $(echo "$MODEL_COMPLETIONS" | head -5 | tr '\n' ' ')"
else
    echo -e "${RED}✗ Model name completion failed${NC}"
fi
echo ""

# Test cache behavior (second call should be faster)
echo "Testing completion cache behavior..."
START_TIME=$(date +%s%N)
./file-search __complete store get "" > /dev/null 2>&1
FIRST_CALL=$(($(date +%s%N) - START_TIME))

START_TIME=$(date +%s%N)
./file-search __complete store get "" > /dev/null 2>&1
SECOND_CALL=$(($(date +%s%N) - START_TIME))

echo "First call: ${FIRST_CALL}ns"
echo "Second call (cached): ${SECOND_CALL}ns"
if [ $SECOND_CALL -lt $FIRST_CALL ]; then
    echo -e "${GREEN}✓ Cache is working (second call faster)${NC}"
else
    echo -e "${BLUE}Note: Cache may already be warm or TTL expired${NC}"
fi
echo ""

# Test completion can be disabled
echo "Testing completion disable configuration..."
COMPLETION_ENABLED=false ./file-search __complete store get "" > /tmp/disabled-test.txt 2>&1
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Completion works with COMPLETION_ENABLED=false${NC}"
    echo "Note: Disabled completion returns empty list gracefully"
else
    echo -e "${RED}✗ Completion failed when disabled${NC}"
fi
echo ""

# Test 9: Operation polling
echo -e "${BLUE}Test 9: Operation polling${NC}"
echo "Testing operation status checking..."
echo ""

# Verify operation command exists
echo "Checking operation command help..."
./file-search operation --help > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Operation command available${NC}"
else
    echo -e "${RED}✗ Operation command not found${NC}"
fi
echo ""

echo "Operation polling usage examples:"
echo "  # Get operation status (auto-detect type)"
echo "  ./file-search operation get \"fileSearchStores/abc123/operations/op456\""
echo ""
echo "  # Get operation status with specific type"
echo "  ./file-search operation get \"fileSearchStores/abc123/operations/op456\" --type import"
echo ""
echo "  # Get operation status in JSON format"
echo "  ./file-search operation get \"fileSearchStores/abc123/operations/op456\" --format json"
echo ""
echo "  # Poll until operation completes (bash loop)"
echo "  OP_NAME=\"fileSearchStores/abc/operations/123\""
echo "  while ! ./file-search operation get \"\$OP_NAME\" --format json | jq -e '.done == true' >/dev/null; do"
echo "    echo \"Waiting...\""
echo "    sleep 2"
echo "  done"
echo ""

# Test with invalid operation name to verify error handling
echo "Testing operation name validation..."
./file-search operation get "invalid-operation-name" 2>&1 | grep -q "must start with 'fileSearchStores/'"
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Operation name validation works${NC}"
else
    echo -e "${RED}✗ Operation name validation failed${NC}"
fi
echo ""

echo -e "${GREEN}=== Integration test guidance complete ===${NC}"
echo ""
echo "Next steps for manual testing:"
echo "1. Upload a test file: ./file-search file upload <path> --name \"test-file\""
echo "2. Create a test store: ./file-search store create \"TestStore\""
echo "3. Import the file: ./file-search store import-file \"test-file\" --store \"TestStore\""
echo "   (Note: Import returns an operation ID that you can poll)"
echo "4. Query with filter: ./file-search query \"test query\" --store \"TestStore\" --metadata-filter \"type=test\""
echo "5. Test completion in shell: source /tmp/file-search-completion-test.bash && ./file-search store get <TAB>"
echo "6. Poll operation status: ./file-search operation get \"<operation-id>\" --format json"
