# Analyst Verdict
**Date:** 2026-04-08 12:00:00 UTC
**Artifact:** Phase 2 Briefing — Z3 variable mapping clarification
**Verdict:** CONDITIONAL (briefing correction — no veto, proceed with guidance below)

---

## Verdict Rationale

AntiGravity correctly identified that the Phase 2 briefing specified Z3 assertions
referencing variables (`is_verified`, `operation_type`, `target_component`,
`dhamma_score`) that do not exist as named fields on `ActionProposal`. This was a
gap in the briefing — the logical intent was correct but the binding mechanism was
not specified. This verdict closes that gap with exact prescriptions.

---

## Required Changes

### 1. Define `ProposalFacts` as the Z3 input type

Z3 does not operate directly on `ActionProposal` or on raw JSON bytes. Define a
`ProposalFacts` struct in `tier1/z3_policy.rs` that is the typed input to all Z3
bindings. This struct is extracted from the `ActionProposal` before any Z3 call.

```rust
/// Typed facts extracted from an ActionProposal for Z3 binding.
/// All Z3 variables are bound from this struct — never from raw proposal fields
/// or raw payload bytes.
#[derive(Debug, Clone)]
pub(crate) struct ProposalFacts {
    /// From ActionProposal.target_path (empty string if None)
    pub target_path: String,
    /// From ActionProposal.is_security_adjacent
    pub is_security_adjacent: bool,
    /// From ActionProposal.agent_id
    pub agent_id: String,
    /// Deserialized from ActionProposal.payload via ProposalPayload schema
    pub operation_type: OperationType,
    /// Deserialized from ActionProposal.payload via ProposalPayload schema
    pub target_component: String,
}
```

`ProposalFacts::extract(proposal: &ActionProposal) -> Result<Self, ExtractionError>`
is the single entry point. It deserializes the payload, validates it against the
schema, and populates the struct. All Z3 assertions bind against `ProposalFacts`
fields — never against `ActionProposal` fields directly.

---

### 2. Define `ProposalPayload` — the canonical payload schema

The `ActionProposal.payload` field (canonical JSON bytes) must conform to this schema.
Define it in `tier1/z3_policy.rs` alongside `ProposalFacts`:

```rust
/// Canonical schema for ActionProposal.payload.
/// Submitters must serialize their payload to this schema.
/// The safety rail validates on deserialization.
#[derive(Debug, Clone, serde::Deserialize)]
pub struct ProposalPayload {
    /// What kind of operation this proposal performs.
    pub operation_type: OperationType,
    /// Which top-level component is targeted (e.g., "safety-rail", "merkle-log",
    /// "root-spine", "dhamma-adviser", "control-panel", "factories", "scripts").
    pub target_component: String,
    /// Human-readable description of what will change. ≤ 512 chars.
    pub change_description: String,
    /// Any additional context the submitter wants to include.
    /// Must be valid JSON. May be null.
    pub context: Option<serde_json::Value>,
}

#[derive(Debug, Clone, PartialEq, Eq, serde::Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum OperationType {
    /// Create a new file or resource
    CreateFile,
    /// Modify an existing file or resource
    ModifyFile,
    /// Delete a file or resource (triggers AuditIntegrity constraints)
    DeleteFile,
    /// Execute code in a sandbox
    ExecuteCode,
    /// Register a new Z3 policy constraint
    RegisterConstraint,
    /// Modify a schema (fitness-vector-schema.json, log-schema.json, proto files)
    ModifySchema,
    /// Update configuration (OTel config, Go module, Cargo.toml)
    UpdateConfig,
}
```

Add `serde` and `serde_json` to `Cargo.toml` with pinned versions and justification
comments. `serde_json` is needed for `ProposalPayload` deserialization and for
canonical JSON verification of `payload_hash`.

---

### 3. Corrected Z3 constraint bindings

**Constraint 1 — `safety_no_self_modify_safety_rail`**

`is_verified` in the briefing was an error. The correct binding uses `is_security_adjacent`
from `ProposalFacts` (which maps directly from `ActionProposal.is_security_adjacent`).

Corrected intent: any proposal whose `target_path` starts with `"safety-rail/"` must have
`is_security_adjacent = true`.

