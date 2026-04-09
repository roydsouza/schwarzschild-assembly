:- module(mcp_bridge, [
    mcp_call/3
]).

:- use_module(library(process)).
:- use_module(library(http/json)).

/** <module> MCP Bridge
 * 
 * Provides predicates to call the Sati-Central Root Spine MCP host.
 * Uses the scripts/mcp-client.sh shell utility to perform HTTP calls.
 */

%% mcp_call(+Method, +Arguments, -Result) is det.
%
% Invokes an MCP tool or method via the Root Spine MCP host.
% Method is an atom (e.g., 'tools/list', 'tools/call').
% Arguments is a dict representing the JSON parameters.
% Result is a dict representing the JSON-RPC result.
mcp_call(Method, Arguments, Result) :-
    % Locate the mcp-client.sh script relative to the project root
    % We assume the current working directory is the project root
    ClientScript = 'scripts/mcp-client.sh',
    
    % Convert Arguments dict to JSON string
    atom_json_dict(ArgsJSON, Arguments, [width(0)]),
    
    % Prepare process arguments
    ProcArgs = [Method, ArgsJSON],
    
    % Execute the script and capture output
    setup_call_cleanup(
        process_create(path(bash), [ClientScript | ProcArgs], [stdout(pipe(Out))]),
        json_read_dict(Out, Response),
        close(Out)
    ),
    
    % Handle JSON-RPC response
    (   get_dict(error, Response, Error)
    ->  throw(error(mcp_error(Error.message), mcp_call(Method, Arguments)))
    ;   get_dict(result, Response, Result)
    ).
