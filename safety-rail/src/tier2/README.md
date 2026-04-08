# Safety Rail — Tier 2 (rocq-of-rust)

## Status: Scaffolded — awaiting Tier 1 verification

Tier 2 provides machine-checkable Rocq proof certificates for every constraint
admission and every `SafetyVerdict::Safe` result. It is a strict superset of Tier 1.

## Upgrade Path

Tier 2 is not implemented in Phase 2. The self-optimization loop will propose Tier 2
proofs over time as the constraint set stabilizes. Each Tier 2 proof corresponds to
a specific safety invariant that Tier 1's Z3 model checking establishes dynamically
but Tier 2 proves statically.

## Prerequisites

- Tier 1 must be fully implemented, tested, and producing a stable constraint set.
- rocq-of-rust toolchain installed (see bootstrap.sh).
- At least one Z3 proof certificate that has been manually verified by the Analyst Droid.

## Directory Layout (when implemented)

```
tier2/
├── src/
│   ├── mod.rs          # Tier2SafetyRail: implements SafetyRail + generates Rocq terms
│   └── proof_gen.rs    # Z3 model → Rocq term translation
├── proofs/
│   ├── *.v             # Rocq source files (version-controlled)
│   └── *.vo            # Compiled Rocq objects (gitignored, built from .v)
└── README.md           # This file
```

## Proposal Reference

Any implementation work here requires a filed proposal in `proposals/pending/`
referencing the specific Tier 1 constraints being promoted to Tier 2.
See `proposals/README.md` for the proposal format.
