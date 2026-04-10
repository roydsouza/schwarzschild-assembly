:- use_module(library(plunit)).
:- use_module('../meta/introspect').

:- begin_tests(regression).

test(skill_parity) :-
    % Define a ground truth test set
    Inputs = [[1, 3, _]],
    % Test against a known skill (between/3)
    introspect:test_predicate(user:between, Inputs, Results),
    length(Results, 1),
    Results = [success(user:between(1, 3, _))].

:- end_tests(regression).
