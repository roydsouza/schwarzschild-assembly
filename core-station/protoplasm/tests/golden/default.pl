:- module(golden_data, [
    golden_input/2
]).

/** <module> STASIS Golden Test Data
 * 
 * Provides regression-test baselines for critical skills.
 * Used by the 'Evaluate' state of the self-improvement loop.
 */

:- multifile golden_input/2.

% Fibonacci baseline
golden_input(fib, [1, 5, 10, 15, 20]).

% Latency target baseline (slow_skill/1)
golden_input(slow_skill, [1, 2, 3]).

% Security baseline (safe_skill/1)
golden_input(safe_skill, [42, 100]).
