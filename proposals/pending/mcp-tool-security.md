# Proposal: Restricted MCP Tool Approval

## Problem
The current `approve_action` MCP tool accepts a raw string for `signature` and passes it directly to the gRPC layer. For security-adjacent proposals (e.g., config changes, Merkle log schema), this bypasses the Translucent Gate requirement for a valid human cryptographic signature.

## Proposed Solution
Modify `internal/mcp/server.go` to validate the proposal type before allowing `approve_action` tool execution.

### Policy
1. If `IsSecurityAdjacent == true`: The MCP tool must return an error. These proposals MUST be approved via the Sati-Central Control Panel to ensure human signature verification.
2. If `IsSecurityAdjacent == false`: The tool can proceed as normal.

## Impact
Preserves the safety invariants of the Event Horizon architecture while allowing MCP-based automation for low-risk actions.

## Status
PENDING. Implementation targeted for Phase 6.
