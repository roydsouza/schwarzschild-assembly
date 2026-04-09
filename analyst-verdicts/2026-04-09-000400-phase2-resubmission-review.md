# Analyst Verdict
**Date:** 2026-04-09 00:04:00 UTC
**Artifact:** `safety-rail/src/tier1/` (Phase 2 re-submission)
**Verdict:** CONDITIONAL

---

## Verdict Rationale

Three of the six required changes from the previous CONDITIONAL verdict are resolved. Two critical defects remain open and one significant defect remains open. Phase 3 work must not resume until all three are addressed.

### Resolved (do not revisit)

**CRITICAL-1 — Proof certificate is now substantive.**
`verify()` returns `Option<String>` containing the serialized Z3 model on `Sat`. `mod.rs` converts it to bytes, computes a SHA-256 digest, and populates `ProofCertificate { bytes, digest, tier }` correctly. The contract's proof requirement is satisfied.

**CRITICAL-2 — Payload hash verification is now present.**
`verify_proposal` computes `SHA-256(proposal.payload)` before extracting facts and returns `SafetyVerdict::TamperedPayload` on mismatch. The fields `claimed_hash` and `actual_hash` are populated correctly. The trait contract's consistency check is satisfied.

**SIGNIFICANT-5 — Unknown constraints now fail explicitly.**
The `_` arm in `verify()` returns `Err(format!("Unknown constraint type not yet implemented: {}"))` instead of silently dropping. The silent-drop security defect is closed.

---

### Still Open — Must Fix Before Phase 3

**CRITICAL-4 — `unsafe impl Send` and `unsafe impl Sync` must be deleted.**

Lines 63–64 of `z3_policy.rs`:
```rust
unsafe impl Send for Z3PolicyEngine {}
unsafe impl Sync for Z3PolicyEngine {}
```

These are still present. The comment reads: "Z3 Context is only used within a Mutex lock." That comment is an *assertion*, not a *proof*. Here is why it is insufficient:

The z3 crate does not implement `Send + Sync` on `Context` because the underlying libz3 C library has internal mutable state — a global allocator, symbol tables, and other non-reentrant structures — that are not protected by any Rust-visible lock. The `Mutex<Arc<Context>>` you added prevents two Rust threads from racing on the *Rust wrapper*, but it does not prevent libz3 from accessing its own C-side globals from whichever thread acquires the lock. The `Mutex` protects the *pointer*, not the *C library state the pointer touches*.

The correct fix does not require `unsafe` at all. You already create a fresh `Solver` per `verify()` call. The remaining step is: create a fresh `Context` per `verify()` call as well, and drop both at the end of the call. Remove the `ctx` field from `Z3PolicyEngine` entirely. The struct then contains only `constraints: Mutex<Vec<PolicyConstraint>>`, which is `Send + Sync` without any `unsafe impl`.

Z3 context creation is ~0.1ms. You have a 100ms budget. This is not a performance concern.

**Required change:**
```rust
// Remove these two lines entirely:
unsafe impl Send for Z3PolicyEngine {}
unsafe impl Sync for Z3PolicyEngine {}

// Remove ctx field from Z3PolicyEngine:
pub(crate) struct Z3PolicyEngine {
    constraints: Mutex<Vec<PolicyConstraint>>,
}

// In verify(), create context and solver locally:
pub fn verify(&self, facts: &ProposalFacts) -> Result<(Option<ViolationReport>, Option<String>), String> {
    let cfg = Config::new();
    let ctx = Context::new(&cfg);
    let solver = Solver::new(&ctx);
    // ... rest of verify unchanged, ctx/solver drop at end of function
}
```

No `unsafe`. No shared `Context`. No Mutex on Context. `Z3PolicyEngine` becomes `Send + Sync` automatically because `Mutex<Vec<PolicyConstraint>>` is `Send + Sync`.

---

**CRITICAL-3 — Circular protection in `register_constraint` must be verified and confirmed.**

`mod.rs` `register_constraint` (lines 238–316) shows that `self.verify_proposal(&proposal)` is called before `self.z3_engine.add_constraint()`. This structure appears correct.

However, the `verify_proposal` call uses a synthetic `ActionProposal` built from the constraint's metadata, with `is_security_adjacent: true`. The payload hash is computed correctly. The `target_path` is `None` (maps to empty string in `extract_facts`). This proposal will pass the three existing constraints only because `is_security_adjacent: true` satisfies both path-prefix constraints, and `target_component` is not `"merkle-log"` with `operation_type` not `"delete_file"`.

**The latent defect:** The circular protection only blocks if the synthetic proposal *violates* an existing constraint. It does not verify the *new constraint itself* for soundness or confliction. A malicious or buggy constraint that passes all existing checks will be admitted regardless of what it asserts. This is acceptable for Phase 2 — Tier 2 rocq-of-rust proofs are the long-term answer — but the current code path must be confirmed to work end-to-end.

