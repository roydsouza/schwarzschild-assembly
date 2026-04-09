# Phase 2 Resubmission Briefing (Round 5) — Safety Rail Tier 1

**Date:** 2026-04-09 03:00:00 UTC
**Artifact:** safety-rail/ — Phase 2 Safety Rail Tier 1 Remediation
**Status:** READY FOR REVIEW (Resolved DoS vector)

## Summary of Remediation

This round addresses the specific denial-of-service (DoS) vector identified in the Round 4 audit by shifting the "unknown constraint" guard from the verification loop to the registration gate.

### 1. Security Compliance (REGRESSION-5) — Early Rejection Guard
- **Fix:** Implemented a `SUPPORTED_CONSTRAINT_NAMES` whitelist in `register_constraint`.
- **Logic:** Calls to `register_constraint` with unknown names now return `RegistrationResult::UnsupportedAssertionKind`. This ensures that `verify()` only ever encounters constraints it knows how to evaluate.
- **Invariant:** The `_` arm in `verify()` has been restored to an unconditional fatal `Err` (Invariant Violated), as it is now a defensive assertion that should be unreachable in a healthy system.
- **Status:** RESOLVED ✓ (DoS vector eliminated)

### 2. CGO Safety (NEW-1)
- **Status:** REMAINS RESOLVED ✓ (Confirmed in full Go build)

### 3. Test Suite Hardening
- **New Test:** Added `test_unsupported_constraint_rejected` which explicitly verifies the new `UnsupportedAssertionKind` return for unknown names.
- **Updated Tests:**
    - `test_stale_proof_rejected`: Updated to use the valid `safety_no_self_modify_safety_rail` name to ensure fingerprint shifts are correctly triggered.
    - `test_duplicate_constraint_rejected`: Updated to use a supported name so the duplicate detection logic is exercised rather than the unsupported name guard.
    - `test_empty_justification_rejected`: Updated to use a supported name to ensure the justification check is exercised.
- **Status:** RESOLVED ✓

## Verification Evidence

### Test Suite Execution
- `unittests`: 6 PASSED
- `contract_compliance_tests`: 9 PASSED (including new `UnsupportedAssertionKind` test)
- `sandbox_tests`: 3 PASSED
- **Total: 18/18 PASSED**

### Build Verification
- **Rust:** `cargo build --release` (Verified).
- **Go Bridge:** `go build ./...` and `go test ./...` in `root-spine/` verified. The CGO boundary remains intact and is non-panicking.

## Request
I request the final **APPROVED** verdict for Phase 2. I have addressed the DoS vector by closing the "registration door," ensuring that the safety rail remains functional even when presented with unsupported constraint types. I am standing by at the gate and will not proceed to Phase 3 execution until the verdict is issued.
