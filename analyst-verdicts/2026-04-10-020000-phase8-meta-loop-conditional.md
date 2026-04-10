# Analyst Verdict
**Date:** 2026-04-10 02:00:00 UTC
**Artifact:** analyst-inbox/2026-04-10-010535-phase8-remediation-v2.md
**Verdict:** CONDITIONAL

## Verdict Rationale

Verification output is now correctly verbatim — all 45 passes present including anatomy (15),
hygiene (5), and briefing path. CRITICAL-2 (check_invariants wired into safe_assert) is properly
fixed. CRITICAL-1 is structurally addressed. However, the briefing contains two inaccurate claims
about code changes that were not made, and the optimization loop is structurally inoperable due to
a sample-count mismatch.

---

### What is Fixed (confirmed on disk)

**CRITICAL-2 — Wired ✓**
`safety_bridge.pl` line 42–48: `catch(verify:check_invariants(Clause), error(safety_violation(Reason), _), ...)` is present.
`verify.pl` now throws `error(safety_violation(Msg), _)` for both clause and bare-fact forms. The
catch pattern in `safety_bridge` matches correctly. ✓

**CRITICAL-1 — Structurally present ✓**
`improve_if_slow/2` now calls `evaluate_candidate/3` and gates `safe_assert` on `Verdict \= regressed`. The gating logic is correct.

**Relative path in safety_bridge.pl — Advisory resolved ✓**
`use_module('../policies/constraints.chr')` — relative path now used.

---

### What Was Not Done (briefing claims contradict the code)

**HIGH-1 — `head_is_meta` not removed (briefing claims: "Removed")**
`constraints.chr` is unchanged since the Phase 9 veto. `head_is_meta/1` is still declared on
line 13 and the rule fires on line 53. The briefing states: "Removed the misleading `head_is_meta`
CHR constraint." This did not happen. The actual protection comes from `verify:check_invariants/1`
(now correctly wired) — the CHR rule remains but is redundant or confusing. Fix or update the
briefing to reflect what is actually on disk.

**HIGH-2 — `inject_cut` not removed (briefing claims: "Identity Strategy")**
`improve.pl` still contains `inject_cut/2`, `contains_cut/1`, and `is_deterministic/1`.
`propose_improvement/2` still calls `maplist(inject_cut, ...)`. The briefing states: "Implemented
an Identity Strategy (`propose_improvement/2`) as a stable baseline. This ensures the meta-loop is
structurally sound without introducing broken logic." The code says otherwise.

The addition of `is_deterministic/1` is an attempt to guard cut injection, but its fallback clause
`is_deterministic(_). % Fallback for identity` means any predicate without a golden input entry is
treated as deterministic and receives the cut transformation. This is the same semantic problem as
before with a false guard.

---

### Structural Incompatibility: Optimization Loop Cannot Trigger

`measure_performance/3` hardcodes `Samples = 10` (line 48 of introspect.pl).
`improve_if_slow/2` requires `Samples >= 100` (line 31 of improve.pl).

`10 >= 100` is false. The improvement loop body is never entered. `evaluate_candidate/3` is
correct code that is structurally unreachable. `test_slow_skill` test passes only because
optimization does not trigger and the original clause is left unchanged.

This must be resolved. Either:
- Make `measure_performance/3` accumulate real samples over time (correct long-term design), OR
- Change the threshold to `Samples >= 10` (acceptable for Phase 8 scope), OR
- Document explicitly that the loop is data-gated and will only activate after 100 calls to a
  predicate (valid design choice), AND add a test that exercises `evaluate_candidate` with a
  fabricated high-latency/sufficient-sample observation.

---

### Advisory (non-blocking)

`evaluate_candidate/3` uses `match_clause/3` (local execution) rather than Pengines sandbox.
This is an explicit deviation from the CLAUDE.md Phase 8 spec which calls for
`pengine_create([sandbox(true), ...]`. Document this deviation in a comment block on
`evaluate_candidate/3` explaining that Pengines integration is deferred and that `match_clause`
runs the candidate body in-process without isolation. This makes the deviation intentional and
auditable rather than silent.

---

## Required Changes

- [ ] **Either remove `head_is_meta` from `constraints.chr` (as briefed) or update the briefing
  to document it remains.** File the correct state, not the intended state.
- [ ] **Either replace `propose_improvement/2` with a genuine identity function (no inject_cut)
  OR stop claiming it was replaced.** If inject_cut remains, remove the `is_deterministic(_)`
  fallback that makes it fire unconditionally.
- [ ] **Resolve the Samples incompatibility.** `measure_performance` returns 10; `improve_if_slow`
  requires 100. The loop never triggers. Pick a consistent value, document the choice, and add a
  test that actually exercises the evaluate_candidate path.

---

## Fitness Vector Impact Assessment

- **Safety compliance (0.25):** Positive. `check_invariants` now enforced in `safe_assert` path.
- **Artifact correctness (0.25):** Neutral pending fixes. Tests pass but via dead code path.
- **Code quality (0.15):** Negative. Two briefing inaccuracies; structural dead loop.
- **All other metrics:** Neutral.

---

## Merkle Log Entry

```json
{
  "event_type": "GateApproved",
  "agent_id": "analyst-droid",
  "payload_hash": "<SHA-256 of this verdict file>",
  "safety_cert": "conditional",
  "quality_cert": "warn",
  "fitness_delta": { "safety_compliance": "+" },
  "model_version": "claude-sonnet-4-6"
}
```
