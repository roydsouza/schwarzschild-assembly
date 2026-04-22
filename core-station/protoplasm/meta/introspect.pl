:- module(introspect, [
    inspect_predicate/2,
    test_predicate/3,
    measure_performance/3
]).

:- use_module('../core/otel_bridge').

/** <module> Introspection and Performance Measurement
 * 
 * Provides predicates to read current skill definitions and analyze their performance.
 * Hardened for Phase 10 with time limits and module isolation.
 */

%% inspect_predicate(+Head, -Clauses) is det.
inspect_predicate(Head, Clauses) :-
    findall((Head :- Body), clause(Head, Body), Clauses).

%% test_predicate(+Head, +Inputs, -Results) is det.
test_predicate(Head, Inputs, Results) :-
    strip_module(Head, Module, PlainHead),
    functor(PlainHead, Name, Arity),
    findall(Output, (
        member(Input, Inputs),
        (  is_list(Input) 
        -> Goal =.. [Name | Input]
        ;  Arity == 1
        -> Goal =.. [Name, Input]
        ;  copy_term(PlainHead, Goal),
           (  arg(1, Goal, Input) -> true ; true )
        ),
        catch(
            call_with_time_limit(5, Module:Goal),
            time_limit_exceeded,
            Output = error(timeout)
        ),
        ( var(Output) -> Output = success(Module:Goal) ; true )
    ), Results).

%% measure_performance(+Head, -AvgLatency, -Samples) is det.
measure_performance(Head, AvgLatency, Samples) :-
    Samples = 10, % Baseline for introspection
    % Ensure Head is stripped for OTel bridge
    strip_module(Head, Module, PlainHead),
    measure_goal(Module:PlainHead, AvgLatency).
