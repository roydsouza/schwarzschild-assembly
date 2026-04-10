:- module(safety_bridge, [
    safe_assert/1,
    safe_retract/1
]).

:- use_module(merkle_bridge).
:- use_module(otel_bridge).
:- use_module('../meta/verify').

% Load constraints.chr using a relative path
:- use_module('../policies/constraints.chr').

/** <module> Safety Bridge
 * 
 * The authoritative mutation point for the Prolog Substrate knowledge base.
 * Every assertion is gated by the Safety Rail and committed to the audit log.
 */

%% safe_assert_metric(+MetricId, +Value) is det.
%
% Emits a metric, silently ignoring failures when in test_mode (offline).
safe_assert_metric(MetricId, Value) :-
    (   current_prolog_flag(test_mode, true)
    ->  true
    ;   emit_metric(MetricId, Value)
    ).

%% safe_assert(+Clause) is det.
%
% Asserts a clause into the knowledge base after verification.
% Throws safety_violation(Reason) if rejected.
safe_assert(Clause) :-
    % 1. Emit attempt metric (no-op in test_mode)
    safe_assert_metric('sati_central.prolog.skill_assertion_attempt_total', 1),
    
    % 1.5 Local CHR constraint check — runs entirely offline
    % Will throw safety_violation(Msg) if a banned predicate is found in the body
    constraints:check_constraints(Clause),

    % 1.6 Head Redefinition Guard
    % Throws safety_violation(Reason) if Head matches a system predicate
    verify:check_invariants(Clause),
    
    % 2. Safety Rail + Merkle Commit
    % Skip network call in offline test_mode; CHR already validated the clause.
    (   current_prolog_flag(test_mode, true)
    ->  Proof = proof{leaf_hash_hex: 'mock', root_hash_hex: 'mock', tree_size: 1}
    ;   catch(
            merkle_bridge:merkle_commit(skill_added, Clause, Proof),
            error(merkle_error(Reason), _),
            (   safe_assert_metric('sati_central.prolog.skill_rejection_total', 1),
                throw(safety_violation(Reason))
            )
        )
    ),
    
    % 3. Local Assertion
    % Note: assertz/1 is allowed only here in the core safety bridge.
    assertz(Clause),
    
    % 4. Emit success metric (no-op in test_mode)
    safe_assert_metric('sati_central.prolog.skills_added_total', 1),
    
    % 5. Log internal event
    print_message(informational, skill_added(Clause, Proof)).

%% safe_retract(+Clause) is det.
%
% Retracts a clause from the knowledge base.
% In Phase 8, this is a placeholder; real retraction requires audit leaves.
safe_retract(Clause) :-
    retract(Clause),
    safe_assert_metric('sati_central.prolog.skills_removed_total', 1).

% Message formatting
:- multifile prolog:message//1.
prolog:message(skill_added(Clause, Proof)) -->
    [ 'Skill added to KB: ~w'-[Clause], nl,
      'Merkle Proof: ~w'-[Proof] ].
