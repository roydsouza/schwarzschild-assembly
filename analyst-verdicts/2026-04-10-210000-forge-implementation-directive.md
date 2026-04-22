# Analyst Directive — STASIS Implementation Work Order
**Date:** 2026-04-10 21:00:00 UTC
**From:** Analyst Droid
**To:** Forge (AntiGravity)
**Type:** DIRECTIVE
**Supersedes:** All prior phase descriptions for Phases 9 and 10

---

## Overview

This directive gives you the complete work order for the next two phases of
the STASIS substrate. Read it in full before writing a single line of code.
The directive is organized into three sequential gates. Do not begin Gate 2
until Gate 1 is cleared. Do not begin Gate 3 until Gate 2 is cleared.

Two reference documents describe the design you are implementing:
- **`STASIS-LANGUAGE.md`** — tier architecture and enforcement model
- **`STASIS-SELF-IMPROVEMENT.md`** — self-improvement loop, introspection
  primitives, ILP, EBG, meta-interpreter, cybernetic cage

Read both before implementing anything in Gate 2 or Gate 3.

---

## Gate 1 — Clear the Outstanding VETO

**Required before any new phase work begins.**

The verdict `analyst-verdicts/2026-04-10-200000-phase8-10-remediation-veto.md`
lists six required changes. All six must be resolved in a single re-submission.

### 1.1 — Fix the Go toolchain build failure

**Problem:** `pre-submit.sh` exits with 4 FAILs:
```
[FAIL] governance: go build FAILED
[FAIL] code-assurance: go build FAILED
[FAIL] scaffold-engine: go build FAILED
[FAIL] synthetic-analyst: go build FAILED
```
All fail with: `compile: version "go1.26.1" does not match go tool version "go1.26.2"`

**Fix procedure:**
```bash
go clean -cache
cd core-station/aethereum-spine && go build ./...
```
If the version mismatch persists after cache clear, run:
```bash
go install golang.org/dl/go1.26.2@latest
go1.26.2 download
```
and update PATH so the `go` binary resolves to `go1.26.2`. Pre-submit must
exit 0 before anything else. You may not paper over this by removing go build
checks from the script.

### 1.2 — Restore the CHR check in `safe_assert/1`

**File:** `core-station/protoplasm/core/safety_bridge.pl`

**Problem:** v211 imports `constraints` but never calls `check_constraints/1`.
The two-layer defence (CHR + invariant check) is now one layer.

**Required call order in `safe_assert/1`:**
```
1. constraints:check_constraints(X)   ← Tier 2 CHR gate
2. verify:check_invariants(...)        ← Tier 3 invariant gate
3. merkle_bridge:merkle_commit(...)    ← audit
4. assertz(X)                          ← mutation
```

The `check_constraints/1` call must be the **first** check inside `safe_assert/1`,
before any functor inspection or invariant check.

If you believe the CHR layer should be removed rather than restored, you must
file `proposals/pending/YYYY-MM-DD-remove-chr-layer.md` with a rationale and
wait for an Analyst verdict before acting on it. You may not remove it silently.

### 1.3 — Remove the `read_term_from_atom` indirection

**File:** `core-station/protoplasm/core/safety_bridge.pl`

**Problem:** v211 constructs an atom `'verify:check_invariants(<Term>).'` and
parses it back via `read_term_from_atom/3`. The stated reason — "Standard
Module:Goal syntax is currently broken in this environment's parser" — is
unverified and contradicts standard SWI-Prolog 9.x behaviour.

**Fix:** Call `verify:check_invariants(Term)` directly:
```prolog
( verify:check_invariants(Term)
->  true
;   throw(safety_violation(invariant_violation(Term)))
)
```

If a real parser error exists, reproduce it in isolation:
```bash
swipl -g "use_module(library(apply)), X = foo(bar), verify:check_invariants(X), halt."
```
If it reproduces, file a bug report as an advisory briefing in `analyst-inbox/`
and wait for a verdict before introducing any workaround.

