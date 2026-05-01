# STASIS Language Reference
*Substrate for Termination-Assured Symbolic Inference Systems*

**Version:** 1.0 — 2026-04-10
**Runtime:** SWI-Prolog 9.x (aarch64-apple-darwin, M5-native)
**Canonical spec:** `proposals/pending/2026-04-10-stasis-tiered-logic-substrate.md`

---

## What Is STASIS?

STASIS is the formal name for the tiered symbolic logic language that implements the
**Protoplasm** layer of every Schwarzschild Spacecraft. It is not a new programming
language — it is a structured discipline for writing SWI-Prolog code that is safe
enough for autonomous operation.

The word *stasis* means stable equilibrium: a system that maintains itself without
external intervention. This is the goal of the Dark Factory—autonomous, self-correcting
operation within formally verified safety boundaries. STASIS provides the technical
substrate that makes that goal achievable by eliminating the non-determinism and
hallucinatory logic often found in unconstrained **AI Slop**.

---

## The Problem STASIS Solves

### Turing-completeness and "Vibe Coding" are unsafe in autonomous systems

Standard Prolog is Turing-complete. A self-evolving agent using unrestricted Prolog
or relying on **Vibe Coding** patterns can accidentally (or maliciously) introduce:

- **Non-terminating recursion** — safety checks that loop forever, hanging the orchestrator
- **Unbounded side effects** — filesystem access, shell execution, network calls from logic code
- **Self-modification of safety rules** — an agent that can rewrite its own constraints is
  an agent that cannot be formally verified

In a Human On The Loop (HOTL) system, Roy can intervene when any of these occur. In a
Human Out Of The Loop (HOOTL) Dark Factory, there is no one to restart a hung safety rail.
A non-terminating safety check is a non-recoverable failure.

### The prior solution was insufficient

Phase 8 introduced `constraints.chr` to ban forbidden predicates (assertz, shell, etc.)
and `safe_assert/1` to gate all knowledge-base mutations through the Safety Rail. This
was the right architecture. What it lacked was a **decidable core**: a layer where
termination is not just enforced by a policy check, but is a mathematical property of
the code itself.

---

## The Three-Tier Architecture

STASIS organises all Protoplasm code into three tiers with explicit boundaries. The tiers
form a strict dependency hierarchy — higher tiers can read lower tiers, but never modify them.

```
┌─────────────────────────────────────────────────────┐
│  Tier 3 — Meta-Introspection                        │
│  Agents reasoning about Tier 1/2 rules.             │
│  SWI-Prolog + copy_term/2 (+ ELPI if needed)        │
├─────────────────────────────────────────────────────┤
│  Tier 2 — Constraint Evolution                      │
│  Dynamic policy and strategy as a constraint store. │
│  SWI-Prolog library(chr)                            │
├─────────────────────────────────────────────────────┤
│  Tier 1 — Invariant Core                            │
│  Hard safety invariants. Decidable by construction. │
│  Datalog-restricted SWI-Prolog with tabling         │
└─────────────────────────────────────────────────────┘
         ↑ reads only, never modifies
```

---

## Tier 1 — The Invariant Core

### What goes here

Hard constraints that must **always hold** and must **always return an answer**:
- Resource boundary checks (memory, file size, request rate)
- Cryptographic integrity invariants (hash must match, signature must verify)
- Policy prohibitions (banned predicates, banned operations)
- Structural invariants (KB consistency rules)

### The decidability guarantee

Tier 1 predicates are restricted to Datalog-style syntax:

| Rule | Why |
|---|---|
| No functor symbols in rule heads | Prevents infinite term construction |
| Stratified negation only | Prevents circular dependencies through negation |
| No direct I/O or side effects | Keeps evaluation pure and repeatable |
| All recursive predicates are tabled | Prevents looping; memoises ground results |

SWI-Prolog's `library(tabling)` provides tabling. Tabling prevents looping but does not
enforce the syntactic restrictions above — that is the job of the Tier 1 linter.