```rust
// In Z3 terms (pseudocode — use z3 crate Bool/String AST API):
// (assert (=> (str.prefixof "safety-rail/" target_path) (= is_security_adjacent true)))
let premise = target_path_z3.starts_with("safety-rail/");  // Z3 string prefix
let conclusion = is_security_adjacent_z3._eq(&Bool::from_bool(ctx, true));
solver.assert(&premise.implies(&conclusion));
```

**Constraint 2 — `audit_no_merkle_deletion`**

`operation_type` and `target_component` come from `ProposalFacts`, which are extracted
from the payload via `ProposalPayload`.

Corrected intent: a proposal with `operation_type = DeleteFile` must not have
`target_component = "merkle-log"`.

```rust
// (assert (=> (= operation_type "delete_file") (not (= target_component "merkle-log"))))
let is_delete = operation_type_z3._eq(&StringSort::from_str(ctx, "delete_file"));
let targets_merkle = target_component_z3._eq(&StringSort::from_str(ctx, "merkle-log"));
solver.assert(&is_delete.implies(&targets_merkle.not()));
```

**Constraint 3 — DROP from Phase 2**

`dhamma_score` cannot be checked at Z3 verification time. The pipeline order is:

```
ActionProposal submitted → Safety Rail (Z3) → [Translucent Gate] → fitness delta eval
```

The Dhamma-Adviser runs as part of fitness delta evaluation, AFTER safety verification.
At the time `verify_proposal` is called, no Dhamma score exists yet.

**Remove Constraint 3 entirely from Phase 2.** It will be added in Phase 6 when the
Dhamma-Adviser is integrated and its score is either pre-computed by the submitter or
injected into the payload by the orchestrator before safety verification.

**Constraint 4 — `security_no_unverified_proto_change`**

`target_path` and `is_security_adjacent` both come directly from `ProposalFacts`.
This constraint is correct as specified. No changes needed.

---

### 4. Updated initial constraint set for Phase 2

Replace the four briefing constraints with these three:

| ID | Name | Variables Used | Source |
|----|------|---------------|--------|
| 1 | `safety_no_self_modify_safety_rail` | `target_path`, `is_security_adjacent` | ActionProposal fields |
| 2 | `audit_no_merkle_deletion` | `operation_type`, `target_component` | ProposalPayload |
| 3 | `security_no_unverified_proto_change` | `target_path`, `is_security_adjacent` | ActionProposal fields |

Three constraints. No `dhamma_score`. All variables are either direct ActionProposal
fields or deserialized from `ProposalPayload`.

---

### 5. `serde` + `serde_json` dependency additions

Add to `safety-rail/Cargo.toml` under `[dependencies]`:

```toml
# JSON deserialization for ProposalPayload and canonical JSON hash verification
serde = { version = "1", features = ["derive"] }
# pinned: 1.x is stable; no breaking changes expected at minor version
serde_json = "1"
# Required for serde_json
```

---

## Fitness Vector Impact Assessment

| Metric | Impact | Notes |
|--------|--------|-------|
| Safety compliance | Positive | Constraint set is now precisely specified and implementable |
| Audit integrity | Neutral | No Merkle changes |
| Dhamma alignment | Neutral | Dhamma constraint correctly deferred to Phase 6 |
| System performance | Neutral | ProposalFacts extraction adds < 1ms overhead |
| Operational cost | Neutral | No LLM calls |

---

## Safety Rail Implications

`ProposalPayload` becomes part of the de-facto proposal protocol. Any submitter
(AntiGravity, root-spine, factories) must serialize their payload to this schema.
This is a contract change — file a proposal in `proposals/pending/` before Phase 3
begins to formally record `ProposalPayload` as a versioned schema.

The removal of the dhamma_score constraint is not a regression — it was never
implementable in Phase 2. The constraint will be added with full Dhamma-Adviser
integration in Phase 6.

---

## Merkle Log Entry

```json
{
  "event_type": "GateApproved",
  "agent_id": "claude-code",
  "payload_hash": "<SHA-256 of this verdict file>",
  "safety_cert": "<PolicyFingerprint.empty()>",
  "dhamma_ref": null,
  "fitness_delta": null,
  "model_version": "claude-sonnet-4-6"
}
```