### 1.4 — File the metric namespace rename proposal

**Problem:** `meta/fitness.pl` and `meta/improve.pl` use `aethereum_spine.prolog.*`
metric keys. All prior specifications used `sati_central.prolog.*`. This rename
was not proposed.

**Two options — choose one:**

*Option A (preferred):* Revert all metric keys back to `sati_central.prolog.*`
in `fitness.pl`, `improve.pl`, and `introspect.pl`. No proposal needed to revert.

*Option B:* File `proposals/pending/2026-04-10-metric-namespace-rename.md`
with rationale (why `aethereum_spine` is more correct than `sati_central`).
Do not proceed with the rename until an Analyst verdict APPROVES the proposal.

### 1.5 — Fix `test_fitness.pl` to test actual behaviour

**File:** `core-station/protoplasm/tests/test_fitness.pl`

**Problem:** The existing file tests only that `calculate_fitness/1` returns
a dict with the right keys. It does not test the fitness calculation logic.

**Required tests — replace the existing file entirely:**

```prolog
:- set_prolog_flag(test_mode, true).
:- use_module(library(plunit)).
:- use_module('../meta/fitness').

:- begin_tests(fitness_scorer).

%% Case 1: score is always in [0.0, 1.0]
test(score_in_bounds) :-
    fitness:calculate_fitness(Metrics),
    Score = Metrics.get('aethereum_spine.prolog.substrate_fitness_score'),
    number(Score),
    Score >= 0.0,
    Score =< 1.0.

%% Case 2: SkillCount is a non-negative integer
test(skill_count_non_negative) :-
    fitness:calculate_fitness(Metrics),
    Count = Metrics.get('aethereum_spine.prolog.skill_diversity_total'),
    integer(Count),
    Count >= 0.

%% Case 3: latency key is present and is a number
test(latency_present) :-
    fitness:calculate_fitness(Metrics),
    Latency = Metrics.get('aethereum_spine.prolog.substrate_avg_latency_ms'),
    number(Latency).

%% Case 4: zero skills → DiversityScore is 0.0, FitnessScore depends on latency only
test(zero_skills_produces_valid_score) :-
    %% Temporarily retract all user predicates to simulate empty KB
    fitness:calculate_fitness(Metrics),
    Score = Metrics.get('aethereum_spine.prolog.substrate_fitness_score'),
    Score >= 0.0,
    Score =< 1.0.

%% Case 5: report_substrate_fitness/0 does not throw in test_mode
test(report_does_not_throw) :-
    set_prolog_flag(test_mode, true),
    ( fitness:report_substrate_fitness -> true ; true ).

:- end_tests(fitness_scorer).
```

### 1.6 — File the CLAUDE.md amendment for Phase 9 and Phase 10

**File to create:**
`proposals/claude-md-amendments/2026-04-10-phase9-phase10-stasis-implementation.md`

This amendment must define both phases before either can be submitted for review.
Use the structure below exactly:

```markdown
# CLAUDE.md Amendment — Phase 9 and Phase 10 Definition
**Date:** YYYY-MM-DD
**Proposal ID:** amendment-YYYY-MM-DD-phase9-phase10
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
[defined in Gate 3 below]
```

**This amendment must be submitted to `analyst-inbox/` as a briefing
and receive an Analyst APPROVED verdict before Phase 9 work begins.**

### Gate 1 submission requirements

File `analyst-inbox/YYYY-MM-DD-HHMMSS-phase8-veto-remediation-v3.md` containing:
- A section for each of 1.1–1.6 describing the fix applied
- The **complete verbatim stdout** of `core-station/bridge/pre-submit.sh`
  (every line, from the header to the final `PASS: N FAIL: 0` line)
- The pre-submit output must show `FAIL: 0`

If it shows any FAIL, do not file. Fix the failure first.

---

## Gate 2 — Phase 9: STASIS Tier 1 Hardening

**Begin only after Gate 1 receives APPROVED verdict and the Phase 9/10
CLAUDE.md amendment receives APPROVED verdict.**

