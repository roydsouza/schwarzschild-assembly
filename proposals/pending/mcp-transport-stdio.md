# Proposal: Native stdio MCP Transport

## Problem
The current MCP host only supports HTTP POST. CLAUDE.md Phase 3 requires "stdio (local)" transport to support CLI-based agents (like `claude-code`) directly without network overhead.

## Proposed Solution
Implement a `transport_stdio.go` handler that reads from `os.Stdin` and writes to `os.Stdout`.

### Interface
```go
type StdioTransport struct {
    logger *zap.Logger
    host   *Host
}

func (t *StdioTransport) Run(ctx context.Context) error
```

### Impact
Permits local agent binding via `command = ["root-spine", "--mcp"]` in agent configuration.

## Status
PENDING. Implementation targeted for Phase 6.
