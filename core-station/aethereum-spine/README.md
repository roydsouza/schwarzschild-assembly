# Root Spine

Go orchestrator — the central nervous system of Aethereum-Spine. MCP host, gRPC control
plane, Merkle log writer, WebSocket signaling plane.

## What it does

- Implements the Orchestrator gRPC service defined in `proto/orchestrator.proto`
- Manages factory lifecycle (create, stop, list)
- Routes proposals through the safety pipeline (safety-rail → Translucent Gate → Merkle)
- Writes RFC 6962 SHA-256 Merkle leaves for every significant event
- Writes `otel-snapshots/latest.json` via OTel file exporter for cold reads
- Serves WebSocket events to the control-panel

## Depends on

- Safety Rail (Rust, via FFI or subprocess — Phase 3 decision)
- PostgreSQL 16 (state persistence for non-critical assets)
- OTel collector (metric export)

## How to run tests

```bash
cd aethereum-spine && go test ./...
```

## How to generate proto bindings

```bash
protoc --proto_path=proto \
  --go_out=internal/grpc/pb --go_opt=paths=source_relative \
  --go-grpc_out=internal/grpc/pb --go-grpc_opt=paths=source_relative \
  proto/orchestrator.proto
```

## Metrics emitted

| OTel Metric Name | Type | Fitness Vector |
|-----------------|------|----------------|
| `aethereum_spine.audit.consistency_failures_total` | counter | audit_integrity |
| `aethereum_spine.perf.request_duration_ms` | histogram | system_performance |
| `aethereum_spine.perf.requests_total` | counter | system_performance |
| `aethereum_spine.cost.usd_per_day` | gauge | operational_cost |

## Implementation status

Phase 3 — not yet implemented. `proto/orchestrator.proto` and `go.mod` are complete.
AntiGravity implements this after Phase 2 (Safety Rail) is approved.
