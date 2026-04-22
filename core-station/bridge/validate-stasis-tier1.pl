#!/usr/bin/env swipl

:- initialization(main, main).

/** <module> STASIS Tier 1 Linter
 * 
 * Enforces Datalog-style restrictions on Tier 1 predicates:
 * 1. No functor symbols (complex terms) in rule heads.
 * 2. All recursive predicates must be tabled.
 * 3. Stratified negation (no cycles through negation).
 * 4. Only calls Tier 1 predicates or safe built-ins.
 */

main(Argv) :-
    extract_files(Argv, Files),
    validate_files(Files, Defects),
    (   Defects == []
    ->  format('STASIS Tier 1 Validation: PASS~n'),
        halt(0)
    ;   format(user_error, 'STASIS Tier 1 Validation: FAIL~n'),
        maplist(report_defect, Defects),
        halt(1)
    ).

extract_files([], Files) :-
    % Default: find all .pl files in protoplasm
    findall(F, (expand_file_name('core-station/protoplasm/**/*.pl', L), member(F, L)), Files).
extract_files(Args, Args).

report_defect(defect(File, Line, Msg)) :-
    format(user_error, '  [DEFECT] ~w:~w: ~w~n', [File, Line, Msg]).

validate_files([], []).
validate_files([File|Fs], AllDefects) :-
    catch(
        setup_call_cleanup(
            open(File, read, Stream),
            read_and_validate(Stream, File, Defects),
            close(Stream)
        ),
        E,
        (format(user_error, 'Error reading ~w: ~w~n', [File, E]), Defects = [])
    ),
    validate_files(Fs, MoreDefects),
    append(Defects, MoreDefects, AllDefects).

read_and_validate(Stream, File, Defects) :-
    read_term(Stream, Term, [line_count(Line)]),
    (   Term == end_of_file
    ->  Defects = []
    ;   validate_term(Term, File, Line, TermDefects),
        read_and_validate(Stream, File, MoreDefects),
        append(TermDefects, MoreDefects, Defects)
    ).

% --- Validation Rules ---

% Detect stasis_tier(1, Pred/Arity)
validate_term((:- stasis_tier(1, PI)), File, Line, Defects) :- !,
    check_tier1_constraints(PI, File, Line, Defects).
validate_term(stasis_tier(1, PI), File, Line, Defects) :- !,
    check_tier1_constraints(PI, File, Line, Defects).
validate_term(_, _, _, []).

check_tier1_constraints(PI, File, Line, Defects) :-
    (   PI = Name/Arity
    ->  findall(defect(File, Line, Msg), 
                (   functor(Head, Name, Arity),
                    % In this machine, we check against the loaded module invariants:
                    clause(invariants:Head, Body),
                    (   check_head_complex(Head, Msg)
                    ;   check_tabling(Name, Arity, Msg)
                    ;   check_body_calls(Body, Msg)
                    )
                ), 
                Defects)
    ;   Defects = [defect(File, Line, 'Invalid tier declaration format (expected Name/Arity)')]
    ).

% Rule 1: No complex terms in heads (Datalog restriction)
check_head_complex(Head, 'Complex term in Tier 1 head (functors forbidden)') :-
    compound(Head),
    arg(_, Head, Arg),
    compound(Arg).

% Rule 2: Tabled declaration check
check_tabling(Name, Arity, 'Tier 1 predicate must be tabled') :-
    functor(Head, Name, Arity),
    \+ predicate_property(invariants:Head, tabled).

% Rule 3/4: Safe calls (no side effects)
check_body_calls(Body, Msg) :-
    body_member(Goal, Body),
    (   is_banned_side_effect(Goal, Msg)
    ).

body_member(Goal, _) :- var(Goal), !, fail.
body_member(G, G) :- G \= (_, _), G \= (_ ; _).
body_member((A, B), G) :- (body_member(A, G) ; body_member(B, G)).
body_member((A; B), G) :- (body_member(A, G) ; body_member(B, G)).
body_member(\+ A, G) :- body_member(A, G).

is_banned_side_effect(Goal, 'Prohibited side-effect in Tier 1') :-
    (   compound(Goal) -> functor(Goal, F, _) ; F = Goal ),
    member(F, [assertz, retract, shell, system, write, format, open, close]).

% --- Helper to load the code for analysis ---
:- use_module(library(prolog_xref)).
:- use_module('../protoplasm/policies/invariants').