**Required action:** Add a test that confirms a constraint with `justification: ""` is rejected with `RegistrationResult::MissingJustification`, and a test that confirms a duplicate constraint ID is rejected with `RegistrationResult::Duplicate`. These test the guards in `add_constraint` *through* the `register_constraint` public interface, which is the path that runs the circular protection. Confirm with `cargo test --features tier1`.

---

**SIGNIFICANT-6 — OTel meter provider is still not initialized.**

`Tier1SafetyRail::new()` calls `global::meter("sati-central-safety-rail")` for all five metric instruments. If no `SdkMeterProvider` has been set as the global provider, every metric emission is a no-op. This is the worst failure mode: the code compiles, the counters increment in Rust address space, and nothing appears in `otel-snapshots/latest.json`.

The standard pattern from `analyst-inbox/2026-04-09-000300-code-quality-standards.md` Section 6:
```rust
// In Tier1SafetyRail::new(), before creating metrics:
use opentelemetry_otlp::WithExportConfig;
use opentelemetry_sdk::metrics::SdkMeterProvider;
use opentelemetry_sdk::runtime;

let exporter = opentelemetry_otlp::new_exporter()
    .tonic()
    .with_endpoint("http://localhost:4317");
let provider = opentelemetry_otlp::new_pipeline()
    .metrics(runtime::Tokio)
    .with_exporter(exporter)
    .build()?;
opentelemetry::global::set_meter_provider(provider);
```

Alternatively, accept a `SdkMeterProvider` as a constructor argument so the caller controls the exporter (preferred for testability):
```rust
pub fn new(meter_provider: Option<opentelemetry_sdk::metrics::SdkMeterProvider>) -> Result<Self, String> {
    if let Some(provider) = meter_provider {
        opentelemetry::global::set_meter_provider(provider);
    }
    // ... rest of new()
}
```

**Required change:** Either initialize the provider in `new()` or accept it as a parameter. After fixing, run the observability smoke test and confirm the five `sati_central.safety.*` metric names appear in `otel-snapshots/latest.json`.

---

### Advisory — Not a Blocker

**TEST-7 — `contract_tests` compliance helpers.**
`safety-rail/src/lib.rs` defines a `contract_tests` module with five helper functions:
- `assert_tampered_payload_rejected`
- `assert_timeout_enforced`
- `assert_safe_proof_nonempty`
- `assert_unsafe_has_violation`
- `assert_policy_fingerprint_stable`

These are listed as "compliance helpers that Tier1 tests MUST invoke." The Phase 2 re-submission tests in `z3_policy.rs` test constraint logic directly but do not invoke these helpers. Before Phase 2 is APPROVED, add a test module in `tier1/` that calls all five helpers against a live `Tier1SafetyRail` instance. This confirms the implementation satisfies the trait contract, not just the internal Z3 logic.

---

## Required Changes Summary

- [ ] Delete `unsafe impl Send for Z3PolicyEngine {}` and `unsafe impl Sync for Z3PolicyEngine {}`. Remove `ctx` field. Create fresh `Config` + `Context` inside each `verify()` call. No `unsafe` should appear in `z3_policy.rs`.
- [ ] Add `register_constraint` integration tests: `MissingJustification` rejection path, `Duplicate` rejection path. Run `cargo test --features tier1` and confirm both pass.
- [ ] Initialize OTLP meter provider in `Tier1SafetyRail::new()` or accept it as a constructor parameter. Verify `sati_central.safety.*` metrics appear in `otel-snapshots/latest.json` after running the smoke test.
- [ ] (Advisory) Add `contract_tests` compliance invocations in `tier1/` test module before the final APPROVED verdict.

---

## Fitness Vector Impact Assessment

- **Safety compliance:** Three contract violations closed. Two remain. The `unsafe impl` issue is a soundness defect — until it is removed, the safety rail itself is not provably safe under concurrent access. No regression possible after the fix.
- **Audit integrity:** No impact on Merkle log at this phase.
- **Dhamma alignment:** Not applicable to this layer.
- **System performance:** Fresh-Context-per-verify adds ~0.1ms. 100ms budget is not threatened.
- **Operational cost:** No impact.

---

## Safety Rail Implications

The `unsafe impl Send/Sync` removal is a safety-rail-internal change. It does not alter the public trait contract or any policy constraint. It does not require a proposal or Translucent Gate approval — it is a correctness fix within Phase 2's scope. No policy fingerprint change results.

---

## Merkle Log Entry

```json
{
  "event_type": "VetoIssued",
  "agent_id": "analyst-droid",
  "artifact": "safety-rail/src/tier1/",
  "verdict": "CONDITIONAL",
  "payload_hash": "<SHA-256 of this verdict file>",
  "safety_cert": null,
  "dhamma_ref": null,
  "fitness_delta": null,
  "model_version": "claude-sonnet-4-6",
  "note": "Phase 2 re-submission: 3/6 prior items resolved; 2 critical + 1 significant still open"
}
```
