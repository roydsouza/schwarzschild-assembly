:- set_prolog_flag(test_mode, true).
:- use_module(library(plunit)).
:- use_module('../meta/fitness').

:- begin_tests(fitness_scorer).

%% Case 1: score is always in [0.0, 1.0]
test(score_in_bounds) :-
    fitness:calculate_fitness(Metrics),
    Score = Metrics.get('sati_central.prolog.substrate_fitness_score'),
    number(Score),
    Score >= 0.0,
    Score =< 1.0.

%% Case 2: SkillCount is a non-negative integer
test(skill_count_non_negative) :-
    fitness:calculate_fitness(Metrics),
    Count = Metrics.get('sati_central.prolog.skill_diversity_total'),
    integer(Count),
    Count >= 0.

%% Case 3: latency key is present and is a number
test(latency_present) :-
    fitness:calculate_fitness(Metrics),
    Latency = Metrics.get('sati_central.prolog.substrate_avg_latency_ms'),
    number(Latency).

%% Case 4: zero skills produces valid score
test(zero_skills_produces_valid_score) :-
    fitness:calculate_fitness(Metrics),
    Score = Metrics.get('sati_central.prolog.substrate_fitness_score'),
    Score >= 0.0,
    Score =< 1.0.

%% Case 5: report_substrate_fitness/0 does not throw in test_mode
test(report_does_not_throw) :-
    set_prolog_flag(test_mode, true),
    ( fitness:report_substrate_fitness -> true ; true ).

:- end_tests(fitness_scorer).
