:- module(fitness, [
    report_substrate_fitness/0,
    calculate_fitness/1
]).

:- use_module('../core/otel_bridge').
:- use_module('introspect').

/** <module> Substrate Fitness Scorer
 * 
 * Aggregates substrate-level metrics (skill diversity, latency) 
 * into a unified fitness score for the global Root Spine dashboard.
 */

%% report_substrate_fitness is det.
%
% Calculates current fitness and reports it via the OTel bridge.
report_substrate_fitness :-
    calculate_fitness(Metrics),
    emit_metrics_batch(Metrics).

%% calculate_fitness(-Metrics) is det.
%
% Computes the fitness footprint of the Prolog substrate.
calculate_fitness(Metrics) :-
    % 1. Skill Diversity (Count of runtime predicates)
    findall(P/A, (current_predicate(P/A), \+ predicate_property(P/A, built_in)), AllPreds),
    length(AllPreds, SkillCount),
    
    % 2. Baseline Latency (Sampled)
    % Use a goal that is guaranteed to be grounded and exists.
    (   measure_performance(introspect:inspect_predicate(user:foo, _), Latency, Samples), Samples > 0
    ->  true
    ;   Latency = 0.001 % Default floor
    ),
    
    % 3. Compute Composite Fitness Score
    LatencyScore is max(0.0, 1.0 - Latency / 1.0),
    DiversityScore is min(1.0, SkillCount / 10.0),
    FitnessScore is (LatencyScore + DiversityScore) / 2.0,

    Metrics = json{
        'aethereum_spine.prolog.substrate_fitness_score': FitnessScore,
        'aethereum_spine.prolog.skill_diversity_total': SkillCount,
        'aethereum_spine.prolog.substrate_avg_latency_ms': Latency
    }.
