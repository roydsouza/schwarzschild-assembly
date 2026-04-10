:- module(mcp_bridge, [
    mcp_call/3
]).

:- use_module(library(process)).
:- use_module(library(json)).

/** <module> MCP Bridge
 * 
 * Provides an interface to the Root Spine MCP host.
 * Calls mcp-spine-client binary to execute tools.
 */

mcp_call(Method, Params, Result) :-
    (   current_prolog_flag(test_mode, true)
    ->  % Mock success response in test mode
        Result = _{isError: false, content: [_{text: "{}"}]}
    ;   setup_call_cleanup(
            process_create(path('mcp-spine-client'), [Method, json(Params)], [stdout(pipe(Out))]),
            (   json_read_dict(Out, Result)
            ),
            close(Out)
        )
    ).
