#!/usr/bin/env bash
# mcp-client.sh — Lightweight bridge to call Root Spine MCP tools from Prolog.
#
# Usage: ./scripts/mcp-client.sh <tool_name> [arguments_json]
# Example: ./scripts/mcp-client.sh tools/list
# Example: ./scripts/mcp-client.sh tools/call '{"name": "create_spec", "arguments": {...}}'

set -euo pipefail

MCP_URL="http://localhost:8082/mcp"
METHOD="${1:-tools/list}"
ARGUMENTS="${2:-{}}"

# Create JSON-RPC request body
REQUEST_BODY=$(cat <<EOF
{
  "jsonrpc": "2.0",
  "id": "prolog-$(date +%s)",
  "method": "${METHOD}",
  "params": ${ARGUMENTS}
}
EOF
)

# Send request via curl
curl -s -X POST "${MCP_URL}" \
     -H "Content-Type: application/json" \
     -d "${REQUEST_BODY}"
