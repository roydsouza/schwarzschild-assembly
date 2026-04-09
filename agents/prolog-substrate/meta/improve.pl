:- module(improve, [
    optimize_skill/1,
    improve_if_slow/2
]).

:- use_module('../core/safety_bridge').
:- use_module('../core/otel_bridge').
:- use_module('introspect').
:- use_module('verify').

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
    (   (AvgLatency >= Threshold, Samples >= 10)
    ->  % 2. Introspect: Get current definition
        inspect_predicate(SkillHead, CurrentClauses),
        % 3. Construct: Generate candidate optimization
        propose_improvement(CurrentClauses, NewClauses),
        % 4. Evaluate & Propose: Verify invariants and assert
        forall(member(NewClause, NewClauses), (
             (check_invariants(NewClause), \+ clause_exists(NewClause))
             -> safe_assert(NewClause)
             ;  true
        ))
    ;   true
    ).

clause_exists((Head :- Body)) :- !, clause(Head, Body).
clause_exists(Head) :- clause(Head, true).

%% propose_improvement(+CurrentClauses, -NewClauses) is det.
%
% Generates improved candidates. Currently implements simple reordering 
% and cut injection as a baseline strategy.
propose_improvement(CurrentClauses, NewClauses) :-
    % For now, we simulate an optimization by adding a cut to the first clause
    % if it doesn't already have one, or just re-proposing the clauses.
    maplist(inject_cut, CurrentClauses, NewClauses).

inject_cut((Head :- Body), (Head :- (Body, !))) :-
    \+ contains_cut(Body), !.
inject_cut(Clause, Clause).

contains_cut(!) :- !.
contains_cut((A, _)) :- contains_cut(A), !.
contains_cut((_, B)) :- contains_cut(B), !.

% Message formatting
:- multifile prolog:message//1.
prolog:message(proposing_improvement(Clause)) -->
    [ 'Proposing self-improvement: ~w'-[Clause] ].
prolog:message(optimized(Head)) -->
    [ 'Executing optimized version of ~w'-[Head] ].
