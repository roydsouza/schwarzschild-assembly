%% test_mode flag must be set BEFORE safety_bridge is loaded so
%% the module-load-time source_file/2 directive can see it.
:- set_prolog_flag(test_mode, true).

:- use_module(library(plunit)).
:- use_module('../core/safety_bridge').

/** <module> Safety Bridge Tests
 * 
 * Offline unit tests for safe_assert/1 and safe_retract/1.
 * Runs without a live Root Spine MCP host.
 * The test_mode flag (set above) bypasses network calls; local CHR
 * constraints in policies/constraints.chr enforce the safety policy.
 */

:- begin_tests(safety_bridge).

test(safe_assertion, [setup(retractall(safety_bridge:test_skill(_)))]) :-
    % This clause is structurally safe — should be accepted
    safe_assert(test_skill(42)),
    safety_bridge:test_skill(X),
    assertion(X == 42).

test(unsafe_assertion_banned_predicate, [throws(safety_violation(_))]) :-
    % assertz is in the banned-predicate list — CHR must throw
    safe_assert((bad_skill :- assertz(oops))).

test(unsafe_assertion_shell, [throws(safety_violation(_))]) :-
    % shell/1 is in the banned-predicate list — CHR must throw
    safe_assert((evil_skill :- shell('rm -rf /'))).

:- end_tests(safety_bridge).
