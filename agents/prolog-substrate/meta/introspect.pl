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
    functor(PlainHead, Name, Arity),
    findall(Output, (
        member(Input, Inputs),
        % Construct the goal in the correct module
        (  is_list(Input) 
        -> Goal =.. [Name | Input]
        ;  Arity == 1
        -> Goal =.. [Name, Input]
        ;  % Fallback for complex heads (like fib/2)
           copy_term(PlainHead, Goal),
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
%
% Baseline performance measurement. 
measure_performance(Head, AvgLatency, Samples) :-
    Samples = 100, % Per CLAUDE.md §3 Phase 8 requirements
    % Ensure Head is instantiated enough to run
    (  \+ ground(Head)
    -> strip_module(Head, _, PlainHead),
       functor(PlainHead, Name, _),
       % Use a safe catch-all to find golden_input if it exists
       ( catch(current_predicate(golden_input/2), _, fail) -> 
         ( catch(golden_data:golden_input(Name, [FirstInput|_]), _, fail) ->
           ( is_list(FirstInput) -> Head =.. [Name|FirstInput] ; Head =.. [Name, FirstInput] )
         ; true
         )
       ; true
       )
    ;  true
    ),
    get_time(T1),
    forall(between(1, Samples, _), (Head -> true ; true)),
    get_time(T2),
    AvgLatency is (T2 - T1) / Samples.

% Hook to find golden inputs if needed for performance measurement
:- multifile golden_data:golden_input/2.
