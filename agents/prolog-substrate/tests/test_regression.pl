:- use_module(library(plunit)).
:- use_module('../meta/introspect').

:- begin_tests(regression).

test(skill_parity) :-
    % Define a ground truth test set
    Inputs = [1, 2, 3],
    % Test against a known skill
    introspect:test_predicate(user:between(1, 3, _), Inputs, Results),
    length(Results, 3),
    forall(member(R, Results), R = success(_)).

:- end_tests(regression).
