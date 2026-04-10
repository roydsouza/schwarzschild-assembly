# CLAUDE.md Amendment Proposal
**Date:** 2026-04-10
**Proposal ID:** amendment-2026-04-10-stasis-tiered-logic-substrate
**Proposed by:** Roy Peter D'Souza (Human), incorporating Analyst Droid technical review
**Status:** APPROVED — explicit operator directive
**Supersedes:** PROPOSAL-04-10.md (AntiGravity draft, now deleted)

---

## Summary

Formalises the **STASIS** tiered logic substrate as the official name and technical
specification for the Protoplasm layer of every Schwarzschild Spacecraft. Incorporates
the three-tier architecture proposed in PROPOSAL-04-10.md with corrections from the
Analyst Droid's review, and connects the Datalog decidability requirement explicitly to
the HOOTL / Dark Factory trajectory.

---

## STASIS — Name and Definition

**STASIS** is used as a proper noun. No backronym is required. As a word, *stasis*
means stable equilibrium — a system that maintains itself without external intervention.
This maps exactly to the Dark Factory goal: autonomous, self-correcting stability
operating within formally verified safety constraints.

If a subtitle is needed for documentation headers:

> **STASIS** — *Substrate for Termination-Assured Symbolic Inference Systems*

The key load-bearing word is **Termination-Assured** — this is what the Tier 1 Datalog
core actually provides, and it is the prerequisite for HOOTL operation.

---

## The Three-Tier Architecture

### Tier 1 — The Invariant Core (Datalog)

**Purpose:** Hard safety invariants — resource boundaries, checksum integrity, policy
prohibitions that must never be violated and must always return an answer.

**Guarantee:** Decidable termination. A safety check that can loop is incompatible with
autonomous operation. Tier 1 is the technical prerequisite for removing Roy from the loop.

**Implementation in SWI-Prolog:**
- Predicates tagged `[stasis_tier(1)]` are restricted to Datalog-style syntax:
  no functor symbols in rule heads (only constants and variables), stratified negation
  only, no side effects.
- SWI-Prolog's tabling (`use_module(library(tabling))`, `:- table pred/n.`) prevents
  looping for tabled predicates, but tabling alone does not enforce Datalog syntax.
- **Enforcement is required:** A `scripts/validate-stasis-tier1.pl` script must verify
  at submit-time that all `[stasis_tier(1)]` predicates conform to the syntactic
  restrictions. Tabling + syntactic enforcement together provide the decidability
  guarantee. Tabling alone does not.

**Critical distinction:** "Won't loop" (tabling) ≠ "definitely terminates in bounded
time regardless of input" (true Datalog). The safety argument requires the latter.

### Tier 2 — The Constraint Evolution Layer (CHR)

**Purpose:** Dynamic policy and domain-specific strategy. The Protoplasm's "current
beliefs" about how to behave, expressed as a constraint store that evolves with experience.

**Guarantee:** Transparent state transition audit. Every constraint propagation and
simplification step is an observable, logged event. This is intrinsic introspection.

**Implementation in SWI-Prolog:**
- `library(chr)` is mature, deeply integrated, and compiles to optimised Prolog code.
- CHR rules in Tier 2 **may only call Tier 1 predicates** in their guards and bodies.
  Calling arbitrary Prolog from a CHR rule escapes the tier boundary and voids the
  Tier 1 decidability guarantee. This restriction is enforced by the Safety Rail Z3
  policy `stasis_tier2_calls_tier1_only`.
- The existing `agents/prolog-substrate/policies/constraints.chr` is a Tier 2 artifact.

### Tier 3 — The Meta-Introspection Layer

**Purpose:** Agents reasoning about Tier 1 and Tier 2 rules — reading, constructing,
and proposing amendments to the Protoplasm's own logic.

**Implementation options:**

*Option A — ELPI (for genuine higher-order unification):*
ELPI is a mature, embeddable λProlog interpreter. It provides lexical scoping and
higher-order unification, which correctly handles variable binding when constructing
new logic fragments at runtime. Required if the meta-improvement loop generates clauses
with complex binder structures.
- Cost: ELPI is a separate interpreter; every SWI→ELPI call has a context-switch cost.
  A benchmark on M5 hardware is **required before adoption** on any hot path.

*Option B — SWI-Prolog `copy_term/2` + `numbervars/3` (for variable capture safety):*
Variable capture in generated terms — the stated motivation for λProlog — can be
handled in standard Prolog using `copy_term/2` (creates a fresh copy with new variables)
and `numbervars/3` (binds variables to distinguishable terms before comparison). This
covers the common cases in `meta/improve.pl` without a foreign interpreter dependency.

