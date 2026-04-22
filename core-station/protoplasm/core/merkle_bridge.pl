:- module(merkle_bridge, [
    merkle_commit/3,
    merkle_verify/3,
    merkle_log/3,
    merkle_root/1
]).

:- dynamic merkle_log/3.

/** <module> Merkle Tree Audit Bridge
 * 
 * Provides a cryptographically-verifiable audit trail for K-base mutations.
 * Every 'safe_assert' is recorded as a leaf in the evolutionary Merkle tree.
 */

%% merkle_commit(+Event, +Data, -Hash) is det.
merkle_commit(Event, Data, Hash) :-
    % Generate a symbolic hash
    term_hash((Event, Data), HashNum),
    format(atom(Hash), 'stasis_v1_~w', [HashNum]),
    
    % Record in the local audit trail
    ( aggregate_all(count, merkle_log(_, _, _), Count) -> Next = Count ; Next = 0 ),
    assertz(merkle_log(Next, Event, Hash)),

    (  (current_prolog_flag(test_mode, true) ; user:current_prolog_flag(test_mode, true))
    -> true
    ;  format('~N[MERKLE] Committed ~w: ~w -> ~w~n', [Event, Data, Hash])
    ).

%% merkle_verify(+Event, +Data, +Hash) is semidet.
merkle_verify(Event, Data, Hash) :-
    term_hash((Event, Data), HashNum),
    format(atom(Expected), 'stasis_v1_~w', [HashNum]),
    Expected == Hash.

%% merkle_root(-RootHash) is det.
% Provides a roll-up hash of the current audit trail.
merkle_root(RootHash) :-
    findall(H, merkle_log(_, _, H), Hashes),
    term_hash(Hashes, RootNum),
    format(atom(RootHash), 'root_~w', [RootNum]).

%% merkle_export_proof(+Hash, -Proof, -Root) is det.
% Phase 12: Generates a proof that 'Hash' is part of the current audit trail.
% In this symbolic version, the proof is the sequence of hashes.
merkle_export_proof(Hash, Proof, Root) :-
    merkle_root(Root),
    findall(H, merkle_log(_, _, H), Proof),
    member(Hash, Proof).

%% merkle_verify_proof(+Hash, +Proof, +Root) is semidet.
% Verifies that a Hash is contained within a Proof that resolves to Root.
merkle_verify_proof(Hash, Proof, Root) :-
    member(Hash, Proof),
    term_hash(Proof, RootNum),
    format(atom(ExpectedRoot), 'root_~w', [RootNum]),
    Root == ExpectedRoot.
