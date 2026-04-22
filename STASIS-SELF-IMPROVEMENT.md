# STASIS Self-Improvement: Agentic Reasoning and Skill Evolution

*Companion document to [STASIS-LANGUAGE.md](./STASIS-LANGUAGE.md)*

**Version:** 1.0 — 2026-04-10
**Runtime:** SWI-Prolog 9.x (aarch64-apple-darwin, M5-native)

---

## The Core Thesis

A STASIS agent is a **homoiconic reasoner**. Its skills, its beliefs about those
skills, its effectiveness measurements, and its improvement proposals are all
represented as Prolog terms inside a single knowledge base. The tier architecture
(described in `STASIS-LANGUAGE.md`) defines what the agent is *allowed* to do with
that knowledge base. This document describes how it actually *learns*.

The central insight is that introspection in Prolog is not a special feature — it is
a direct consequence of homoiconicity. `clause(my_skill(X), Body)` retrieves the
agent's own source code as a first-class data structure. The agent can read it,
transform it, evaluate the transformation in a sandbox, and commit the improvement
via `safe_assert/1` — all without leaving the logic layer.

This is what distinguishes STASIS from a scripted rule engine: the improvement loop
is written in the same language as the skills it improves.

---

## The Self-Improvement Loop

The loop is a deterministic state machine. Each state transition is a logged event
in the Merkle log. A loop iteration that does not commit a change is still auditable —
the MEASURE and EVALUATE states emit OTel metrics even when no improvement is found.

```
  ┌──────────────────────────────────────────────────────────────────┐
  │                                                                  │
  ▼                                                                  │
OBSERVE ──→ MEASURE ──→ INTROSPECT ──→ HYPOTHESIZE ──→ EVALUATE ──→ COMMIT
                                                          │
                                               (regressed or unsafe)
                                                          │
                                                        DISCARD
```

After COMMIT, the canary window (default 24 hours) runs. If fitness regresses during
the canary window, the loop enters:

```
COMMIT ──→ (canary window) ──→ REFLECT ──→ ROLLBACK
                                         (or CONFIRM if fitness holds)
```

### OBSERVE

The observation phase is continuous — it does not block the agent. The instrumented
meta-interpreter (§5) emits OTel metrics on every goal call. Between improvement
cycles, the agent accumulates a latency histogram, a success/failure ratio, and a
cost counter for each of its predicates.

**Trigger condition:** A skill is nominated for improvement when any of its fitness
dimensions crosses a threshold over the last N calls:
- Latency p99 > threshold (configurable per skill)
- Success rate < threshold
- Cost per call > threshold
- No nomination if fewer than 100 samples (insufficient data)

```prolog
%% nominate_for_improvement(+SkillName) is semidet.
%% Succeeds if the skill's fitness crosses a trigger threshold.
nominate_for_improvement(SkillName) :-
    skill_fitness(SkillName, Fitness),
    (   Fitness.latency_p99_ms > Fitness.latency_threshold
    ;   Fitness.success_rate   < Fitness.success_threshold
    ),
    Fitness.sample_count >= 100.
```

### MEASURE

Take a snapshot of the fitness vector before any change. This is the baseline the
REFLECT phase compares against.

```prolog
%% measure_fitness_snapshot(+SkillName, -Snapshot) is det.
measure_fitness_snapshot(SkillName, Snapshot) :-
    skill_fitness(SkillName, CurrentFitness),
    get_time(TS),
    Snapshot = snapshot{skill: SkillName, ts: TS, fitness: CurrentFitness}.
```

### INTROSPECT

Read the skill's current clause set using SWI-Prolog's meta-predicates:

```prolog
%% introspect_skill(+SkillName, -ClauseSet) is det.
introspect_skill(SkillName, ClauseSet) :-
    functor(Head, SkillName, _),
    findall((Head :- Body), clause(Head, Body), ClauseSet).
```

The clause set is a plain Prolog list of `(Head :- Body)` terms. Every item is
inspectable, transformable, and comparable using standard unification.

### HYPOTHESIZE

Generate one or more candidate clause sets using one of the three modification modes
(§7 below). Each candidate is annotated with a hypothesis tag:

```prolog
Candidate = candidate{
    clauses:    NewClauses,
    mode:       tuning,          % tuning | specialization | synthesis
    rationale:  "Reduced threshold from 500ms to 350ms based on p95 latency."
}.
```

### EVALUATE

Run each candidate in a pengines sandbox (§13 below) against the skill's golden test
set. Record: test pass rate, latency improvement vs. baseline, regression flag.

