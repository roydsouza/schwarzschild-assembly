:- module(merkle_bridge, [
    merkle_commit/3
]).

:- use_module(mcp_bridge).

/** <module> Merkle Bridge
 * 
 * Provides predicates to commit events to the Merkle audit log.
 * In Phase 8, this primarily handles skill updates.
 */

%% merkle_commit(+EventType, +Payload, -Proof) is det.
%
% Commits a payload to the Merkle log.
% EventType is an atom (currently primarily 'skill_added').
% Payload is the Prolog term that was added.
% Proof is a dict containing the MerkleProof (leaf_hash, root_hash, tree_size).
merkle_commit(skill_added, Clause, Proof) :-
    % We use the agent_id 'prolog-substrate' and skill_name 'self-modification'
    % for now. A more advanced version would track individual agent IDs.
    AgentId = 'prolog-substrate',
    SkillName = 'autonomous-skill',
    
    % Convert Clause to atom for transmission
    term_to_atom(Clause, ClauseAtom),
    
    mcp_call('tools/call', _{
        name: 'submit_skill_proposal',
        arguments: _{
            agent_id: AgentId,
            skill_name: SkillName,
            clause: ClauseAtom
        }
    }, Result),
    
    % Verify success and extract proof
    (   get_dict(isError, Result, true)
    ->  get_dict(content, Result, [Content|_]),
        get_dict(text, Content, ErrorMsg),
        throw(error(merkle_error(ErrorMsg), merkle_commit(skill_added, Clause)))
    ;   get_dict(content, Result, [Data|_]),
        get_dict(text, Data, ProofJSON),
        atom_json_dict(ProofJSON, Proof, [])
    ).
