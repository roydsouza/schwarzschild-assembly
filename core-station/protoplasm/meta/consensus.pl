:- module(consensus, [
    propose_shared_skill/2,
    evaluate_consensus/2,
    reconcile_substrate/1,
    conflict_resolver/3
]).

:- use_module('../core/merkle_bridge').
:- use_module('../core/safety_bridge').
:- use_module('fitness').

/** <module> STASIS Phase 12: Collective Intelligence
 * 
 * Implements the multi-agent consensus and reconciliation protocol.
 * Allows spacecraft to share verified knowledge across the Aethereum-Spine.
 */

%% propose_shared_skill(+Term, -Proposal) is det.
% Packages a local skill with its Merkle proof and fitness metrics.
propose_shared_skill(Term, Proposal) :-
    % 1. Verify local safety first
    safety_bridge:test_safety(Term),
    
    % 2. Calculate local fitness
    fitness:calculate_fitness(Metrics),
    Fitness = Metrics.get('sati_central.prolog.substrate_fitness_score'),
    
    % 3. Retrieve Merkle proof (Phase 12 stubs)
    merkle_bridge:merkle_root(Root),
    
    Proposal = consensus_proposal{
        term: Term,
        fitness: Fitness,
        source_root: Root,
        timestamp: T
    },
    get_time(T).

%% evaluate_consensus(+Proposal, -Verdict) is det.
% A ship evaluates a remote proposal.
evaluate_consensus(Proposal, Verdict) :-
    % 1. Tier 1/2 Safety check
    (  catch(safety_bridge:test_safety(Proposal.term), _, fail)
    -> (  Proposal.fitness > 0.6  % Quorum threshold
       -> Verdict = accept
       ;  Verdict = reject(low_fitness)
       )
    ;  Verdict = reject(safety_violation)
    ).

%% reconcile_substrate(+Proposal) is det.
% Merges a verified proposal into the station-wide protoplasm.
reconcile_substrate(Proposal) :-
    % Must pass final safety gate before mutation
    safety_bridge:safe_assert(Proposal.term).

%% conflict_resolver(+PropA, +PropB, -Winner) is det.
% Phase 12 Requirement: Resolves collisions between competing logic.
conflict_resolver(PropA, PropB, Winner) :-
    (  PropA.fitness >= PropB.fitness
    -> Winner = PropA
    ;  Winner = PropB
    ).
