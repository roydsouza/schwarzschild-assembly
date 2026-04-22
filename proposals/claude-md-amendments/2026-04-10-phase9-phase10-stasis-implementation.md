# CLAUDE.md Amendment — Phase 9 and Phase 10 Definition
**Date:** 2026-04-10
**Proposal ID:** amendment-2026-04-10-phase9-phase10
**Status:** PENDING

## Phase 9 — STASIS Tier 1 Hardening

**Goal:** Make the decidability guarantee concrete. Implement the Tier 1 linter
and migrate all hard safety invariants into properly-tagged Tier 1 predicates.

**Deliverables:**
- `core-station/bridge/validate-stasis-tier1.pl` — Tier 1 syntactic linter
- `core-station/protoplasm/policies/invariants.pl` — Tier 1-tagged hard invariants
- Integration of linter into `core-station/bridge/pre-submit.sh`
- Tests: `core-station/protoplasm/tests/test_tier1_linter.pl`

**Completion checklist:**
✓ ls core-station/bridge/validate-stasis-tier1.pl
✓ grep -c 'stasis_tier(1' core-station/protoplasm/policies/invariants.pl   # >= 3
✓ grep -c 'validate-stasis-tier1' core-station/bridge/pre-submit.sh        # >= 1
✓ swipl -g "consult('core-station/bridge/validate-stasis-tier1.pl'), halt."
✓ cd core-station/aethereum-spine && go build ./...
✓ cd core-station/security && cargo test --features tier1

## Phase 10 — STASIS Self-Improvement Infrastructure

**Goal:** Implement the self-improvement loop and supporting infrastructure
described in STASIS-SELF-IMPROVEMENT.md.

**Deliverables:**
- `core-station/protoplasm/meta/meta_interpreter.pl` — instrumented meta-interpreter
- `core-station/protoplasm/meta/ebg.pl` — explanation-based generalization
- `core-station/protoplasm/meta/abduction.pl` — abductive diagnosis
- Upgrade `core-station/protoplasm/meta/improve.pl` — full 6-state loop
- Upgrade `core-station/protoplasm/meta/introspect.pl` — skill_record/2
- Upgrade `core-station/protoplasm/meta/fitness.pl` — composite fitness with counter-metrics
- Tests for each new module

**Completion checklist:**
- [3.1] Meta-interpreter `solve/2` with OTel hooks and < 1001 depth limit.
- [3.2] `skill_record/2` with SHA-256 versioning.
- [3.3] `improve.pl` with 6-state machine and 5 mandatory OTel metrics.
- [3.4] EBG skeleton `generalize_success/3`.
- [3.5] Abduction skeleton `diagnose_regression/3`.
- [3.6] Fitness upgrade `improvement_score/3` with 3x regression penalty.
