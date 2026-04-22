:- set_prolog_flag(test_mode, true).
:- use_module(library(plunit)).
:- use_module('../meta/improve').
:- use_module('../meta/introspect').
:- use_module('golden/default').

:- begin_tests(self_improvement).

% Setup dummy skills in 'user' module
:- dynamic user:slow_skill/1.
:- dynamic user:safe_skill/1.
:- dynamic user:fib/2.

% Explicitly hook into golden_data for test visibility
:- multifile golden_data:golden_input/2.
golden_data:golden_input(safe_skill, [1, 42]).
golden_data:golden_input(fib, [10]).

test(improvement_trigger, [setup(retractall(user:slow_skill(_))), 
                           cleanup(retractall(user:slow_skill(_)))]) :-
    assertz((user:slow_skill(X) :- sleep(0.001), X = 1)),
    % Threshold 0.0 should always trigger
    improve:improve_if_slow(user:slow_skill(_), 0.0),
    user:slow_skill(Y),
    assertion(Y == 1).

test(regression_veto, [setup(retractall(user:safe_skill(_))),
                        cleanup(retractall(user:safe_skill(_)))]) :-
    assertz(user:safe_skill(42)),
    % Propose a candidate that would regress logic (only succeeds for 1, whereas reality is 42)
    Candidate = (user:safe_skill(1) :- true),
    improve:evaluate_candidate(user:safe_skill(_), Candidate, Verdict),
    assertion(Verdict == regressed).

test(logic_parity, [setup(retractall(user:fib/2)),
                    cleanup(retractall(user:fib/2))]) :-
    assertz(user:fib(10, 55)),
    % Candidate is equivalent but differently shaped
    Candidate = (user:fib(10, 55) :- true),
    improve:evaluate_candidate(user:fib(_,_), Candidate, Verdict),
    assertion(Verdict == equivalent).

:- end_tests(self_improvement).
