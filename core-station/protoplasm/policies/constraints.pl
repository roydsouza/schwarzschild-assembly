:- module(constraints, [
    check_constraints/1
]).

:- use_module(library(chr)).

/** <module> Safety Rail CHR Policy
 * 
 * Defines declarative constraints on the shape of Prolog clauses.
 */

:- chr_constraint check_constraints/1, banned_predicate/1.

% 1. Body Scanning
check_constraints((_Head :- Body)) :- !,
    find_banned(Body, BannedList),
    forall(member(P, BannedList), banned_predicate(P)).
check_constraints(_).

% 2. Trigger
banned_predicate(P) <=>
    format(atom(Msg), 'banned predicate usage detected in body: ~w', [P]),
    throw(safety_violation(Msg)).

find_banned(Body, Banned) :-
    findall(P, (member(P, [shell, system, assert, retract, asserta, assertz]), 
                contains_predicate(Body, P)), Banned).

contains_predicate(P, P) :- !.
contains_predicate((A, _), P) :- contains_predicate(A, P), !.
contains_predicate((_, B), P) :- contains_predicate(B, P), !.
contains_predicate((A; _), P) :- contains_predicate(A, P), !.
contains_predicate((_; B), P) :- contains_predicate(B, P), !.
contains_predicate(\+ A, P) :- contains_predicate(A, P), !.
