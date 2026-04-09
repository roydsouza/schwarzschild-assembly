:- use_module(library(plunit)).
:- use_module('../meta/introspect').

% Define a dummy skill for testing
:- dynamic test_skill/1.
test_skill(a).
test_skill(b).

:- begin_tests(introspect).

test(inspect_predicate) :-
    introspect:inspect_predicate(test_skill(_), Clauses),
    length(Clauses, 2),
    member((test_skill(a) :- true), Clauses),
    member((test_skill(b) :- true), Clauses).

test(test_predicate_success) :-
    introspect:test_predicate(test_skill(_), [a, b], Results),
    member(success(a), Results),
    member(success(b), Results).

test(measure_performance) :-
    introspect:measure_performance(test_skill(a), Latency, Samples),
    number(Latency),
    Samples =:= 10.

:- end_tests(introspect).