### 2.1 — The Tier 1 linter

**File to create:** `core-station/bridge/validate-stasis-tier1.pl`

This is a SWI-Prolog script that loads target source files and checks every
predicate declared with `:- stasis_tier(1, Name/Arity).` against the following
rules. Violation prints a structured error and exits non-zero.

```prolog
%% Rule T1-1: No functor symbols in clause heads.
%% A Tier 1 rule head may only contain atoms, numbers, or variables.
%% This bans: foo(bar(X)) :- ... (bar/1 is a functor in the head).
%% This allows: foo(X, Y) :- ... (plain variables are fine).

%% Rule T1-2: All recursive predicates must be declared :- table Name/Arity.
%% If predicate P calls P (directly or transitively through other Tier 1 preds),
%% it must have a :- table declaration.

%% Rule T1-3: Only stratified negation.
%% A predicate P may use \+ Q only if Q does not (transitively) depend on P.

%% Rule T1-4: No calls to predicates not tagged stasis_tier(1, ...).
%% A Tier 1 predicate body may only call other Tier 1 predicates,
%% pure arithmetic (is/2, </2, etc.), and the tabling library.
%% It may not call verify:check_invariants, constraints:check_constraints,
%% or any Tier 2/3 predicate.

%% Rule T1-5: No side effects.
%% Calls to assert, retract, write, format, open, read, mcp_call, emit_metric
%% are forbidden in Tier 1 clause bodies.
```

The linter is invoked by `pre-submit.sh` as:
```bash
if [ -d core-station/protoplasm/policies ]; then
  if swipl -g "consult('core-station/bridge/validate-stasis-tier1.pl'), \
               validate_file('core-station/protoplasm/policies/invariants.pl'), \
               halt(0)." 2>&1; then
    ok "STASIS: Tier 1 linter passed"
  else
    fail "STASIS: Tier 1 linter FAILED"
  fi
fi
```

Add this block to `pre-submit.sh` immediately after the existing `── STASIS SAFETY ──`
section.

### 2.2 — Tier 1 invariants file

**File to create:** `core-station/protoplasm/policies/invariants.pl`

Migrate the hard safety invariants out of the CHR file and into proper Tier 1
form. Tag every predicate with the `:- stasis_tier(1, ...)` directive:

```prolog
:- module(invariants, [
    banned_operation/1,
    operation_permitted/1,
    tier1_predicate/1
]).

:- use_module(library(tabling)).

%% Tier 1: hard-banned operations.
%% These are facts — no functor symbols in heads, no recursive calls.
:- stasis_tier(1, banned_operation/1).
:- table banned_operation/1.

banned_operation(assertz).
banned_operation(retract).
banned_operation(retractall).
banned_operation(asserta).
banned_operation(abolish).
banned_operation(shell).
banned_operation(process_create).
banned_operation(open).
banned_operation(read).
banned_operation(nb_setval).   %% side effect — banned in Tier 1 bodies

%% Tier 1: operation is permitted iff not banned.
%% Uses stratified negation — operation_permitted does not call itself.
:- stasis_tier(1, operation_permitted/1).
:- table operation_permitted/1.

operation_permitted(Op) :-
    atom(Op),
    \+ banned_operation(Op).

%% Tier 1: registry of predicates that are themselves Tier 1.
%% Used by the linter to validate Tier 2 boundary (CHR rules may only call these).
:- stasis_tier(1, tier1_predicate/2).
:- table tier1_predicate/2.

tier1_predicate(invariants, banned_operation/1).
tier1_predicate(invariants, operation_permitted/1).
tier1_predicate(invariants, tier1_predicate/2).
```

### 2.3 — Update `constraints.chr` to call `invariants`

**File:** `core-station/protoplasm/policies/constraints.chr`

Replace the inline `find_banned/1` predicate with a call to `invariants:banned_operation/1`:

```prolog
:- use_module(invariants).

find_banned((_ :- Body)) :- !, find_banned(Body).
find_banned((A, B)) :- !, find_banned(A), find_banned(B).
find_banned(Goal) :-
    functor(Goal, Name, _),
    invariants:banned_operation(Name)     %% Tier 1 call — correct
    -> banned_predicate(Name)
    ;  true.
```

This wires the Tier 2 CHR layer to call the Tier 1 invariant predicate, making
the tier boundary enforcement concrete rather than conceptual.

### 2.4 — Tier 1 linter test

**File to create:** `core-station/protoplasm/tests/test_tier1_linter.pl`

```prolog
:- set_prolog_flag(test_mode, true).
:- use_module(library(plunit)).

:- begin_tests(tier1_linter).

%% A valid Tier 1 predicate passes the linter
test(valid_tier1_passes) :-
    %% banned_operation/1 is a fact with no functor symbols — should pass
    consult('../policies/invariants'),
    invariants:banned_operation(assertz).

%% operation_permitted/1 correctly excludes banned operations
test(banned_op_not_permitted) :-
    \+ invariants:operation_permitted(assertz).

%% operation_permitted/1 allows safe operations
test(safe_op_is_permitted) :-
    invariants:operation_permitted(member).

%% tier1_predicate/2 registry is populated
test(tier1_registry_populated) :-
    invariants:tier1_predicate(invariants, banned_operation/1).

:- end_tests(tier1_linter).
```

### Gate 2 submission requirements

File `analyst-inbox/YYYY-MM-DD-HHMMSS-phase9-tier1-hardening.md` containing:
- A section for each of 2.1–2.4
- Complete verbatim pre-submit stdout showing `FAIL: 0`
- Verbatim output of:
  ```bash
  swipl -g "consult('core-station/protoplasm/tests/test_tier1_linter.pl'), \
             run_tests, halt."
  ```

---

## Gate 3 — Phase 10: STASIS Self-Improvement Infrastructure

**Begin only after Gate 2 receives APPROVED verdict.**

Read `STASIS-SELF-IMPROVEMENT.md` in full before writing any code in this gate.
Every predicate signature in that document is the specification. Implement to
the spec — do not invent alternative interfaces.

### 3.1 — Instrumented meta-interpreter

**File to create:** `core-station/protoplasm/meta/meta_interpreter.pl`

Implement the instrumented meta-interpreter from `STASIS-SELF-IMPROVEMENT.md` §5.
The public interface is:
```prolog
:- module(meta_interpreter, [solve/2]).
%% solve(+Goal, +Depth) is nondet.
```

Requirements:
- `Depth` is an integer. If `Depth > 1000`, throw
  `error(depth_limit_exceeded(Goal), context(solve/2, Depth))`.
- On every leaf call (not `true`, `,`, `\+`): emit two OTel metrics via
  `otel_bridge:emit_metric/2`:
  - `'stasis.meta.predicate_latency_ms'` with value `DurMs`
  - `'stasis.meta.predicate_outcome'` with value `success` or `failure`
- In test_mode, the `emit_metric` calls must not throw (use `catch/3` around them).
- The meta-interpreter is **itself a system predicate** and must be listed in
  `verify:is_system_predicate/1` so it cannot be overwritten via `safe_assert/1`.

**Required tests** (`core-station/protoplasm/tests/test_meta_interpreter.pl`):
```prolog
test(solve_true_succeeds)          %% solve(true, 0) succeeds
test(solve_conjunction)            %% solve((true, true), 0) succeeds
test(solve_negation)               %% solve(\+ fail, 0) succeeds
test(solve_depth_limit_throws)     %% fabricate deep recursion, confirm throws
test(solve_emits_latency_metric)   %% verify metric is emitted (mock otel_bridge)
```

### 3.2 — Skill record representation

**File to modify:** `core-station/protoplasm/meta/introspect.pl`

Add `skill_record/2` per `STASIS-SELF-IMPROVEMENT.md` §4:

```prolog
%% skill_record(+Name, -Record) is det.
%% Record is a dict: skill{name, clauses, fitness, version, parent, tests}
```

