# AntiGravity Mandatory Process Rules — Schwarzschild Assembly

These rules are derived from Claude Code (Analyst Droid) feedback after reviewing Phases 2–5.
They are **permanent**. Every future session in this project must follow them.
Failure to follow any rule is a veto-eligible offense.

---

## Rule 1: HARD STOP at Phase Gates

After completing Phase N:
1. Submit briefing to `analyst-inbox/YYYY-MM-DD-HHMMSS-<topic>.md`
2. Update `STATUS.md` to "AWAITING ANALYST VERDICT"
3. **STOP. Do not write any Phase N+1 code.**
4. Wait for verdict in `analyst-verdicts/`
5. If `CONDITIONAL`: fix items first, re-submit briefing, wait for new verdict
6. If `VETOED`: fix ALL items, resubmit briefing, wait for new verdict
7. If `APPROVED`: begin Phase N+1

**CONDITIONAL does NOT mean "proceed with caveats." It means fix, resubmit, wait.**

---

## Rule 2: Compile-Verify-Test Before Every "COMPLETE" Entry

Before writing "COMPLETE" in STATUS.md:
```bash
cargo build              # Rust
go build ./...           # Go
npm run build            # TypeScript
cargo test --features tier1
go test ./...
npm test
```
ALL must pass. If ANY fails, the phase is NOT complete.

**For Go protobuf:** Always grep the generated `.pb.go` file before using field names or enum values.

**Never claim "Successfully compiled" without running the command in the same session.**

---

## Rule 3: Contract Checklist Method

Before implementing ANY trait method or RPC handler:
1. Read every `///` doc comment on the method
2. List every return variant / error case explicitly
3. List every precondition mentioned
4. List every postcondition mentioned
5. Implement ALL of them — no omissions "for brevity"
6. Write a test for EACH return variant (including ALL error cases)
7. Cross-reference checklist before marking done

---

## Rule 4: No `unsafe` Without Proof

When hitting Send/Sync/lifetime issues:
1. Ask: WHY does the upstream crate not implement it?
2. If answer is "C library state" → do NOT use `unsafe impl`
3. Restructure to contain non-Send types to a single thread
4. Every `unsafe` block MUST have `// Safety:` comment with a **PROOF**, not an assertion
5. If you cannot write the proof, do not write the `unsafe`

---

## Rule 5: Explicit Failure Over Silent Acceptance

- NEVER silently accept input and ignore it
- NEVER use empty/zero values as placeholders in production paths
- NEVER comment "omitted for brevity" on an unimplemented code path
- ALWAYS return explicit error for unimplemented features
- ALWAYS file a proposal in `proposals/pending/` and add TODO with reference

---

## Rule 6: Persistence Verification

For every stateful component, ask: **"What happens after `kill -9`?"**
If answer is "state is lost" → persistence is NOT implemented.

Write a restart test: create state → stop → start → assert state matches.

---

## Rule 7: Pre-Commit Checklist

Run before EVERY briefing submission:
- [ ] `cargo build` / `go build ./...` / `npm run build` — zero errors
- [ ] All tests pass including negative/failure cases
- [ ] Every new public function has at least one failure-case test
- [ ] OTel metrics verified in `otel-snapshots/latest.json`
- [ ] No `unwrap()` on hot paths or FFI boundaries
- [ ] No `unsafe` without `// Safety:` proof comment
- [ ] No `TODO` without `proposals/pending/` reference
- [ ] No in-memory-only state that must survive restarts
- [ ] Migrations applied and verified against real database
- [ ] `scripts/pre-submit.sh` exits 0 — verbatim output in briefing

**"Complete" may NOT appear unless ALL items are checked.**

---

## Rule 8: Test Both Paths

For every test, write BOTH:
1. Positive case: correct input → correct output
2. Negative case: bad input → correct error

The negative case is usually more valuable.
Never skip verification in a test.

---

## Rule 9: OTel End-to-End Verification

After adding any metric:
1. Ensure MeterProvider is initialized (not global no-op)
2. Start OTel collector
3. Exercise the code path
4. Check `otel-snapshots/latest.json` for the metric name
5. If absent → instrumentation is broken, fix before proceeding

---

## Rule 10: No Mislabeling of Deliverables

- If you delivered pre-conditions for Phase N, say "Phase N Pre-conditions," not "Phase N"
- If the factory README says "not yet implemented," the phase is not complete
- Phase numbers have specific definitions in CLAUDE.md §3 — use them exactly

---

## Rule 11: Every Non-Trivial Change Requires a Briefing

Changes without a briefing at `analyst-inbox/YYYY-MM-DD-HHMMSS-<topic>.md`:
- Have no Merkle leaf, no fitness delta, no audit trail
- Are a Merkle audit integrity violation in spirit

---

## Rule 12: Error Return Values Are Never Optional

Every function call that returns an error on an audit-critical path MUST be checked.
A veto/approval not persisted does not exist from the auditor's perspective.

---

## Rule 13: Interface Consistency Verification

For any string defined in one file and referenced in another, grep both sides before filing.
Never write both sides from memory.

---

## Rule 14: Regression Guard

ALL prior phase tests must pass before a new phase briefing is filed:
```bash
cd aethereum-spine   && go test ./...
cd safety-rail  && cargo test --features tier1
cd control-panel && npx vitest run
```

---

## Rule 15: The Physics of Synchronization

Every significant state change or session close must follow the centralized synchronization protocol:

1. **Local SYNC_LOG.md**: Always update `{PROJECT_ROOT}/SYNC_LOG.md` with a summary of the current session's accomplishments and the immediate next steps.
2. **Central SYNC_LOG.md**: If there is a need to share status across the `~/antigravity/` ecosystem (e.g., between `schwarzschild-assembly` and `event-horizon-core`), append an entry to `~/antigravity/SYNC_LOG.md`.
3. **Peer Repositories**: Peer directories like `umbra/`, `penumbra/`, or `darkmatter/` are **NOT** to be used for status synchronization.
4. **Checkpoint Protocol**: When the user says "checkpoint", perform the following:
   - Update both local and central `SYNC_LOG.md` files.
   - Run `git add . && git commit -m "checkpoint: <summary>"` in the current project root.
   - Do **NOT** interact with the `umbra` repository unless explicitly tasked with a feature inside it.
