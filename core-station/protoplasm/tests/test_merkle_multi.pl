:- set_prolog_flag(test_mode, true).
:- use_module(library(plunit)).
:- use_module('../core/merkle_bridge').
:- use_module('../core/safety_bridge').

:- begin_tests(merkle_multi_ship).

test(merkle_isolation, [setup(retractall(merkle_bridge:merkle_log(_,_,_)))]) :-
    % Ship A mutation
    safety_bridge:safe_assert(ship_skill(a, 1)),
    merkle_bridge:merkle_root(RootA),
    assertion(merkle_bridge:merkle_log(0, skill_added, _)),

    % Simulate Ship B by clearing local log (in a real scenario, these are separate processes)
    retractall(merkle_bridge:merkle_log(_,_,_)),
    safety_bridge:safe_assert(ship_skill(b, 2)),
    merkle_bridge:merkle_root(RootB),
    
    assertion(RootA \== RootB),
    assertion(merkle_bridge:merkle_log(0, skill_added, _)).

test(merkle_verification, [setup(retractall(merkle_bridge:merkle_log(_,_,_)))]) :-
    % Ship A adds knowledge
    Data = skill(navigation, 100),
    safety_bridge:safe_assert(ship_skill(a, Data)),
    merkle_bridge:merkle_log(0, skill_added, HashA),
    
    % Ship B receives Ship A's data and hash and verifies it
    merkle_bridge:merkle_verify(skill_added, ship_skill(a, Data), HashA).

:- end_tests(merkle_multi_ship).
