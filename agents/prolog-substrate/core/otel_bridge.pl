:- module(otel_bridge, [
    emit_metric/2
]).

:- use_module(mcp_bridge).

/** <module> OTel Bridge
 * 
 * Provides predicates to emit OpenTelemetry metrics from Prolog.
 * Proxies calls through the Root Spine MCP 'report_metrics' tool.
 */

%% emit_metric(+MetricId, +Value) is det.
%
% Reports a telemetry metric to the global fitness vector.
% MetricId is an atom (e.g., 'sati_central.prolog.skills_added_total').
% Value is a number.
emit_metric(MetricId, Value) :-
    FactoryId = 'prolog-substrate',
    FactoryType = 'logic-engine',
    % We wrap the single metric into the dict format 'report_metrics' expects
    MetricsDict = _{'id': MetricId, 'value': Value}, % The Tool expects map[string]float64
    % Adjusting to match the MCP tool schema: metrics: map[string]float64
    MetricsMap = json{}.put(MetricId, Value),
    
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
        throw(error(otel_error(ErrorMsg), emit_metric(MetricId, Value)))
    ;   true
    ).
