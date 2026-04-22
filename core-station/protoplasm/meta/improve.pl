:- module(improve, [
    optimize_skill/1,
    improve_if_slow/2,
    evaluate_candidate/3
]).

:- use_module('../core/safety_bridge').
:- use_module('../core/otel_bridge').
:- use_module('introspect').
:- use_module('verify').
:- use_module('../tests/golden/default').

/** <module> STASIS Self-Improvement Loop (Phase 10)
 * 
 * Implements the 6-state evolutionary loop defined in STASIS-SELF-IMPROVEMENT.md:
 * 1. Observe: Identify candidate skill
 * 2. Measure: Record performance baseline (Latency)
 * 3. Introspect: Analyze current implementation
 * 4. Hypothesize: Propose optimized variant (logic-tuner)
 * 5. Evaluate: Sandbox verification (non-regression)
 * 6. Commit: Authorized mutation via safety_bridge
 */

%% optimize_skill(+SkillHead) is det.
optimize_skill(SkillHead) :-
    improve_if_slow(SkillHead, 0.0).

%% improve_if_slow(+SkillHead, +Threshold) is det.
improve_if_slow(SkillHead, Threshold) :-
    % 1. Observe & 2. Measure
    measure_performance(SkillHead, AvgLatency, _Samples),
    (   AvgLatency >= Threshold
    ->  % 3. Introspect
        inspect_predicate(SkillHead, CurrentClauses),
        emit_metric('stasis.improvement.observation_total', 1),
        
        % 4. Hypothesize
        propose_improvement(CurrentClauses, NewClauses),
        
        % 5. Evaluate & 6. Commit
        forall(member(NewClause, NewClauses), (
             (check_invariants(NewClause), \+ clause_exists(NewClause))
             -> (   evaluate_candidate(SkillHead, NewClause, Verdict),
                    (  Verdict \= regressed
                    -> ( safe_assert(NewClause),
                         emit_metric('stasis.improvement.success_total', 1)
                       )
                    ;  ( emit_metric('stasis.improvement.regression_veto_total', 1),
                         print_message(warning, regression_detected(NewClause))
                       )
                    )
                )
             ;  true
        ))
    ;   true
    ).

%% evaluate_candidate(+Head, +NewClause, -Verdict) is det.
evaluate_candidate(Head, NewClause, Verdict) :-
    % 1. Setup Golden Baseline
    strip_module(Head, Module, PlainHead),
    functor(PlainHead, Name, _),
    (  catch(golden_data:golden_input(Name, Inputs), _, fail)
    -> true
    ;  Inputs = []
    ),
    
    % 2. Measure Baseline Success Set
    test_predicate(Module:Head, Inputs, OldResults),
    
    % 3. Measure Candidate Success Set (Sandbox)
    evaluate_in_sandbox(PlainHead, NewClause, Inputs, NewResults),
    
    % 4. Non-Regression Verdict
    % Comparison ignores module prefixes to allow sandbox/user parity.
    maplist(strip_result_module, OldResults, StrippedOld),
    maplist(strip_result_module, NewResults, StrippedNew),
    sort(StrippedOld, SortedOld),
    sort(StrippedNew, SortedNew),
    (   SortedNew \= SortedOld
    ->  Verdict = regressed
    ;   Verdict = equivalent
    ).

% Helper: success(user:fib(10,55)) -> success(fib(10,55))
strip_result_module(success(M:Term), success(Term)) :- !, nonvar(M).
strip_result_module(Other, Other).

%% evaluate_in_sandbox(+PlainHeadSpec, +Clause, +Inputs, -Results) is det.
evaluate_in_sandbox(_HeadSpec, Clause, Inputs, Results) :-
    % Determine HeadTemplate and Body from Clause
    ( Clause = (HeadTemplate :- Body) -> true ; ( HeadTemplate = Clause, Body = true ) ),
    strip_module(HeadTemplate, _, PlainHeadTemplate),
    functor(PlainHeadTemplate, _Name, Arity),
    findall(Output, (
        member(Input, Inputs),
        copy_term((PlainHeadTemplate, Body), (H, B)),
        % Bind input to first arg
        ( Arity > 0 -> (arg(1, H, Input) -> true ; true ) ; true ),
        ( catch(call_with_time_limit(1, B), _, fail)
        -> Output = success(H)
        ;  Output = fail
        )
    ), Results).

clause_exists((Head :- Body)) :- !, clause(Head, Body).
clause_exists(Head) :- clause(Head, true).

%% propose_improvement(+CurrentClauses, -NewClauses) is det.
propose_improvement(Clauses, Clauses).

% Message formatting
:- multifile prolog:message//1.
prolog:message(regression_detected(Clause)) -->
    [ 'STASIS Regression Veto: Candidate discarded: ~w'-[Clause] ].