Only candidates with `regression = false` and `tests_passing >= baseline_tests_passing`
proceed to COMMIT.

### COMMIT

Commit the winning candidate via `safe_assert/1`. This runs the full pipeline:
CHR check → invariant check → Safety Rail Z3 → Merkle leaf → `assertz`.

```prolog
%% commit_candidate(+Candidate, +BaselineSnapshot) is det.
commit_candidate(Candidate, BaselineSnapshot) :-
    maplist(safe_retract_if_exists, Candidate.clauses_to_replace),
    maplist(safe_assert, Candidate.clauses),
    merkle_commit(skill_improved, Candidate, _Proof),
    emit_metric('stasis.improvement.committed_total', 1),
    schedule_canary_check(Candidate.skill, BaselineSnapshot, 86400). % 24h
```

### REFLECT (canary window close)

After the canary window, compare the current fitness vector to the pre-commit snapshot:

```prolog
reflect_on_commit(SkillName, BaselineSnapshot) :-
    measure_fitness_snapshot(SkillName, CurrentSnapshot),
    fitness_delta(BaselineSnapshot.fitness, CurrentSnapshot.fitness, Delta),
    (   Delta.improvement_score >= 0.0
    ->  confirm_commit(SkillName, CurrentSnapshot)
    ;   rollback_to_snapshot(SkillName, BaselineSnapshot)
    ).
```

Rollback is a `safe_retract`/`safe_assert` sequence that reconstructs the prior
clause set from the Merkle log. The rollback itself is a Merkle event — it is
auditable.

---

## Introspection Primitives

These are the SWI-Prolog builtins that constitute the agent's self-reflection
capability. All are available in SWI-Prolog 9.x without additional packages.

| Primitive | What it's for |
|---|---|
| `clause/2` | Read own source clauses: `clause(head(X), Body)` |
| `clause/3` | Read clause with reference for subsequent retract |
| `predicate_property/2` | Query metadata: `tabled`, `dynamic`, `discontiguous`, `imported_from` |
| `current_predicate/1` | Enumerate all predicates in scope |
| `functor/3` | Decompose/construct term heads: `functor(foo(a,b), foo, 2)` |
| `arg/3` | Extract argument by position |
| `=..` (univ) | Convert between term and list form: `foo(a,b) =.. [foo,a,b]` |
| `copy_term/2` | Fresh variable copy — essential before any term manipulation |
| `numbervars/3` | Bind unbound variables to `'$VAR'(N)` for ground comparison |
| `term_to_atom/2` | Serialize term to atom for Merkle hashing |
| `portray_clause/1` | Human-readable term output for briefing generation |
| `strip_module/3` | Remove module qualifier: `m:foo(X)` → module `m`, term `foo(X)` |
| `assert_attributed_variable/2` | Attach provenance metadata to a variable |

**The variable capture hazard.** Before manipulating any clause retrieved via
`clause/2`, always call `copy_term/2`:

```prolog
% WRONG — variables in the retrieved clause are shared with the original
clause(my_skill(X), Body),
Body2 = (Body, extra_step).        % mutates the original clause's variables

% CORRECT — work with a fresh copy
clause(my_skill(X), Body),
copy_term((my_skill(X) :- Body), (FreshHead :- FreshBody)),
FreshBody2 = (FreshBody, extra_step).
```

After `copy_term/2`, `FreshHead` and `FreshBody` have new variables that share no
identity with the original. All manipulation proceeds on the copy.

---

## Effectiveness as a First-Class Object

In STASIS, a **skill** is not just a predicate. It is a structured object whose
clauses, fitness history, provenance, and test set are all first-class:

```prolog
%% skill_record(+Name, -Record) is det.
%% Retrieves the full record for a skill.
%
% Record is:
%   skill{
%     name:     Name,            % atom
%     clauses:  ClauseSet,       % list of (Head :- Body) terms
%     fitness:  FitnessDict,     % fitness{latency_p99_ms, success_rate, ...}
%     version:  MerkleLeafHash,  % hex atom — identifies this exact clause set
%     parent:   ParentHash,      % null or prior version's hash
%     tests:    TestRecord        % test{positive: Pos, negative: Neg}
%   }
skill_record(Name, Record) :-
    introspect_skill(Name, ClauseSet),
    skill_fitness(Name, Fitness),
    skill_version(Name, Version),
    skill_parent(Name, Parent),
    skill_tests(Name, TestRecord),
    Record = skill{
        name:    Name,
        clauses: ClauseSet,
        fitness: Fitness,
        version: Version,
        parent:  Parent,
        tests:   TestRecord
    }.
```

