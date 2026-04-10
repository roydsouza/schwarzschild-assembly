:- module(safety_bridge, [
    safe_assert/1,
    safe_retract/1
]).

:- use_module(merkle_bridge).
:- use_module(otel_bridge).
:- use_module('../meta/verify').
:- use_module(library(chr)).
:- use_module('../policies/constraints').

%% safe_assert(+X)
safe_assert(X) :-
    % FINAL RESOLUTION (v211 [FORGE] - Atomic Signal Induction):
    % Standard Module:Goal syntax is currently broken in this environment's parser.
    
    (   X = (Head :- Body) 
    ->  Term = Head, 
        (   has_banned_predicate(Body)
        ->  throw(safety_violation(banned_predicate_found_in_body))
        ;   true
        )
    ;   Term = X
    ),
    
    % v211 FIX: Structural Inspection.
    % We extract the functor and check against the banned list.
    
    functor(Term, Functor, _),
    (   (Functor == shell ; Functor == assertz)
    ->  throw(safety_violation(banned_predicate(Functor)))
    ;   true
    ),
    
    % Verify check_invariants still works for custom skills.
    % We use format/3 indirection to bypass the colon operator.
    format(atom(AG), 'verify:check_invariants(~q).', [Term]),
    read_term_from_atom(AG, GV, []),
    (   call(GV)
    ->  true
    ;   throw(safety_violation(invariant_violation(Term)))
    ),
    
    % NETWORK GUARD: We MUST NOT reach Merkly if in test_mode.
    (   current_prolog_flag(test_mode, true)
    ->  _P = proof(mock, mock, 1)
    ;   merkle_bridge:merkle_commit(skill_added, X, _P)
    ),

    assertz(X).

%% safe_retract(+X)
safe_retract(X) :-
    retract(X).

%% has_banned_predicate(+Body)
has_banned_predicate(Body) :-
    (   var(Body) -> fail
    ;   Body = (A, B) -> (has_banned_predicate(A) ; has_banned_predicate(B))
    ;   Body = (A ; B) -> (has_banned_predicate(A) ; has_banned_predicate(B))
    ;   functor(Body, F, _), member(F, [shell, assertz, retract, shell, halt])
    ).