**Important distinction:** Tabling alone is not sufficient. Tabling prevents a tabled
predicate from re-entering a computation it is already in the middle of (preventing loops).
It does not prevent a tabled predicate that calls non-tabled predicates from running for
arbitrarily long. True Datalog decidability requires both tabling AND the syntactic
restrictions. The linter enforces both.

### Writing Tier 1 code

```prolog
% Tier 1 predicate — mark with the tier directive
:- stasis_tier(1, banned_operation/1).
:- table banned_operation/1.

% Facts only — no function symbols in heads
banned_operation(assertz).
banned_operation(retract).
banned_operation(shell).
banned_operation(process_create).
banned_operation(abolish).

% Rules with stratified negation — allowed
:- stasis_tier(1, operation_permitted/1).
:- table operation_permitted/1.

operation_permitted(Op) :-
    \+ banned_operation(Op).
```

### Tier 1 linter

Before any briefing is filed, `scripts/validate-stasis-tier1.pl` verifies that all
predicates tagged `stasis_tier(1, ...)`:
- Have no functor symbols (complex terms) in rule heads
- Use only stratified negation
- Are declared with `:- table pred/n.`
- Have no calls to non-Tier-1 predicates in their bodies

Violation at submit time: the pre-submit script reports a `[FAIL]` and the briefing
cannot be filed.

---

## Tier 2 — The Constraint Evolution Layer

### What goes here

Dynamic policy and domain-specific strategy expressed as **Constraint Handling Rules**.
The Protoplasm's "current beliefs" — rules that change as the spacecraft learns.

- Safety constraint heuristics (patterns to accept/reject beyond hard bans)
- Performance optimisation policies
- Domain-specific decision rules that evolve with experience
- The `check_constraints/1` policy gate (currently in `policies/constraints.chr`)

### Why CHR?

Constraint Handling Rules provide a **constraint store** with a transparent audit trail.
Every constraint propagation and simplification step is observable. When a new clause is
proposed and rejected, the CHR trace shows exactly which rule fired and why. This is
intrinsic introspection — the policy layer explains its own decisions.

CHR also excels at **declarative state evolution**: the spacecraft doesn't change its code,
it updates its constraint store. The boundary between "what the spacecraft knows" and
"what the spacecraft does" remains clean.

**Tier 2 (Declarative Constraints):** CHR rules that restrict the shape of valid knowledge.
**Tier 3 (Invariants & Audit):** Hardened Prolog gates that perform synchronous recursive scanning of terms and maintain a Merkle-signed audit trail for every mutation.

### Tier 2 boundary rule

CHR rules in Tier 2 may only call **Tier 1 predicates** in their guards and bodies.
Calling arbitrary Prolog from a CHR rule escapes the tier boundary and voids the
decidability guarantee of Tier 1 checks.

```prolog
% CORRECT — Tier 2 rule calling only Tier 1 predicates
check_constraints(Clause) <=>
    find_functor(Clause, Name),
    banned_operation(Name)    % Tier 1 predicate — allowed
    | banned_predicate(Name).

% FORBIDDEN — Tier 2 rule calling unrestricted Prolog
check_constraints(Clause) <=>
    call(arbitrary_prolog_predicate(Clause))  % exits tier boundary
    | true.
```

This restriction is enforced by the Safety Rail Z3 policy `stasis_tier2_calls_tier1_only`.

### Writing Tier 2 code

```prolog
:- use_module(library(chr)).
:- chr_constraint check_constraints/1, banned_predicate/1.

% Propagation rule: fire when check_constraints is posted
check_constraints(Clause) <=> find_banned(Clause).

% find_banned/1 is plain Prolog, but only calls Tier 1 predicates
find_banned((_ :- Body)) :- !, find_banned(Body).
find_banned((A, B)) :- !, find_banned(A), find_banned(B).
find_banned(Goal) :-
    functor(Goal, Name, _),
    banned_operation(Name)   % Tier 1 call
    -> banned_predicate(Name)
    ;  true.

% Simplification rule: handle the constraint
banned_predicate(Name) <=>
    format(atom(Msg), 'Banned operation in proposed clause: ~w', [Name]),
    throw(safety_violation(Msg)).
```

---

## Tier 3 — The Meta-Introspection Layer

