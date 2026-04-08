# Safety Rail

Formal verification layer for Sati-Central. All agent-proposed actions pass through
this component before execution.

## What it does

- Verifies `ActionProposal` instances against a Z3 SMT constraint set (Tier 1)
- Executes verified artifacts in a Wasmtime sandbox with memory and time limits
- Provides a policy fingerprint written to every Merkle leaf for auditability
- Protects itself: `register_constraint` passes through `verify_proposal` before admission

## Depends on

- `z3` crate (SMT solver) — Tier 1
- `wasmtime` crate (WASM runtime) — Tier 1
- `opentelemetry` crate — metric emission

## How to run tests

```bash
cargo test --manifest-path safety-rail/Cargo.toml
# With Tier 1 features:
cargo test --manifest-path safety-rail/Cargo.toml --features tier1
```

## Metrics emitted

| OTel Metric Name | Type | Fitness Vector |
|-----------------|------|----------------|
| `sati_central.safety.violations_total` | counter | safety_compliance |
| `sati_central.safety.verifications_total` | counter | safety_compliance |
| `sati_central.safety.verification_duration_ms` | histogram | system_performance |
| `sati_central.safety.constraints_total` | gauge | audit_integrity |

## Tier upgrade path

Tier 2 (rocq-of-rust proofs) is scaffolded in `src/tier2/`. See that directory's
README for the upgrade path. Tier 2 work requires a proposal in `proposals/pending/`.
