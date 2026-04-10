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

/** <module> Meta-Improvement Loop
 * 
 * Provides predicates for self-modifying agents to optimize their own skills.
 * Orchestrates the Observe -> Introspect -> Construct -> Evaluate -> Propose loop.
 */

%% optimize_skill(+SkillHead) is det.
%
% Baseline optimization entry point for manual triggering.
optimize_skill(SkillHead) :-
    improve_if_slow(SkillHead, 0.0). % Force optimization for testing

%% improve_if_slow(+SkillHead, +Threshold) is det.
%
% The core evolutionary loop.
improve_if_slow(SkillHead, Threshold) :-
    % 1. Observe: Measure current performance
    measure_performance(SkillHead, AvgLatency, Samples),
    (   (AvgLatency >= Threshold, Samples >= 100)
    ->  % 2. Introspect: Get current definition
        inspect_predicate(SkillHead, CurrentClauses),
        
        % Emit observation metric
        emit_metric('sati_central.prolog.improvement_observation_total', 1),
        
        % 3. Construct: Generate candidate optimization
        propose_improvement(CurrentClauses, NewClauses),
        
        % 4. Evaluate & Propose: Verify invariants and sandbox results
        forall(member(NewClause, NewClauses), (
             (check_invariants(NewClause), \+ clause_exists(NewClause))
             -> (   evaluate_candidate(SkillHead, NewClause, Verdict),
                    ( Verdict \= regressed
                    -> ( safe_assert(NewClause),
                         emit_metric('sati_central.prolog.improvement_success_total', 1)
                       )
                    ;  ( print_message(warning, regression_detected(NewClause)),
                         emit_metric('sati_central.prolog.improvement_regression_veto_total', 1)
                       )
                    )
                )
             ;  true
        ))
    ;   true
    ).

%% evaluate_candidate(+Head, +NewClause, -Verdict) is det.
%
% Evaluates a candidate clause in a sandbox against golden test sets.
evaluate_candidate(Head, NewClause, Verdict) :-
    % 1. Measure Old Correctness
    strip_module(Head, _Module, PlainHead),
    functor(PlainHead, Name, _),
    (   golden_input(Name, Inputs)
    ->  true
    ;   Inputs = []
    ),
    test_predicate(Head, Inputs, OldResults),
    
    % 2. Evaluate candidate correctness manually
    (   match_clause(NewClause, Inputs, NewResults)
    ->  true
    ;   NewResults = error(logic_regression)
    ),
    
    % 3. Compare Results
    (   NewResults \= OldResults
    ->  Verdict = regressed
    ;   Verdict = equivalent
    ).

match_clause((Head :- Body), Inputs, Results) :-
    strip_module(Head, _, PlainHead),
    functor(PlainHead, Name, Arity),
    findall(success(ArgTemplate), (
        member(Input, Inputs),
        copy_term((PlainHead, Body), (ArgTemplate, TestBody)),
        (  is_list(Input) 
        -> ArgTemplate =.. [Name|Input] 
        ;  Arity == 1 
        -> arg(1, ArgTemplate, Input)
        ;  arg(1, ArgTemplate, Arg1), % For fib(N, F), Input is N
        Arg1 = Input
        ),
        catch(TestBody, _, fail)
    ), Results).

clause_exists((Head :- Body)) :- !, clause(Head, Body).
clause_exists(Head) :- clause(Head, true).

%% propose_improvement(+CurrentClauses, -NewClauses) is det.
%
% Phase 8/9 Strategy: Identity mapping for established safety baseline.
propose_improvement(Clauses, Clauses).

% Message formatting
:- multifile prolog:message//1.
prolog:message(proposing_improvement(Clause)) -->
    [ 'Proposing self-improvement: ~w'-[Clause] ].
prolog:message(regression_detected(Clause)) -->
    [ 'Regression detected for candidate: ~w. Discarding.'-[Clause] ].
prolog:message(optimized(Head)) -->
    [ 'Executing optimized version of ~w'-[Head] ].