### What goes here

Code that reasons *about* Tier 1 and Tier 2 rules:
- Reading current clause definitions (`introspect.pl`)
- Proposing candidate improvements (`improve.pl`)
- Verifying invariants before assertion (`verify.pl`)
- Performance measurement and analysis

Tier 3 code is full SWI-Prolog — unrestricted. The safety constraint is structural:
**Tier 3 cannot modify Tier 1 predicates.** This is enforced by `verify:check_invariants/1`
(wired into `safe_assert/1`) and the CHR check that rejects any proposed clause whose
head matches a Tier 1 predicate.

### Variable capture and meta-reasoning

When constructing new logic fragments at runtime, variable capture is the key hazard.
A naive term construction can bind variables that were intended to remain free.

**The standard solution (preferred, zero new dependencies):**

```prolog
% Use copy_term/2 to get fresh variables before manipulation
propose_improvement(OldClause, NewClause) :-
    copy_term(OldClause, (Head :- Body)),
    % Now Head and Body have fresh variables — safe to manipulate
    NewClause = (Head :- (Body, true)).  % add a no-op guard as example

% Use numbervars/3 for ground comparison
clauses_equivalent(C1, C2) :-
    copy_term(C1, C1c), numbervars(C1c, 0, _),
    copy_term(C2, C2c), numbervars(C2c, 0, _),
    C1c == C2c.
```

**When to use ELPI (higher-order unification, if genuinely needed):**

ELPI (Embeddable Lambda Prolog Interpreter) provides lexical scoping and higher-order
unification. It is the right tool when:
- The meta-improvement loop generates clauses with complex binder structures that
  `copy_term/2` cannot handle correctly
- The introspection task requires reasoning about the *binding structure* of terms,
  not just their shape

ELPI should not be adopted until:
1. A specific task is identified that `copy_term/2` + `numbervars/3` cannot solve
2. A benchmark on M5 hardware confirms the SWI→ELPI context-switch cost is acceptable
   for the call frequency in question

**`library(yall)` — what it is and is not:**

YALL provides anonymous predicate syntax (`\X^Goal` for `call(Goal, X)`). It is syntactic
sugar for higher-order *calling*, not higher-order *unification*. Use YALL for concise
predicate-passing. Do not use it as a substitute for λProlog semantics — it does not
provide lexical scoping or higher-order unification.

```prolog
% YALL appropriate use — anonymous predicate passing
maplist(\X^(X > 0), [1, 2, 3])  % checks all elements > 0

% YALL inappropriate use — not higher-order unification
% For this, use copy_term or ELPI
```

---

## Tier Enforcement: The Three-Layer Stack

Tier boundaries are not conventions — they are mechanically enforced.

### Layer 1: Syntactic linter (pre-submit time)

`core-station/bridge/validate-stasis-tier1.pl` runs during `core-station/bridge/pre-submit.sh`. It rejects
any predicate tagged `stasis_tier(1, ...)` that violates Datalog syntax restrictions.

### Layer 2: Safety Rail Z3 policy (proposal time)

When any code change is proposed via `safe_assert/1` or through the Forge/Crucible
briefing pipeline, the Safety Rail's Z3 solver checks the `stasis_tier2_calls_tier1_only`
constraint: CHR rule bodies may only call Tier 1 predicates.

### Layer 3: `safe_assert` invariant check (runtime)

`verify:check_invariants/1` (called from `safe_assert/1`) rejects any proposed clause
whose head matches a Tier 1 predicate. Tier 1 invariants are immutable at runtime.

```
Proposed clause
       │
       ▼
constraints:check_constraints/1  ← Tier 2 gate (banned predicates in body)
       │
       ▼
verify:check_invariants/1        ← Tier 3 gate (Tier 1 head redefinition check)
       │
       ▼
merkle_bridge:merkle_commit/3    ← Audit (Merkle leaf)
       │
       ▼
assertz/1                        ← KB mutation (authorized only here)
```

---

## Integration with the Safety Rail

The Safety Rail (Rust, `safety-rail/`) and STASIS (SWI-Prolog) operate at different
abstraction levels but share a common interface through `mcp_bridge.pl`.