*Option C — `library(yall)` (for anonymous predicates only):*
YALL provides `\X^Goal` syntax as sugar for `call`. It does **not** provide higher-order
unification and is **not** a substitute for λProlog. Use YALL only for anonymous
predicate passing, not for meta-reasoning over clause structure.

**Decision point for implementation:** Start with Option B (zero new dependencies).
Adopt ELPI only if a specific meta-reasoning task is demonstrably impossible with
`copy_term/2` + `numbervars/3`, and only after a benchmark confirms acceptable latency.

---

## Tier Boundary Enforcement

The tier architecture is only meaningful if boundaries are enforced. Three enforcement
layers are required:

1. **Syntactic linter** (`scripts/validate-stasis-tier1.pl`) — run at pre-submit time.
   Rejects any Tier 1 predicate that uses functor symbols in heads, non-stratified
   negation, or calls non-tabled predicates.

2. **Z3 Safety Rail policy** — `stasis_tier2_calls_tier1_only` constraint blocks any
   CHR rule body that references a predicate not tagged `[stasis_tier(1)]`.

3. **`safe_assert` CHR check** — the existing `constraints:check_constraints/1` in Tier 2
   is extended to reject any proposed clause whose head matches a Tier 1 predicate
   (modification of invariants at runtime is forbidden).

The dependency graph must be strictly one-way: Tier 3 reads/reasons about Tier 1 and 2,
never modifies them. Tier 2 calls Tier 1 predicates only. Tier 1 has no external calls.

---

## Connection to HOOTL / Dark Factory

The HOOTL transition (documented in PROCESS.md) requires that the assembly can make
safety decisions without a human in the loop. This is only safe when:

1. Safety checks are **guaranteed to terminate** (can't hang the orchestrator)
2. Safety checks are **formally verifiable** (can be proven correct, not just tested)
3. The **self-evolution loop cannot modify the safety checks themselves**

Tier 1 (Datalog decidability) satisfies requirement 1.
Tier 1 + Safety Rail Z3 proofs satisfy requirement 2.
The tier boundary enforcement (Tier 3 cannot modify Tier 1) satisfies requirement 3.

STASIS Tier 1 is therefore not an optimisation — it is a technical prerequisite for the
Dark Factory. Any HOOTL timeline that does not include a verified Tier 1 Datalog core
is operating on an unsafe assumption.

---

## Migration Path from Phase 8 Deliverables

The existing Phase 8 code maps onto the STASIS tier model as follows:

| Existing artifact | STASIS tier |
|---|---|
| `policies/constraints.chr` | Tier 2 — CHR constraint store |
| `core/safety_bridge.pl` | Infrastructure (cross-tier — manages assertions) |
| `meta/verify.pl` | Tier 3 — meta-invariant checking |
| `meta/improve.pl` | Tier 3 — meta-improvement loop |
| `meta/introspect.pl` | Tier 3 — clause introspection |
| Future: hard safety invariants | Tier 1 — to be defined in Phase 9 |

No existing code needs to be deleted. The migration is additive: Tier 1 predicates are
written fresh as the safety invariant set is formalised, tagged with `[stasis_tier(1)]`,
and placed under the Tier 1 linter.

---

## Changes to CLAUDE.md

1. **§3 Phase 8 title:** "Prolog Self-Enhancement Framework" → "STASIS Self-Enhancement Framework"
2. **§3 Phase 7 language table:** "Prolog (SWI-Prolog)" → "STASIS (SWI-Prolog runtime)"
3. **§3 Phase 8 body:** update "In Prolog…" prose to name STASIS as the language
4. **§9 Runtime section:** retitle "SWI-Prolog" entry as "SWI-Prolog / STASIS" and add tier reference
5. **§10 Phase 8 checklist heading:** "Prolog Self-Enhancement" → "STASIS Self-Enhancement"
6. **Domain fitness extension:** "Prolog Substrate" → "STASIS Substrate"

---

## New Deliverable

`STASIS-LANGUAGE.md` — top-level reference document in `schwarzschild-assembly/`.
Defines: the problem STASIS solves, the three-tier architecture, tier enforcement,
SWI-Prolog implementation details, code examples, integration with Safety Rail and
Merkle log, HOOTL trajectory, and file organization.

---

## No Regression

This amendment does not change the Safety Rail trait contract, the Merkle leaf schema,
`scripts/pre-submit.sh`, or any Go/Rust source. The `agents/prolog-substrate/` directory
retains its name (filesystem rename deferred; prose updated throughout). All prior
APPROVED verdicts remain valid.