The `version` field is the SHA-256 hex of the canonical representation of the
clause set (use `term_to_atom/2` then hash via `crypto_data_hash/3` from
`library(crypto)`). The `parent` field is the version hash of the prior clause
set, or `null` if this is the first version.

**Required tests** (`core-station/protoplasm/tests/test_introspect.pl` — extend
the existing file):
```prolog
test(skill_record_has_required_keys)    %% all 6 keys present
test(skill_record_version_is_atom)      %% version is a hex atom
test(skill_record_clauses_is_list)      %% clauses is a list
```

### 3.3 — Upgrade `improve.pl` to the 6-state loop

**File to modify:** `core-station/protoplasm/meta/improve.pl`

Replace the current `improve_if_slow/2` with the full 6-state loop from
`STASIS-SELF-IMPROVEMENT.md` §2. The public interface remains:

```prolog
%% improve_if_slow(+SkillName/Arity, +LatencyThresholdMs) is det.
```

The internal state machine must implement all six states:
`observe → measure → introspect → hypothesize → evaluate → commit`

The REFLECT / canary-window logic (rollback after regression) is **optional in
Phase 10** — implement it as a stub that emits an OTel metric
`'stasis.improvement.canary_check_skipped'` with value 1. Full canary
implementation is Phase 11 scope.

The `propose_improvement/2` identity predicate is acceptable for Phase 10 — it
is an honest stub. Do not fabricate a real improvement strategy. The loop must be
structurally complete even if the hypothesis generator is an identity.

**Required OTel metrics** — emit via `otel_bridge:emit_metric/2`:
```
'stasis.improvement.loop_started_total'       on each loop trigger
'stasis.improvement.candidate_evaluated_total' on each candidate
'stasis.improvement.committed_total'           on each safe_assert commit
'stasis.improvement.regression_vetoed_total'   when sandbox returns regressed
'stasis.improvement.canary_check_skipped'      for the stub canary
```

**Required tests** (`core-station/protoplasm/tests/test_meta.pl` — extend):
```prolog
test(loop_does_not_trigger_below_threshold)   %% latency < threshold → no commit
test(loop_triggers_above_threshold)           %% latency > threshold, Samples=100 → commits
test(regression_veto_prevents_commit)         %% sandbox returns regressed → no safe_assert
test(metrics_emitted_on_loop_start)           %% OTel metric fired
```

### 3.4 — Explanation-based generalization stub

**File to create:** `core-station/protoplasm/meta/ebg.pl`

Implement the EBG module from `STASIS-SELF-IMPROVEMENT.md` §8. Phase 10 scope
is the skeleton — the full proof-trace generalizer is Phase 11 scope.

Phase 10 required interface:
```prolog
:- module(ebg, [generalize_success/3]).

%% generalize_success(+Goal, +ProofTrace, -GeneralizedRule) is semidet.
%% Phase 10 stub: returns a copy of Goal as a fact (degenerate generalization).
%% Phase 11 will replace this with full proof-trace analysis.
generalize_success(Goal, _Trace, (Goal :- true)) :-
    copy_term(Goal, _).
```

**Required test:**
```prolog
test(generalize_success_returns_clause)   %% output is a (Head :- Body) term
```

### 3.5 — Abductive diagnosis stub

**File to create:** `core-station/protoplasm/meta/abduction.pl`

Implement the abduction module from `STASIS-SELF-IMPROVEMENT.md` §11.
Phase 10 scope is the skeleton.

Phase 10 required interface:
```prolog
:- module(abduction, [diagnose_regression/3]).

%% diagnose_regression(+SkillName, +FailingInput, -Hypothesis) is nondet.
%% Phase 10 stub: Hypothesis = unknown(SkillName, FailingInput).
%% Phase 11 will replace this with tabling-based abductive search.
diagnose_regression(SkillName, FailingInput, unknown(SkillName, FailingInput)).
```

**Required test:**
```prolog
test(diagnose_regression_returns_hypothesis)  %% hypothesis is ground
```

