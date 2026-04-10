:- module(golden_data, [
    golden_input/2
]).

% golden_input(+PredicateName, -Inputs)
golden_input(test_slow_skill, [a]).
golden_input(fib, [10, 20]).
golden_input(between, [[1, 10, _]]).
