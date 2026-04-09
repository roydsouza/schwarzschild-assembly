:- use_module(library(plunit)).
:- use_module('../meta/improve').
:- use_module('../core/safety_bridge').

% Define a target skill for optimization in the user module
:- dynamic user:test_slow_skill/1.

:- begin_tests(meta_improvement).

test(improve_if_slow_trigger) :-
    % Set test_mode to bypass Merkle network calls
    set_prolog_flag(test_mode, true),

    % 1. Setup skill with NO cut in the user module
    retractall(user:test_slow_skill(_)),
    safe_assert((user:test_slow_skill(a) :- wait_a_bit)),
    assertz((user:wait_a_bit :- sleep(0.01))),
    
    % 2. Trigger optimization on user:test_slow_skill(a)
    improve:improve_if_slow(user:test_slow_skill(a), 0.0),
    
    % 3. Verify that an optimized version was asserted
    % We use strip_module to handle the safety_bridge: prefix if present
    findall(B, clause(user:test_slow_skill(a), B), Bodies),
    ( (member(Body, Bodies), strip_module(Body, _, (wait_a_bit, !))) -> true ; 
      format('Actual Bodies for user:test_slow_skill(a): ~w~n', [Bodies]), fail 
    ).

:- end_tests(meta_improvement).
