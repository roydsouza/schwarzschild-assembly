:- use_module(library(plunit)).
:- use_module('../meta/fitness').
:- use_module('../core/otel_bridge').

:- begin_tests(fitness_integration).

test(calculate_fitness_structure) :-
    fitness:calculate_fitness(Metrics),
    get_dict('aethereum_spine.prolog.substrate_fitness_score', Metrics, Score),
    number(Score),
    get_dict('aethereum_spine.prolog.skill_diversity_total', Metrics, Count),
    integer(Count).

test(report_substrate_fitness_mock) :-
    % In test_mode, mcp_call should be handled or mocked.
    % We verify it doesn't crash.
    set_prolog_flag(test_mode, true),
    ( fitness:report_substrate_fitness -> true ; fail ).

:- end_tests(fitness_integration).
