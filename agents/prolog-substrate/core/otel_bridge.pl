:- module(otel_bridge, [
    emit_metric/2,
    emit_metrics_batch/1
]).

:- use_module(mcp_bridge).

/** <module> OTel Bridge
 * 
 * Provides predicates to emit OpenTelemetry metrics from Prolog.
 * Proxies calls through the Root Spine MCP 'report_metrics' tool.
 */

%% emit_metric(+MetricId, +Value) is det.
%
% Reports a single telemetry metric to the global fitness vector.
emit_metric(MetricId, Value) :-
    emit_metrics_batch(json{}.put(MetricId, Value)).

%% emit_metrics_batch(+MetricsMap) is det.
%
% Reports a batch of telemetry metrics.
% MetricsMap is a JSON object/dict of MetricId: Value pairs.
emit_metrics_batch(MetricsMap) :-
    FactoryId = 'prolog-substrate',
    FactoryType = 'logic-engine',
    
    mcp_call('tools/call', _{
        name: 'report_metrics',
        arguments: _{
            factory_id: FactoryId,
            factory_type: FactoryType,
            metrics: MetricsMap
        }
    }, Result),
    
    % Verify success
    (   get_dict(isError, Result, true)
    ->  get_dict(content, Result, [Content|_]),
        get_dict(text, Content, ErrorMsg),
        throw(error(otel_error(ErrorMsg), emit_metrics_batch(MetricsMap)))
    ;   true
    ).