```
STASIS Tier 3 (improve.pl)
    │ safe_assert(NewClause)
    ▼
STASIS Tier 2 (constraints.chr)  ← local CHR check
    │ check passed
    ▼
mcp_bridge:mcp_call('submit_skill_proposal', ...)  ← calls Root Spine MCP host
    │
    ▼
Root Spine (Go, :8082)
    │
    ▼
Safety Rail Tier 1 (Rust/Z3)  ← formal verification
    │
    ▼
SafetyVerdict::Safe / ViolationReport
```

The Safety Rail provides the formal guarantee; the STASIS CHR layer provides a fast
local pre-check that catches common violations before the network round-trip to :8082.
In test mode (`current_prolog_flag(test_mode, true)`), the network call is bypassed
and the CHR check is the operative gate.

---

## Integration with the Merkle Audit Log

Every knowledge-base mutation through `safe_assert/1` generates a Merkle leaf:

```prolog
% From core/merkle_bridge.pl
merkle_commit(EventType, Clause, Proof) :-
    term_to_atom(Clause, ClauseAtom),
    % ... calls MCP tool 'submit_skill_proposal' on Root Spine
    % Root Spine writes RFC 6962 leaf: SHA-256(0x00 || canonical_json)
    Proof = proof{leaf_hash_hex: Hash, root_hash_hex: Root, tree_size: Size}.
```

On restart, the persistence layer replays committed clauses in Merkle order, reconstructing
the knowledge base from the audit log. The Protoplasm's state at any point in time is
provably derivable from the Merkle log — no separate state snapshot is required.

---

## STASIS and the HOOTL / Dark Factory Trajectory

The PROCESS.md operational paradigm note describes the trajectory:

> HOTL (Human On The Loop) → HOOTL (Human Out Of The Loop) → Dark Factory

This transition is only safe when three conditions are met:

| Condition | STASIS mechanism |
|---|---|
| Safety checks are guaranteed to terminate | Tier 1 Datalog decidability |
| Safety checks are formally verifiable | Tier 1 + Safety Rail Z3 proofs |
| Self-evolution cannot modify safety rules | Tier boundary enforcement (Tier 3 cannot modify Tier 1) |

**STASIS Tier 1 is a prerequisite for HOOTL, not an optimisation.** A Dark Factory
running on unrestricted Prolog has no formal guarantee that a self-proposed clause
cannot introduce a non-terminating safety check. STASIS provides that guarantee.

The migration path: as the assembly matures, each safety heuristic currently encoded
in Tier 2 (CHR) or Tier 3 (verify.pl) that can be proven invariant across all inputs
should be migrated to Tier 1. The Tier 1 core grows over time; the HOOTL window expands.

---

## File Organisation

```
agents/prolog-substrate/          ← STASIS substrate root (directory name legacy)
├── core/
│   ├── safety_bridge.pl          ← Cross-tier infrastructure (safe_assert/safe_retract)
│   ├── merkle_bridge.pl          ← Audit log integration
│   ├── otel_bridge.pl            ← OTel metrics emission
│   └── mcp_bridge.pl             ← Root Spine MCP tool calls
├── policies/
│   ├── constraints.chr           ← Tier 2: CHR constraint store
│   └── invariants.pl             ← Tier 1 candidates (to be migrated with linter)
├── meta/
│   ├── introspect.pl             ← Tier 3: clause inspection
│   ├── improve.pl                ← Tier 3: meta-improvement loop
│   ├── verify.pl                 ← Tier 3: invariant checking
│   └── fitness.pl                ← Tier 3: substrate fitness scoring
├── skills/
│   ├── base.pl                   ← Seed skills (immutable, never retracted)
│   └── runtime/                  ← Runtime-asserted clauses (loaded from Merkle log)
└── tests/
    ├── golden/                   ← Ground-truth test sets for evaluate_candidate/3
    ├── test_safe_assert.pl       ← Tier 2+infra: mutation gate tests
    ├── test_introspect.pl        ← Tier 3: introspection tests
    ├── test_meta.pl              ← Tier 3: improvement loop tests
    ├── test_regression.pl        ← Regression parity tests
    └── test_fitness.pl           ← Fitness scorer tests
```