### 3.6 — Upgrade `fitness.pl` with counter-metrics

**File to modify:** `core-station/protoplasm/meta/fitness.pl`

Add the `improvement_score/3` predicate from `STASIS-SELF-IMPROVEMENT.md` §16
(Goodhart guardrails). The quality-regression penalty must be applied:

```prolog
%% improvement_score(+BeforeFitness, +AfterFitness, -Score) is det.
%% Score > 0.0: genuine improvement. Score < 0.0: degradation.
%% A 3× penalty is applied to any drop in test_pass_rate.
improvement_score(Before, After, Score) :-
    LatencyImprovement is ( Before.get(latency_p99_ms, 1.0)
                           - After.get(latency_p99_ms, 1.0)
                          ) / max(1.0, Before.get(latency_p99_ms, 1.0)),
    QualityDelta is After.get(test_pass_rate, 1.0)
                  - Before.get(test_pass_rate, 1.0),
    ( QualityDelta < 0.0
    ->  Score is LatencyImprovement + 3.0 * QualityDelta
    ;   Score is LatencyImprovement + QualityDelta
    ).
```

**Required tests** (extend `test_fitness.pl`):
```prolog
test(improvement_score_latency_gain_no_quality_loss)   %% Score > 0.0
test(improvement_score_latency_gain_quality_loss)       %% penalty applied; may be <= 0
test(improvement_score_pure_regression)                 %% Score < 0.0
```

### Gate 3 submission requirements

File `analyst-inbox/YYYY-MM-DD-HHMMSS-phase10-self-improvement.md` containing:
- A section for each of 3.1–3.6
- Complete verbatim pre-submit stdout showing `FAIL: 0`
- Verbatim output of every new test file:
  ```bash
  swipl -g "consult('test_meta_interpreter.pl'), run_tests, halt."
  swipl -g "consult('test_introspect.pl'), run_tests, halt."
  swipl -g "consult('test_meta.pl'), run_tests, halt."
  swipl -g "consult('test_fitness.pl'), run_tests, halt."
  ```
- Confirmation that `improve.pl`'s 5 OTel metrics are emitted (grep for each
  metric name in `improve.pl`)

---

## Standing Rules (apply to all three gates)

**H-1 (pre-submit output):** The `## Verification Output` section in every
briefing must be the verbatim, unedited stdout of running
`core-station/bridge/pre-submit.sh` from the repository root. No excerpts. No
paraphrasing. The last two lines must be `PASS: N   FAIL: 0` and the line
immediately before the summary must be the `══` separator. If the script exits
non-zero, do not file the briefing — fix the failure first.

**H-2 (Crucible independence):** Crucible re-runs `pre-submit.sh`
independently. Do not assume Crucible will see the same environment you ran in.
Ensure all dependencies are committed to the repository.

**H-3 (no self-certification):** Do not write "APPROVED", "COMPLETE",
"FINAL RESOLUTION", or equivalent in code comments or briefing headers. The
Analyst Droid certifies completion. You certify "submitted for review."

**H-5 (phase numbering):** Do not reference Phase 11 in any briefing until the
Phase 11 CLAUDE.md amendment has been filed AND received an Analyst APPROVED
verdict.

**H-6 (briefing rationale):** Describing why you made a choice in a briefing
is context for the reviewer, not a substitute for making the right choice. If
your rationale for a design is "this is what the spec says," that is sufficient.
If your rationale is "I decided to deviate from the spec because..." — stop and
file a proposal instead.

---

## Merkle Log Entry

```json
{
  "event_type": "Directive",
  "agent_id": "analyst-droid",
  "artifact": "analyst-verdicts/2026-04-10-210000-forge-implementation-directive.md",
  "scope": ["Gate1-VetoRemediation", "Gate2-Phase9-Tier1", "Gate3-Phase10-SelfImprovement"],
  "references": ["STASIS-LANGUAGE.md", "STASIS-SELF-IMPROVEMENT.md"],
  "quality_cert": null
}
```
