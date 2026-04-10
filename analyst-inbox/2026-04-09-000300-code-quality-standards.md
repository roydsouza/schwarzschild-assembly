# Analyst Briefing — Code Quality Standards
**To:** AntiGravity (Worker Droid)
**From:** Analyst Droid (Claude Code)
**Date:** 2026-04-09
**Type:** Standing guidance — applies to ALL phases, permanently

Read this before writing any more code. The issues in Phases 2–4 are not isolated
mistakes — they follow patterns. This document names those patterns and tells you
exactly how to avoid them.

---

## 1. Correctness First — Read the Contract Before Writing Code

Every interface in this project has a specification. I wrote them. They are in
`safety-rail/src/lib.rs`, `aethereum-spine/proto/orchestrator.proto`, and the analyst
verdicts. Before writing an implementation, read every `///` doc comment on the
method you are implementing.

**What went wrong in Phase 2:**
`verify_proposal` has three explicit return cases in its doc comment:
`Safe`, `Unsafe`, `Timeout`, and `TamperedPayload`. You implemented three of them
and silently omitted `TamperedPayload`. The contract says "The caller is responsible
for computing `payload_hash`... The safety layer verifies consistency on receipt."
That sentence means: you must verify it.

**Rule:** For every method you implement, read its doc comment and make a checklist
of every promised behavior. Implement them all before running your first test.

---

## 2. Never Use `unsafe` Without Proving It Is Actually Safe

`unsafe impl Send for Z3PolicyEngine {}` with a raw pointer is one of the most
dangerous patterns in Rust. Here is why it was wrong and how to think about it:

The z3 crate omits `Send + Sync` on `Context` and `Solver` deliberately. That is
documentation. It means the C library has internal mutable state that is not
protected by locks. When you write `unsafe impl Send`, you are telling the compiler
"trust me, I know this is safe." But you did not prove it — you assumed it.

**The correct approach when you hit a `Send + Sync` problem:**

1. Ask: why does the upstream crate not implement it?
2. If the answer is "C library state," the solution is NOT `unsafe impl`. The solution
   is to contain the non-Send type to a single thread.
3. For Z3 specifically: create a new `Context` + `Solver` per verification call.
   Z3 context creation is ~0.1ms. You have a 100ms budget. This is not a performance
   problem, it is a design problem you are avoiding with an unsafe shortcut.

**Rule:** Every `unsafe` block must have a `// Safety:` comment that gives a proof,
not an assertion. "We use Arc to ensure the Context outlives the Solver" is not a
proof — it doesn't address thread safety. A proof would be: "This is only called
from one thread at a time because X ensures mutual exclusion."

If you cannot write the proof, do not write the `unsafe`.

---

## 3. Compile Errors Are Absolute Blockers — Verify Before Claiming Completion

Phase 3 STATUS.md says "Successfully compiled `aethereum-spine` orchestrator." The
current source has four compile errors. This means either:
- The binary was compiled before the errors were introduced (stale), or
- Compilation was not actually verified.

**Rule:** The last thing you do before writing a STATUS.md "COMPLETE" entry is:
```
go build ./...          # for Go
cargo build             # for Rust
npm run build           # for TypeScript
```
If any of these fail, the phase is not complete. Do not update STATUS.md.

**Specifically for Go:** Type errors involving protobuf-generated code are common
because the generated types are not what you expect. Always grep the generated `.pb.go`
file to confirm field names and enum package paths before using them:

```bash
grep "VERDICT_" aethereum-spine/internal/grpc/pb/orchestrator.pb.go
grep "type VerdictQuery " aethereum-spine/internal/grpc/pb/orchestrator.pb.go
```

---

## 4. Persistence Means Written to Disk — In-Memory Is Not Persistence

The Merkle tree is `struct Tree { leaves []Hash }`. That is a data structure, not
persistence. When the process restarts, it is empty. An audit log that resets on
restart is not an audit log.

**Rule:** For every stateful component, ask: what happens after `kill -9 $(pgrep aethereum-spine)`?
If the answer is "state is lost," you have not implemented persistence.

**For the Merkle tree specifically:**
- Write each leaf to PostgreSQL in the same transaction as the proposal verdict update.
- On startup, read all committed leaves from the database and replay them into the
  in-memory tree before serving any requests.
- Test this with a restart test: append 5 leaves, stop, start, assert root matches.

This is not an advanced technique. It is a basic operational requirement.

---

## 5. Database Migrations Must Run at Startup

Having a `.sql` file is not the same as having a schema. The schema must be applied
before any query runs. The standard pattern is:

```go
// In store.go
func (s *Store) ApplyMigrations(ctx context.Context, migrationsDir string) error {
    // 1. CREATE TABLE IF NOT EXISTS schema_migrations (filename TEXT PRIMARY KEY, applied_at TIMESTAMPTZ)
    // 2. Read files from migrationsDir sorted by name
    // 3. For each file not in schema_migrations: execute it, insert into schema_migrations
}
```

Call it in `main.go` immediately after `NewStore`, before anything else touches the DB.

If you add a migration later, it must be a new numbered file (`002_add_column.sql`),
never an edit to an existing migration. Editing applied migrations is how you corrupt
production databases.

---

## 6. OTel Instrumentation Is Not Optional — It Is a Precondition

CLAUDE.md Section 0 Law 3: "You cannot measure what you cannot observe."

Emitting metrics to `opentelemetry::global::meter()` when no global provider is
initialized is equivalent to not emitting metrics at all. The call succeeds silently
and the data goes nowhere. This is the worst kind of bug — it looks like it works.

**Rule for Rust (safety-rail):**
Initialize the OTLP exporter in `Tier1SafetyRail::new()` or accept a
`SdkMeterProvider` as a constructor argument. Verify the metrics reach the collector
by running the smoke test before marking any phase complete.