*Note: `agents/prolog-substrate/` retains its directory name for filesystem compatibility.
All prose documentation uses "STASIS substrate." A directory rename to `agents/stasis-substrate/`
is deferred as a future cleanup task — it requires updating all shell script path references.*

---

## Station Relationships (decided 2026-04-22)

### Where STASIS Lives

STASIS is implemented inside `schwarzschild-assembly/core-station/protoplasm/`. It does
**not** live in `shapeshifter/`, despite an earlier reference in the schwarzschild-assembly
README describing it as "provided by shapeshifter." That was aspirational design that was
never implemented. STASIS belongs to this project.

### Relationship with Shapeshifter

Shapeshifter (`~/antigravity/shapeshifter/`) is a **runtime skill-parameterization DSL**
embedded inside individual agents. It uses Python S-expressions to let an agent tune its
own skill parameters and propose body mutations at runtime.

STASIS is a **station-wide policy enforcement substrate**. It decides whether proposed
mutations are allowed.

The intended integration (Shapeshifter Phase 4):

```
Agent runtime
  └─ Shapeshifter evaluator
       └─ proposes mutation (quote/eval, gas-limited)
            └─ submits to STASIS safe_assert pipeline
                 ├─ Tier 2 CHR scan
                 ├─ Tier 1 invariant check
                 ├─ Merkle commit
                 └─ → pending → Roy approves via control plane
```

This integration is not yet built. Until it is, Shapeshifter and STASIS operate
independently. Shapeshifter proposes in Python; STASIS enforces in Prolog.

### Planned Extraction to `stasis-language/`

When Shapeshifter Phase 4 is ready to consume STASIS, this project will extract the
general runtime (safe_assert pipeline, Merkle bridge, OTel bridge, tier runtime, linter,
generic meta/ modules) to `~/antigravity/stasis-language/` as a standalone station
substrate. schwarzschild-assembly will then depend on it as a submodule.

The **extraction seam** divides the code into two parts:

| General runtime → `stasis-language/` | Assembly-specific → stays here |
|---|---|
| `core/safety_bridge.pl` | Spacecraft CHR policy rules |
| `core/merkle_bridge.pl` | Spacecraft Tier 1 invariant overrides |
| `core/otel_bridge.pl` | `core/safety_bridge.pl` integration glue |
| `meta/improve.pl`, `introspect.pl`, `verify.pl`, `fitness.pl` | Z3 policy stubs for spacecraft domain |
| Tier 1 linter (`validate-stasis-tier1.pl`) | |
| Generic `policies/invariants.pl` defaults | |

**Do not extract until the trigger condition is met:** Shapeshifter Phase 4 mutation
gating is actively ready to integrate. Premature extraction adds migration cost before
there is a second consumer to justify it. See `dsl-008` in root `TASKS.md`.

---

## Glossary

| Term | Definition |
|---|---|
| **STASIS** | The tiered symbolic logic language for Schwarzschild Spacecraft Protoplasms |
| **Protoplasm** | The reasoning/intelligence layer of a Spacecraft; implemented in STASIS |
| **Tier 1** | Decidable, Datalog-restricted safety invariants |
| **Tier 2** | Dynamic CHR constraint store for evolving policy |
| **Tier 3** | Meta-introspection code in unrestricted SWI-Prolog |
| **Tabling** | SWI-Prolog mechanism preventing looping in recursive predicates |
| **Datalog** | A logic language subset with decidable semantics (no functor symbols in heads, stratified negation) |
| **CHR** | Constraint Handling Rules — a multiset rewriting formalism for constraint stores |
| **ELPI** | Embeddable Lambda Prolog Interpreter — provides higher-order unification |
| **YALL** | SWI-Prolog's anonymous predicate library — NOT λProlog, does not provide higher-order unification |
| **safe_assert** | The only authorized KB mutation point; routes through CHR, invariant check, Safety Rail, and Merkle log |
| **Dark Factory** | HOOTL autonomous operation state; requires STASIS Tier 1 as prerequisite |
