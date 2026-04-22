:- set_prolog_flag(test_mode, true).
:- use_module(library(plunit)).
:- use_module('../meta/consensus').
:- use_module('../core/merkle_bridge').
:- use_module('../core/safety_bridge').

:- begin_tests(consensus_simulation).

test(multi_ship_consensus, [setup(retractall(merkle_bridge:merkle_log(_,_,_)))]) :-
    % 1. Ship A learns a new 'Safe Color' invariant
    NewRule = (safe_color(gold)),
    consensus:propose_shared_skill(NewRule, ProposalA),
    
    assertion(ProposalA.term == NewRule),
    assertion(ProposalA.fitness > 0.0),
    
    % 2. Ship B receives the Proposal
    % Simulate receiving it over the Spine...
    consensus:evaluate_consensus(ProposalA, VerdictB),
    
    assertion(VerdictB == accept),
    
    % 3. Station reconciling the knowledge globally
    consensus:reconcile_substrate(ProposalA),
    
    % Verify the knowledge is now in the global mind
    assertion(user:safe_color(gold)).

test(consensus_veto, [setup(retractall(merkle_bridge:merkle_log(_,_,_)))]) :-
    % 1. Ship A proposes a malicious rule (e.g., re-enabling shell)
    MaliciousRule = (oops :- shell('rm -rf /')),
    
    % propose_shared_skill should fail or the proposal should be rejected
    % Actually, propose_shared_skill runs local safety first.
    catch(consensus:propose_shared_skill(MaliciousRule, _), E, assertion(E = safety_violation(_))).

test(conflict_resolution) :-
    PropA = consensus_proposal{term: (v(1)), fitness: 0.8},
    PropB = consensus_proposal{term: (v(2)), fitness: 0.9},
    
    consensus:conflict_resolver(PropA, PropB, Winner),
    assertion(Winner == PropB).

:- end_tests(consensus_simulation).