The fitness dictionary is a dict (SWI-Prolog's `library(dicts)`) because dicts
support partial access, pattern matching, and `.get(Key, Default)` semantics.
This allows the meta-improvement loop to read any fitness dimension without
knowing the full set of dimensions — extensible without code changes.

**Counters are session-persistent.** `nb_setval/nb_getval` are used for OTel
counters because they survive backtracking — backtracking an attempted improvement
does not rewind the observation counts.

```prolog
%% increment_call_counter(+SkillName) is det.
increment_call_counter(SkillName) :-
    atom_concat('stasis.skill.calls.', SkillName, Key),
    (nb_current(Key, N) -> N1 is N + 1 ; N1 = 1),
    nb_setval(Key, N1).
```

---

## The Instrumented Meta-Interpreter

The meta-interpreter is the observability foundation for the self-improvement loop.
The classical four-clause Prolog meta-interpreter:

```prolog
solve(true) :- !.
solve((A, B)) :- !, solve(A), solve(B).
solve(\+ G) :- !, \+ solve(G).
solve(G) :- call(G).
```

Extended for STASIS with OTel hooks and depth tracking:

```prolog
%% solve(+Goal, +Depth) is nondet.
%% Interprets Goal with full instrumentation.
:- module(meta_interpreter, [solve/2]).

solve(true, _) :- !.

solve((A, B), Depth) :- !,
    solve(A, Depth),
    solve(B, Depth).

solve(\+ G, Depth) :- !,
    \+ solve(G, Depth).

solve(G, Depth) :-
    functor(G, Name, Arity),
    get_time(T0),
    ( call(G) -> Outcome = success ; Outcome = failure, fail ),
    get_time(T1),
    DurMs is (T1 - T0) * 1000.0,
    emit_metric(predicate_latency_ms, Name/Arity, DurMs),
    emit_metric(predicate_outcome,    Name/Arity, Outcome),
    Depth1 is Depth + 1,
    ( Depth1 > 1000 ->
        throw(error(depth_limit_exceeded(G), context(solve/2, Depth1)))
    ; true
    ).
```

**Two operational modes:**

*Shadow mode* — used during EVALUATE. The meta-interpreter runs the candidate
clause set in a pengines sandbox in parallel with the production set, comparing
outcomes. Neither run has side effects; the sandbox prevents KB mutation.

*Instrumented live mode* — used during OBSERVE. Runs over the production clause
set on every call within the improvement loop's scope. Depth limit prevents
runaway recursion. Overhead is acceptable because the loop runs on a background
thread, not on the hot path.

The meta-interpreter is **Tier 3 code** and is itself immutable (protected by
`verify:check_invariants/1` as a system predicate). An agent cannot rewrite the
interpreter that is watching it — this prevents an adversarial skill from
disabling its own observation.

---

## Three Modes of Self-Modification

Ordered from lowest to highest risk. The agent should attempt the lowest
applicable mode first.

### Mode A — Tuning

**What:** Modify numeric constants in existing clauses. Thresholds, weights,
retry counts, timeout values.

**Why it's safe:** The clause shape is unchanged. The Tier 1 linter passes
trivially because no new predicates or structures are introduced. The CHR check
passes for the same reason.

**Mechanism:**
```prolog
%% tune_threshold(+SkillName, +ArgPosition, +NewValue) is det.
%% Replaces a numeric constant at ArgPosition in all clauses of SkillName.
tune_threshold(SkillName, ArgPos, NewValue) :-
    introspect_skill(SkillName, Clauses),
    maplist(replace_arg(ArgPos, NewValue), Clauses, NewClauses),
    maplist(safe_retract, Clauses),
    maplist(safe_assert, NewClauses).

replace_arg(Pos, Val, (H :- B), (H2 :- B)) :-
    H =.. List,
    nth1(Pos, List, _, Rest),
    nth1(Pos, NewList, Val, Rest),
    H2 =.. NewList.
```

**Typical triggers:** Latency p99 consistently above threshold, meaning the
timeout or batch size is too conservative; calibrate it down.

### Mode B — Specialization

**What:** Generate a faster version of a general predicate for the most common
input pattern. The general clause is retained; the specialized clause fires first.

**Why it works:** Prolog clause lookup is deterministic first-match. A
specialized clause with more specific argument patterns fires before the general
clause. For hot paths with a known common case, this is a 2–5× speedup with
zero risk of semantic change.

**Mechanism:**

1. Inspect the call histogram to identify the most common argument pattern.
2. Construct the specialized clause using `copy_term/2` to bind the common
   arguments while leaving uncommon ones as variables.
3. Partially evaluate the body by resolving known-true guards and simplifying
   arithmetic.
4. Sandbox-evaluate the specialized clause against the golden test set.
5. If no regression, commit via `safe_assert/1`.

```prolog
%% specialize_for_pattern(+SkillName, +CommonPattern) is det.
%% Generates and commits a specialized clause for CommonPattern.
specialize_for_pattern(SkillName, CommonPattern) :-
    introspect_skill(SkillName, [GeneralClause | _]),
    copy_term(GeneralClause, (GeneralHead :- GeneralBody)),
    %% Bind the common pattern arguments
    GeneralHead =.. [SkillName | Args],
    CommonPattern =.. [SkillName | PatternArgs],
    Args = PatternArgs,           % unify — binds common args
    %% Partially evaluate: simplify arithmetic, fold known-true guards
    partial_eval_body(GeneralBody, SpecializedBody),
    SpecializedClause = (GeneralHead :- SpecializedBody),
    safe_assert(SpecializedClause).
```

**The specialization registry** — each skill's specialized variants are tracked
by fitness profile so the improvement loop can compare them:

```prolog
:- dynamic skill_variant/3.
%% skill_variant(SkillName, VariantId, FitnessSnapshot)
```

### Mode C — Synthesis

**What:** Generate entirely new clauses from examples using Inductive Logic
Programming. The agent does not transform existing code — it learns rules from
positive and negative examples.

**When to use:** An entirely new capability is needed, or the existing clauses
cannot be improved by tuning or specialization alone.

**Risk:** Synthesis has the broadest search space. Generated clauses must be
aggressively sandboxed and require Translucent Gate approval if security-adjacent.

**Mechanism:** See §9 (ILP) below.

---

## Inductive Logic Programming for Novel Skill Learning

ILP is the systematic search for logic rules that cover all positive examples and
no negative examples. In STASIS it is the mechanism for Mode C synthesis.

### The covering algorithm

```
ILP_loop(Positives, Negatives, ExistingRules, NewRule):
    1. Find an uncovered positive example E.
    2. Generate the most specific rule covering E (bottom clause).
    3. Generalise until the rule covers some, but not all, positives.
    4. Remove negatives covered by the rule (prune).
    5. If rule covers no negatives: candidate.
    6. Repeat until all positives covered or search depth exceeded.
```

### Bottom clause construction in Prolog

The bottom clause is the most specific rule — it captures exactly one example,
including all relevant context from the background knowledge:

```prolog
%% bottom_clause(+Example, +BackgroundKnowledge, -BottomClause) is det.
bottom_clause(Example, BK, (Example :- Body)) :-
    prove_with_trace(Example, BK, ProofTrace),
    trace_to_body(ProofTrace, Body).

%% prove_with_trace(+Goal, +BK, -Trace) is det.
%% Proves Goal using background knowledge, recording which BK rules fired.
prove_with_trace(Goal, BK, [Goal | SubTraces]) :-
    member((Goal :- Body), BK),
    comma_to_list(Body, Goals),
    maplist(prove_with_trace_elem(BK), Goals, SubTraces).
```

### Generalization via inverse resolution

Once a bottom clause is available, generalization proceeds by anti-unification
(computing the least general generalization of two terms):

```prolog
%% lgg(+Term1, +Term2, -LGG) is det.
%% Least General Generalization of two terms.
lgg(T1, T2, T1) :- T1 == T2, !.
lgg(T1, T2, V) :-
    ( \+ compound(T1) ; \+ compound(T2) ), !,
    gensym(v, V).   % introduce a new variable
lgg(T1, T2, LGG) :-
    T1 =.. [F | Args1],
    T2 =.. [F | Args2],
    length(Args1, N), length(Args2, N), !,
    maplist(lgg, Args1, Args2, LGGArgs),
    LGG =.. [F | LGGArgs].
lgg(_, _, V) :- gensym(v, V).
```

### Integration with the safety pipeline

Every candidate rule from the ILP loop is:
1. Checked by `validate-stasis-tier1.pl` if tagged as Tier 1 material
2. Evaluated against the full negative example set in a pengines sandbox
3. Submitted to `safe_assert/1` — which runs the CHR + invariant + Safety Rail
   pipeline before committing

Security-adjacent skills (those with names matching the Safety Rail pattern) are
auto-routed to the Translucent Gate regardless of sandbox result.

---

## Explanation-Based Generalization

EBG caches the proof of a success as a new rule, allowing the same conclusion
to be reached faster next time by short-circuiting the proof search.

### The EBG pattern

1. A goal G succeeds. The meta-interpreter has recorded its proof trace.
2. Replay the proof trace, but with variables substituted for the specific
   constants that are *arguments* to the top-level call (not guards).
3. The resulting generalized trace, compiled into a clause, is the new rule.

```prolog
%% ebg(+Goal, +ProofTrace, -GeneralizedRule) is det.
%% Produces a generalized rule from a proof of Goal.
ebg(Goal, Trace, GeneralizedRule) :-
    copy_term(Goal, GoalTemplate),
    trace_to_body(Trace, SpecificBody),
    copy_term(SpecificBody, GeneralBody),
    %% Variables that appear in GoalTemplate are now free in GeneralBody too
    %% — they have been correctly generalized.
    GeneralizedRule = (GoalTemplate :- GeneralBody).
```

**When this fires:** After a skill succeeds on an input that has not been seen
before. The generalized rule fires first on future similar inputs, skipping the
full proof search.

**Caution:** EBG generates a lot of clauses over time. The fitness loop must
monitor clause count and prune clauses that have low utilization (few cache
hits) relative to their maintenance overhead.

---

## Partial Evaluation and Specialization

Partial evaluation computes a specialized version of a program given partial
knowledge of its input. In STASIS the common case is: the skill has a fixed
component (a database, a configuration, a constant background) that changes
infrequently. Partially evaluating with that component frozen produces a
faster residual program.

### Manually implemented in Prolog

```prolog
%% partially_evaluate(+Clause, +BoundArgs, -ResidualClause) is det.
%% Given a clause and some arguments already bound, produce the residual.
partially_evaluate((Head :- Body), BoundArgs, (SpecHead :- ResidualBody)) :-
    Head =.. [Name | Args],
    apply_bindings(Args, BoundArgs, BoundHead),
    SpecHead =.. [Name | BoundHead],
    reduce_body(Body, BoundHead, ResidualBody).

%% reduce_body(+Body, +Bindings, -ReducedBody) is det.
%% Evaluates deterministic sub-goals that are now ground; retains the rest.
reduce_body((A, B), Bindings, Reduced) :-
    reduce_body(A, Bindings, RA),
    reduce_body(B, Bindings, RB),
    ( RA == true -> Reduced = RB
    ; RB == true -> Reduced = RA
    ; Reduced = (RA, RB)
    ).
reduce_body(Goal, _, true) :-
    ground(Goal), catch(call(Goal), _, fail), !.  % evaluate now
reduce_body(Goal, _, Goal).                        % keep for runtime
```

### `library(apply_macros)` integration

SWI-Prolog's `apply_macros` optimizes away the overhead of `maplist/call` when
the predicate argument is known at compile time. For hot list pipelines inside
skills, loading this library and using `maplist` correctly gives compile-time
unrolling.

### The specialization dispatcher

When a skill has both a general clause and one or more specialized variants,
use a dispatcher to route based on argument shape:

```prolog
%% sort_skill(+List, -Sorted) is det.
%% Dispatcher: routes to specialized or general implementation.
sort_skill([], []) :- !.                           % trivial case
sort_skill([_], [_]) :- !.                        % trivial case
sort_skill([A,B], Sorted) :- !,                   % 2-element specialization
    sort_2(A, B, Sorted).
sort_skill(List, Sorted) :-                        % general path
    sort_general(List, Sorted).
```

The dispatcher is generated by the improvement loop and committed via
`safe_assert/1`. The general clause is never removed — the dispatcher is an
additional clause that fires first when the pattern matches.

---

## Constraint Logic Programming for Evolving Policy

CLP allows the agent to express policy as constraints that propagate
automatically rather than as if-then-else code.

### `library(clpfd)` for resource policy

```prolog
:- use_module(library(clpfd)).

%% within_resource_budget(+TaskCost, +BudgetRemaining) is semidet.
%% Succeeds if the task can be started within the remaining budget.
within_resource_budget(TaskCost, BudgetRemaining) :-
    TaskCost #=< BudgetRemaining,
    BudgetRemaining #>= 0.
```

The constraint `TaskCost #=< BudgetRemaining` propagates automatically — as
soon as either variable is bound, the constraint is re-evaluated. Adding a new
resource type means adding a new CLP constraint, not rewriting the policy engine.

### `library(clpq)` for continuous optimization

For continuous domains (latency targets, quality thresholds, cost weights):

```prolog
:- use_module(library(clpq)).

%% acceptable_quality(+ActualScore, +MinAcceptable) is semidet.
acceptable_quality(Actual, Min) :-
    {Actual >= Min, Actual =< 1.0}.
```

CLP constraints are Tier 2 material — they belong in `policies/constraints.chr`
alongside the CHR rules, since both are constraint-store-based reasoning.

---

## Abduction via Tabling — "Why Did This Fail?"

When a skill regresses after an update, the agent needs to diagnose why.
Abduction (reasoning backwards from an observation to its causes) is the
mechanism.

### The diagnostic query

```prolog
%% diagnose_regression(+SkillName, +FailingInput, -Hypothesis) is nondet.
%% For each possible cause of SkillName failing on FailingInput, unify Hypothesis
%% with a minimal set of database facts whose removal enables success.
diagnose_regression(SkillName, FailingInput, Hypothesis) :-
    Goal =.. [SkillName | FailingInput],
    abduction(Goal, Hypothesis).

%% abduction(+Goal, -AbductiveHypothesis) is nondet.
%% Uses tabling to avoid re-exploring the same hypothesis twice.
:- table abduction/2.

abduction(Goal, []) :- call(Goal), !.   % already true — no assumptions needed
abduction(Goal, [Assumption | Rest]) :-
    abduce_assumption(Goal, Assumption),
    assert(Assumption),
    catch(abduction(Goal, Rest), _, fail),
    retract(Assumption).
```

With tabling, `abduction/2` will not re-enter the same goal with the same
partially-bound hypothesis — memoization prevents infinite loops. This gives
a tractable bounded-depth abductive search.

### Use in the improvement loop

After REFLECT detects a regression, the loop calls `diagnose_regression/3` to
produce a hypothesis, which is attached to the Crucible briefing. The briefing
becomes:

> Skill `sort_skill` regressed from 97% pass rate to 83%. Abductive analysis
> suggests the failure is caused by the new `sort_2` specialization applying
> incorrectly to equal-element inputs. Recommended: add test case `sort_skill([A,A], Sorted)`
> to the golden set and re-evaluate the specialization.

---

## Advanced SWI-Prolog Features for Agentic Patterns

### Attribute variables — metadata that survives unification

Attribute variables attach metadata to unbound variables. When the variable is
unified, a hook fires, allowing the agent to enforce constraints on variable
binding without CHR overhead:

```prolog
%% annotate_with_tier(+Var, +Tier) is det.
%% Attaches tier metadata to an unbound variable.
%% When Var is later unified, verify the binding respects the tier.
annotate_with_tier(Var, Tier) :-
    put_attr(Var, stasis_tier, Tier).

:- multifile attr_unify_hook/2.
attr_unify_hook(Tier, Value) :-
    ( tier_compatible(Value, Tier) -> true
    ; throw(error(tier_violation(Value, Tier), attr_unify_hook/2))
    ).
```

This makes it impossible to accidentally unify a Tier 1 predicate reference
with a value that belongs in Tier 3 — the hook catches it at unification time.

### `freeze/2` and `when/2` — speculative computation

`freeze(Var, Goal)` delays Goal until Var is bound. This is the Prolog equivalent
of a lazy future — schedule the work, don't pay for it until needed:

```prolog
%% prefetch_fitness_if_needed(+SkillName) is det.
%% Schedules fitness recomputation when the fitness flag becomes stale.
prefetch_fitness_if_needed(SkillName) :-
    fitness_stale_flag(SkillName, Flag),
    freeze(Flag, recompute_fitness(SkillName)).
```

`when/2` is a generalization: the goal fires when a cond about multiple
variables becomes true. Useful for triggering improvements when two conditions
are met simultaneously (e.g., high latency AND high call volume).

### `library(delcont)` — sandboxed checkpoints

Delimited continuations allow the agent to "run this section of code, but
if it fails, reset to this point and try the alternative":

```prolog
:- use_module(library(delcont)).

%% try_improvement_with_rollback(+Candidate) is det.
try_improvement_with_rollback(Candidate) :-
    reset(                         % establish checkpoint
        apply_candidate(Candidate),
        rollback,
        _
    ).

apply_candidate(Candidate) :-
    maplist(safe_assert, Candidate.clauses),
    run_canary(Candidate.skill, Outcome),
    ( Outcome = regressed -> shift(rollback) ; true ).
```

This is cleaner than manual `assert/retract` for rollback — the delimited
continuation automatically handles the state reset.

---

## Sandboxed Candidate Evaluation via Pengines

Every candidate clause (from any modification mode) is evaluated in a pengines
sandbox before any commitment. Pengines provide true isolation: the candidate
runs in a separate engine with no shared KB state, no filesystem access, no
network, and a hard CPU time limit.

```prolog
%% evaluate_candidate_in_sandbox(+SkillName, +Candidate, -Verdict) is det.
evaluate_candidate_in_sandbox(SkillName, Candidate, Verdict) :-
    skill_tests(SkillName, Tests),
    term_string(Candidate.clauses, CandidateSrc),
    pengine_create([
        application(stasis_sandbox),
        src_text(CandidateSrc),
        sandbox(true),
        time_limit(5)
    ], PengineId),
    run_test_set_in_pengine(PengineId, Tests.positive, PosResults),
    run_test_set_in_pengine(PengineId, Tests.negative, NegResults),
    pengine_destroy(PengineId),
    score_results(PosResults, NegResults, Verdict).

%% score_results(+PosResults, +NegResults, -Verdict) is det.
score_results(PosResults, NegResults, Verdict) :-
    include(==(pass), PosResults, PosPass),
    include(==(fail), NegResults, NegFail),  % failures on negatives are expected
    length(PosResults, TotalPos),
    length(NegResults, TotalNeg),
    length(PosPass,  PassedPos),
    length(NegFail,  PassedNeg),
    PosRate is PassedPos / max(1, TotalPos),
    NegRate is PassedNeg / max(1, TotalNeg),
    baseline_test_rate(SkillName, BaselinePos, BaselineNeg),
    ( PosRate >= BaselinePos, NegRate >= BaselineNeg ->
        Verdict = pass(_{pos_rate: PosRate, neg_rate: NegRate})
    ;   Verdict = regressed(_{pos_rate: PosRate, neg_rate: NegRate,
                               baseline_pos: BaselinePos, baseline_neg: BaselineNeg})
    ).
```

**The sandbox application** (`stasis_sandbox`) is a pre-configured pengines
environment that includes background knowledge (Tier 1 facts, helper predicates)
but excludes the live KB, the OTel bridge, and the Safety Rail bridge. The
candidate runs against a frozen snapshot of background knowledge.

---

## The Cybernetic Cage

Every self-modification that survives the sandbox must still pass through
`safe_assert/1`. This is the inner wall of the cybernetic cage.

```
Candidate clause
      │
      ▼
pengines sandbox evaluation          ← outer wall: no live KB access
      │ pass
      ▼
constraints:check_constraints/1      ← Tier 2 CHR: banned predicates
      │ no violations
      ▼
verify:check_invariants/1            ← Tier 3: no Tier 1 head redefinition
      │ no violations
      ▼
mcp_bridge:submit_skill_proposal/2   ← Safety Rail Z3: formal policy check
      │ SafetyVerdict::Safe
      ▼
merkle_bridge:merkle_commit/3        ← Audit: immutable, append-only leaf
      │
      ▼
assertz/1                            ← only authorized mutation point
```

The improvement loop itself — `meta/improve.pl`, `meta/introspect.pl`,
`meta/verify.pl`, `meta/fitness.pl` — is protected by `is_system_predicate/1`
in `verify.pl`. Any attempt to redefine these predicates is caught at the
`check_invariants` step. The agent cannot modify the cage from within.

---

## Goodhart Hazards and Metric Design

**Goodhart's Law:** When a measure becomes a target, it ceases to be a good
measure. A self-improving agent optimizing naively against a single fitness
metric will learn to game that metric rather than improve the underlying
capability.

### Counter-metrics

Every primary optimization metric must have at least one counter-metric that
degrades if the agent cheats:

| Primary metric | Counter-metric |
|---|---|
| Latency p99 ↓ | Test pass rate (unchanged or better) |
| Success rate ↑ | Silent failure count (catching and hiding errors is not success) |
| Clause count ↓ | Coverage on golden negative set (fewer clauses may mean less specificity) |
| Memory usage ↓ | Tabling cache hit rate (flushing caches reduces memory but costs latency) |

### Composite fitness score

The fitness function used in REFLECT must be a composite that weights the
counter-metric:

```prolog
%% improvement_score(+Before, +After, -Score) is det.
%% Score > 0.0 means genuine improvement. Score < 0.0 means degradation.
improvement_score(Before, After, Score) :-
    LatencyImprovement is (Before.latency_p99_ms - After.latency_p99_ms)
                            / max(1.0, Before.latency_p99_ms),
    QualityDelta       is After.test_pass_rate - Before.test_pass_rate,
    %% Penalize hard if test quality dropped
    ( QualityDelta < 0.0 ->
        Score is LatencyImprovement + 3.0 * QualityDelta  % 3× penalty
    ;   Score is LatencyImprovement + QualityDelta
    ).
```

A 15% latency improvement that costs 5% test quality scores −0.0 (neutral to
negative after the 3× penalty). This prevents the agent from trading correctness
for speed.

### Never optimize without a baseline

Every COMMIT requires a `BaselineSnapshot` from the MEASURE phase. There is no
opt-out. If the snapshot is missing, the COMMIT is rejected.

---

## What STASIS Adds Beyond Stock SWI-Prolog

| Capability | Stock SWI-Prolog | STASIS addition |
|---|---|---|
| Clause introspection | `clause/2` | + fitness metadata dict, provenance hash, golden test association |
| KB mutation | `assertz/retract` (unrestricted) | `safe_assert/safe_retract` — CHR + invariant + Z3 + Merkle |
| Decidability | `library(tabling)` prevents loops | + Tier 1 syntactic linter: functor symbols banned from heads |
| Candidate evaluation | `library(pengines)` available | + fitness delta comparison, golden test scoring, regression guard |
| Self-improvement loop | ad-hoc patterns possible | Formalised 6-state state machine + canary window + rollback |
| Audit | None built in | RFC 6962 Merkle log — every mutation is a verifiable leaf |
| Persistence | `library(persistency)` (file-based) | Merkle-log replay on restart — state is provably derivable |
| Policy evolution | No standard pattern | Tier 2 CHR constraint store + `check_constraints/1` gate |
| Abduction | Possible but ad-hoc | Tabling + `abduction/2` meta-predicate — tractable diagnosis |
| Fitness as data | None | `skill_fitness/2` dict — queryable by any Prolog goal |

---

## Open Questions for Future Design

These are deliberately deferred — the answers depend on operational experience
that has not yet accumulated.

1. **Fitness aggregation across domains.** When multiple factories report fitness
   vectors, the global optimisation objective is unclear: weighted sum vs. Pareto
   frontier. Pareto is the principled answer but harder to implement.

2. **Concurrent self-improvement.** When two agents attempt to improve the same
   skill simultaneously, who wins? Optimistic locking on the Merkle leaf version
   hash is the obvious mechanism.

3. **Federated skill libraries.** Spacecraft learn new skills. Can one spacecraft
   share a skill with another? Provenance chains in the Merkle log make this
   auditable; the trust model is the open design question.

4. **Semantic preservation proofs.** Specialization and EBG preserve semantics
   informally. A formal proof (via `safety-rail/src/tier2/` rocq proofs, when
   implemented) would be needed for HOOTL-grade assurance on those transformations.

5. **Improvement loop termination.** The improvement loop itself must terminate.
   A Tier 1-certified termination proof for the loop predicate is needed before
   HOOTL operation. Currently the loop terminates because `nominate_for_improvement`
   requires `Fitness.sample_count >= 100` and improvement candidates are consumed —
   but this is a policy argument, not a formal proof.

---

## References and Prior Art

- **E. Shapiro, *Algorithmic Program Debugging* (1982)** — divide-and-query
  localization; the foundational method for Prolog diagnostic reasoning.

- **T. Mitchell, R. Keller, S. Kedar-Cabelli, *Explanation-Based Generalization:
  A Unifying View* (1986)** — the EBG algorithm described in §8.

- **S. Muggleton & L. De Raedt, *Inductive Logic Programming: Theory and Methods*
  (1994)** — the theoretical foundation for ILP (§9).

- **SWI-Prolog 9.x Manual §8 (Meta-predicates), §§ on tabling, CHR, pengines,
  delimited continuations, attribute variables** — all primitives referenced here
  are documented there.

- **A. Morales, M. Carro, M. Hermenegildo, *Practical Global Stack-Size
  Analysis for Logic Languages* (1992)** — context for understanding why
  tabling alone does not give the decidability guarantee (§3 of STASIS-LANGUAGE.md).

- **`STASIS-LANGUAGE.md`** in this repository — the tier architecture, boundary
  enforcement, and safety pipeline that this document builds on.

---

*This document describes design intent and patterns. Implementation tracks the
design: `core-station/protoplasm/meta/improve.pl` et al. are the current
implementation baseline. Deviations between this document and the implementation
are implementation gaps to be resolved via the standard Forge/Crucible briefing
cycle — this document takes precedence.*
