:- set_prolog_flag(test_mode, true).
:- use_module(library(plunit)).
:- use_module('../core/safety_bridge').

:- begin_tests(safety_bridge).

test(safe_assertion, [setup(retractall(user:test_skill(_)))]) :-
    safety_bridge:safe_assert(test_skill(42)),
    user:test_skill(X),
    assertion(X == 42).

test(unsafe_assertion_banned_predicate, [throws(safety_violation(_))]) :-
    safety_bridge:safe_assert((bad_skill :- assertz(oops))).

test(unsafe_assertion_shell, [throws(safety_violation(_))]) :-
    safety_bridge:safe_assert((evil_skill :- shell('rm -rf /'))).

:- end_tests(safety_bridge).
