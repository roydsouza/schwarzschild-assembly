# Proposals — Lifecycle and Process

Every significant system change flows through this directory before implementation.
This is not bureaucracy — it is the audit trail that makes the self-optimization loop safe.

---

## Directory Layout

```
proposals/
├── pending/          # Submitted, awaiting verdict
├── approved/         # Approved by Analyst Droid (+ Translucent Gate if security-adjacent)
├── rejected/         # Rejected; file preserved with rejection rationale
└── claude-md-amendments/  # Approved amendments to CLAUDE.md specifically
```

---

## File Naming Convention

`YYYY-MM-DD-HHMMSS-<agent>-<short-slug>.md`

Examples:
- `2026-04-08-143022-antigravity-tier1-z3-policy-expansion.md`
- `2026-04-08-161500-claude-code-fitness-weight-rebalance.md`

---

## Required Fields (all proposals)

```markdown
# Proposal: <Short Title>
**ID:** YYYY-MM-DD-HHMMSS-<agent>-<slug>  (matches filename without .md)
**Author:** <AntiGravity | Claude Code>
**Date:** YYYY-MM-DD HH:MM:SS UTC
**Category:** <Security-Adjacent | Performance | Factory-Logic | RAG-Strategy | UI | Schema | Amendment>
**Security-Adjacent:** <yes | no>  (yes = requires Translucent Gate regardless of fitness delta)

## Summary
<One paragraph. What changes. Why. What doesn't change.>

## Affected Components
- <path/to/component> — <what changes>

## Fitness Vector Impact
| Metric | Projected Delta | Confidence |
|--------|----------------|------------|
| Safety compliance | ... | ... |
| Audit integrity | ... | ... |
| Dhamma alignment | ... | ... |
| System performance | ... | ... |
| Operational cost | ... | ... |

## Verification Plan
<How will this be tested? What benchmark proves the claim? What rollback triggers exist?>

## Revert Artifact
<Path to revert script or description of revert procedure>
```

---

## Lifecycle

```
[submitted to pending/]
        ↓
[Analyst Droid reviews] → VETOED → stays in pending/ with verdict attached → can be revised
        ↓ APPROVED (or CONDITIONAL with required changes complete)
[Security-Adjacent?]
  YES → Translucent Gate (human approval required)
  NO  → auto-approve if fitness delta positive, no global metric regresses
        ↓
[moved to approved/]
[RevertArtifact written to approved/ alongside proposal]
[Merkle leaf committed: event_type=GateApproved or FactoryCommit]
        ↓
[implemented]
[24-hour canary window begins for non-security changes]
```

---

## CLAUDE.md Amendment Flow

Amendments to `CLAUDE.md` are the most sensitive proposal type.

1. File in `proposals/claude-md-amendments/` (not `pending/`) directly.
2. Analyst Droid reviews on next invocation.
3. If approved: Analyst Droid writes the amendment directly to `CLAUDE.md` and
   commits with message `docs: accept amendment <proposal-id>`.
4. Merkle leaf: `event_type=AmendmentAccepted`.
5. If rejected: Analyst Droid writes rejection rationale to the proposal file and
   moves it to `proposals/rejected/`.

**No amendment takes effect without an explicit Analyst Droid approval written to
`proposals/claude-md-amendments/`.**

---

## Anti-Slop Rule

A `TODO` comment in any source file must reference a proposal ID:

```rust
// TODO(proposals/pending/2026-04-08-143022-antigravity-tier2-proofs.md): implement Tier 2
```

Orphaned TODOs (no corresponding proposal file) are veto-eligible on review.
