#!/bin/bash
set -e

# Generate Windows configuration from the main extension config
# Uses jq to append .exe to the command

jq '.mcpServers["file-search"].command += ".exe"' gemini-extension.json > gemini-extension.windows.json
