:- use_module(library(plunit)).
:- use_module('../meta/introspect').

% Define a dummy skill for testing
:- dynamic user:test_skill/1.

:- begin_tests(introspect).

test(inspect_predicate) :-
    retractall(user:test_skill(_)),
    assertz(user:test_skill(a)),
    assertz(user:test_skill(b)),
    introspect:inspect_predicate(user:test_skill(_), Clauses),
    length(Clauses, 2),
    member((user:test_skill(a) :- true), Clauses),
    member((user:test_skill(b) :- true), Clauses).

test(test_predicate_success) :-
    retractall(user:test_skill(_)),
    assertz(user:test_skill(a)),
    assertz(user:test_skill(b)),
    introspect:test_predicate(user:test_skill(_), [a, b], Results),
    member(success(user:test_skill(a)), Results),
    member(success(user:test_skill(b)), Results).

test(measure_performance) :-
    retractall(user:test_skill(_)),
    assertz(user:test_skill(a)),
    introspect:measure_performance(user:test_skill(a), Latency, Samples),
    number(Latency),
    Samples =:= 100.

:- end_tests(introspect).
