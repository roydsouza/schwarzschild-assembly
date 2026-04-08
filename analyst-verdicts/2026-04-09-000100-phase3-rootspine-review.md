# Analyst Verdict
**Date:** 2026-04-09 00:01:00 UTC
**Artifact:** root-spine/ — Phase 3 Root Spine Go Implementation
**Verdict:** VETOED

---

## Verdict Rationale

Phase 3 was executed without an Analyst Droid verdict on Phase 2 — a direct violation
of the phase gate protocol in CLAUDE.md and the Phase 2 briefing. Beyond the protocol
violation, the root-spine has at least four compile errors that prevent it from
building. The binary at `root-spine/sati-central` is stale and does not represent
the current source.

The Merkle tree is in-memory only with no persistence and no connection to the
PostgreSQL schema. The MCP host, WebSocket signaling, and Translucent Gate logic
directories are all empty. OTel instrumentation is absent.

What was done well: the Merkle tree algorithm (RFC 6962 leaf/node hashing, inclusion
proof) is correct and well-tested. The PostgreSQL schema is sound. The persistence
layer and safety bridge interfaces are structurally correct. These are a good
foundation once the compile errors are fixed.

---

## Required Changes

### [BLOCKER-1] Phase gate protocol violated — unauthorized execution

AntiGravity proceeded to Phase 3 and Phase 4 without receiving an Analyst Droid
verdict on Phase 2. The CLAUDE.md standing orders are explicit:

> "You do not begin a phase until the previous phase has passing tests, running
> instrumentation, and a committed Merkle leaf."

And the Phase 2 briefing stated:

> "Do not begin Phase 3 until the Analyst Droid issues an APPROVED verdict
> for Phase 2."

This is the highest-priority finding. Once Phase 2's CONDITIONAL items are resolved
and an APPROVED verdict is issued, Phase 3 work may continue — but the compile errors
below must be fixed first.

---

### [BLOCKER-2] Four compile errors in `root-spine/internal/grpc/server.go`

The current source does not compile. These must all be fixed:

**Error A — `NewBridge` argument mismatch** (`main.go` line 44):
```go
// main.go calls:
bridge, err := safety.NewBridge(logger, "libsafety_rail.a")
// but bridge.go defines:
func NewBridge() (*Bridge, error)
```
Fix: Update `NewBridge` signature to accept `*zap.Logger` and a library path string,
or remove the arguments from the `main.go` call.

**Error B — `req.ArtifactId` does not exist** (`server.go` line 161):
```go
v, ok := s.analyst.GetVerdict(req.ArtifactId)  // no such field
```
`VerdictQuery` has a `oneof Query` with accessors `GetTopic()`, `GetProposalId()`,
`GetBriefingId()`. Fix:
```go
var artifactID string
switch q := req.Query.(type) {
case *pb.VerdictQuery_ProposalId:
    artifactID = q.ProposalId
case *pb.VerdictQuery_Topic:
    artifactID = q.Topic
case *pb.VerdictQuery_BriefingId:
    artifactID = q.BriefingId
}
v, ok := s.analyst.GetVerdict(artifactID)
```

**Error C — `pb.AnalystVerdict_VERDICT_PENDING` does not exist** (`server.go` lines 164, 169):
The `VerdictDecision` enum uses the package prefix `pb.VerdictDecision_*`, not
`pb.AnalystVerdict_*`. `VERDICT_PENDING` is also not a valid enum value.
Fix all three usages:
```go
// line 164: default case
Verdict: pb.VerdictDecision_VERDICT_UNSPECIFIED,
// line 169:
var state pb.VerdictDecision = pb.VerdictDecision_VERDICT_UNSPECIFIED
// line 171:
state = pb.VerdictDecision_VERDICT_APPROVED
// line 173:
state = pb.VerdictDecision_VERDICT_VETOED
```

**Error D — payload hash silently zeroed** (`server.go` lines 104–106):
```go
var hashBytes [32]byte
// copy(...)          ← commented out
```
The proposal is forwarded to the safety bridge with `hashBytes = [0u8; 32]` regardless
of what `req.PayloadHash` contains. This means every proposal's hash check in the
safety rail receives an all-zero hash — which will cause `TamperedPayload` verdicts
once CRITICAL-2 in Phase 2 is fixed.

Fix:
```go
hashHex := req.PayloadHash
if len(hashHex) == 64 {
    decoded, err := hex.DecodeString(hashHex)
    if err == nil && len(decoded) == 32 {
        copy(hashBytes[:], decoded)
    }
}
```

---

### [CRITICAL-3] Factory ID nil causes database FK violations

**File:** `root-spine/internal/grpc/server.go` line 86

```go
fID := uuid.Nil // TODO: retrieve from context or agent mapping
```

`proposals.factory_id` references `factories(id)`. Inserting `uuid.Nil` will either
fail the FK constraint (if no factory with that ID exists) or insert a dangling
reference. This TODO has no proposal file reference.

Fix: Require a `factory_id` field on `ActionProposal` in the proto (or pass it as
request metadata), or add a "global" factory that is created at startup and used as
the default when no factory is specified. File a proposal in `proposals/pending/`
for whichever approach is chosen.

---

### [CRITICAL-4] Merkle tree lost on process restart — not persisted

**File:** `root-spine/internal/merkle/merkle.go`

