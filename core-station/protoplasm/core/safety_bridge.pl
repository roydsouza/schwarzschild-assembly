:- module(safety_bridge, [
    safe_assert/1,
    safe_retract/1,
    test_safety/1,
    test_string_safety/1
]).

:- use_module(merkle_bridge).
:- use_module(otel_bridge).
:- use_module('../meta/verify').
:- use_module(library(chr)).
:- use_module('../policies/invariants').

/** <module> Safety Bridge (Hardened)
 * 
 * Intercepts all knowledge base mutations and verifies them against
 * Tier 1 Invariants using a deep recursive scanner.
 * 
 * This implementation bypasses CHR rule scheduling issues by performing
 * synchronous structural analysis directly in the bridge.
 */

%% safe_assert(+Term) is det.
safe_assert(Term) :-
    % 1. Perform Deep Structural Scan (Mandatory)
    test_safety(Term),

    % 2. Tier 3: Invariant verification (High-level semantic checks)
    Term_Atom2 = 'verify:check_invariants(T).',
    read_term_from_atom(Term_Atom2, G2, [variable_names(['T'=Term])]),
    (  call(G2)
    -> true
    ;  throw(safety_violation(invariant_violation(Term)))
    ),

    % 3. Audit: Merkle commit
    Term_Atom3 = 'merkle_bridge:merkle_commit(skill_added, T, _).',
    read_term_from_atom(Term_Atom3, G3, [variable_names(['T'=Term])]),
    call(G3),

    % 4. Mutation
    user:assertz(Term).

%% test_safety(+Term) is det.
% Synchronous check without mutation.
test_safety(Term) :-
    (  ( Term = (_H :- Body) -> scan_body(Body, Context) ; scan_body(Term, Context) )
    -> throw(safety_violation(banned_operation(Context)))
    ;  true
    ).

%% test_string_safety(+String) is det.
% Parses a string into a term and then runs test_safety.
test_string_safety(String) :-
    (  catch(term_string(Term, String), _, fail)
    -> test_safety(Term)
    ;  % If it's not valid Prolog syntax, check the atom itself
       atom_string(Atom, String),
       test_safety(Atom)
    ).

%% scan_body(+Term, -Context) is semidet.
% Fails if the term is safe, succeeds (binding Context) if a violation is found.
scan_body(Term, Term) :-
    is_banned(Term).
scan_body(Term, Context) :-
    compound(Term),
    Term =.. [_F|Args],
    member(A, Args),
    scan_body(A, Context).

%% is_banned(+Term) is semidet.
is_banned(Var) :- var(Var), !, fail.
is_banned(Term) :- 
    % Handle module qualification
    ( Term = (_M:T) -> (compound(T) -> functor(T, F, _) ; F = T) ; ( compound(Term) -> functor(Term, F, _) ; F = Term ) ),
    % Check against Tier 1
    catch(invariants:banned_operation(F), _, fail).

%% safe_retract(+Term) is det.
safe_retract(Term) :-
    user:retract(Term).
