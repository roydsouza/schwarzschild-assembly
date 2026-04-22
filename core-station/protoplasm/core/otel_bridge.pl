:- module(otel_bridge, [
    emit_metric/2,
    emit_metrics_batch/1,
    trace_goal/2,
    measure_goal/2
]).

/** <module> OTel Instrumentation Bridge
 * 
 * Provides unified entry points for performance measurement and metric emission.
 * Integrates with Phase 9 safety rails to ensure observability without side effects.
 */

%% emit_metric(+MetricName, +Value) is det.
emit_metric(Name, Value) :-
    % Currently log-buffered for M5 performance; will connect to gRPC in Phase 13.
    get_time(Now),
    (  (current_prolog_flag(test_mode, true) ; user:current_prolog_flag(test_mode, true))
    -> true
    ;  format('~N[OTEL] ~w metric ~w: ~w~n', [Now, Name, Value])
    ).

%% emit_metrics_batch(+Metrics:dict) is det.
emit_metrics_batch(Metrics) :-
    % Supports dict-based metric batches from fitness_scorer.
    dict_keys(Metrics, Keys),
    forall(member(K, Keys), (
        V = Metrics.get(K),
        emit_metric(K, V)
    )).

%% trace_goal(:Goal, -Latency) is det.
trace_goal(Goal, Latency) :-
    get_time(T1),
    % unshielded call to ensure we measure real execution behavior
    ( catch(Goal, E, (print_message(error, E), fail)) -> true ; true ),
    get_time(T2),
    Latency is T2 - T1.

%% measure_goal(:Goal, -AvgLatency) is det.
measure_goal(Goal, AvgLatency) :-
    Samples = 10, % Baseline for introspection to avoid stalling the loop
    get_time(T1),
    forall(between(1, Samples, _), (Goal -> true ; true)),
    get_time(T2),
    AvgLatency is (T2 - T1) / Samples.