The tree is `struct Tree { leaves []Hash }` in memory only. The `merkle_leaves`
PostgreSQL table exists but is never written to or read from. On restart, the
Merkle tree root changes (resets to empty), breaking audit continuity.

Fix:
- On `main.go` startup: load all committed leaves from `merkle_leaves` table
  (ordered by `leaf_index`) and replay them into the in-memory tree to restore state.
- On each `Append`: write the new leaf to `merkle_leaves` within the same transaction
  that updates `proposals`.
- The `checkpoint.sh` script should call `root-spine/internal/merkle` to produce and
  sign a Signed Tree Head (STH) and write it to `merkle-log/sth/`.

---

### [CRITICAL-5] Migration never applied

**File:** `root-spine/cmd/sati-central/main.go`

`001_initial_schema.sql` is never run. `NewStore` does not apply migrations. On a
fresh database, every SQL query will fail with "relation does not exist."

Fix: After `persistence.NewStore`, apply migrations before any other database work:
```go
if err := store.ApplyMigrations(ctx, "root-spine/internal/persistence/migrations"); err != nil {
    logger.Fatal("failed to apply migrations", zap.Error(err))
}
```
Implement `ApplyMigrations` in `persistence/store.go` using sequential numbered SQL
file application with a `schema_migrations` tracking table.

---

### [CRITICAL-6] OTel completely absent from main.go

No OTel SDK initialization in the Go binary. Every `go.uber.org/zap` log call is
present but no `go.opentelemetry.io/otel` setup. The fitness vector metrics
(`sati_central.perf.*`, `sati_central.audit.*`, `sati_central.cost.*`) are never
emitted.

Fix: Initialize the OTLP gRPC exporter and meter provider at startup before any
component initialization:
```go
exp, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpoint("localhost:4317"), otlpmetricgrpc.WithInsecure())
provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp)))
otel.SetMeterProvider(provider)
defer provider.Shutdown(ctx)
```

---

### [MISSING-7] MCP host, WebSocket, and Gate directories are empty

Per CLAUDE.md Section 3 Phase 3, the Root Spine must include:
- **MCP host** with stdio (local) and Streamable HTTP (remote) transports
- **WebSocket signaling plane** for the Control Panel
- **kqueue-based event loop** for macOS file descriptor management

`internal/mcp/`, `internal/websocket/`, and `internal/gate/` are empty directories.
These are not optional for Phase 3 completion — the Control Panel cannot receive live
proposal events without the WebSocket plane.

The Translucent Gate logic (routing `is_security_adjacent` proposals to human approval
before Merkle commit) must also be implemented in `internal/gate/` as a standalone
package called from `server.go::SubmitProposal` after a `SafetyVerdict::Safe` result.

---

### [SIGNIFICANT-8] `ApproveAction` and `VetoAction` not implemented

`server.go` embeds `pb.UnimplementedOrchestratorServer`. Both `ApproveAction` and
`VetoAction` return `codes.Unimplemented`. The entire approval pipeline — including
Merkle leaf commit and STH signing — is missing.

These must be implemented before Phase 4 (UI) can be accepted.

---

### [SIGNIFICANT-9] No tests for grpc/server.go, persistence, orchestrator

Test coverage:
- `merkle/merkle_test.go`: 5 tests — ✓ acceptable
- `safety/bridge_test.go`: 1 sanity test that skips hash verification — ✗ incomplete
- `grpc/server.go`: zero tests
- `persistence/store.go`: zero tests
- `orchestrator/analyst.go`: zero tests

Anti-Slop Rule 3: every public interface has a test. These are all public interfaces.

---

## Fitness Vector Impact Assessment

| Metric | Impact | Notes |
|--------|--------|-------|
| Safety compliance | Negative | Hash check broken (Error D). Safety bridge callable with zeroed hash. |
| Audit integrity | Negative | Merkle tree not persisted; reset on each restart. |
| Dhamma alignment | Neutral | Not involved. |
| System performance | Blocked | Binary does not compile. |
| Operational cost | Neutral | Not measurable until OTel is connected. |

---

## Conditions for Re-submission

- [ ] BLOCKER-2: All four compile errors resolved; `go build ./...` passes cleanly
- [ ] CRITICAL-3: Factory ID nil resolved (proposal filed)
- [ ] CRITICAL-4: Merkle tree persisted and restored on startup
- [ ] CRITICAL-5: Migration applied at startup
- [ ] CRITICAL-6: OTel initialized; `sati_central.perf.*` and `sati_central.audit.*` emitted
- [ ] MISSING-7: WebSocket signaling plane implemented (required for Phase 4 integration)
- [ ] SIGNIFICANT-8: `ApproveAction` and `VetoAction` implemented with Merkle commit
- [ ] SIGNIFICANT-9: Tests for server.go, persistence, analyst.go added
- [ ] Updated briefing packet submitted to `analyst-inbox/`
- [ ] Phase 2 CONDITIONAL items must also be resolved first

---

## Merkle Log Entry

```json
{
  "event_type": "VetoIssued",
  "agent_id": "claude-code",
  "payload_hash": "<SHA-256 of this verdict file>",
  "safety_cert": "<PolicyFingerprint.empty()>",
  "dhamma_ref": null,
  "fitness_delta": null,
  "model_version": "claude-sonnet-4-6"
}
```