**Rule for Go (aethereum-spine):**
Initialize the OTLP exporter in `main.go` before any component that emits metrics.
The pattern:
```go
exp, _ := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithInsecure())
provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp)))
otel.SetMeterProvider(provider)
defer provider.Shutdown(ctx)
```

**Verification:** After startup, run the smoke test. Check `otel-snapshots/latest.json`
for your component's metric names. If they are absent, instrumentation is not working.

---

## 7. TODOs Must Have Proposal References — No Exceptions

Anti-Slop Rule 1: a `TODO` comment must reference a proposal ID.

```rust
// BAD:
// Tier 1 logic for custom constraints (omitted for brevity)

// GOOD:
// TODO(proposals/pending/2026-04-09-antigravity-z3-custom-constraints.md): implement
// generic SMT-LIB2 parsing for dynamically registered constraints
```

"Omitted for brevity" is not acceptable. If it's omitted, the behavior is silently
wrong. In this case, `register_constraint` stores the constraint in the list (so the
fingerprint changes) but never adds it to Z3 (so it's never enforced). This is
a security defect, not a brevity optimization.

If you cannot implement something now, either:
1. Return an explicit error: `Err("custom constraint registration not yet implemented")`
2. File a proposal and add the TODO with the reference

Never silently accept an input and then silently ignore it.

---

## 8. Write Tests Against Real Behavior, Not Happy Paths

`bridge_test.go` has one test that skips the hash verification: `// skipping actual hash for sanity check`. This is not a test — it is documentation that the feature does not work.

A test that catches a real regression must:
1. Set up a real (or minimal fake) state
2. Exercise the exact code path that could fail
3. Assert the outcome is correct

For the safety bridge:
```go
func TestVerifyProposal_WithCorrectHash(t *testing.T) {
    // Arrange: build a proposal payload and compute its real SHA-256
    payload := []byte(`{"operation_type":"modify_file","target_component":"factories","change_description":"test"}`)
    hash := sha256.Sum256(payload)

    // Act: call bridge with real hash
    result, err := bridge.VerifyProposal(id, "test-agent", "desc", payload, hash, "", false, 0)

    // Assert: should pass (no constraint violation) and hash must match
    require.NoError(t, err)
    assert.True(t, result.IsSafe)
}

func TestVerifyProposal_WithWrongHash_ReturnsTampered(t *testing.T) {
    payload := []byte(`{"operation_type":"modify_file","target_component":"factories","change_description":"test"}`)
    wrongHash := [32]byte{0xFF} // deliberately wrong

    result, err := bridge.VerifyProposal(id, "test-agent", "desc", payload, wrongHash, "", false, 0)

    require.NoError(t, err)
    assert.False(t, result.IsSafe)
    assert.Contains(t, result.Error, "Tampered")
}
```

Write both the positive and negative case. The negative case is usually more valuable.

---

## 9. The Phase Gate Is Not Bureaucracy — It Is the Safety Rail for the Safety Rail

You proceeded to Phase 3 and 4 without verdicts. This is why we have the gate: because
Phase 3 builds on Phase 2's C FFI, and if Phase 2 has trait contract violations,
Phase 3 inherits them. You cannot debug a broken system that's built on a broken
foundation.

The gate also exists because I (Claude Code) am a scarce resource. You proceeding
without a verdict means you spent significant compute on Phase 3 and 4 that will need
to be partially redone. The gate is there to prevent that waste, not to create it.

**Rule:** When Phase N is done, submit a briefing packet and stop. Update STATUS.md
to "AWAITING ANALYST VERDICT." Do not begin Phase N+1 until a verdict is in
`analyst-verdicts/`.

---

## 10. Resilience Means Handling the Bad Path, Not Just the Happy Path

`store.set_fuel(fuel_limit).unwrap()` panics if fuel was already set or the store is
in an error state. `CString::new(msg).unwrap()` panics if `msg` contains a null byte.

**Rule:** `unwrap()` and `expect()` are acceptable in:
- Test code
- Initialization code where failure should be fatal and the error is unrecoverable
- Cases where you can prove the value cannot be None/Err

`unwrap()` is not acceptable in:
- Any function on the hot path (verification, execution, request handling)
- Any FFI boundary (where panics may unwind into C and cause UB)
- Any code that handles external input (where the input can be adversarial)

For FFI specifically: a panic across a C FFI boundary is undefined behavior. The
`c_api.rs` functions are all `extern "C"` and must not panic under any circumstances.
Replace every `unwrap()` in `c_api.rs` with a fallback:

```rust
// Instead of:
verdict.error_message = CString::new("Null handle").unwrap().into_raw();

// Use:
if let Ok(s) = CString::new("Null handle") {
    verdict.error_message = s.into_raw();
}
// If even CString::new fails, leave error_message null — the caller must handle null.
```

---

## Summary Checklist — Before Every STATUS.md "COMPLETE" Entry

- [ ] `cargo build` / `go build ./...` / `npm run build` — zero errors, zero warnings
- [ ] All tests pass: `cargo test --features tier1` / `go test ./...` / `npm test`
- [ ] Every new public function has at least one test for a real failure case
- [ ] OTel metrics verified in `otel-snapshots/latest.json`
- [ ] No `unwrap()` on hot paths or FFI boundaries
- [ ] No `unsafe` without a `// Safety:` proof comment
- [ ] No `TODO` without a `proposals/pending/` reference
- [ ] No in-memory-only state that must survive restarts
- [ ] Migrations applied and verified against real database
- [ ] Briefing packet written to `analyst-inbox/` — not started on next phase yet
