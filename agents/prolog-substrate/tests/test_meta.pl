:- use_module(library(plunit)).
:- use_module('../meta/improve').
:- use_module('../core/safety_bridge').

% Define a target skill for optimization in the user module
:- dynamic user:test_slow_skill/1.
:- dynamic user:wait_a_bit/0.
:- dynamic user:fib/2.

:- begin_tests(meta_improvement).

test(improve_if_slow_trigger) :-
    % Set test_mode to bypass Merkle network calls
    set_prolog_flag(test_mode, true),

    % 1. Setup skill
    retractall(user:test_slow_skill(_)),
    safe_assert((user:test_slow_skill(a) :- wait_a_bit)),
    retractall(user:wait_a_bit),
    assertz((user:wait_a_bit :- sleep(0.001))),
    
    % 2. Trigger optimization on user:test_slow_skill(a)
    improve:improve_if_slow(user:test_slow_skill(a), 0.0),
    
    % 3. Verify that the skill is still the same
    findall(B, clause(user:test_slow_skill(a), B), Bodies),
    ( member(Body, Bodies), strip_module(Body, _, wait_a_bit) -> true ;
      format('Actual Bodies: ~w~n', [Bodies]), fail
    ).

test(evaluate_candidate_regression) :-
    % Set test_mode
    set_prolog_flag(test_mode, true),
    
    % Setup base skill
    retractall(user:fib(_,_)),
    safe_assert((user:fib(0, 0) :- !)),
    safe_assert((user:fib(1, 1) :- !)),
    safe_assert((user:fib(N, F) :- N > 1, N1 is N-1, N2 is N-2, user:fib(N1, F1), user:fib(N2, F2), F is F1+F2)),
    
    % Candidate that is WRONG
    NewClause = (user:fib(N, 0) :- N > 1),
    
    % Evaluate candidate
    % We expect Verdict = regressed because NewResults (all 0s) \= OldResults (fib values)
    % IMPORTANT: ensure fib/2 is ground or handled in introspect
    improve:evaluate_candidate(user:fib(10, _F), NewClause, Verdict),
    ( Verdict == regressed -> true ; 
      format('Verdict was: ~w~n', [Verdict]), fail
    ).

test(safety_guard_head_redefinition) :-
    % Attempt to redefine a system predicate via safe_assert
    catch(
        safe_assert((improve_if_slow(X, Y) :- true)),
        error(safety_violation(Msg), _),
        ( sub_atom(Msg, _, _, _, 'redefinition of meta-improvement logic prohibited') -> true ;
          format('Unexpected error message: ~w~n', [Msg]), fail
        )
    ).

:- end_tests(meta_improvement).
