:- module(invariants, [
    stasis_tier/2,
    banned_operation/1,
    operation_permitted/1
]).

/** <module> STASIS Tier 1: Invariant Core
 * 
 * Defines hard safety invariants that are decidable by construction.
 * Rules in this module must follow Datalog-style restrictions:
 * - No functor symbols in rule heads.
 * - Stratified negation only.
 * - No direct I/O or side effects.
 * - All recursive predicates must be tabled.
 */

:- dynamic stasis_tier/2.

% Declarative tier markers
stasis_tier(1, banned_operation/1).
stasis_tier(1, operation_permitted/1).

% --- Tier 1 Definitions ---

:- table banned_operation/1.
banned_operation(shell).
banned_operation(system).
banned_operation(assert).
banned_operation(retract).
banned_operation(asserta).
banned_operation(assertz).
banned_operation(process_create).
banned_operation(abolish).

:- table operation_permitted/1.
operation_permitted(Op) :-
    \+ banned_operation(Op).
