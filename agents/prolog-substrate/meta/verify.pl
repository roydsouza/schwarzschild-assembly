:- module(verify, [
    check_invariants/1
]).

/** <module> Invariant Verification
 * 
 * Provides predicates to ensure proposed optimizations do not violate core skill invariants.
 */

%% check_invariants(+ProposedClause) is semidet.
%
% Baseline invariant check: Ensures the proposed clause has a valid structure.
% Future versions will run semantic checks against ground-truth test sets.
check_invariants((Head :- _Body)) :-
    % 1. Head must be an atom or a compound term
    (atom(Head) ; compound(Head)),
    % Extract the module if present
    strip_module(Head, _Module, PlainHead),
    % 2. Prohibit redefinition of system-level predicates
    functor(PlainHead, Name, _),
    ( is_system_predicate(Name) 
    -> format(atom(Msg), 'redefinition of meta-improvement logic prohibited: ~w', [Name]),
       throw(error(safety_violation(Msg), _))
    ;  true
    ).

check_invariants(Head) :-
    (atom(Head) ; compound(Head)),
    strip_module(Head, _Module, PlainHead),
    functor(PlainHead, Name, _),
    ( is_system_predicate(Name) 
    -> format(atom(Msg), 'redefinition of meta-improvement logic prohibited: ~w', [Name]),
       throw(error(safety_violation(Msg), _))
    ;  true
    ).

is_system_predicate(safe_assert).
is_system_predicate(safe_retract).
is_system_predicate(mcp_call).
is_system_predicate(merkle_commit).
is_system_predicate(emit_metric).
is_system_predicate(inspect_predicate).
is_system_predicate(test_predicate).
is_system_predicate(measure_performance).
is_system_predicate(check_invariants).
is_system_predicate(propose_improvement).
is_system_predicate(improve_if_slow).
is_system_predicate(evaluate_candidate).
