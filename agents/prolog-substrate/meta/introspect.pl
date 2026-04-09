:- module(introspect, [
    inspect_predicate/2,
    test_predicate/3,
    measure_performance/3
]).

:- use_module('../core/otel_bridge').

/** <module> Introspection and Performance Measurement
 * 
 * Provides predicates to read current skill definitions and analyze their performance.
 */

%% inspect_predicate(+Head, -Clauses) is det.
%
% Retrieves all clauses defining the predicate matching Head.
% Clauses is a list of (Head :- Body) terms.
inspect_predicate(Head, Clauses) :-
    findall((Head :- Body), clause(Head, Body), Clauses).

%% test_predicate(+Head, +Inputs, -Results) is det.
%
% Runs a predicate against a set of Inputs and returns the Results.
% Uses a time limit to prevent hung evaluations in the sandbox.
test_predicate(Head, Inputs, Results) :-
    % Ensure Head is stripped of potential module qualification for functor/3
    strip_module(Head, Module, PlainHead),
    functor(PlainHead, Name, _Arity),
    findall(Output, (
        member(Input, Inputs),
        % Construct the goal in the correct module
        (  is_list(Input) 
        -> Goal =.. [Name | Input]
        ;  Goal =.. [Name, Input]
        ),
        catch(
            call_with_time_limit(5, Module:Goal),
            time_limit_exceeded,
            Output = error(timeout)
        ),
        ( var(Output) -> Output = success(Input) ; true )
    ), Results).

%% measure_performance(+Head, -AvgLatency, -Samples) is det.
%
% Baseline performance measurement. 
measure_performance(Head, AvgLatency, Samples) :-
    Samples = 10,
    get_time(T1),
    forall(between(1, Samples, _), (Head -> true ; true)),
    get_time(T2),
    AvgLatency is (T2 - T1) / Samples.
